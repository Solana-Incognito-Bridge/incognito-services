package serializers

import (
	"math"
)

type PagingResp struct {
	Records     interface{}
	Page        *int
	Limit       *int
	TotalRecord *int
	TotalPage   *int
}

func MakePagingResp(records interface{}, page, limit *int, totalRecord *int) *PagingResp {
	result := PagingResp{}
	result.Records = records
	result.Page = page
	result.Limit = limit
	result.TotalRecord = totalRecord
	if totalRecord != nil && limit != nil {
		r := int(math.Ceil(float64(*totalRecord) / float64(*limit)))
		result.TotalPage = &r
	}
	return &result
}
