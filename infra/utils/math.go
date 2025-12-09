package utils

import (
	"math"
	"math/rand/v2"
)

// Round 四舍五入函数
// num: 要处理的浮点数
// precision: 要保留的小数位数
func Round(num float64, precision int) float64 {
	// 处理特殊情况
	if math.IsNaN(num) || math.IsInf(num, 0) {
		return num
	}

	// 计算10的precision次方
	pow := math.Pow10(precision)

	// 四舍五入
	rounded := math.Round(num*pow) / pow

	return rounded
}

// 随机一个浮点型的数字，范围为 [min, max)
func RandFloat(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + rand.Float64()*(max-min)
}
