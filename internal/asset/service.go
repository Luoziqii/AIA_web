package asset

import (
	"mime"
	"path/filepath"
	"strings"
)

type Service struct {
	repo    *Repository
	storage Storage // 核心：注入接口
}

func NewService(repo *Repository, storage Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

// ProcessUpload 处理上传逻辑
func (s *Service) ProcessUpload(originalName, scope, nameStem string) (*AssetRecord, string) {
	extWithDot := filepath.Ext(originalName)
	ext := strings.TrimPrefix(extWithDot, ".")
	contentType := mime.TypeByExtension(extWithDot)
	filename := nameStem + extWithDot
	
	// 1. Path (用于数据库标识和传入参数): /storage/assets/scope/filename
	path := "/storage/assets/" + scope + "/" + filename
	
	// 2. SavedPath (用于磁盘操作): storage/assets/scope/filename
	savedPath := "storage/assets/" + scope + "/" + filename
	
	// 3. URL (用于前端访问): http://localhost:8080/assets/scope/filename
	accessURL := s.storage.GetBaseURL() + "/" + scope + "/" + filename
	
	kind := "shared"
	markdownValue := "/assets/" + scope + "/" + filename // 推荐 Markdown 路径
	if strings.HasPrefix(scope, "article-") || len(scope) > 20 {
		kind = "article"
		markdownValue = filename
	}

	record := &AssetRecord{
		Path:          path,
		URL:           accessURL,    
		MarkdownValue: markdownValue,
		Scope:         scope,
		Kind:          kind,
		Filename:      filename,
		Name:          nameStem,
		Ext:           ext,
		ContentType:   contentType,
		SavedPath:     savedPath,
	}

	return record, savedPath
}