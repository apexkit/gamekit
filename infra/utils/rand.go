package utils

import (
	"math/rand/v2"
	"strconv"
)

func init() {
	//rand.Seed(time.Now().UnixNano())
}

func RandString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := RandomInt(33, 126+1)
		bytes[i] = byte(b)
	}
	return string(bytes)
}

/*
* 随机数字  (start,  end]  不包含end
**/
func RandomInt(start int, end int) int {
	random := rand.IntN(end - start)
	random = start + random
	return random
}

func RandVerifyCode(len uint) string {
	ret := ""
	for i := 0; i < int(len); i++ {
		ret += strconv.Itoa(RandomInt(0, 10))
	}
	return ret
}
