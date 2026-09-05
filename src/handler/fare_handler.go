// Package handler は、HTTP リクエスト・レスポンスの処理を行います。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rkarsnk/commute2invoice/src/models"
	"github.com/rkarsnk/commute2invoice/src/service"
)

// FareHandler は、運賃区間に関する HTTP エンドポイントを処理します。
type FareHandler struct {
	service *service.FareService
}

// NewFareHandler は、FareHandler のコンストラクタです。
func NewFareHandler(svc *service.FareService) *FareHandler {
	return &FareHandler{service: svc}
}

// CreateFareEvidence は、新しい区間運賃情報を登録するエンドポイント。
// POST /setup/fare-evidence
func (h *FareHandler) CreateFareEvidence(c *gin.Context) {
	var req models.CreateFareEvidenceRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	id, err := h.service.CreateFareEvidence(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
			Code:  "CREATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}

// ListValidFareEvidences は、有効な全区間一覧を取得するエンドポイント。
// GET /setup/fare-evidence
func (h *FareHandler) ListValidFareEvidences(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "date query parameter is required",
			Code:  "MISSING_PARAM",
		})
		return
	}

	fares, err := h.service.ListValidFareEvidenceForDate(c.Request.Context(), dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
			Code:  "LIST_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fare_evidences": fares,
	})
}

// ListAllFareEvidences は、登録済みの全区間一覧を取得するエンドポイント。
// GET /fare-evidences
func (h *FareHandler) ListAllFareEvidences(c *gin.Context) {
	fares, err := h.service.ListAllFareEvidence(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
			Code:  "LIST_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fare_evidences": fares,
	})
}

// DeleteFareEvidence は、未使用の区間情報を削除するエンドポイント。
// DELETE /fare-evidences/:id
func (h *FareHandler) DeleteFareEvidence(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid ID format",
			Code:  "INVALID_ID",
		})
		return
	}
	if err := h.service.DeleteFareEvidence(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: err.Error(),
			Code:  "DELETE_FAILED",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateFareEvidenceValidity は、区間の有効終了日を更新するエンドポイント。
// PATCH /fare-evidences/:id/validity
func (h *FareHandler) UpdateFareEvidenceValidity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid ID format", Code: "INVALID_ID"})
		return
	}
	var req models.UpdateFareEvidenceValidityRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format", Code: "INVALID_REQUEST"})
		return
	}
	if err := h.service.UpdateFareEvidenceValidity(c.Request.Context(), id, req.ValidUntil); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "UPDATE_FAILED"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetFareEvidenceImage は、区間の証跡画像を取得するエンドポイント。
// GET /setup/fare-evidence/:id/image
func (h *FareHandler) GetFareEvidenceImage(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid ID format",
			Code:  "INVALID_ID",
		})
		return
	}

	imageBase64, mimeType, err := h.service.GetFareEvidenceImageAsBase64(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: err.Error(),
			Code:  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_base64": imageBase64,
		"mime_type":    mimeType,
	})
}
