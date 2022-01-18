/**
 * 抽奖系统数据处理（包括数据库，也包括缓存等其他形式数据）
 */
package services

import (
	"encoding/json"
	"log"
	"lottery/common"
	"lottery/dao"
	"lottery/datasource"
	"lottery/models"
	"strconv"
	"strings"
)

type GiftService interface {
	GetAll(useCache bool) []models.LtGift
	CountAll() int64
	//Search(country string) []models.LtGift
	Get(id int, useCache bool) *models.LtGift
	Delete(id int) error
	Update(data *models.LtGift, columns []string) error
	Create(data *models.LtGift) error
	GetAllUse(useCache bool) []models.ObjGiftPrize
	IncrLeftNum(id, num int) (int64, error)
	DecrLeftNum(id, num int) (int64, error)
}

type giftService struct {
	dao *dao.GiftDao
}

func NewGiftService() GiftService {
	return &giftService{
		dao: dao.NewGiftDao(datasource.InstanceDbMaster()),
	}
}

func (s *giftService) GetAll(useCache bool) []models.LtGift {
	if !useCache {
		// 直接读取数据库的方式
		return s.dao.GetAll()
	}

	// 先读取缓存
	gifts := s.getAllByCache()
	if len(gifts) < 1 {
		// 再读取数据库
		gifts = s.dao.GetAll()
		s.setAllByCache(gifts)
	}
	return gifts
}

func (s *giftService) CountAll() int64 {
	// 直接读取数据库的方式
	//return s.dao.CountAll()

	// 缓存优化之后的读取方式
	gifts := s.GetAll(true)
	return int64(len(gifts))
}

//func (s *giftService) Search(country string) []models.LtGift {
//	return s.dao.Search(country)
//}

func (s *giftService) Get(id int, useCache bool) *models.LtGift {
	if !useCache {
		// 直接读取数据库的方式
		return s.dao.Get(id)
	}

	// 缓存优化之后的读取方式
	gifts := s.GetAll(true)
	for _, gift := range gifts {
		if gift.Id == id {
			return &gift
		}
	}
	return nil
}

func (s *giftService) Delete(id int) error {
	// 先更新缓存
	data := &models.LtGift{Id: id}
	s.updateByCache(data, nil)
	// 再更新数据库
	return s.dao.Delete(id)
}

func (s *giftService) Update(data *models.LtGift, columns []string) error {
	// 先更新缓存
	s.updateByCache(data, columns)
	// 再更新数据库
	return s.dao.Update(data, columns)
}

func (s *giftService) Create(data *models.LtGift) error {
	// 先更新缓存
	s.updateByCache(data, nil)
	// 再更新数据库
	return s.dao.Create(data)
}

// 获取到当前可以获取的奖品列表
// 有奖品限定，状态正常，时间期间内
// gtype倒序， displayorder正序
func (s *giftService) GetAllUse(useCache bool) []models.ObjGiftPrize {
	list := make([]models.LtGift, 0)
	if !useCache {
		// 直接读取数据库的方式
		list = s.dao.GetAllUse()
	} else {
		// 缓存优化之后的读取方式
		now := common.NowUnix()
		gifts := s.GetAll(true)
		for _, gift := range gifts {
			if gift.Id > 0 && gift.SysStatus == 0 &&
				gift.PrizeNum >= 0 &&
				gift.TimeBegin <= now &&
				gift.TimeEnd >= now {
				list = append(list, gift)
			}
		}
	}

	if list != nil {
		gifts := make([]models.ObjGiftPrize, 0)
		for _, gift := range list {
			codes := strings.Split(gift.PrizeCode, "-")
			if len(codes) == 2 {
				// 设置了获奖编码范围 a-b 才可以进行抽奖
				codeA := codes[0]
				codeB := codes[1]
				a, e1 := strconv.Atoi(codeA)
				b, e2 := strconv.Atoi(codeB)
				if e1 == nil && e2 == nil && b >= a && a >= 0 && b < 10000 {
					data := models.ObjGiftPrize{
						Id:           gift.Id,
						Title:        gift.Title,
						PrizeNum:     gift.PrizeNum,
						LeftNum:      gift.LeftNum,
						PrizeCodeA:   a,
						PrizeCodeB:   b,
						Img:          gift.Img,
						Displayorder: gift.Displayorder,
						Gtype:        gift.Gtype,
						Gdata:        gift.Gdata,
					}
					gifts = append(gifts, data)
				}
			}
		}
		return gifts
	} else {
		return []models.ObjGiftPrize{}
	}
}

func (s *giftService) IncrLeftNum(id, num int) (int64, error) {
	return s.dao.IncrLeftNum(id, num)
}

func (s *giftService) DecrLeftNum(id, num int) (int64, error) {
	return s.dao.DecrLeftNum(id, num)
}

// 从缓存中获取全部的奖品
func (s *giftService) getAllByCache() []models.LtGift {
	// 集群模式，redis缓存
	key := "allgift"
	rds := datasource.InstanceCache()
	// 读取缓存
	rs, err := rds.Do("GET", key)
	if err != nil {
		log.Println("gift_service.getAllByCache GET key=", key, ", error=", err)
		return nil
	}
	str := common.GetString(rs, "")
	if str == "" {
		return nil
	}
	// 将json数据反序列化
	datalist := []map[string]interface{}{}
	err = json.Unmarshal([]byte(str), &datalist)
	if err != nil {
		log.Println("gift_service.getAllByCache json.Unmarshal error=", err)
		return nil
	}
	// 格式转换
	gifts := make([]models.LtGift, len(datalist))
	for i := 0; i < len(datalist); i++ {
		data := datalist[i]
		id := common.GetInt64FromMap(data, "Id", 0)
		if id <= 0 {
			gifts[i] = models.LtGift{}
		} else {
			gift := models.LtGift{
				Id:           int(id),
				Title:        common.GetStringFromMap(data, "Title", ""),
				PrizeNum:     int(common.GetInt64FromMap(data, "PrizeNum", 0)),
				LeftNum:      int(common.GetInt64FromMap(data, "LeftNum", 0)),
				PrizeCode:    common.GetStringFromMap(data, "PrizeCode", ""),
				PrizeTime:    int(common.GetInt64FromMap(data, "PrizeTime", 0)),
				Img:          common.GetStringFromMap(data, "Img", ""),
				Displayorder: int(common.GetInt64FromMap(data, "Displayorder", 0)),
				Gtype:        int(common.GetInt64FromMap(data, "Gtype", 0)),
				Gdata:        common.GetStringFromMap(data, "Gdata", ""),
				TimeBegin:    int(common.GetInt64FromMap(data, "TimeBegin", 0)),
				TimeEnd:      int(common.GetInt64FromMap(data, "TimeEnd", 0)),
				//PrizeData:    common.GetStringFromMap(data, "PrizeData", ""),
				PrizeBegin: int(common.GetInt64FromMap(data, "PrizeBegin", 0)),
				PrizeEnd:   int(common.GetInt64FromMap(data, "PrizeEnd", 0)),
				SysStatus:  int(common.GetInt64FromMap(data, "SysStatus", 0)),
				SysCreated: int(common.GetInt64FromMap(data, "SysCreated", 0)),
				SysUpdated: int(common.GetInt64FromMap(data, "SysUpdated", 0)),
				SysIp:      common.GetStringFromMap(data, "SysIp", ""),
			}
			gifts[i] = gift
		}
	}
	return gifts
}

// 将奖品的数据更新到缓存
func (s *giftService) setAllByCache(gifts []models.LtGift) {
	// 集群模式，redis缓存
	strValue := ""
	if len(gifts) > 0 {
		datalist := make([]map[string]interface{}, len(gifts))
		// 格式转换
		for i := 0; i < len(gifts); i++ {
			gift := gifts[i]
			data := make(map[string]interface{})
			data["Id"] = gift.Id
			data["Title"] = gift.Title
			data["PrizeNum"] = gift.PrizeNum
			data["LeftNum"] = gift.LeftNum
			data["PrizeCode"] = gift.PrizeCode
			data["PrizeTime"] = gift.PrizeTime
			data["Img"] = gift.Img
			data["Displayorder"] = gift.Displayorder
			data["Gtype"] = gift.Gtype
			data["Gdata"] = gift.Gdata
			data["TimeBegin"] = gift.TimeBegin
			data["TimeEnd"] = gift.TimeEnd
			//data["PrizeData"] = gift.PrizeData
			data["PrizeBegin"] = gift.PrizeBegin
			data["PrizeEnd"] = gift.PrizeEnd
			data["SysStatus"] = gift.SysStatus
			data["SysCreated"] = gift.SysCreated
			data["SysUpdated"] = gift.SysUpdated
			data["SysIp"] = gift.SysIp
			datalist[i] = data
		}
		str, err := json.Marshal(datalist)
		if err != nil {
			log.Println()
		}
		strValue = string(str)
	}
	key := "allgift"
	rds := datasource.InstanceCache()
	// 更新缓存
	_, err := rds.Do("SET", key, strValue)
	if err != nil {
		log.Println("gift_service.setAllByCache SET key=", key,
			", value=", strValue, ", error=", err)
	}
}

// 数据更新，需要更新缓存，直接清空缓存数据
func (s *giftService) updateByCache(data *models.LtGift, columns []string) {
	if data == nil || data.Id <= 0 {
		return
	}
	// 集群模式，redis缓存
	key := "allgift"
	rds := datasource.InstanceCache()
	// 删除redis中的缓存
	rds.Do("DEL", key)
}
