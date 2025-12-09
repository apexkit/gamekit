package utils

import "strings"

var langMap = map[string]string{
	"zh-CN": "zh", "en-US": "en", "th-TH": "th", "vi-VN": "vi",
	"id-ID": "id", "hi-IN": "hi", "ta-IN": "ta", "my-MM": "my",
	"ja-JP": "ja", "ms-MY": "ms", "ko-KR": "ko", "bn-IN": "bn",
	"es-AR": "es", "pt-BR": "pt", "it-IT": "it", "sv-SE": "sv",
	"de-DE": "de", "da-DK": "da", "ro-RO": "ro", "nl-NL": "nl",
	"tr-TR": "tr", "ru-RU": "ru", "el-GR": "el", "fr-FR": "fr",
}

// 反向映射（简写 -> 复杂）
var reverseLangMap = func() map[string]string {
	m := make(map[string]string)
	for k, v := range langMap {
		m[v] = k
	}
	return m
}()

// GetLangAbbr 获取语言缩写（zh-CN -> zh）
// 如果输入已是缩写（en / zh），直接返回；否则返回默认 "en"
func GetLangAbbr(lang string) string {
	lang = strings.TrimSpace(lang)
	if len(lang) == 2 {
		return strings.ToLower(lang)
	}
	if abbr, ok := langMap[lang]; ok {
		return abbr
	}
	return "en"
}

// GetFullLang 获取完整语言标识（zh -> zh-CN）
// 若找不到匹配，默认返回 "en-US"
func GetFullLang(abbr string) string {
	abbr = strings.TrimSpace(strings.ToLower(abbr))
	if full, ok := reverseLangMap[abbr]; ok {
		return full
	}
	return "en-US"
}
