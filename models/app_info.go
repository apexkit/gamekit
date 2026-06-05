package models

import (
	"time"

	"gorm.io/gorm"
)

// AppInfo 商户应用信息，对应 game.app_info 表。
type AppInfo struct {
	Id uint64 `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`

	Name        string `gorm:"column:name;type:longtext;comment:商户名称" json:"name"`
	AppId       string `gorm:"column:app_id;type:varchar(32);uniqueIndex:idx_app_info_app_id;comment:应用ID" json:"appId"`
	CallBackUrl string `gorm:"column:call_back_url;type:longtext;comment:api回调接口,需要商户提供" json:"callBackUrl"`
	Currency        string `gorm:"column:currency;type:longtext;comment:货币类型" json:"currency"`
	AccessKeySecret string `gorm:"column:access_secret;type:longtext;comment:访问密钥" json:"accessKeySecret"`
	Country     string `gorm:"column:country;type:longtext;comment:国家如中国cn,美国us" json:"country"`
	TimeZone    string `gorm:"column:time_zone;type:varchar(191);default:Asia/Kolkata;comment:时区" json:"timeZone"`
	Rtp         string `gorm:"column:rtp;type:varchar(191);default:95;comment:默认rtp" json:"rtp"`

	State            uint8    `gorm:"column:state;type:tinyint unsigned;default:0;comment:状态,0正常,1禁用" json:"state"`
	Rate             *float64 `gorm:"column:rate;type:double;comment:费率" json:"rate"`
	Note             string   `gorm:"column:note;type:longtext;comment:备注" json:"note"`
	TriggerWinIfZero *uint8   `gorm:"column:trigger_win_if_zero;type:tinyint(1);comment:派奖为0是否回调：0否, 1是" json:"triggerWinIfZero"`

	CreateTime        time.Time      `gorm:"column:create_time;type:datetime(3);autoCreateTime;comment:创建时间" json:"createTime"`
	UpdateTime        time.Time      `gorm:"column:update_time;type:datetime(3);autoUpdateTime;comment:更新时间" json:"updateTime"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index:idx_app_info_deleted_at" json:"-"`
	ShardingState     uint8          `gorm:"column:sharding_state;type:tinyint unsigned;default:0;comment:分表状态,0否,1是" json:"shardingState"`
	ShardingStartDate *time.Time     `gorm:"column:sharding_start_date;type:datetime(3);comment:分表开始日期" json:"shardingStartDate"`
	AppGroupId        *uint64        `gorm:"column:app_group_id;type:bigint unsigned;index:idx_app_info_app_group_id;comment:总商户ID" json:"appGroupId"`
}

func (AppInfo) TableName() string {
	return "app_info"
}
