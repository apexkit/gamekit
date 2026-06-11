package sharding

import (
	"testing"
	"time"
)

func TestExtractAppGroupIDFromAppID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		appID   string
		want    string
		wantErr bool
	}{
		{name: "prefix before underscore", appID: "1001_demo", want: "1001"},
		{name: "hyphen group id", appID: "zm-test_demo", want: "zm-test"},
		{name: "no underscore", appID: "1001", want: "1001"},
		{name: "empty prefix", appID: "_demo", wantErr: true},
		{name: "invalid characters", appID: "zm.test_demo", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ExtractAppGroupIDFromAppID(tt.appID)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestAppGameRecordRouter_UsesGroupIDTables(t *testing.T) {
	t.Parallel()

	router := NewAppGameRecordRouter()
	shardStart := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	router.UpdateRule("zm-test_demo", MainAndHistory, &shardStart)

	mainTable := router.GetMainTable("zm-test_demo")
	if mainTable != "app_game_record_zm-test" {
		t.Fatalf("expected main table app_game_record_zm-test, got %s", mainTable)
	}

	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Now().AddDate(0, 0, -45)
	tables := router.GetQueryTables("zm-test_demo", start, end)
	if len(tables) == 0 {
		t.Fatalf("expected history tables, got none")
	}
	for _, table := range tables {
		if len(table) >= len("app_game_record_") && table[:len("app_game_record_")] != "app_game_record_" {
			t.Fatalf("unexpected table prefix: %s", table)
		}
		if table == "app_game_record_zm-test_demo" {
			t.Fatalf("expected group-id shard table, got app-id table %s", table)
		}
	}
}
