package gocolly

import (
	"config"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/globalsign/mgo/bson"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"helpers"
	"regexp"
	"services/mongodb"
	"strconv"
	"strings"
)

// 房源详情|无小区数据版本
func DetailX(cityId int) {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 开始时间
	startTime := helpers.GetMicroTimestamp()
	// 实例化
	c := colly.NewCollector()
	// URL队列
	q, _ := queue.New(
		config.ThreadsNum["detail"],
		&queue.InMemoryQueueStorage{MaxSize: 100000},
	)
	// 获取URL
	for _, v := range houseDetailUrl(cityId) {
		// 添加队列
		q.AddURL(v)
	}
	// 抓取数据|房源在线
	c.OnHTML("div.wrapper", func(e *colly.HTMLElement) {
		// 房源标题
		titleDom := e.DOM.Find("div.content > p.content__title")
		title := strings.TrimSpace(titleDom.Text())
		// 房源ID
		houseUrl := e.Request.URL.String()
		urlSplit := strings.Split(houseUrl, "/")
		houseUrl = urlSplit[len(urlSplit)-1]
		houseId := houseUrl[0 : len(houseUrl)-5]
		// 发布日期
		postedDateDom := e.DOM.Find("div.content > div.content__subtitle")
		postedDateStr := strings.TrimSpace(postedDateDom.Text())
		postedDateReg := regexp.MustCompile(`20\d{2}-\d{2}-\d{2}`)
		postedDate := postedDateReg.FindString(postedDateStr)
		// 租金
		rentDom := e.DOM.Find("div.content > div.content__aside > p.content__aside--title > span")
		rent, _ := strconv.ParseInt(rentDom.Text(), 10, 64)
		// 付款方式
		paymentDom := e.DOM.Find("div.content > div.content__aside > p.content__aside--title")
		paymentStr := strings.TrimSpace(paymentDom.Text())
		paymentReg := regexp.MustCompile(`[\p{Han}]{2,3}价`)
		payment := paymentReg.FindString(paymentStr)
		if len(payment) != 0 {
			payment = payment[0 : len(payment)-3]
		} else {
			payment = "未知"
		}
		// 标签
		tagDom := e.DOM.Find("div.content > div.content__aside > p.content__aside--tags > i")
		tagStr := []string{}
		tagDom.Each(func(i int, selection *goquery.Selection) {
			tagStr = append(tagStr, strings.TrimSpace(selection.Text()))
		})
		// 租赁方式
		rentType := "整租"
		// 户型
		houseStruct := ""
		// 卧室数量
		roomAmount := int64(0)
		// 客厅数量
		hallAmount := int64(0)
		// 卫生间数量
		toiletAmount := int64(0)
		// 面积
		houseArea := int64(0)
		// 朝向
		houseOrient := ""
		// 匹配
		otherReg := regexp.MustCompile(`(<i(.*)>(.*)</i>| )`)
		otherDom := e.DOM.Find("div.content > div.content__aside > ul.content__aside__list > p.content__article__table > span")
		otherDom.Each(func(i int, selection *goquery.Selection) {
			otherStr := otherReg.ReplaceAllString(selection.Text(), "")
			switch i {
			case 0:
				// 租赁方式
				if otherStr == "合租" {
					rentType = otherStr
				}
			case 1:
				// 户型
				houseStruct = otherStr
				// 室厅卫
				structReg := regexp.MustCompile(`\d居室`)
				structRegFind := structReg.FindString(houseStruct)
				if len(structRegFind) > 0 {
					roomAmount, _ = strconv.ParseInt(structRegFind[0:1], 10, 64)
				} else {
					structReg = regexp.MustCompile(`\d室\d厅\d卫`)
					structRegFind = structReg.FindString(houseStruct)
					if len(structRegFind) > 0 {
						roomAmount, _ = strconv.ParseInt(structRegFind[0:1], 10, 64)
						hallAmount, _ = strconv.ParseInt(structRegFind[4:5], 10, 64)
						toiletAmount, _ = strconv.ParseInt(structRegFind[8:9], 10, 64)
					}
				}
			case 2:
				// 面积
				houseArea, _ = strconv.ParseInt(strings.Replace(otherStr, "㎡", "", 1), 10, 64)
			case 3:
				// 朝向
				houseOrient = otherStr
			}
		})
		// 楼层
		houseFloor := ""
		floorType := ""
		// 电梯
		houseElevator := ""
		// 基本信息
		infoDom := e.DOM.Find("div.content > div.content__article > div.content__article__info > ul > li")
		infoDom.Each(func(i int, selection *goquery.Selection) {
			infoStr := selection.Text()
			infoData := strings.Split(infoStr, "：")
			if len(infoData) == 2 && infoData[0] == "楼层" {
				houseFloor = infoData[1]
				// 高中低划分
				floorReg := regexp.MustCompile(`[高中低]楼层`)
				floorRegFind := floorReg.FindString(houseFloor)
				if len(floorRegFind) > 0 {
					floorType = floorRegFind[0:3]
				} else {
					floorReg = regexp.MustCompile(`\d{1,2}/\d{1,2}`)
					floorRegFind = floorReg.FindString(houseFloor)
					if len(floorRegFind) > 0 {
						floorTypeArr := strings.Split(floorRegFind, "/")
						if len(floorTypeArr) == 2 {
							floorCurr, _ := strconv.ParseInt(floorTypeArr[0], 10, 64)
							floorTotal, _ := strconv.ParseInt(floorTypeArr[1], 10, 64)
							if floorCurr > 0 && floorTotal > 0 {
								floorValue := float64(floorCurr) / float64(floorTotal)
								if floorValue <= 0.33 {
									floorType = "低"
								} else if floorValue <= 0.67 {
									floorType = "中"
								} else {
									floorType = "高"
								}
							}
						}
					}
				}
			}
			if len(infoData) == 2 && infoData[0] == "电梯" {
				houseElevator = infoData[1]
			}
		})
		// 入库
		err := S.GetC("house").Update(bson.M{"house_id": houseId}, bson.M{"$set": bson.M{
			"title":         title,
			"posted_date":   postedDate,
			"rent_amount":   rent,
			"payment_type":  payment,
			"tags":          tagStr,
			"rent_type":     rentType,
			"struct":        houseStruct,
			"room_amount":   roomAmount,
			"hall_amount":   hallAmount,
			"toilet_amount": toiletAmount,
			"area":          houseArea,
			"orient":        houseOrient,
			"floor":         houseFloor,
			"floor_type":    floorType,
			"elevator":      houseElevator,
			"update_time":   helpers.GetTimestamp(),
		}})
		fmt.Println("[Error]:", err, "[houseId]:", houseId)
	})
	// 抓取数据|房源下线
	c.OnHTML("div.offline > p.offline__title", func(e *colly.HTMLElement) {
		// 已下线
		if len(strings.TrimSpace(e.Text)) != 0 {
			// 房源ID
			houseUrl := e.Request.URL.String()
			urlSplit := strings.Split(houseUrl, "/")
			houseUrl = urlSplit[len(urlSplit)-1]
			houseId := houseUrl[0 : len(houseUrl)-5]
			// 查询
			result := house{}
			err := S.GetC("house").Find(bson.M{"house_id": houseId}).One(&result)
			// 更新
			if err == nil && result.OfflineDate == "" {
				err := S.GetC("house").Update(bson.M{"house_id": houseId}, bson.M{"$set": bson.M{"offline_date": helpers.GetDate(), "update_time": helpers.GetTimestamp()}})
				fmt.Println("[Error]:", err, "[houseId]:", houseId)
			}
		}
	})
	//
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("[Detail-URL]:", r.URL.String())
	})
	//
	c.OnError(func(r *colly.Response, err error) {
		// 错误日志
		helpers.AddFailedLogs(5, cityId, string(r.Request.URL.String()))
		fmt.Println("[Request]:", r.Request.URL, "\n[Response]:", r, "\n[Error]:", err)
	})
	//
	q.Run(c)
	//
	fmt.Printf("[%s][EXECUTE_TIMES: %d ms]\n", helpers.GetTime(), helpers.GetMicroTimestamp()-startTime)
}
