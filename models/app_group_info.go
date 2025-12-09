package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	AppGroupInfoStatusDisabled uint8 = 0 // 禁用
	AppGroupInfoStatusEnabled  uint8 = 1 // 启用
)

// AppGroupInfo 总商户信息，对应 game.app_group_info 表。
type AppGroupInfo struct {
	Id uint64 `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`

	GroupId         string  `gorm:"column:group_id;type:varchar(64);not null;uniqueIndex:idx_app_group_info_group_id;comment:总商户编号" json:"groupId"`
	Name            string  `gorm:"column:name;type:varchar(128);not null;comment:总商户名称" json:"name"`
	AccessKey       string  `gorm:"column:access_key;type:varchar(128);comment:总商户Key" json:"accessKey"`
	AccessSecret    string  `gorm:"column:access_secret;type:varchar(255);comment:总商户密钥" json:"accessSecret"`
	CallBackUrl     string  `gorm:"column:call_back_url;type:longtext;comment:回调地址" json:"callBackUrl"`
	Currency        string  `gorm:"column:currency;type:varchar(64);comment:货币类型" json:"currency"`
	Rtp             string  `gorm:"column:rtp;type:varchar(32);comment:默认RTP" json:"rtp"`
	Rate            *float64 `gorm:"column:rate;type:double;comment:费率" json:"rate"`
	TriggerWinIfZero uint8  `gorm:"column:trigger_win_if_zero;type:tinyint unsigned;not null;default:0;comment:派奖为0是否回调：0否,1是" json:"triggerWinIfZero"`
	Status          uint8   `gorm:"column:status;type:tinyint unsigned;default:1;comment:状态,0禁用,1启用" json:"status"`
	Note            string  `gorm:"column:note;type:text;comment:备注" json:"note"`

	CreateTime time.Time      `gorm:"column:create_time;type:datetime(3);autoCreateTime;comment:创建时间" json:"createTime"`
	UpdateTime time.Time      `gorm:"column:update_time;type:datetime(3);autoUpdateTime;comment:更新时间" json:"updateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index:idx_app_group_info_deleted_at" json:"-"`
}

func (AppGroupInfo) TableName() string {
	return "app_group_info"
}
