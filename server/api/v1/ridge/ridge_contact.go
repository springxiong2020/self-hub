package ridge

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/springbear2020/self-hub/server/global"
	"github.com/springbear2020/self-hub/server/model/common/response"
	"github.com/springbear2020/self-hub/server/model/ridge"
	"github.com/springbear2020/self-hub/server/model/ridge/request"
	"github.com/springbear2020/self-hub/server/utils"
	"go.uber.org/zap"
)

type RidgeContactApi struct{}

func (ridgeContactApi *RidgeContactApi) CreateRidgeContact(c *gin.Context) {
	var ridgeContact ridge.RidgeContact
	err := c.ShouldBindJSON(&ridgeContact)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	uid := utils.GetUserID(c)
	err = ridgeContactService.CreateRidgeContact(uid, &ridgeContact)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage("创建成功", c)
}

func (ridgeContactApi *RidgeContactApi) DeleteRidgeContact(c *gin.Context) {
	id := c.Query("id")
	uid := utils.GetUserID(c)
	err := ridgeContactService.DeleteRidgeContact(uid, id)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage("删除成功", c)
}

func (ridgeContactApi *RidgeContactApi) DeleteRidgeContactByIds(c *gin.Context) {
	ids := c.QueryArray("ids[]")
	uid := utils.GetUserID(c)
	err := ridgeContactService.DeleteRidgeContactByIds(uid, ids)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage("批量删除成功", c)
}

func (ridgeContactApi *RidgeContactApi) UpdateRidgeContact(c *gin.Context) {
	var ridgeContact ridge.RidgeContact
	err := c.ShouldBindJSON(&ridgeContact)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	uid := utils.GetUserID(c)
	err = ridgeContactService.UpdateRidgeContact(uid, ridgeContact)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage("更新成功", c)
}

func (ridgeContactApi *RidgeContactApi) FindRidgeContact(c *gin.Context) {
	id := c.Query("id")
	uid := utils.GetUserID(c)
	ridgeContact, err := ridgeContactService.GetRidgeContact(uid, id)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}

	response.OkWithData(ridgeContact, c)
}

func (ridgeContactApi *RidgeContactApi) GetRidgeContactList(c *gin.Context) {
	var pageInfo request.RidgeContactSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	pageInfo.PageInfo.Check()

	uid := utils.GetUserID(c)
	list, total, err := ridgeContactService.GetRidgeContactInfoList(uid, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}

	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

func (ridgeContactApi *RidgeContactApi) GetRidgeContactStat(c *gin.Context) {
	data, err := ridgeContactService.GetRidgeContactStat()
	if err != nil {
		global.GVA_LOG.Error("获取联系人统计失败", zap.Error(err))
		response.FailWithMessage("获取联系人统计失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (ridgeContactApi *RidgeContactApi) GetUpcomingContactBirthdays(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "30"))
	if err != nil || (days != 7 && days != 30 && days != 90) {
		response.FailWithMessage("生日提醒范围仅支持7、30或90天", c)
		return
	}
	data, err := ridgeContactService.GetUpcomingContactBirthdays(days, time.Now())
	if err != nil {
		global.GVA_LOG.Error("获取生日提醒失败", zap.Error(err))
		response.FailWithMessage("获取生日提醒失败", c)
		return
	}
	response.OkWithData(data, c)
}
