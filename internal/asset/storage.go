package asset

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Storage interface {
	Save(savedPath string, reader io.Reader) error
	Delete(savedPath string) error
	GetBaseURL() string
}

type LocalStorage struct {
	// RootPath 传 "." 即可，表示项目根目录
	RootPath string 
	// BaseURL 传 "http://localhost:8080/assets"
	BaseURL  string 
}

func (l *LocalStorage) Save(savedPath string, reader io.Reader) error {
	// 1. 确保 savedPath 不带前缀斜杠 (如 storage/assets/...)
	cleanPath := strings.TrimPrefix(savedPath, "/")
	// 2. 拼接物理路径 (./storage/assets/...)
	fullPath := filepath.Join(l.RootPath, cleanPath)

	// 3. 创建目录
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	// 4. 创建并写入
	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, reader)
	return err
}

func (l *LocalStorage) Delete(savedPath string) error {
	cleanPath := strings.TrimPrefix(savedPath, "/")
	fullPath := filepath.Join(l.RootPath, cleanPath)
	return os.Remove(fullPath)
}

func (l *LocalStorage) GetBaseURL() string {
	return l.BaseURL
}