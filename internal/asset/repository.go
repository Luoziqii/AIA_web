package asset

import (
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	_ = db.AutoMigrate(&AssetRecord{})
	return &Repository{db: db}
}

func (r *Repository) Create(a *AssetRecord) error {
	err := r.db.Create(a).Error
	if err != nil {
		fmt.Printf("数据库插入失败原因: %v\n", err)
	}
	return err
}

func (r *Repository) Update(a *AssetRecord) error {
	return r.db.Save(a).Error
}

func (r *Repository) FindByPath(path string) (*AssetRecord, error) {
	var a AssetRecord
	err := r.db.Where("path = ?", path).First(&a).Error
	if err != nil {
		return nil, err 
	}
	return &a, nil
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&AssetRecord{}, id).Error
}

func (r *Repository) List(scope, kind, keyword string, page, pageSize int) ([]AssetRecord, int64, error) {
	var items []AssetRecord
	var total int64
	db := r.db.Model(&AssetRecord{})

	if scope != "" { db = db.Where("scope = ?", scope) }
	if kind != "" { db = db.Where("kind = ?", kind) }
	if keyword != "" { db = db.Where("name LIKE ?", "%"+keyword+"%") }

	db.Count(&total)
	err := db.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}