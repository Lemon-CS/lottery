package utils

import (
	"log"
	"lottery/services"
)

// 发奖，指定的奖品是否还可以发出来奖品
func PrizeGift(id, leftNum int) bool {
	ok := false
	// 更新数据库，减少奖品的库存
	giftService := services.NewGiftService()
	rows, err := giftService.DecrLeftNum(id, 1)
	if rows < 1 || err != nil {
		log.Println("prizedata.PrizeGift giftService.DecrLeftNum error=", err, ", rows=", rows)
		// 数据更新失败，不能发奖
		return false
	}
	return ok
}

// 优惠券类的发放
func PrizeCodeDiff(id int, codeService services.CodeService) string {
	lockUid := 0 - id - 1000000000
	LockLucky(lockUid)
	defer UnLockLucky(lockUid)

	return ""
}
