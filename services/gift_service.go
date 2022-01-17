/**
 * 抽奖系统数据处理（包括数据库，也包括缓存等其他形式数据）
 */
package services

import (
	"lottery/dao"
	"lottery/models"
)

type GiftService interface {
	GetAll() []models.LtGift
	CountAll() int64
	Get(id int) *models.LtGift
	Delete(id int) error
	Update(data *models.LtGift, columns []string) error
	Create(data *models.LtGift) error
}

type giftService struct {
	dao *dao.GiftDao
}

func NewGiftService() GiftService {
	return &giftService{
		dao: dao.NewGiftDao(nil),
	}
}

func (g *giftService) GetAll() []models.LtGift {
	return g.dao.GetAll()
}

func (g *giftService) CountAll() int64 {
	return g.dao.CountAll()
}

func (g *giftService) Get(id int) *models.LtGift {
	return g.dao.Get(id)
}

func (g *giftService) Delete(id int) error {
	return g.dao.Delete(id)
}

func (g *giftService) Update(data *models.LtGift, columns []string) error {
	return g.dao.Update(data, columns)
}

func (g *giftService) Create(data *models.LtGift) error {
	return g.dao.Create(data)
}
