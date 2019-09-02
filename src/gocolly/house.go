package gocolly

import (
	"config"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"helpers"
	"services/mongodb"
	"strconv"
	"strings"
)

type house struct {
	CityId       int      `bson:"city_id"`
	DistrictId   string   `bson:"district_id"`
	BizcircleId  string   `bson:"bizcircle_id"`
	CommunityId  string   `bson:"community_id"`
	HouseId      string   `bson:"house_id"`
	BrandName    string   `bson:"barnd_name"`
	PostedDate   string   `bson:"posted_date"`
	OfflineDate  string   `bson:"offline_date"`
	Area         int      `bson:"area"`
	Elevator     string   `bson:"elevator"`
	Floor        string   `bson:"floor"`
	FloorType    string   `bson:"floor_type"`
	Orient       string   `bson:"orient"`
	HallAmount   int      `bson:"hall_amount"`
	RoomAmount   int      `bson:"room_amount"`
	ToiletAmount int      `bson:"toilet_amount"`
	PaymentType  string   `bson:"payment_type"`
	RentAmount   int      `bson:"rent_amount"`
	RentType     string   `bson:"rent_type"`
	HouseStruct  string   `bson:"struct"`
	Tags         []string `bson:"tags"`
	Title        string   `bson:"title"`
	ImageUrl     string   `bson:"image_url"`
	CreateTime   int64    `bson:"create_time"`
	UpdateTime   int64    `bson:"update_time"`
}

// 抓取房源
func House(cityId int) {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 开始时间
	startTime := helpers.GetMicroTimestamp()
	// 实例化
	c := colly.NewCollector()
	// URL队列
	q, _ := queue.New(
		config.ThreadsNum["house"],
		&queue.InMemoryQueueStorage{MaxSize: 100000},
	)
	// 获取URL
	for _, v := range housePageUrl(cityId) {
		// 添加队列
		q.AddURL(v)
	}
	// 抓取数据
	c.OnHTML("div.content__article > div.content__list > div.content__list--item", func(e *colly.HTMLElement) {
		// 区域
		districtId := ""
		districtDom := e.DOM.Find("div.content__list--item--main > p.content__list--item--des > a:nth-of-type(1)")
		districtIdStr, exists := districtDom.Attr("href")
		if !exists || len(districtIdStr) <= 8 || districtIdStr[1:7] != "zufang" {
			fmt.Println("[Error] districtIdStr:", districtIdStr)
			return
		}
		districtId = districtIdStr[8 : len(districtIdStr)-1]
		// 商圈
		bizcircleId := ""
		bizcircleDom := e.DOM.Find("div.content__list--item--main > p.content__list--item--des > a:nth-of-type(2)")
		bizcircleIdStr, exists := bizcircleDom.Attr("href")
		if !exists || len(bizcircleIdStr) <= 8 || bizcircleIdStr[1:7] != "zufang" {
			fmt.Println("[Error] bizcircleIdStr:", bizcircleIdStr)
			return
		}
		bizcircleId = bizcircleIdStr[8 : len(bizcircleIdStr)-1]
		// 社区
		communityId := ""
		communityDom := e.DOM.Find("div.content__list--item--main > p.content__list--item--des > a:nth-of-type(3)")
		communityIdStr, exists := communityDom.Attr("href")
		if !exists || len(communityIdStr) <= 9 || communityIdStr[1:7] != "zufang" {
			fmt.Println("[Error] communityIdStr:", communityIdStr)
			return
		}
		communityId = communityIdStr[9 : len(communityIdStr)-1]
		// 品牌
		brandDom := e.DOM.Find("div.content__list--item--main > p.content__list--item--brand")
		brandName := strings.TrimSpace(brandDom.Text())
		// 图片
		imageDom := e.DOM.Find("a.content__list--item--aside > img")
		imageUrl, _ := imageDom.Attr("data-src")
		// 房源
		houseDom := e.DOM.Find("a.content__list--item--aside")
		houseUrl, _ := houseDom.Attr("href")
		urlSplit := strings.Split(houseUrl, "/")
		houseUrl = urlSplit[len(urlSplit)-1]
		urlSplit = strings.Split(houseUrl, "?")
		houseUrl = urlSplit[0]
		// 房间URL格式校验
		if len(houseUrl) < 5 || houseUrl[len(houseUrl)-5:] != ".html" {
			fmt.Println("[Error] houseUrl", houseUrl)
			return
		}
		houseId := houseUrl[0 : len(houseUrl)-5]
		// 房间ID格式校验
		if houseId[0:2] != config.HousePrefixMap[cityId] {
			fmt.Println("[Error] houseId:", houseId)
			return
		}
		// 入库
		err := S.GetC("house").Insert(house{
			CityId:      cityId,
			DistrictId:  districtId,
			BizcircleId: bizcircleId,
			CommunityId: communityId,
			HouseId:     houseId,
			BrandName:   brandName,
			ImageUrl:    imageUrl,
			CreateTime:  helpers.GetTimestamp(),
			UpdateTime:  int64(0),
		})
		// 日志
		if err != nil {
			fmt.Println("[Error]:", err, "[houseId]:", houseId)
		}
	})
	//
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("[House-URL]:", r.URL.String())
	})
	//
	c.OnError(func(r *colly.Response, err error) {
		// 错误日志
		helpers.AddFailedLogs(4, cityId, string(r.Request.URL.String()))
		fmt.Println("[Request]:", r.Request.URL, "\n[Response]:", r, "\n[Error]:", err)
	})
	//
	q.Run(c)
	//
	fmt.Printf("[%s][EXECUTE_TIMES: %d ms]\n", helpers.GetTime(), helpers.GetMicroTimestamp()-startTime)
}

// 房源分页URL
func housePageUrl(cityId int) []string {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 切片
	var data []string
	// 失败记录
	for _, v := range helpers.GetFailedLogs(4, cityId) {
		data = append(data, v)
	}
	// 实例化
	c := colly.NewCollector()
	// URL队列
	q, _ := queue.New(
		config.ThreadsNum["house"],
		&queue.InMemoryQueueStorage{MaxSize: 100000},
	)
	// 失败记录
	failed := helpers.GetFailedLogs(3, cityId)
	for _, v := range failed {
		q.AddURL(v)
	}
	// 时间范围
	endTime := helpers.GetZeroTimestamp(0, 0, 0)
	// 查询小区
	result := []community{}
	err := S.GetC("community").Find(bson.M{"city_id": cityId, "update_time": bson.M{"$lte": endTime}}).All(&result)
	if err != nil {
		return data
	}
	// 循环
	for _, v := range result {
		// 当前小区
		communityId := v.CommunityId
		// 当前页面
		currUrl := fmt.Sprintf("https://%s.zu.ke.com/zufang/c%s/", strings.ToLower(config.HousePrefixMap[cityId]), communityId)
		// 添加队列
		q.AddURL(currUrl)
	}
	// 抓取数据
	c.OnHTML("div.content__article", func(e *colly.HTMLElement) {
		// 总条数
		totalNum, _ := strconv.Atoi(e.DOM.Find("p.content__title > span.content__title--hl").Text())
		if totalNum == 0 || totalNum > 2000 {
			return
		}
		// 总页数
		totalPage := 0
		totalPageDom, exists := e.DOM.Find("div.content__pg").Attr("data-totalpage")
		if exists {
			totalPage, _ = strconv.Atoi(totalPageDom)
		}
		// 数据异常
		if totalPage == 0 {
			// 错误日志
			helpers.AddFailedLogs(3, cityId, e.Request.URL.String())
		} else {
			// 完整URL
			for i := 1; i <= totalPage; i++ {
				data = append(data, strings.Replace(e.Request.URL.String(), "zufang/c", fmt.Sprintf("zufang/pg%dc", i), 1))
			}
		}
	})
	//
	c.OnError(func(r *colly.Response, err error) {
		// 错误日志
		helpers.AddFailedLogs(3, cityId, string(r.Request.URL.String()))
		fmt.Println("[Request]:", r.Request.URL, "\n[Response]:", r, "\n[Error]:", err)
	})
	//
	c.OnRequest(func(r *colly.Request) {
		// 小区ID
		communitySplit := strings.Split(strings.Trim(r.URL.String(), "/"), "c")
		if len(communitySplit) < 1 {
			return
		}
		communityId := communitySplit[len(communitySplit)-1]
		// 更新
		S.GetC("community").Update(bson.M{"community_id": communityId}, bson.M{"$set": bson.M{"update_time": helpers.GetTimestamp()}})
		fmt.Println("[Community-URL]:", r.URL.String())
	})
	//
	q.Run(c)
	//
	return data
}
