package models

import (
	"math"
	"sync"

	"github.com/jinzhu/gorm"
)

type Model struct {
	gorm.Model
	mtx sync.Mutex
}

func (m *Model) Lock() {
	m.mtx.Lock()
}

func (m *Model) Unlock() {
	m.mtx.Unlock()
}

type PagingModel struct {
	Records     interface{}
	Page        *int
	Limit       *int
	TotalRecord *int
	LastPage    *int
}

func MakePagingModel(records interface{}, page, limit *int, totalRecord *int) *PagingModel {
	result := PagingModel{}
	result.Records = records
	result.Page = page
	result.Limit = limit
	result.TotalRecord = totalRecord
	if totalRecord != nil && limit != nil {
		r := int(math.Ceil(float64(*totalRecord) / float64(*limit)))
		result.LastPage = &r
	}
	return &result
}
