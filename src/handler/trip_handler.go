package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rkarsnk/commute2invoice/src/models"
	"github.com/rkarsnk/commute2invoice/src/service"
)

// TripHandler は、利用実績に関する HTTP エンドポイントを処理します。
type TripHandler struct {
	service *service.TripService
}

// NewTripHandler は、TripHandler のコンストラクタです。
func NewTripHandler(svc *service.TripService) *TripHandler {
	return &TripHandler{service: svc}
}

// CreateTrip は、新しい利用実績を登録するエンドポイント。
// POST /calendar/:date/trips
func (h *TripHandler) CreateTrip(c *gin.Context) {
	date := c.Param("date")

	var req models.CreateTripRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	id, err := h.service.CreateTrip(c.Request.Context(), date, &req)
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

// ListTripsByDate は、指定日付の利用実績一覧を取得するエンドポイント。
// GET /calendar/:date/trips
func (h *TripHandler) ListTripsByDate(c *gin.Context) {
	date := c.Param("date")

	trips, err := h.service.GetDailyTripDetails(c.Request.Context(), date)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
			Code:  "LIST_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, models.DailyTripsResponse{
		Date:  date,
		Trips: dereferenceTripDetails(trips),
	})
}

func dereferenceTripDetails(details []*models.TripDetail) []models.TripDetail {
	result := make([]models.TripDetail, 0, len(details))
	for _, detail := range details {
		result = append(result, *detail)
	}
	return result
}

// UpdateTrip は、利用実績を編集するエンドポイント。
// PUT /trips/:id
func (h *TripHandler) UpdateTrip(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid ID format",
			Code:  "INVALID_ID",
		})
		return
	}

	var req models.UpdateTripRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.service.UpdateTrip(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
			Code:  "UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Trip updated successfully",
	})
}

// DeleteTrip は、利用実績を削除するエンドポイント。
// DELETE /trips/:id
func (h *TripHandler) DeleteTrip(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid ID format",
			Code:  "INVALID_ID",
		})
		return
	}

	if err := h.service.DeleteTrip(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: err.Error(),
			Code:  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Trip deleted successfully",
	})
}
