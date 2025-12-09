package grpc

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/metadata"
	log "github.com/sirupsen/logrus"
)

type PRCHelper struct {
}

func GetMd(md metadata.Metadata, key string) (string, bool) {
	v, ok := md[strings.ToLower(key)]
	if !ok || len(v) == 0 {
		return "", false

	}
	return v[0], true
}

// 判断是否是从admin请求过来的消息
func (r *PRCHelper) IsAdmin(ctx context.Context) bool {
	md, ok := metadata.FromServerContext(ctx)
	if ok {

		uidStr, exit := GetMd(md, "x-md-global-is_admin")
		if !exit {
			return false
		}
		if uidStr == "true" {
			return true
		}
	} else {
		log.Debug("get metadata failed")
	}
	return false
}

func (r *PRCHelper) GetUid(ctx context.Context) uint64 {
	md, ok := metadata.FromServerContext(ctx)

	if ok {
		uidStr, exit := GetMd(md, "x-md-global-uid")
		if !exit {
			return 0
		}
		if uid, err := strconv.ParseUint(uidStr, 10, 64); err == nil {
			return uid
		}
	} else {
		log.Debug("get metadata failed")
	}
	return 0
}

func (r *PRCHelper) GetRemoteIp(ctx context.Context) string {
	md, ok := metadata.FromServerContext(ctx)
	if ok {
		ip, exit := GetMd(md, "x-md-global-remote_ip")
		if !exit {
			return ""
		}

		return ip
	} else {
		log.Debug("get metadata failed")
	}
	return ""
}

func (r *PRCHelper) GetAppId(ctx context.Context) string {
	return r.GetExtra(ctx, "x-md-global-appid")
}

func (r *PRCHelper) GetDeviceId(ctx context.Context) string {
	return r.GetExtra(ctx, "x-md-global-deviceId")
}

func (r *PRCHelper) GetPlatform(ctx context.Context) uint64 {
	platform := r.GetExtra(ctx, "x-md-global-platform")
	if platform == "" {
		return 0
	}
	num, err := strconv.ParseUint(platform, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func (r *PRCHelper) GetExtra(ctx context.Context, key string) string {
	md, ok := metadata.FromServerContext(ctx)
	if ok {
		ip, exit := GetMd(md, key)
		if !exit {
			return ""
		}

		return ip
	} else {
		log.Debugf("get metadata key:%s failed", key)
	}
	return ""
}
