package gocolly

import (
	"config"
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"helpers"
	"services/mongodb"
	"strings"
)

type community struct {
	CityId        int    `bson:"city_id"`
	DistrictId    string `bson:"district_id"`
	BizcircleId   string `bson:"bizcircle_id"`
	CommunityId   string `bson:"community_id"`
	CommunityName string `bson:"community_name"`
	CreateTime    int64  `bson:"create_time"`
	UpdateTime    int64  `bson:"update_time"`
}

// 抓取小区
func Community(cityId int) {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 开始时间
	startTime := helpers.GetMicroTimestamp()
	// 实例化
	c := colly.NewCollector()
	// URL队列
	q, _ := queue.New(
		config.ThreadsNum["community"],
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)
	// 获取URL
	for _, v := range communityPageUrl(cityId) {
		// 添加队列
		q.AddURL(v)
	}
	// 抓取数据
	c.OnHTML("ul.listContent > li.xiaoquListItem > div.info", func(e *colly.HTMLElement) {
		// 小区
		communityDom := e.DOM.Find("div.title > a")
		communityUrl, _ := communityDom.Attr("href")
		// 分隔
		communitySplit := strings.Split(strings.Trim(communityUrl, "/"), "/")
		if len(communitySplit) < 1 {
			return
		}
		// 小区数据
		communityId := communitySplit[len(communitySplit)-1]
		communityName := communityDom.Text()
		// 区域
		districtDom := e.DOM.Find("div.positionInfo > a.district")
		districtUrl, _ := districtDom.Attr("href")
		// 分隔
		districtSplit := strings.Split(strings.Trim(districtUrl, "/"), "/")
		if len(districtSplit) < 1 {
			return
		}
		// 区域ID
		districtId := districtSplit[len(districtSplit)-1]
		// 商圈
		bizcircleDom := e.DOM.Find("div.positionInfo > a.bizcircle")
		bizcircleUrl, _ := bizcircleDom.Attr("href")
		// 分隔
		bizcircleSplit := strings.Split(strings.Trim(bizcircleUrl, "/"), "/")
		if len(bizcircleSplit) < 1 {
			return
		}
		// 商圈ID
		bizcircleId := bizcircleSplit[len(bizcircleSplit)-1]
		// 入库
		err := S.GetC("community").Insert(community{
			CityId:        cityId,
			DistrictId:    districtId,
			BizcircleId:   bizcircleId,
			CommunityId:   communityId,
			CommunityName: communityName,
			CreateTime:    helpers.GetTimestamp(),
			UpdateTime:    int64(0),
		})
		// 日志
		if err != nil {
			fmt.Println("[Error]:", err, "[communityId]:", communityId)
		}
	})
	//
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("[Community-URL]:", r.URL.String())
	})
	//
	c.OnError(func(r *colly.Response, err error) {
		// 错误日志
		helpers.AddFailedLogs(2, cityId, string(r.Request.URL.String()))
		fmt.Println("[Request]:", r.Request.URL, "\n[Response]:", r, "\n[Error]:", err)
	})
	//
	q.Run(c)
	//
	fmt.Printf("[%s][EXECUTE_TIMES: %d ms]\n", helpers.GetTime(), helpers.GetMicroTimestamp()-startTime)
}

// 小区分页URL
func communityPageUrl(cityId int) []string {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 切片
	var data []string
	// 失败记录
	for _, v := range helpers.GetFailedLogs(2, cityId) {
		data = append(data, v)
	}
	// 总页数
	totalPage := 0
	// 实例化
	c := colly.NewCollector()
	// URL队列
	q, _ := queue.New(
		config.ThreadsNum["community"],
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)
	// 失败记录
	failed := helpers.GetFailedLogs(1, cityId)
	for _, v := range failed {
		q.AddURL(v)
	}
	// 时间范围
	endTime := helpers.GetZeroTimestamp(0, 0, 0)
	// 查询商圈
	result := []bizcircle{}
	err := S.GetC("bizcircle").Find(bson.M{"city_id": cityId, "update_time": bson.M{"$lte": endTime}}).All(&result)
	if err != nil {
		return data
	}
	// 循环
	for _, v := range result {
		// 当前商圈
		bizcircleId := v.BizcircleId
		// 当前页面
		currUrl := fmt.Sprintf("https://%s.ke.com/xiaoqu/%s/", strings.ToLower(config.HousePrefixMap[cityId]), bizcircleId)
		// 添加队列
		q.AddURL(currUrl)
	}
	// 抓取数据
	c.OnHTML("div.content > div.leftContent > div.contentBottom > div.page-box > div.house-lst-page-box", func(e *colly.HTMLElement) {
		// Json-string To Struct
		var page helpers.Pager
		if err := json.Unmarshal([]byte(e.Attr("page-data")), &page); err == nil {
			totalPage = page.TotalPage
		}
		// 获取页数失败
		if totalPage == 0 {
			// 错误日志
			helpers.AddFailedLogs(1, cityId, e.Request.URL.String())
		} else {
			// 完整URL
			for i := 1; i <= totalPage; i++ {
				data = append(data, fmt.Sprintf("%spg%d/", e.Request.URL.String(), i))
			}
		}
	})
	//
	c.OnError(func(r *colly.Response, err error) {
		// 错误日志
		helpers.AddFailedLogs(1, cityId, string(r.Request.URL.String()))
		fmt.Println("[Request]:", r.Request.URL, "\n[Response]:", r, "\n[Error]:", err)
	})
	//
	c.OnRequest(func(r *colly.Request) {
		// 商圈ID
		bizcircleSplit := strings.Split(strings.Trim(r.URL.String(), "/"), "/")
		if len(bizcircleSplit) < 1 {
			return
		}
		bizcircleId := bizcircleSplit[len(bizcircleSplit)-1]
		// 更新
		S.GetC("bizcircle").Update(bson.M{"bizcircle_id": bizcircleId}, bson.M{"$set": bson.M{"update_time": helpers.GetTimestamp()}})
		fmt.Println("[Bizcircle-URL]:", r.URL.String())
	})
	//
	q.Run(c)
	//
	return data
}
