package asset

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ErrorResponse 统一错误响应
type ErrorResponse struct {
	Error string `json:"error" example:"错误描述"`
}

// ListQuery 定义获取列表时的查询参数
type ListQuery struct {
	Scope    string `form:"scope"`
	Kind     string `form:"kind" binding:"omitempty,oneof=article shared"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize,default=20" binding:"omitempty,min=1,max=100"`
}

// UploadFile 处理静态资源上传
// @Summary 上传静态资源
// @Tags Assets
// @Accept multipart/form-data
// @Produce json
// @Param scope formData string true "第一层路径 (文章ID或公共目录)"
// @Param name formData string true "文件名主干 (不含扩展名)"
// @Param overwrite formData bool false "是否覆盖同路径资源"
// @Param file formData file true "文件流"
// @Success 201 {object} AssetRecord
// @Router /assets [post]
func (h *Handler) UploadFile(c *gin.Context) {
	fmt.Println("--- 开始处理文件上传 ---")

	scope := c.PostForm("scope")
	nameStem := c.PostForm("name")
	overwrite := c.PostForm("overwrite") == "true"
	fileHeader, err := c.FormFile("file")

	if err != nil {
		fmt.Printf("获取文件失败: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "未接收到文件"})
		return
	}

	// 1. 预处理
	record, path := h.svc.ProcessUpload(fileHeader.Filename, scope, nameStem)
	record.Size = fileHeader.Size
	fmt.Printf("预处理成功，准备保存到: %s\n", path)

	// 2. 物理保存
	file, _ := fileHeader.Open()
	defer file.Close()
	if err := h.svc.storage.Save(record.SavedPath, file); err != nil {
		fmt.Printf("物理文件保存失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "磁盘写入失败"})
		return
	}
	fmt.Println("物理文件保存成功")

	// 3. 数据库插入
	fmt.Println("准备写入数据库...")
	existing, _ := h.svc.repo.FindByPath(path)
	if existing != nil {
		if overwrite {
			record.ID = existing.ID
			err = h.svc.repo.Update(record)
			fmt.Println("执行了更新(Update)")
		} else {
			fmt.Println("文件已存在且未开启覆盖，跳过写入")
			c.JSON(http.StatusConflict, gin.H{"error": "资源已存在"})
			return
		}
	} else {
		err = h.svc.repo.Create(record)
		if err != nil {
			fmt.Printf("数据库插入报错: %v\n", err)
		} else {
			fmt.Println("数据库插入完成，ID 为:", record.ID)
		}
	}

	c.JSON(http.StatusCreated, record)
}

// ListAssets 获取资源列表
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

// DeleteAsset 删除静态资源
// @Summary 删除静态资源
// @Tags Assets
// @Param path query string true "资源完整路径 (如 /images/capoo.webp)"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Router /assets [delete]
func (h *Handler) DeleteAsset(c *gin.Context) {
	// 获取参数: /storage/assets/images/capoo.jpg
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "必须提供 path 参数"})
		return
	}

	// 1. 数据库查询 (必须精确匹配 /storage/assets/...)
	record, err := h.svc.repo.FindByPath(path)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// 2. 物理删除
	_ = h.svc.storage.Delete(record.SavedPath)

	// 3. 数据库删除
	if err := h.svc.repo.Delete(record.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "记录删除失败"})
		return
	}

	c.Status(http.StatusNoContent)
}