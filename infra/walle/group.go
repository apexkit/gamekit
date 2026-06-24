package walle

import (
	"fmt"
	"regexp"
	"strings"
)

const maxGroupNameLen = 64

var groupNamePattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// ParseGroupName validates a single group name.
func ParseGroupName(raw string) (string, error) {
	if strings.Contains(raw, ",") {
		return "", fmt.Errorf("only one group is allowed")
	}
	names, err := ParseGroupNames(raw)
	if err != nil {
		return "", err
	}
	if len(names) != 1 {
		return "", fmt.Errorf("group name is required")
	}
	return names[0], nil
}

// ParseGroupNames splits and validates group query values.
// Accepts comma-separated names in one string or multiple strings.
func ParseGroupNames(parts ...string) ([]string, error) {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		for _, raw := range strings.Split(part, ",") {
			name := strings.TrimSpace(raw)
			if name == "" {
				continue
			}
			name = strings.ToLower(name)
			if len(name) > maxGroupNameLen {
				return nil, fmt.Errorf("group name too long: %q", name)
			}
			if !groupNamePattern.MatchString(name) {
				return nil, fmt.Errorf("invalid group name: %q", name)
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}
	return out, nil
}

// SelectGameGroup picks the group matching want; if want is empty, returns the first item.
func SelectGameGroup(groups []GameGroup, want string) (*GameGroup, error) {
	if len(groups) == 0 {
		if want == "" {
			return nil, fmt.Errorf("walle returned no game groups for %q", want)
		}
		return nil, fmt.Errorf("game group %q not found in walle response (0 groups returned)", want)
	}
	want = strings.ToLower(strings.TrimSpace(want))
	if want == "" {
		g := groups[0]
		return &g, nil
	}
	for i := range groups {
		if strings.ToLower(groups[i].GroupName) == want {
			g := groups[i]
			return &g, nil
		}
	}
	return nil, fmt.Errorf("game group %q not found in walle response (%d groups returned)", want, len(groups))
}
