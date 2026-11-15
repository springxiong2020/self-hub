package ridge

import (
	"github.com/gin-gonic/gin"
	"github.com/springbear2020/self-hub/server/middleware"
)

type RidgeContactRouter struct{}

func (s *RidgeContactRouter) InitRidgeContactRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	ridgeContactRouter := Router.Group("ridgeContact").Use(middleware.OperationRecord())
	ridgeContactRouterWithoutRecord := Router.Group("ridgeContact")
	ridgeContactRouterWithoutAuth := PublicRouter.Group("ridgeContact")
	{
		ridgeContactRouter.POST("createRidgeContact", ridgeContactApi.CreateRidgeContact)             // 新建社交关系
		ridgeContactRouter.DELETE("deleteRidgeContact", ridgeContactApi.DeleteRidgeContact)           // 删除社交关系
		ridgeContactRouter.DELETE("deleteRidgeContactByIds", ridgeContactApi.DeleteRidgeContactByIds) // 批量删除社交关系
		ridgeContactRouter.PUT("updateRidgeContact", ridgeContactApi.UpdateRidgeContact)              // 更新社交关系
	}
	{
		ridgeContactRouterWithoutRecord.GET("getUpcomingContactBirthdays", ridgeContactApi.GetUpcomingContactBirthdays) // 邻近生日查询
		ridgeContactRouterWithoutRecord.GET("getRidgeContactStat", ridgeContactApi.GetRidgeContactStat)                 // 获取联系人统计
		ridgeContactRouterWithoutRecord.GET("findRidgeContact", ridgeContactApi.FindRidgeContact)                       // 根据ID获取社交关系
		ridgeContactRouterWithoutRecord.GET("getRidgeContactList", ridgeContactApi.GetRidgeContactList)                 // 获取社交关系列表
	}
	{
		_ = ridgeContactRouterWithoutAuth
	}
}
