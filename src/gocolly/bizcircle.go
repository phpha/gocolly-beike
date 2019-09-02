package gocolly

import (
	"config"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/gocolly/colly"
	"helpers"
	"services/mongodb"
	"strings"
)

type bizcircle struct {
	CityId        int    `bson:"city_id"`
	DistrictId    string `bson:"district_id"`
	BizcircleId   string `bson:"bizcircle_id"`
	BizcircleName string `bson:"bizcircle_name"`
	CreateTime    int64  `bson:"create_time"`
	UpdateTime    int64  `bson:"update_time"`
}

// 商圈
func Bizcircle(cityId int) {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 开始时间
	startTime := helpers.GetMicroTimestamp()
	// 时间范围
	endTime := helpers.GetZeroTimestamp(0, 0, 0)
	// 查询区域
	data := []bizcircle{}
	err := S.GetC("district").Find(bson.M{"city_id": cityId, "update_time": bson.M{"$lte": endTime}}).All(&data)
	if err != nil {
		return
	}
	// 循环
	for _, v := range data {
		// 当前区域
		districtId := v.DistrictId
		// 实例化
		c := colly.NewCollector()
		// 抓取数据
		c.OnHTML("div.m-filter > div.position > dl > dd > div[data-role=\"ershoufang\"] > div:nth-of-type(2) > a", func(e *colly.HTMLElement) {
			// 商圈
			bizcircleUrl := e.Attr("href")
			// 分隔
			bizcircleSplit := strings.Split(strings.Trim(bizcircleUrl, "/"), "/")
			if len(bizcircleSplit) < 1 {
				return
			}
			// 商圈数据
			bizcircleId := bizcircleSplit[len(bizcircleSplit)-1]
			bizcircleName := strings.TrimSpace(e.Text)
			// 入库
			err := S.GetC("bizcircle").Insert(bizcircle{
				CityId:        cityId,
				DistrictId:    districtId,
				BizcircleId:   bizcircleId,
				BizcircleName: bizcircleName,
				CreateTime:    helpers.GetTimestamp(),
				UpdateTime:    int64(0),
			})
			// 日志
			if err != nil {
				fmt.Println("[Error]:", err, "[bizcircleId]:", bizcircleId)
			}
		})
		//
		c.OnRequest(func(r *colly.Request) {
			// 更新
			S.GetC("district").Update(bson.M{"district_id": districtId}, bson.M{"$set": bson.M{"update_time": helpers.GetTimestamp()}})
			fmt.Println("Visiting", r.URL.String())
		})
		//
		c.OnError(func(r *colly.Response, err error) {
			fmt.Println("[Request]:", r.Request.URL, "\n[Response]:", r, "\n[Error]:", err)
		})
		//
		c.Visit(fmt.Sprintf("https://%s.ke.com/xiaoqu/%s/", strings.ToLower(config.HousePrefixMap[cityId]), districtId))
	}
	//
	fmt.Printf("[%s][EXECUTE_TIMES: %d ms]\n", helpers.GetTime(), helpers.GetMicroTimestamp()-startTime)
}
