package request

import (
	"github.com/springbear2020/self-hub/server/model/common/request"
)

type RidgeContactSearch struct {
	RemarkName      *string `json:"remarkName" form:"remarkName"`
	Gender          *int    `json:"gender" form:"gender"`
	CategoryId      *int    `json:"categoryId" form:"categoryId"`
	TagId           *int    `json:"tagId" form:"tagId"`
	Status          *int    `json:"status" form:"status"`
	IsExempt        *int    `json:"isExempt" form:"isExempt"`
	SuggestedAction *int    `json:"suggestedAction" form:"suggestedAction"`
	request.PageInfo
}
