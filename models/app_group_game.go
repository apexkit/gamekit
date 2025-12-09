package models

import (
	"time"
)

const (
	AppGroupGameStatusEnabled  uint8 = 0 // 启用
	AppGroupGameStatusDisabled uint8 = 1 // 禁用
)

// AppGroupGame 总商户游戏配置，对应 app_group_game 表。
type AppGroupGame struct {
	ID         int64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement" json:"id"`
	AppGroupID string    `gorm:"column:app_group_id;type:varchar(64);not null;comment:'总商户编号';uniqueIndex:idx_app_group_game_group_game,priority:1" json:"app_group_id"`
	GameID     int64     `gorm:"column:game_id;type:bigint;not null;comment:'关联game表id';uniqueIndex:idx_app_group_game_group_game,priority:2" json:"game_id"`
	Status     uint8     `gorm:"column:status;type:tinyint unsigned;not null;default:0;comment:'状态：0启用,1禁用'" json:"status"`
	Rtp        string    `gorm:"column:rtp;type:varchar(32);default:NULL" json:"rtp"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'db更新时间'" json:"updated_at"`
}

func (AppGroupGame) TableName() string {
	return "app_group_game"
}
