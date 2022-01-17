package datasource

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"lottery/config"
	"sync"
	"xorm.io/xorm"
)

var dbLock sync.Mutex
var masterInstance *xorm.Engine

// 得到唯一的主库实例
func InstanceDbMaster() *xorm.Engine {
	if masterInstance != nil {
		return masterInstance
	}
	dbLock.Lock()
	defer dbLock.Unlock()
	if masterInstance != nil {
		return masterInstance
	}
	return NewDbMaster()
}

func NewDbMaster() *xorm.Engine {
	sourcename := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8",
		config.DbMaster.User,
		config.DbMaster.Pwd,
		config.DbMaster.Host,
		config.DbMaster.Port,
		config.DbMaster.Database)
	instance, err := xorm.NewEngine(config.DriverName, sourcename)

	if err != nil {
		log.Fatal("dbhelper.InstanceDbMaster NewEngine error ", err)
		return nil
	}
	instance.ShowSQL(true)
	//instance.ShowSQL(false)
	masterInstance = instance
	return instance
}
