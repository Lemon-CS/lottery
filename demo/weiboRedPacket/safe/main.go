/**
 * 微博抢红包
 * 两个步骤
 * 1 抢红包，设置红包总金额，红包个数，返回抢红包的地址
 * curl "http://localhost:8080/set?uid=1&money=100&num=100"
 * 2 抢红包，先到先得，随机得到红包金额
 * curl "http://localhost:8080/get?id=3171875376&uid=1"
 * 注意：
 * 线程不安全1，红包列表 packageList map 的并发读写会产生异常
 * 测试方法： wrk -t10 -c10 -d5  "http://localhost:8080/set?uid=1&money=100&num=10"
 * fatal error: concurrent map writes
 * 线程不安全2，红包里面的金额切片 packageList map[uint32][]uint 并发读写不安全，虽然不会报错
 */

package main

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

// 文件日志
var logger *log.Logger

// 当前有效红包列表，int64是红包唯一ID，[]uint是红包里面随机分到的金额（单位分）
//var packageList map[uint32][]uint = make(map[uint32][]uint)

var packageList *sync.Map = new(sync.Map)

func main() {
	app := newApp()
	app.Run(iris.Addr(":8080"))
}

// 初始化Application
func newApp() *iris.Application {
	app := iris.New()
	mvc.New(app.Party("/")).Handle(&lotteryController{})
	return app
}

// 初始化日志
func initLog() {
	f, _ := os.Create("/Users/fangpeng/logs/lottery_demo.log")
	logger = log.New(f, "", log.Ldate|log.Lmicroseconds)
}

// 抽奖的控制器
type lotteryController struct {
	Ctx iris.Context
}

// 返回全部红包地址
// GET http://localhost:8080/
func (c lotteryController) Get() map[uint32][2]int {
	res := make(map[uint32][2]int)

	/*for id, list := range packageList {
		var money int
		for _, v := range list {
			money += int(v)
		}
		res[id] = [2]int{len(list), money}
	}*/

	packageList.Range(func(key, value interface{}) bool {
		id := key.(uint32)
		list := value.([]uint)
		var money int
		for _, v := range list {
			money += int(v)
		}
		res[id] = [2]int{len(list), money}
		return true
	})
	return res
}

// 发红包
// GET http://localhost:8080/set?uid=1&money=100&num=100
func (c lotteryController) GetSet() string {
	uid, errUid := c.Ctx.URLParamInt("uid")
	money, errMoney := c.Ctx.URLParamFloat64("money")
	num, errNum := c.Ctx.URLParamInt("num")
	if errUid != nil || errMoney != nil || errNum != nil {
		return fmt.Sprintf("参数格式异常，errUid=%s, errMoney=%s, errNum=%s\n", errUid, errMoney, errNum)
	}
	moneyTotal := int(money * 100)
	if uid < 1 || moneyTotal < num || num < 1 {
		return fmt.Sprintf("参数数值异常，uid=%d, money=%d, num=%d\n", uid, money, num)
	}

	// 金额分配算法
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rMax := 0.55 // 随机分配最大比例
	if num > 1000 {
		rMax = 0.01
	} else if num >= 100 {
		rMax = 0.1
	} else if num >= 10 {
		rMax = 0.3
	}
	list := make([]uint, num)
	leftMoney := moneyTotal
	leftNum := num

	// 大循环开始，只要还有没分配的名额，继续分配
	for leftNum > 0 {
		if leftNum == 1 {
			// 最后一个名额，把剩余的全部给它
			list[num-1] = uint(leftMoney)
			break
		}

		// 剩下的最多只能分配到1分钱时，不用再随机
		if leftMoney == leftNum {
			for i := num - leftNum; i < num; i++ {
				list[i] = 1
			}
			break
		}

		// 每次对剩余金额的1%-55%随机，最小1，最大就是剩余金额55%（需要给剩余的名额留下1分钱的生存空间）
		rMoney := int(float64(leftMoney-leftNum) * rMax)
		m := r.Intn(rMoney)
		if m < 1 {
			m = 1
		}

		list[num-leftNum] = uint(m)
		leftMoney -= m
		leftNum--
	}
	// 最后再来一个红包的唯一ID
	id := r.Uint32()
	//packageList[id] = list
	packageList.Store(id, list)
	// 返回抢红包的URL
	return fmt.Sprintf("/get?id=%d&uid=%d&num=%d\n", id, uid, num)
}

// 抢红包
// GET http://localhost:8080/get?id=1&uid=1
func (c lotteryController) GetGet() string {
	uid, errUid := c.Ctx.URLParamInt("uid")
	id, errId := c.Ctx.URLParamInt("id")
	if errUid != nil || errId != nil {
		return fmt.Sprintf("参数格式异常，errUid=%s, errId=%s\n", errUid, errId)
	}
	if uid < 1 || id < 1 {
		return fmt.Sprintf("参数数值异常，uid=%d, id=%d\n", uid, id)
	}

	//list, ok := packageList[uint32(id)]
	value, ok := packageList.Load(uint32(id))
	list := value.([]uint)
	if !ok || len(list) < 1 {
		return fmt.Sprintf("红包不存在,id=%d\n", id)
	}

	// 分配的随机数
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 从红包金额中随机得到一个
	i := r.Intn(len(list))
	money := list[i]
	// 更新红包列表中的信息

	if len(list) > 1 {
		if i == len(list)-1 {
			//packageList[uint32(id)] = list[:i]
			packageList.Store(uint32(id), list[:i])
		} else if i == 0 {
			//packageList[uint32(id)] = list[1:]
			packageList.Store(uint32(id), list[1:])
		} else {
			//packageList[uint32(id)] = append(list[:i], list[i+1:]...)
			packageList.Store(uint32(id), append(list[:i], list[i+1:]...))
		}
	} else {
		//delete(packageList, uint32(id))
		packageList.Delete(uint32(id))
	}
	return fmt.Sprintf("恭喜你抢到一个红包，金额为:%d\n", money)
}
