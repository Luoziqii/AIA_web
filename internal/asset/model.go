package asset

import (
	"errors"
	"time"
)

var ErrUnsafePath = errors.New("scope 或 name 含有非法路径字符")

type AssetRecord struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	Path          string    `gorm:"size:512;uniqueIndex;not null" json:"path"`
	URL           string    `gorm:"size:512;not null" json:"url"`
	MarkdownValue string    `gorm:"size:512;not null" json:"markdownValue"`
	Scope         string    `gorm:"size:100;index;not null" json:"scope"`
	Kind          string    `gorm:"size:20;index;not null" json:"kind"`
	Filename      string    `gorm:"size:255;not null" json:"filename"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	Ext           string    `gorm:"size:20;not null" json:"ext"`
	ContentType   string    `gorm:"size:100" json:"contentType"`
	Size          int64     `json:"size"`
	SavedPath     string    `gorm:"size:512;not null" json:"-"`
	UploadedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"uploadedAt"`
}

type ListResponse struct {
	Items    []AssetRecord `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}
