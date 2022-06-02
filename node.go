package btree

import (
	"bytes"
	"fmt"
	"sort"
)

func insertAt[T ~*[]E, E any](slc T, idx int, elem E) {
	var null E
	*slc = append(*slc, null)
	if idx+1 <= len(*slc) {
		copy((*slc)[idx+1:], (*slc)[idx:])
	}
	(*slc)[idx] = elem
}

func removeAt[T ~*[]E, E any](slc T, idx int) (ret E) {
	var null E
	ret = (*slc)[idx]
	copy((*slc)[idx:], (*slc)[idx+1:])
	(*slc)[len(*slc)-1] = null
	*slc = (*slc)[:len(*slc)-1]
	return
}

func truncateAt[T ~*[]E, E any](slc T, idx int) {
	var null E
	for i := idx; i < len(*slc); i++ {
		(*slc)[i] = null
	}
	*slc = (*slc)[:idx]
}

type Comparator[E any]func(k1, k2 E) int

type item[K, V any] struct {
	key K
	val V
}

func (i *item[K, V]) String() string {
	return fmt.Sprintf("%v:%v", i.key, i.val)
}

type node[K, V any] struct {
	items    []item[K, V]
	children []*node[K, V]
}

func (n *node[K, V]) String() string {
	var sb = new(bytes.Buffer)
	for i := 0; i < len(n.items); i++ {
		sb.WriteString(fmt.Sprintf("|%v", n.items[i].String()))
	}
	sb.WriteString("|")
	return sb.String()
}

func (n *node[K, V]) leaf() bool {
	return len(n.children) == 0
}

func searchItems[K, V any](cmp Comparator[K], items []item[K, V], key K) (index int, ok bool) {
	index = sort.Search(len(items), func(i int) bool {
		return cmp(items[i].key, key) >= 0
	})
	if index < len(items) && cmp(items[index].key, key) == 0 {
		ok = true
	}
	return
}

func (n *node[K, V]) split(i int) (item item[K, V], right *node[K, V]) {
	item = n.items[i]
	right = new(node[K, V])
	right.items = append(right.items, n.items[i+1:]...)
	truncateAt(&n.items, i)
	if !n.leaf() {
		right.children = append(right.children, n.children[i+1:]...)
		truncateAt(&n.children, i+1)
	}
	return
}

func (n *node[K, V]) trySplitChild(cdx int, maxItems int) (ok bool) {
	child := n.children[cdx]
	if len(child.items) < maxItems {
		return false
	}
	item, right := child.split(maxItems / 2)
	insertAt(&n.items, cdx, item)
	insertAt(&n.children, cdx+1, right)
	return true
}

func (n *node[K, V]) insert(cmp Comparator[K], maxItems int, key K, val V) (old V, ok bool) {
	index, ok := searchItems(cmp, n.items, key)
	if ok {
		old = n.items[index].val
		n.items[index].val = val
		return
	}
	if n.leaf() {
		insertAt(&n.items, index, item[K, V]{key, val})
		return
	}
	if n.trySplitChild(index, maxItems) {
		c := cmp(n.items[index].key, key)
		if c == 0 {
			ok = true
			old = n.items[index].val
			n.items[index].val = val
			return
		} else if c < 0 {
			index++
		}
	}
	return n.children[index].insert(cmp, maxItems, key, val)
}

func (n *node[K, V]) tryMergeChild(cdx int, minItems int) bool {
	if len(n.children[cdx].items) > minItems {
		return false
	}
	if cdx > 0 && len(n.children[cdx-1].items) > minItems {
		prev, curr := n.children[cdx-1], n.children[cdx]
		insertAt(&curr.items, 0, n.items[cdx-1])
		n.items[cdx-1] = removeAt(&prev.items, len(prev.items)-1)
		if !curr.leaf() {
			insertAt(&curr.children, 0, removeAt(&prev.children, len(prev.children)-1))
		}
	} else if cdx+1 < len(n.children) && len(n.children[cdx+1].items) > minItems {
		curr, next := n.children[cdx], n.children[cdx+1]
		insertAt(&curr.items, len(curr.items), n.items[cdx])
		n.items[cdx] = removeAt(&next.items, 0)
		if !curr.leaf() {
			insertAt(&curr.children, len(curr.children), removeAt(&next.children, 0))
		}
	} else {
		if cdx == len(n.children)-1 {
			cdx--
		}
		curr, next := n.children[cdx], removeAt(&n.children, cdx+1)
		curr.items = append(curr.items, removeAt(&n.items, cdx))
		curr.items = append(curr.items, next.items...)
		curr.children = append(curr.children, next.children...)
	}
	return true
}

func (n *node[K, V]) max() item[K, V] {
	for !n.leaf() {
		n = n.children[len(n.children)-1]
	}
	return n.items[len(n.items)-1]
}

func (n *node[K, V]) min() item[K, V] {
	for !n.leaf() {
		n = n.children[0]
	}
	return n.items[0]
}

func (n *node[K, V]) remove(cmp Comparator[K], minItems int, key K) (old V, ok bool) {
	index, ok := searchItems(cmp, n.items, key)
	if n.leaf() {
		if ok {
			old = removeAt(&n.items, index).val
		}
		return
	}
	if n.tryMergeChild(index, minItems) {
		return n.remove(cmp, minItems, key)
	}
	if ok {
		old = n.items[index].val
		max := n.children[index].max()
		n.remove(cmp, minItems, max.key)
		n.items[index] = max
		return
	}
	return n.children[index].remove(cmp, minItems, key)
}

func (n *node[K, V]) get(cmp Comparator[K], key K) (val V, ok bool) {
	for {
		index, ok := searchItems(cmp, n.items, key)
		if ok {
			return n.items[index].val, true
		}
		if n.leaf() {
			return val, false
		}
		n = n.children[index]
	}
}

func (n *node[K, V]) iterate(cmp Comparator[K], from K, inclusive, asc bool, act func(key K, val V) bool) bool {
	index, ok := searchItems(cmp, n.items, from)
	if asc {
		if !ok || !inclusive {
			if !inclusive {
				index++
			}
			if !n.children[index].iterate(cmp, from, inclusive, asc, act) {
				return false
			}
		}
		for i := index; i < len(n.items); i++ {
			if !act(n.items[i].key, n.items[i].val) {
				return false
			}
			if !n.children[i+1].iterate(cmp, from, inclusive, asc, act) {
				return false
			}
		}
	} else {
		if !ok || !inclusive {
			if !n.children[index].iterate(cmp, from, inclusive, asc, act) {
				return false
			}
			index--
		}
		for i := index; i >= 0; i-- {
			if !act(n.items[i].key, n.items[i].val) {
				return false
			}
			if !n.children[i].iterate(cmp, from, inclusive, asc, act) {
				return false
			}
		}
	}
	return true
}
