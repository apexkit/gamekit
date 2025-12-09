package utils

import (
	"testing"
)

//测试案例
func Test_trie(t *testing.T) {
	ti := NewTrie()

	ti.Inster("脏话")

	ti.Inster("mb")

	s := ti.Replace("不能说脏话也不能说mb")
	t.Log(ti.HasDirty("不能说脏话也不能说mb"))
	t.Log(s)
	s = ti.Replace("不能说脏也不能说m")
	t.Log(ti.HasDirty("不能说脏也不能说m"))
	t.Log(s)
}
