package model

import "time"

/**
* 游戏的数据
**/
type JDBGameInfo struct {
	Id uint `gorm:"primaryKey;comment:id;"`

	TypeId   string `gorm:"comment:游戏类型id;"`
	GameId   uint   `gorm:"comment:游戏id;uniqueIndex;"`
	GameName string `gorm:"comment:游戏的名称;"`

	GameType string `gorm:"comment:游戏类型:slot,fish,table,crash;"` // 游戏类型:slot,fish,table,crash

	GameResVersion string `gorm:"comment:游戏资源版本;"` // 游戏资源版本

	Data string `gorm:"comment:info数据;"`

	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间;"`
	UpdateTime time.Time `gorm:"autoCreateTime;comment:创建时间;"`
}

// 备忘录
func (JDBGameInfo) TableName() string {
	return "jdb_info"
}
