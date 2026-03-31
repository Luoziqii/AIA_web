package asset

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ListQuery struct {
	Scope    string `form:"scope"`
	Kind     string `form:"kind" binding:"omitempty,oneof=article shared"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize,default=20" binding:"omitempty,min=1,max=100"`
}

// @Summary 上传静态资源
// @Tags Assets
// @Accept multipart/form-data
// @Produce json
// @Param scope formData string true "第一层路径"
// @Param name formData string true "文件名主干"
// @Param overwrite formData bool false "是否允许覆盖同路径资源"
// @Param file formData file true "文件"
// @Success 201 {object} AssetRecord
// @Router /assets [post]
func (h *Handler) UploadFile(c *gin.Context) {
	scope := c.PostForm("scope")
	nameStem := c.PostForm("name")
	overwrite := c.PostForm("overwrite") == "true"

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "未接收到有效文件"})
		return
	}

	if scope == "" || nameStem == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "scope 和 name 为必填项"})
		return
	}

	record, savedPath, err := h.svc.ProcessUpload(fileHeader.Filename, scope, nameStem)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	record.Size = fileHeader.Size

	existing, err := h.svc.repo.FindByPath(record.Path)
	if err == nil && existing != nil {
		if !overwrite {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "资源已存在，如需覆盖请开启 overwrite 模式"})
			return
		}
		record.ID = existing.ID
		record.UploadedAt = existing.UploadedAt
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "无法读取文件流"})
		return
	}
	defer file.Close()

	if err := h.svc.storage.Save(savedPath, file); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "物理存储写入失败: " + err.Error()})
		return
	}

	if record.ID > 0 {
		if err := h.svc.repo.Update(record); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "更新数据库记录失败"})
			return
		}
	} else {
		if err := h.svc.repo.Create(record); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "创建数据库记录失败"})
			return
		}
	}

	c.JSON(http.StatusCreated, record)
}

// @Summary 获取静态资源列表
// @Tags Assets
// @Produce json
// @Param scope query string false "路径范围"
// @Param kind query string false "类型" enums(article,shared)
// @Param page query int false "页码"
// @Router /assets [get]
func (h *Handler) ListAssets(c *gin.Context) {
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "查询参数格式错误"})
		return
	}

	items, total, err := h.svc.repo.List(q.Scope, q.Kind, q.Keyword, q.Page, q.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取列表失败"})
		return
	}

	c.JSON(http.StatusOK, ListResponse{
		Items:    items,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

// @Summary 删除静态资源
// @Tags Assets
// @Param path query string true "资源完整路径 (如 /images/capoo.webp)"
// @Success 204 "No Content"
// @Router /assets [delete]
func (h *Handler) DeleteAsset(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "必须提供 path 参数"})
		return
	}

	record, err := h.svc.repo.FindByPath(path)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// 仅执行数据库删除，保留物理文件（孤儿文件）
	if err := h.svc.repo.Delete(record.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "记录删除失败"})
		return
	}

	c.Status(http.StatusNoContent)
}
