package asset

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
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

type COSStorage struct {
	BucketURL string // 格式: https://<BucketName-APPID>.cos.<Region>.myqcloud.com
	SecretID  string // 腾讯云控制台获取
	SecretKey string // 腾讯云控制台获取
}

// newClient 私有辅助方法：初始化腾讯云 SDK 客户端
func (c *COSStorage) newClient() *cos.Client {
	u, _ := url.Parse(c.BucketURL)
	b := &cos.BaseURL{BucketURL: u}
	return cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  c.SecretID,
			SecretKey: c.SecretKey,
		},
	})
}

func (c *COSStorage) Save(savedPath string, reader io.Reader) error {
	client := c.newClient()
	key := strings.TrimPrefix(savedPath, "/")

	_, err := client.Object.Put(context.Background(), key, reader, nil)
	return err
}

func (c *COSStorage) Delete(savedPath string) error {
	client := c.newClient()
	key := strings.TrimPrefix(savedPath, "/")

	// 调用官方 SDK 的 Delete 方法
	_, err := client.Object.Delete(context.Background(), key)
	return err
}

func (c *COSStorage) GetBaseURL() string {
	// 返回 Bucket 域名，生成的 URL 将会是 BucketURL + savedPath
	return c.BucketURL
}