package models

import "time"

// AppGameRecord 商家游戏记录，对应 game.app_game_record_{appId} 分表。
type AppGameRecord struct {
	ID                  int64      `gorm:"column:id;type:bigint;primaryKey;autoIncrement" json:"id"`
	AppID               string     `gorm:"column:app_id;type:varchar(32);not null;uniqueIndex:udx_app_transactionid,priority:1;index:idx_app_round,priority:1;index:idx_app_player_game_round,priority:1;index:idx_appid_createdat,priority:1" json:"app_id"`
	PlayerID            string     `gorm:"column:player_id;type:varchar(64);not null;index:idx_app_player_game_round,priority:2" json:"player_id"`
	GameID              string     `gorm:"column:game_id;type:varchar(32);not null;index:idx_app_player_game_round,priority:4;comment:厂商游戏ID" json:"game_id"`
	GameBrand           string     `gorm:"column:game_brand;type:varchar(32);not null;index:idx_app_player_game_round,priority:3;comment:游戏厂商" json:"game_brand"`
	PlatformGameID      int64      `gorm:"column:platform_game_id;type:bigint;comment:平台游戏ID(game表主键id)" json:"platform_game_id"`
	GameType            string     `gorm:"column:game_type;type:varchar(32)" json:"game_type"`
	RoundID             string     `gorm:"column:round_id;type:varchar(64);index:idx_app_round,priority:2;index:idx_app_player_game_round,priority:5" json:"round_id"`
	PreRoundID          string     `gorm:"column:pre_round_id;type:varchar(64)" json:"pre_round_id"`
	TransactionID       string     `gorm:"column:transaction_id;type:varchar(64);uniqueIndex:udx_app_transactionid,priority:2" json:"transaction_id"`
	PreTransactionID    string     `gorm:"column:pre_transaction_id;type:varchar(64);comment:前置交易ID" json:"pre_transaction_id"`
	Currency            string     `gorm:"column:currency;type:varchar(32)" json:"currency"`
	Rtp                 string     `gorm:"column:rtp;type:varchar(32)" json:"rtp"`
	Bet                 float64    `gorm:"column:bet;type:decimal(16,2)" json:"bet"`
	Win                 float64    `gorm:"column:win;type:decimal(16,2)" json:"win"`
	Status              string     `gorm:"column:status;type:varchar(10);default:ENABLE" json:"status"`
	BetTime             *time.Time `gorm:"column:bet_time;type:datetime(3)" json:"bet_time"`
	WinTime             *time.Time `gorm:"column:win_time;type:datetime(3)" json:"win_time"`
	StatusTime          *time.Time `gorm:"column:status_time;type:datetime(3)" json:"status_time"`
	TraceID             string     `gorm:"column:trace_id;type:varchar(64)" json:"trace_id"`
	CreatedAt           time.Time  `gorm:"column:created_at;type:datetime(3);autoCreateTime;index:idx_appid_createdat,priority:2;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;type:datetime(3);autoUpdateTime;comment:db更新时间" json:"updated_at"`
	GameData            []byte     `gorm:"column:game_data;type:longblob" json:"game_data"`
	RoundModel          string     `gorm:"column:round_model;type:varchar(32)" json:"round_model"`
	WinTransactionId    string     `gorm:"column:win_transaction_id;type:varchar(64)" json:"win_transaction_id"`
	RefundTransactionId string     `gorm:"column:refund_transaction_id;type:varchar(64)" json:"refund_transaction_id"`
	PreBalance          float64    `gorm:"column:pre_balance;type:decimal(16,2);comment:投注前余额" json:"pre_balance"`
	PostBalance         float64    `gorm:"column:post_balance;type:decimal(16,2)" json:"post_balance"`
	WinBalance          float64    `gorm:"column:win_balance;type:decimal(16,2);comment:结算后余额" json:"win_balance"`
	RefundBalance       float64    `gorm:"column:refund_balance;type:decimal(16,2);comment:退款后余额" json:"refund_balance"`
	IsFree              bool       `gorm:"column:is_free;type:tinyint(1);default:0;comment:是否免费模式: 0否，1是" json:"is_free"`
	Note                string     `gorm:"column:note;type:varchar(256);comment:订单备注" json:"note"`
}

// AppGameRecord 状态常量定义
const (
	AppGameRecordStatusInit     = "INIT"     // 初始化
	AppGameRecordStatusBet      = "BET"      // 下注完成
	AppGameRecordStatusSettled  = "SETTLED"  // 结算完成
	AppGameRecordStatusCanceled = "CANCELED" // 撤销完成
	AppGameRecordStatusError    = "ERROR"    // 故障
)

// 可选：定义状态描述映射
var AppGameRecordStatusDesc = map[string]string{
	AppGameRecordStatusInit:     "初始化",
	AppGameRecordStatusBet:      "下注完成",
	AppGameRecordStatusSettled:  "结算完成",
	AppGameRecordStatusCanceled: "撤销完成",
	AppGameRecordStatusError:    "故障",
}

// TableName 设置表名
//func (AppGameRecord) TableName() string {
//	return "app_game_record"
//}
