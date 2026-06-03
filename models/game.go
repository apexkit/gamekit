package models

import (
	"time"
)

const (
	ProxyModel_Local = "Local" // 本地模式

	GameStatusEnabled  uint8 = 0 // 启用
	GameStatusDisabled uint8 = 1 // 禁用
)

type Game struct {
	PlatformGameID int64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement;comment:'id, 游戏编码'" json:"id"`
	GameId         string    `gorm:"column:game_id;type:varchar(32);not null;comment:'游戏id';index:game_game_id_udx,priority:2" json:"game_id"`
	GameName       string    `gorm:"column:game_name;type:varchar(64);default:NULL;comment:'游戏名称'" json:"game_name"`
	GameFullName   string    `gorm:"column:game_full_name;type:varchar(512);default:NULL;comment:'游戏Full名称'" json:"game_full_name"`
	GameIcon       string    `gorm:"column:game_icon;type:varchar(512);default:NULL;comment:'游戏icon'" json:"game_icon"`
	GameType       string    `gorm:"column:game_type;type:varchar(32);default:NULL;comment:'游戏类型:slot'" json:"game_type"`
	GameBrand      string    `gorm:"column:game_brand;type:varchar(32);default:NULL;comment:'游戏厂商:jili,pg';index:game_game_id_udx,priority:1" json:"game_brand"`
	Status         uint8     `gorm:"column:status;type:tinyint unsigned;not null;default:0;comment:'状态：0启用,1禁用'" json:"status"`
	Rtp            string    `gorm:"column:rtp;type:varchar(32);default:NULL" json:"rtp"`
	RtpModel       string    `gorm:"column:rtp_model;type:varchar(32);default:NULL;comment:'RTP模式'" json:"rtp_model"`
	ProxyModel     string    `gorm:"column:proxy_model;type:varchar(16);default:NULL;comment:'代理模式, 默认是Local'" json:"proxy_model"`
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'db更新时间'" json:"updated_at"`
}

func (Game) TableName() string {
	return "game"
}
