package controller

import "lottery/common"

// localhost:8080/lucky
func (c *IndexController) GetLucky() map[string]interface{} {
	rs := make(map[string]interface{})
	rs["code"] = 0
	rs["msg"] = ""
	// 1 验证登录用户
	loginuser := common.GetLoginUser(c.Ctx.Request())
	if loginuser == nil || loginuser.Uid < 1 {
		rs["code"] = 101
		rs["msg"] = "请先登录，再来抽奖"
		return rs
	}
	ip := common.ClientIP(c.Ctx.Request())
	api := &LuckyApi{}
	code, msg, gift := api.luckyDo(loginuser.Uid, loginuser.Username, ip)
	rs["code"] = code
	rs["msg"] = msg
	rs["gift"] = gift
	return rs
}
