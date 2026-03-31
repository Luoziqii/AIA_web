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

// Storage 接口中彻底移除了 Delete 方法
type Storage interface {
	Save(savedPath string, reader io.Reader) error
	GetBaseURL() string
}

type LocalStorage struct {
	RootPath string
	BaseURL  string
}

func (l *LocalStorage) Save(savedPath string, reader io.Reader) error {
	cleanPath := strings.TrimPrefix(savedPath, "/")
	fullPath := filepath.Join(l.RootPath, cleanPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, reader)
	return err
}

func (l *LocalStorage) GetBaseURL() string {
	return l.BaseURL
}

type COSStorage struct {
	BucketURL string
	SecretID  string
	SecretKey string
}

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

func (c *COSStorage) GetBaseURL() string {
	return c.BucketURL
}
