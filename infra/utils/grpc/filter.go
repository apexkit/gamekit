package grpc

import (
	"context"

	"github.com/go-kratos/kratos/v2/selector"
)

func MetadataFilter(key, value string) selector.NodeFilter {
	return func(ctx context.Context, nodes []selector.Node) []selector.Node {
		out := make([]selector.Node, 0, len(nodes))
		for _, n := range nodes {
			md := n.Metadata()
			if md == nil {
				continue
			}
			if v, ok := md[key]; ok && v == value {
				out = append(out, n)
			}
		}
		return out
	}
}
