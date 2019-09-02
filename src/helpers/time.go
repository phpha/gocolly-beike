// 时间戳相关
package helpers

import "time"

// 获取当前时间|字符串
func GetTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 获取当前日期|字符串
func GetDate() string {
	return time.Now().Format("2006-01-02")
}

// 获取相对时间|字符串
func GetRelativeTime(years int, months int, days int) string {
	return time.Now().AddDate(years, months, days).Format("2006-01-02 15:04:05")
}

// 获取当前时间戳|秒
func GetTimestamp() int64 {
	return time.Now().UnixNano() / 1e9
}

// 获取当前时间戳|毫秒
func GetMicroTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

// 获取相对时间戳|秒
func GetRelativeTimestamp(years int, months int, days int) int64 {
	timeStr := GetRelativeTime(years, months, days)
	timeLocation, _ := time.LoadLocation("Asia/Chongqing")
	timeParse, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, timeLocation)
	return timeParse.Unix()
}

// 获取相对日期[00:00:00]时间戳|秒
func GetZeroTimestamp(years int, months int, days int) int64 {
	timeStr := time.Now().AddDate(years, months, days).Format("2006-01-02 00:00:00")
	timeLocation, _ := time.LoadLocation("Asia/Chongqing")
	timeParse, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, timeLocation)
	return timeParse.Unix()
}

// 时间字符串转时间戳|秒
func TimeToTimestamp(timeStr string) int64 {
	timeLocation, _ := time.LoadLocation("Asia/Chongqing")
	timeParse, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, timeLocation)
	return timeParse.Unix()
}
