package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"lottery/common"
	"lottery/models"
	"lottery/services"
)

type AdminUserController struct {
	Ctx            iris.Context
	ServiceUser    services.UserService
	ServiceGift    services.GiftService
	ServiceCode    services.CodeService
	ServiceResult  services.ResultService
	ServiceUserday services.UserdayService
	ServiceBlackip services.BlackipService
}

// GET /admin/user/
func (c *AdminUserController) Get() mvc.Result {
	page := c.Ctx.URLParamIntDefault("page", 1)
	size := 100
	pagePrev := ""
	pageNext := ""
	// 数据列表
	datalist := c.ServiceUser.GetAll(page, size)
	total := (page - 1) + len(datalist)
	// 数据总数
	if len(datalist) >= size {
		total = c.ServiceUser.CountAll()
		pageNext = fmt.Sprintf("%d", page+1)
	}
	if page > 1 {
		pagePrev = fmt.Sprintf("%d", page-1)
	}
	return mvc.View{
		Name: "admin/user.html",
		Data: iris.Map{
			"Title":    "管理后台",
			"Channel":  "user",
			"Datalist": datalist,
			"Total":    total,
			"Now":      common.NowUnix(),
			"PagePrev": pagePrev,
			"PageNext": pageNext,
		},
		Layout: "admin/layout.html",
	}
}

// GET /admin/user/black?id=1&time=0
func (c *AdminUserController) GetBlack() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	t := c.Ctx.URLParamIntDefault("time", 0)
	if err == nil {
		if t > 0 {
			t = t*86400 + common.NowUnix()
		}
		c.ServiceUser.Update(&models.LtUser{Id: id, Blacktime: t, SysUpdated: common.NowUnix()},
			[]string{"blacktime"})
	}
	return mvc.Response{
		Path: "/admin/user",
	}
}
