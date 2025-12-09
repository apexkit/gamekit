package app

import (
	"fmt"
	"strings"
)

// AppGroupIDFromAppID 从子商户 app_id 解析总商户编号（取下划线前一段）。
func AppGroupIDFromAppID(appId string) (string, error) {
	segment := appId
	if idx := strings.Index(appId, "_"); idx >= 0 {
		segment = appId[:idx]
	}
	if segment == "" {
		return "", fmt.Errorf("app_id %s: missing app_group_id segment", appId)
	}
	return segment, nil
}
