package main

import (
	"config"
	"flag"
	"fmt"
	"gocolly"
	"helpers"
)

func main() {
	// 获取参数
	cityId := flag.Int("cityid", 0, "")
	// 解析参数
	flag.Parse()
	// 参数校验
	if _, ok := config.HousePrefixMap[*cityId]; !ok {
		helpers.OutputError("PARAMS_ERROR: `cityid`")
		return
	}
	// 不同版本
	if *cityId >= 101 {
		// [1]区域
		fmt.Printf("[%s]==========[1/5]==========\n", helpers.GetTime())
		gocolly.DistrictX(*cityId)
		// [2]商圈
		fmt.Printf("[%s]==========[2/5]==========\n", helpers.GetTime())
		gocolly.BizcircleX(*cityId)
		// [4]房间ID
		fmt.Printf("[%s]==========[4/5]==========\n", helpers.GetTime())
		gocolly.HouseX(*cityId)
		// [5]房间详情
		fmt.Printf("[%s]==========[5/5]==========\n", helpers.GetTime())
		gocolly.DetailX(*cityId)
	} else {
		// [1]区域
		fmt.Printf("[%s]==========[1/5]==========\n", helpers.GetTime())
		gocolly.District(*cityId)
		// [2]商圈
		fmt.Printf("[%s]==========[2/5]==========\n", helpers.GetTime())
		gocolly.Bizcircle(*cityId)
		// [3]社区
		fmt.Printf("[%s]==========[3/5]==========\n", helpers.GetTime())
		gocolly.Community(*cityId)
		// [4]房间ID
		fmt.Printf("[%s]==========[4/5]==========\n", helpers.GetTime())
		gocolly.House(*cityId)
		// [5]房间详情
		fmt.Printf("[%s]==========[5/5]==========\n", helpers.GetTime())
		gocolly.Detail(*cityId)
	}
}
