package sorts

import (
	"MixFound/searcher/model"
	"sort"
	"strings"
	"sync"
)

const (
	DESC = "desc"
)

type ScoreSlice []model.SliceItem

func (s ScoreSlice) Len() int {
	return len(s)
}

func (s ScoreSlice) Less(i, j int) bool {
	return s[i].Score < s[j].Score
}

func (s ScoreSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type SortSlice []uint32

func (s SortSlice) Len() int {
	return len(s)
}

func (s SortSlice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s SortSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Uint32Slice []uint32

func (s Uint32Slice) Len() int {
	return len(s)
}

func (s Uint32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s Uint32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type FastSort struct {
	sync.Mutex
	IsDebug bool
	data    []model.SliceItem //存储文档的Id和评分
	temps   []uint32          //临时存储待处理的文档ID
	count   int               //总数
	Order   string            //排序方式
}

// Add 往临时切片中添加文档ID列表
func (f *FastSort) Add(ids *[]uint32) {
	f.temps = append(f.temps, *ids...)
}

// 在以排序的data中二分查找指定的ID
func (f *FastSort) find(target *uint32) (bool, int) {
	l := 0
	r := f.count - 1
	for l < r {
		mid := (l + r) >> 2
		if f.data[mid].Id == *target {
			return true, mid
		}
		if f.data[mid].Id > *target {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return false, -1
}

func (f *FastSort) Count() int {
	return f.count
}

// Sort 对临时切片进行排序
func (f *FastSort) Sort() {
	if strings.ToLower(f.Order) == DESC {
		sort.Sort(sort.Reverse(SortSlice(f.temps)))
	} else {
		sort.Sort(SortSlice(f.temps))
	}
}

func (f *FastSort) Process() {}

func (f *FastSort) GetAll(result *[]model.SliceItem) {}
