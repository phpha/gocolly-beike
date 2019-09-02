package gocolly

import (
	"config"
	"fmt"
	"github.com/gocolly/colly"
	"helpers"
	"services/mongodb"
	"strings"
)

type district struct {
	CityId       int    `bson:"city_id"`
	DistrictId   string `bson:"district_id"`
	DistrictName string `bson:"district_name"`
	CreateTime   int64  `bson:"create_time"`
	UpdateTime   int64  `bson:"update_time"`
}

// 区域
func District(cityId int) {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 开始时间
	startTime := helpers.GetMicroTimestamp()
	// 实例化
	c := colly.NewCollector()
	// 抓取数据
	c.OnHTML("div.m-filter > div.position > dl > dd > div[data-role=\"ershoufang\"] > div > a.CLICKDATA", func(e *colly.HTMLElement) {
		// 区域URL
		districtUrl := e.Attr("href")
		// 过滤
		if districtUrl[0:4] == "http" {
			return
		}
		// 分隔
		districtSplit := strings.Split(strings.Trim(districtUrl, "/"), "/")
		if len(districtSplit) < 1 {
			return
		}
		// 区域数据
		districtId := districtSplit[len(districtSplit)-1]
		districtName := strings.TrimSpace(e.Text)
		// 入库
		S.GetC("district").Insert(district{
			CityId:       cityId,
			DistrictId:   districtId,
			DistrictName: districtName,
			CreateTime:   helpers.GetTimestamp(),
			UpdateTime:   int64(0),
		})
	})
	//
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("[Request]:", r.Request.URL, "\n[Response]:", r, "\n[Error]:", err)
	})
	//
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	//
	c.Visit(fmt.Sprintf("https://%s.ke.com/xiaoqu/", strings.ToLower(config.HousePrefixMap[cityId])))
	//
	fmt.Printf("[%s][EXECUTE_TIMES: %d ms]\n", helpers.GetTime(), helpers.GetMicroTimestamp()-startTime)
}
