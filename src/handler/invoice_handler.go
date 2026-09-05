package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rkarsnk/commute2invoice/src/models"
	"github.com/rkarsnk/commute2invoice/src/service"
)

// InvoiceHandler は、請求書に関する HTTP エンドポイントを処理します。
type InvoiceHandler struct {
	service *service.InvoiceService
}

// NewInvoiceHandler は、InvoiceHandler のコンストラクタです。
func NewInvoiceHandler(svc *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{service: svc}
}

// GetInvoicePreview は、請求書のプレビューを取得するエンドポイント。
// GET /invoices/preview
func (h *InvoiceHandler) GetInvoicePreview(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")
	billingTarget := c.Query("billing_target")

	if yearStr == "" || monthStr == "" || billingTarget == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "year, month, billing_target query parameters are required",
			Code:  "MISSING_PARAM",
		})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid year format",
			Code:  "INVALID_PARAM",
		})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid month format",
			Code:  "INVALID_PARAM",
		})
		return
	}

	preview, err := h.service.GetInvoicePreview(c.Request.Context(), year, month, billingTarget)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
			Code:  "PREVIEW_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// ConfirmInvoice は、請求書を確定するエンドポイント。
// POST /invoices/confirm
func (h *InvoiceHandler) ConfirmInvoice(c *gin.Context) {
	var req models.ConfirmInvoiceRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	id, generatedAt, err := h.service.ConfirmInvoice(c.Request.Context(), req.Year, req.Month, req.BillingTarget)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
			Code:  "CONFIRM_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, models.ConfirmInvoiceResponse{
		ID:          id,
		ConfirmedAt: generatedAt,
		PDFURL:      fmt.Sprintf("/invoices/%d/%d/%s/pdf", req.Year, req.Month, req.BillingTarget),
	})
}

// GetInvoicePDF は、確定済み請求書の PDF をダウンロードするエンドポイント。
// GET /invoices/:year/:month/:billing_target/pdf
func (h *InvoiceHandler) GetInvoicePDF(c *gin.Context) {
	yearStr := c.Param("year")
	monthStr := c.Param("month")
	billingTarget := c.Param("billing_target")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid year format",
			Code:  "INVALID_PARAM",
		})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid month format",
			Code:  "INVALID_PARAM",
		})
		return
	}

	pdf, err := h.service.GetInvoicePDF(c.Request.Context(), year, month, billingTarget)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: err.Error(),
			Code:  "NOT_FOUND",
		})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\"invoice.pdf\"")
	c.Data(http.StatusOK, "application/pdf", pdf)
}
