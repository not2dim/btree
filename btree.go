package btree

import (
	"bytes"
	"fmt"
)

type BTree[K, V any] struct {
	cmp    Comparator[K]
	degree int
	root   *node[K, V]
}

func NewBTree[K, V any](cmp Comparator[K], degree int) *BTree[K, V] {
	return &BTree[K, V]{
		cmp:    cmp,
		degree: degree,
		root:   new(node[K, V]),
	}
}

func (bt *BTree[K, V]) String() string {
	var sb = new(bytes.Buffer)
	var queue = []*node[K, V]{bt.root}
	for len(queue) > 0 {
		local := make([]*node[K, V], 0, bt.degree*len(queue))
		for i := 0; i < len(queue); i++ {
			sb.WriteString(fmt.Sprintf("%v  ", queue[i].String()))
			local = append(local, queue[i].children...)
		}
		sb.WriteString("\n")
		queue = local
	}
	return sb.String()
}

func minItems(degree int) int {
	return degree - 1
}

func maxItems(degree int) int {
	return degree*2 - 1
}

func (bt *BTree[K, V]) Get(key K) (v V, ok bool) {
	return bt.root.get(bt.cmp, key)
}

func (bt *BTree[K, V]) Put(key K, val V) (old V, ok bool) {
	maxItems := maxItems(bt.degree)
	old, ok = bt.root.insert(bt.cmp, maxItems, key, val)
	if len(bt.root.items) > maxItems {
		oldRt := bt.root
		bt.root = new(node[K, V])
		bt.root.children = append(bt.root.children, oldRt)
		bt.root.trySplitChild(0, maxItems)
	}
	return
}

func (bt *BTree[K, V]) Del(key K) (old V, ok bool) {
	minItems := minItems(bt.degree)
	old, ok = bt.root.remove(bt.cmp, minItems, key)
	if len(bt.root.items) == 0 && len(bt.root.children) == 1 {
		bt.root = bt.root.children[0]
	}
	return
}

func (bt *BTree[K, V]) Max() (key K, val V, ok bool) {
	if bt.root.leaf() && len(bt.root.items) == 0 {
		return key, val, false
	}
	max := bt.root.max()
	key, val = max.key, max.val
	ok = true
	return
}

func (bt *BTree[K, V]) Min() (key K, val V, ok bool) {
	if bt.root.leaf() && len(bt.root.items) == 0 {
		return key, val, false
	}
	min := bt.root.min()
	key, val = min.key, min.val
	ok = true
	return
}
