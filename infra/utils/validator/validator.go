package validator

func IsTime(s string) bool {
	return timeRegex.MatchString(s)
}

func IsHexColor(s string) bool {
	return hexcolorRegex.MatchString(s)
}

func IsEmail(s string) bool {
	return emailRegex.MatchString(s)
}

func IsPhone(s string) bool {
	if len(s) < 6 {
		return false
	}
	if !phoneRegex.MatchString(s) {
		return false
	}
	if s[0] != '+' && len(s) != 11 {
		return false
	}
	if s[0] == '+' && s[1] == '8' && s[2] == '6' && len(s) != 14 {
		return false
	}
	return true
}

func IsBankCard(s string) bool {
	if len(s) < 15 || len(s) > 20 {
		return false
	}
	data := []byte(s)
	len := len(data)
	nCheck := int(data[len-1] - '0')
	sum := 0

	j := 0
	for i := len - 2; i >= 0; i-- {
		k := int(data[i] - '0')
		if j%2 == 0 {
			k *= 2
			k = k/10 + k%10
		}
		sum = (sum + k) % 10
		j++
	}
	if sum%10 != 0 {
		sum = 10 - sum
	}
	return sum == nCheck
}
