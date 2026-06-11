package sharding

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/apexkit/gamekit/models"
)

// ShardMode 表示分片模式
type ShardMode int

const (
	OnlyMain       ShardMode = iota // 只有主表（小商户）
	MainAndHistory                  // 主表 + 历史表（大商户）
)

// AppGameRecordShardRule 商户分表规则
type AppGameRecordShardRule struct {
	AppID             string
	Mode              ShardMode
	shardingStartDate *time.Time
}

// AppGameRecordRouter 分表路由器
type AppGameRecordRouter struct {
	TablePrefix string
	Rules       map[string]*AppGameRecordShardRule
	mu          sync.RWMutex
}

var appGameRecordShardKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// NewRouter 初始化
func NewAppGameRecordRouter() *AppGameRecordRouter {
	return &AppGameRecordRouter{
		TablePrefix: "app_game_record",
		Rules:       make(map[string]*AppGameRecordShardRule),
	}
}

// InitializeRules 根据AppInfo列表初始化分片规则
func (r *AppGameRecordRouter) InitializeRules(appInfos []*models.AppInfo) {
	tempRules := make(map[string]*AppGameRecordShardRule)
	for _, appInfo := range appInfos {
		mode := OnlyMain // 默认模式
		if appInfo.ShardingState == 1 {
			mode = MainAndHistory
		}
		tempRules[appInfo.AppId] = &AppGameRecordShardRule{
			AppID:             appInfo.AppId,
			Mode:              mode,
			shardingStartDate: appInfo.ShardingStartDate,
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Rules = tempRules
}

// UpdateRule 更新或添加指定 appID 的分片规则
func (r *AppGameRecordRouter) UpdateRule(appID string, mode ShardMode, shardingStartDate *time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rule, exists := r.Rules[appID]; exists {
		rule.Mode = mode
	} else {
		r.Rules[appID] = &AppGameRecordShardRule{
			AppID:             appID,
			Mode:              mode,
			shardingStartDate: shardingStartDate,
		}
	}
}

// ExtractAppGroupIDFromAppID returns the record shard key from appID.
// The shard key is the segment before the first underscore, or the full appID
// when no underscore is present.
func ExtractAppGroupIDFromAppID(appID string) (string, error) {
	segment := appID
	if idx := strings.Index(appID, "_"); idx >= 0 {
		segment = appID[:idx]
	}
	if segment == "" {
		return "", fmt.Errorf("app_id %q: missing app_group_id segment", appID)
	}
	if !appGameRecordShardKeyPattern.MatchString(segment) {
		return "", fmt.Errorf("app_id %q: invalid app_group_id segment %q", appID, segment)
	}
	return segment, nil
}

func mustExtractAppGroupIDFromAppID(appID string) string {
	groupID, err := ExtractAppGroupIDFromAppID(appID)
	if err != nil {
		panic(err)
	}
	return groupID
}

// GetMainTable 获取写入表名（永远写主表）
func (r *AppGameRecordRouter) GetMainTable(appID string) string {
	groupID := mustExtractAppGroupIDFromAppID(appID)
	return fmt.Sprintf("%s_%s", r.TablePrefix, groupID)
}

// GetQueryTables 获取查询表列表
// includeMain 控制是否包含主表（当前表），对于查询历史数据可以设置 false
func (r *AppGameRecordRouter) GetQueryTables(appID string, start, end time.Time) []string {
	r.mu.RLock()
	rule, ok := r.Rules[appID]
	r.mu.RUnlock()
	if !ok || rule.Mode == OnlyMain {
		// 默认或小商户：只查询主表
		return []string{r.GetMainTable(appID)}
	}
	// 大商户：返回主表 + 历史表
	return r.getTablesWithHistory(appID, rule, start, end)
}

// getTablesWithHistory 获取大商户查询表列表
func (r *AppGameRecordRouter) getTablesWithHistory(appID string, rule *AppGameRecordShardRule, start, end time.Time) []string {
	var tables []string
	now := time.Now()
	groupID := mustExtractAppGroupIDFromAppID(appID)

	// 当前表（主表）只在 includeMain 为 true 且 end 日期在最近 30 天内时包含
	if end.After(now.AddDate(0, 0, -30)) {
		tables = append(tables, r.GetMainTable(appID))
	}
	// 确定实际的起始时间：如果 start 比 shardingStartDate 更早，则使用 shardingStartDate
	actualStart := start
	if rule.shardingStartDate != nil && rule.shardingStartDate.After(start) {
		actualStart = *rule.shardingStartDate
	}
	// 历史表按月生成，历史时间范围在主表前
	startMonth := time.Date(actualStart.Year(), actualStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)

	for d := startMonth; !d.After(endMonth); d = d.AddDate(0, 1, 0) {
		// 历史表只覆盖主表前的数据（30天前）
		if d.Before(now.AddDate(0, 0, -30)) {
			historyTable := fmt.Sprintf("%s_%s_%s", r.TablePrefix, groupID, d.Format("200601"))
			tables = append(tables, historyTable)
		}
	}

	return tables
}

// BuildUnionSQL 根据查询表列表生成 UNION ALL SQL
// fields: 查询字段列表，例如 "id, user_id, score"
// whereClause: 公共 WHERE 条件，例如 "user_id=123 AND created_at BETWEEN '2025-07-01' AND '2025-09-20'"
func BuildAppGameRecordUnionSQL(tables []string, fields, whereClause string) string {
	if len(tables) == 0 {
		return ""
	}
	var sqlParts []string
	for _, t := range tables {
		part := fmt.Sprintf("SELECT %s FROM %s WHERE %s", fields, t, whereClause)
		sqlParts = append(sqlParts, part)
	}
	return strings.Join(sqlParts, " UNION ALL ")
}
