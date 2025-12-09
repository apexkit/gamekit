package macro

type AppInfo struct {
	AppId      string //由qingdou分配的appid，必填
	DeviceId   string //设备id, 选填
	AppVersion string //app版本号， 必填
	AppTime    string //时间戳, 必填
	Platform   int32  //获取平台， app，小程序还是h5 定义在 Common.pb.go中
	Os         int32  //定义在 Common.pb.go中
	OsVersion  string
	Language   string
}
