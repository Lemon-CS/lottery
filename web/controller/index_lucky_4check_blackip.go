package controller

import (
	"lottery/models"
	"time"
)

// 验证当前用户的IP是否存在黑名单限制
func (c *IndexController) checkBlackip(ip string) (bool, *models.LtBlackip) {
	info := c.ServiceBlackip.GetByIp(ip)
	if info == nil || info.Ip == "" {
		return true, nil
	}
	if info.Blacktime > int(time.Now().Unix()) {
		// IP黑名单存在，而且没有过去
		return false, info
	}
	return true, info
}
