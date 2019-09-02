package gocolly

import (
	"config"
	"fmt"
	"github.com/gocolly/colly"
	"helpers"
	"services/mongodb"
	"strings"
)

// 区域|无小区数据版本
func DistrictX(cityId int) {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 开始时间
	startTime := helpers.GetMicroTimestamp()
	// 实例化
	c := colly.NewCollector()
	// 抓取数据
	c.OnHTML("div.wrapper > div > div.filter > div.filter__wrapper > ul[data-target=\"area\"] > li[data-type=\"district\"] > a", func(e *colly.HTMLElement) {
		// 区域URL
		districtUrl := e.Attr("href")
		// 区域名称
		districtName := strings.TrimSpace(e.Text)
		// 过滤
		if districtUrl[0:4] == "http" || districtName == "不限" {
			return
		}
		// 分隔
		districtSplit := strings.Split(strings.Trim(districtUrl, "/"), "/")
		if len(districtSplit) < 1 {
			return
		}
		// 区域ID
		districtId := districtSplit[len(districtSplit)-1]
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
	c.Visit(fmt.Sprintf("https://%s.zu.ke.com/zufang/", strings.ToLower(config.HousePrefixMap[cityId])))
	//
	fmt.Printf("[%s][EXECUTE_TIMES: %d ms]\n", helpers.GetTime(), helpers.GetMicroTimestamp()-startTime)
}
