package ridge

import (
	"time"
)

type RidgeContact struct {
	Id                  *int       `json:"id" form:"id" gorm:"primarykey;comment:ID;column:id;size:10;"`                                                //ID
	UserId              uint       `json:"userId" form:"userId" gorm:"comment:所属后台用户ID;column:user_id;size:10;"`                                        //用户ID
	RemarkName          *string    `json:"remarkName" form:"remarkName" gorm:"comment:微信备注名，未设置时可填写昵称;column:remark_name;size:128;" binding:"required"` //姓名
	Gender              *int       `json:"gender" form:"gender" gorm:"comment:性别：0-未知；1-男；2-女;column:gender;" binding:"required"`                       //性别
	Birthday            *string    `json:"birthday" form:"birthday" gorm:"column:birthday;size:10;comment:生日（所选历法YYYY-MM-DD），未知为空"`
	BirthdayIsLunar     bool       `json:"birthdayIsLunar" form:"birthdayIsLunar" gorm:"column:birthday_is_lunar;not null;default:false;comment:是否农历生日"`
	BirthdayIsLeapMonth bool       `json:"birthdayIsLeapMonth" form:"birthdayIsLeapMonth" gorm:"column:birthday_is_leap_month;not null;default:false;comment:是否农历闰月生日"`
	CategoryId          *int       `json:"categoryId" form:"categoryId" gorm:"comment:分类字典ID，0-未分类;column:category_id;size:10;" binding:"required"`            //分类
	TagId               *int       `json:"tagId" form:"tagId" gorm:"comment:标签字典ID，0-无标签;column:tag_id;size:10;" binding:"required"`                           //标签
	Notes               *string    `json:"notes" form:"notes" gorm:"comment:备注，如认识场景、联系用途、保留原因;column:notes;size:256;" binding:"required"`                     //备注
	SortOrder           *int       `json:"sortOrder" form:"sortOrder" gorm:"comment:排序值，越大越靠前，同值按ID倒序;column:sort_order;size:10;" binding:"required"`          //排序值
	FriendedAt          *time.Time `json:"friendedAt" form:"friendedAt" gorm:"comment:成为微信好友的时间，未知则为空;column:friended_at;" binding:"required"`                 //好友添加时间
	LastChatAt          *time.Time `json:"lastChatAt" form:"lastChatAt" gorm:"comment:最近一次已知聊天时间，空表示无记录，不代表从未聊天;column:last_chat_at;"`                         //上次聊天时间
	LastInteractionAt   *time.Time `json:"lastInteractionAt" form:"lastInteractionAt" gorm:"comment:最近一次已知朋友圈互动时间，空表示无记录，不代表从未互动;column:last_interaction_at;"` //上次互动时间
	ChatOnlySince       *time.Time `json:"chatOnlySince" form:"chatOnlySince" gorm:"comment:本轮仅聊天起始时间，原始时间未知时填写人工确认时间;column:chat_only_since;"`                //置仅聊天时间
	Status              *int       `json:"status" form:"status" gorm:"comment:微信实际状态：0-正常；1-仅聊天；2-已删除;column:status;" binding:"required"`                      //好友状态
	IsExempt            *int       `json:"isExempt" form:"isExempt" gorm:"comment:是否豁免自动分析：0-否；1-是;column:is_exempt;" binding:"required"`                      //是否豁免
	SuggestedAction     uint       `json:"suggestedAction" form:"suggestedAction" gorm:"comment:操作建议：0-无；1-设为仅聊天；2-删除好友；3-补充资料;column:suggested_action;"`      //操作建议
	CreatedAt           *time.Time `json:"createdAt" form:"createdAt" gorm:"comment:后台记录创建时间;column:created_at;"`                                              //创建时间
	UpdatedAt           *time.Time `json:"updatedAt" form:"updatedAt" gorm:"comment:后台记录更新时间;column:updated_at;"`                                              //更新时间
}

func (RidgeContact) TableName() string {
	return "ridge_contacts"
}
