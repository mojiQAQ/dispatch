package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateUUID() string {
	// 使用当前时间作为随机数生成器的种子
	rand.Seed(time.Now().UnixNano())

	// 获取当前时间
	currentTime := time.Now()

	// 根据当前时间生成一个时间格式的随机数
	formattedTime := currentTime.Format("20060102150405") // 时间格式

	// 生成一个随机数并附加到时间格式的字符串末尾
	randomNumber := rand.Intn(1000000)                         // 生成一个0到99之间的随机数
	formattedRandomNumber := fmt.Sprintf("%06d", randomNumber) // 格式化为两位数

	// 将随机数附加到时间格式的字符串末尾
	randomTime := formattedTime + formattedRandomNumber
	return randomTime
}
