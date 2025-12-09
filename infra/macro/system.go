package macro

const (
	ClientType_ANDROID_APP uint64 = 1 << 1
	ClientType_ANDROID_H5  uint64 = 1 << 2
	ClientType_IOS_APP     uint64 = 1 << 3
	ClientType_IOS_H5      uint64 = 1 << 4

	ClientType_ANDROID_WebChatMP uint64 = 1 << 5
	ClientType_IOS_WebChatMP     uint64 = 1 << 6
)

const (
	PayType_PayIn  = 1 << 1
	PayType_PayOut = 1 << 2
)

func hasClientType(t, flag uint64) bool {
	return flag&t == t
}
