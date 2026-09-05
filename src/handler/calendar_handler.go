package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rkarsnk/commute2invoice/src/models"
	"github.com/rkarsnk/commute2invoice/src/service"
)

// CalendarHandler は、カレンダー・月別集計に関する HTTP エンドポイントを処理します。
type CalendarHandler struct {
	tripService    *service.TripService
	invoiceService *service.InvoiceService
}

// NewCalendarHandler は、CalendarHandler のコンストラクタです。
func NewCalendarHandler(tripSvc *service.TripService, invoiceSvc *service.InvoiceService) *CalendarHandler {
	return &CalendarHandler{tripService: tripSvc, invoiceService: invoiceSvc}
}

// GetCalendar は、指定年月のカレンダーデータを取得するエンドポイント。
// GET /calendar
func (h *CalendarHandler) GetCalendar(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	if yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "year and month query parameters are required",
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

	days, err := h.tripService.GetCalendarDays(c.Request.Context(), year, month)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_PARAM",
		})
		return
	}

	companyConfirmed, err := h.invoiceService.IsInvoiceConfirmed(c.Request.Context(), year, month, "自社")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
			Code:  "CALENDAR_FAILED",
		})
		return
	}
	clientConfirmed, err := h.invoiceService.IsInvoiceConfirmed(c.Request.Context(), year, month, "客先")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
			Code:  "CALENDAR_FAILED",
		})
		return
	}

	for i := range days {
		days[i].InvoiceStatus = map[string]bool{
			"自社": companyConfirmed,
			"客先": clientConfirmed,
		}
	}

	calendarResponse := &models.CalendarResponse{
		Year:  year,
		Month: month,
		Days:  days,
	}

	c.JSON(http.StatusOK, calendarResponse)
}
