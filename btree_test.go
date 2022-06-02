package btree

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func numbers(count int) []int {
	nums := make([]int, 0, count)
	for i := 0; i < count; i++ {
		nums = append(nums, i)
	}
	return nums
}

func randPerm(nums []int) {
	for i := len(nums) - 1; i > 0; i-- {
		loc := rand.Int() % (i + 1)
		nums[i], nums[loc] = nums[loc], nums[i]
	}
}

func TestBTree(t *testing.T) {
	count := 1000000
	expect := numbers(count)
	randPerm(expect)
	tree := NewBTree[int, int](func(a, b int) int { return a - b }, 100)
	for _, num := range expect {
		tree.Put(num, 0)
	}
	var actual = make([]int, 0, count)
	for {
		key, _, ok := tree.Min()
		if !ok {
			break
		}
		tree.Del(key)
		actual = append(actual, key)
	}
	sort.Ints(expect)
	if !reflect.DeepEqual(expect, actual) {
		t.FailNow()
	}
}
