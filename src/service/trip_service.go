package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rkarsnk/commute2invoice/src/models"
	"github.com/rkarsnk/commute2invoice/src/repository"
)

// TripService は、利用実績（Trip）に関するビジネスロジックを実装します。
type TripService struct {
	tripRepo repository.TripRepository
	fareRepo repository.FareRepository
}

// NewTripService は、TripService のコンストラクタです。
func NewTripService(tripRepo repository.TripRepository, fareRepo repository.FareRepository) *TripService {
	return &TripService{
		tripRepo: tripRepo,
		fareRepo: fareRepo,
	}
}

// CreateTrip は、新しい利用実績を登録します。
func (s *TripService) CreateTrip(ctx context.Context, date string, req *models.CreateTripRequest) (int, error) {
	// 日付フォーマット検証
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return 0, fmt.Errorf("validation error: invalid date format: %w", err)
	}

	// バリデーション
	if req.BillingTarget != "自社" && req.BillingTarget != "客先" {
		return 0, fmt.Errorf("validation error: invalid billing_target: %s", req.BillingTarget)
	}

	// 参照先の FareEvidence が存在するか確認
	fare, err := s.fareRepo.GetFareEvidenceByID(ctx, req.FareEvidenceID)
	if err != nil {
		return 0, fmt.Errorf("validation error: fare evidence not found: %w", err)
	}

	// 指定日付に対して、FareEvidence が有効か確認
	if fare.ValidFrom > date || (fare.ValidUntil != nil && *fare.ValidUntil < date) {
		return 0, fmt.Errorf("validation error: fare evidence is not valid for date %s", date)
	}

	trip := &models.Trip{
		Date:           date,
		FareEvidenceID: req.FareEvidenceID,
		RoundTrip:      req.RoundTrip,
		BillingTarget:  req.BillingTarget,
		CreatedAt:      time.Now().Format(time.RFC3339),
	}

	id, err := s.tripRepo.CreateTrip(ctx, trip)
	if err != nil {
		return 0, fmt.Errorf("failed to create trip: %w", err)
	}

	return id, nil
}

// GetTripByID は、IDで利用実績を取得します。
func (s *TripService) GetTripByID(ctx context.Context, id int) (*models.Trip, error) {
	if id <= 0 {
		return nil, fmt.Errorf("validation error: id must be > 0")
	}

	trip, err := s.tripRepo.GetTripByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	return trip, nil
}

// ListTripsByDate は、指定日付の利用実績一覧を取得します。
func (s *TripService) ListTripsByDate(ctx context.Context, date string) ([]*models.Trip, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("validation error: invalid date format: %w", err)
	}

	trips, err := s.tripRepo.ListTripsByDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("failed to list trips by date: %w", err)
	}

	return trips, nil
}

// GetCalendarDays は、指定年月の各日に登録された利用実績数を返します。
func (s *TripService) GetCalendarDays(ctx context.Context, year, month int) ([]models.CalendarDayInfo, error) {
	if year < 2000 || year > 2100 {
		return nil, fmt.Errorf("validation error: invalid year: %d", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("validation error: invalid month: %d", month)
	}

	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := firstDay.AddDate(0, 1, -1).Day()
	days := make([]models.CalendarDayInfo, 0, daysInMonth)

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		trips, err := s.tripRepo.ListTripsByDate(ctx, date)
		if err != nil {
			return nil, fmt.Errorf("failed to list trips for %s: %w", date, err)
		}
		days = append(days, models.CalendarDayInfo{
			Date:      date,
			TripCount: len(trips),
		})
	}

	return days, nil
}

// UpdateTrip は、利用実績を編集します。
func (s *TripService) UpdateTrip(ctx context.Context, id int, req *models.UpdateTripRequest) error {
	if id <= 0 {
		return fmt.Errorf("validation error: id must be > 0")
	}

	if req.BillingTarget != "自社" && req.BillingTarget != "客先" {
		return fmt.Errorf("validation error: invalid billing_target: %s", req.BillingTarget)
	}

	trip, err := s.tripRepo.GetTripByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get trip: %w", err)
	}

	trip.RoundTrip = req.RoundTrip
	trip.BillingTarget = req.BillingTarget

	if err := s.tripRepo.UpdateTrip(ctx, trip); err != nil {
		return fmt.Errorf("failed to update trip: %w", err)
	}

	return nil
}

// DeleteTrip は、利用実績を削除します。
func (s *TripService) DeleteTrip(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("validation error: id must be > 0")
	}

	if err := s.tripRepo.DeleteTrip(ctx, id); err != nil {
		return fmt.Errorf("failed to delete trip: %w", err)
	}

	return nil
}

// GetDailyTripDetails は、指定日付の Trip 一覧に FareEvidence 情報を組み込んで返します。
func (s *TripService) GetDailyTripDetails(ctx context.Context, date string) ([]*models.TripDetail, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("validation error: invalid date format: %w", err)
	}

	trips, err := s.tripRepo.ListTripsByDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("failed to list trips by date: %w", err)
	}

	var tripDetails []*models.TripDetail

	for _, trip := range trips {
		fare, err := s.fareRepo.GetFareEvidenceByID(ctx, trip.FareEvidenceID)
		if err != nil {
			continue
		}

		tripDetails = append(tripDetails, &models.TripDetail{
			Trip:         trip,
			FareEvidence: fare,
		})
	}

	return tripDetails, nil
}
