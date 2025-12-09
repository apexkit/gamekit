package app

import "testing"

func TestAppGroupIDFromAppID(t *testing.T) {
	tests := []struct {
		appId   string
		want    string
		wantErr bool
	}{
		{"group001", "group001", false},
		{"group001_sub", "group001", false},
		{"", "", true},
		{"_sub", "", true},
	}
	for _, tt := range tests {
		got, err := AppGroupIDFromAppID(tt.appId)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("AppGroupIDFromAppID(%q) expected error", tt.appId)
			}
			continue
		}
		if err != nil {
			t.Fatalf("AppGroupIDFromAppID(%q) unexpected error: %v", tt.appId, err)
		}
		if got != tt.want {
			t.Fatalf("AppGroupIDFromAppID(%q) = %q, want %q", tt.appId, got, tt.want)
		}
	}
}
