package helpers

import (
	"fmt"
	"github.com/globalsign/mgo/bson"
	"services/mongodb"
)

type logs struct {
	Id         bson.ObjectId `bson:"_id"`
	TypeId     int           `bson:"type_id"`
	CityId     int           `bson:"city_id"`
	Url        string        `bson:"url"`
	CreateTime int64         `bson:"create_time"`
}

type Pager struct {
	TotalPage int `json:"totalPage"`
	CurPage   int `json:"curPage"`
}

// 写入失败记录
func AddFailedLogs(typeId int, cityId int, url string) bool {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 类型[1]小区列表分页失败[2]获取小区ID失败[3]房源列表分页失败[4]获取房源ID失败
	err := S.GetC("logs").Insert(logs{
		TypeId:     typeId,
		CityId:     cityId,
		Url:        url,
		CreateTime: GetTimestamp(),
	})
	// 返回
	if err != nil {
		return false
	}
	return true
}

// 获取失败记录
func GetFailedLogs(typeId int, cityId int) []string {
	// 数据库
	S := mongodb.GetS()
	defer S.Close()
	// 切片
	var data []string
	// 查询
	result := []logs{}
	err := S.GetC("logs").Find(bson.M{"type_id": typeId, "city_id": cityId}).All(&result)
	if err != nil {
		return data
	}
	// 循环
	for _, v := range result {
		// 删除
		err := S.GetC("logs").RemoveId(v.Id)
		// 赋值
		if err != nil {
			data = append(data, v.Url)
		}
	}
	// 返回
	return data
}

// 输出错误
func OutputError(errMsg string) {
	fmt.Printf("[%s][ERROR] %s\n", GetTime(), errMsg)
}
