package pagination

import (
	"math"
)

type Pagination struct {
	Limit     int
	PageCount int
	Total     int
}

func (p *Pagination) Init(limit int, total int) {
	p.Limit = limit
	p.Total = total

	pageCount := math.Ceil(float64(total) / float64(limit))
	p.PageCount = int(pageCount)
}

func (p *Pagination) GetPage(page int) (int, int) {
	if page > p.PageCount {
		page = p.PageCount
	}
	if page < 0 {
		page = 1
	}

	page -= 1

	start := page * p.Limit
	end := start + p.Limit

	if start > p.Total {
		return 0, p.Total - 1
	}
	if end > p.Total {
		end = p.Total
	}
	return start, end
}
