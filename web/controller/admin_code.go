package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"lottery/common"
	"lottery/config"
	"lottery/models"
	"lottery/services"
	"strings"
)

type AdminCodeController struct {
	Ctx            iris.Context
	ServiceUser    services.UserService
	ServiceGift    services.GiftService
	ServiceCode    services.CodeService
	ServiceResult  services.ResultService
	ServiceUserday services.UserdayService
	ServiceBlackip services.BlackipService
}

func (c *AdminCodeController) Get() mvc.Result {
	giftId := c.Ctx.URLParamIntDefault("gift_id", 0)
	page := c.Ctx.URLParamIntDefault("page", 1)
	size := 100
	pagePrev := ""
	pageNext := ""
	// 数据列表
	var datalist []models.LtCode
	var num int
	var cacheNum int
	if giftId > 0 {
		datalist = c.ServiceCode.Search(giftId)
	} else {
		datalist = c.ServiceCode.GetAll(page, size)
	}
	total := (page - 1) + len(datalist)
	// 数据总数
	if len(datalist) >= size {
		if giftId > 0 {
			total = int(c.ServiceCode.CountByGift(giftId))
		} else {
			total = int(c.ServiceCode.CountAll())
		}
		pageNext = fmt.Sprintf("%d", page+1)
	}
	if page > 1 {
		pagePrev = fmt.Sprintf("%d", page-1)
	}
	return mvc.View{
		Name: "admin/code.html",
		Data: iris.Map{
			"Title":    "管理后台",
			"Channel":  "code",
			"GiftId":   giftId,
			"Datalist": datalist,
			"Total":    total,
			"PagePrev": pagePrev,
			"PageNext": pageNext,
			"CodeNum":  num,
			"CacheNum": cacheNum,
		},
		Layout: "admin/layout.html",
	}
}

func (c *AdminCodeController) PostImport() {
	giftId := c.Ctx.URLParamIntDefault("gift_id", 0)
	fmt.Println("PostImport giftId=", giftId)
	if giftId < 1 {
		c.Ctx.Text("没有指定奖品ID，无法进行导入，<a href='' onclick='history.go(-1);return false;'>返回</a>")
		return
	}
	gift := c.ServiceGift.Get(giftId, false)
	if gift == nil || gift.Gtype != config.GtypeCodeDiff {
		c.Ctx.HTML("没有指定的优惠券类型的奖品，无法进行导入，<a href='' onclick='history.go(-1);return false;'>返回</a>")
		return
	}
	codes := c.Ctx.PostValue("codes")
	now := common.NowUnix()
	list := strings.Split(codes, "\n")
	sucNum := 0
	errNum := 0
	for _, code := range list {
		code := strings.TrimSpace(code)
		if code != "" {
			data := &models.LtCode{
				GiftId:     giftId,
				Code:       code,
				SysCreated: now,
			}
			err := c.ServiceCode.Create(data)
			if err != nil {
				errNum++
			} else {
				// 成功导入数据库，还需要导入到缓存中
				// TODO
			}
		}
	}
	c.Ctx.HTML(fmt.Sprintf("成功导入 %d 条，导入失败 %d 条，<a href='/admin/code?gift_id=%d'>返回</a>", sucNum, errNum, giftId))
}

func (c *AdminCodeController) GetDelete() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	if err == nil {
		c.ServiceCode.Delete(id)
	}
	refer := c.Ctx.GetHeader("Referer")
	if refer == "" {
		refer = "/admin/code"
	}
	return mvc.Response{
		Path: refer,
	}
}

func (c *AdminCodeController) GetReset() mvc.Result {
	id, err := c.Ctx.URLParamInt("id")
	if err == nil {
		c.ServiceCode.Update(&models.LtCode{Id: id, SysStatus: 0}, []string{"sys_status"})
	}
	refer := c.Ctx.GetHeader("Referer")
	if refer == "" {
		refer = "/admin/code"
	}
	return mvc.Response{
		Path: refer,
	}
}
