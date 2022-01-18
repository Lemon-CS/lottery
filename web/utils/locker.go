package utils

import (
	"fmt"
	"lottery/datasource"
)

func getLuckyLockKey(uid int) string {
	return fmt.Sprintf("lucky_lock_%d", uid)
}

func LockLucky(uid int) bool {
	key := getLuckyLockKey(uid)
	cache := datasource.InstanceCache()
	rs, _ := cache.Do("SET", key, 1, "EX", 3, "NX")
	if rs == "OK" {
		return true
	} else {
		return false
	}
}

func UnLockLucky(uid int) bool {
	key := getLuckyLockKey(uid)
	cache := datasource.InstanceCache()
	rs, _ := cache.Do("DEL", key)
	if rs == "OK" {
		return true
	} else {
		return false
	}
}
