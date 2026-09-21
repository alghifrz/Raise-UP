package gallery

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/diuk/raiseup/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler exposes gallery HTTP endpoints.
type Handler struct {
	service        *Service
	log            *slog.Logger
	maxUploadBytes int64
}

// NewHandler creates a gallery HTTP handler.
func NewHandler(service *Service, log *slog.Logger, maxUploadBytes ...int64) *Handler {
	maxBytes := int64(10 << 20)
	if len(maxUploadBytes) > 0 && maxUploadBytes[0] > 0 {
		maxBytes = maxUploadBytes[0]
	}
	return &Handler{service: service, log: log, maxUploadBytes: maxBytes}
}

// RegisterRoutes mounts gallery routes behind auth + role middleware.
func (h *Handler) RegisterRoutes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	group := api.Group("/gallery")
	group.Use(middlewares...)

	group.GET("", h.List)
	group.POST("", h.Create)
	group.POST("/upload", h.Upload)
	group.GET("/:id", h.Get)
	group.PATCH("/:id", h.Update)
	group.PUT("/:id/image", h.ReplaceImage)
	group.DELETE("/:id", h.Delete)
}

// RegisterPublicRoutes mounts read-only gallery listing for the public portal.
func (h *Handler) RegisterPublicRoutes(api *gin.RouterGroup) {
	api.GET("/public/gallery", h.List)
}

// List handles GET /api/v1/gallery.
func (h *Handler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.BadRequest(c, "invalid page")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		response.BadRequest(c, "invalid page_size")
		return
	}

	result, err := h.service.List(c.Request.Context(), page, pageSize, c.Query("search"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSONWithMeta(c, http.StatusOK, result.Items, result.Meta)
}

// Get handles GET /api/v1/gallery/:id.
func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// Create handles POST /api/v1/gallery.
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	item, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, item)
}

// Upload handles POST /api/v1/gallery/upload.
func (h *Handler) Upload(c *gin.Context) {
	req, ok := h.bindUpload(c, true)
	if !ok {
		return
	}
	item, err := h.service.CreateUploaded(c.Request.Context(), req.UploadRequest)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, item)
}

// Update handles PATCH /api/v1/gallery/:id.
func (h *Handler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	item, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// ReplaceImage handles PUT /api/v1/gallery/:id/image.
func (h *Handler) ReplaceImage(c *gin.Context) {
	req, ok := h.bindUpload(c, false)
	if !ok {
		return
	}
	item, err := h.service.ReplaceImage(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, item)
}

// Delete handles DELETE /api/v1/gallery/:id.
func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidRequest):
		message := err.Error()
		message = strings.TrimPrefix(message, ErrInvalidRequest.Error()+": ")
		if message == ErrInvalidRequest.Error() {
			message = "Invalid request"
		}
		response.BadRequest(c, message)
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "GALLERY_ITEM_NOT_FOUND", "Gallery item not found")
	case errors.Is(err, ErrStorageUnavailable):
		response.Error(
			c,
			http.StatusServiceUnavailable,
			"STORAGE_UNAVAILABLE",
			"Gallery storage is not configured",
		)
	default:
		h.log.Error("gallery handler error", "error", err)
		response.Internal(c)
	}
}

func (h *Handler) bindUpload(c *gin.Context, create bool) (ReplaceImageRequest, bool) {
	const multipartOverhead = 1 << 20
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadBytes+multipartOverhead)

	header, err := c.FormFile("image")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			response.Error(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Image is too large")
		} else {
			response.BadRequest(c, "image file is required")
		}
		return ReplaceImageRequest{}, false
	}
	file, err := header.Open()
	if err != nil {
		response.BadRequest(c, "cannot read image file")
		return ReplaceImageRequest{}, false
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, h.maxUploadBytes+1))
	if err != nil {
		response.BadRequest(c, "cannot read image file")
		return ReplaceImageRequest{}, false
	}
	if int64(len(data)) > h.maxUploadBytes {
		response.Error(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Image is too large")
		return ReplaceImageRequest{}, false
	}

	sortOrder, hasSortOrder, ok := parseOptionalSortOrder(c.PostForm("sort_order"))
	if !ok {
		response.BadRequest(c, "sort_order must be an integer >= 0")
		return ReplaceImageRequest{}, false
	}
	caption, hasCaption := c.GetPostForm("caption")
	if create {
		hasCaption = true
		hasSortOrder = sortOrder != nil
	}

	return ReplaceImageRequest{
		UploadRequest: UploadRequest{
			FileName:    header.Filename,
			ContentType: header.Header.Get("Content-Type"),
			Data:        data,
			Caption:     caption,
			SortOrder:   sortOrder,
		},
		UpdateCaption:   hasCaption,
		UpdateSortOrder: hasSortOrder,
	}, true
}

func parseOptionalSortOrder(raw string) (*int32, bool, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false, true
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || value < 0 {
		return nil, true, false
	}
	parsed := int32(value)
	return &parsed, true, true
}
