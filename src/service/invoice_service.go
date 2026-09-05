package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/rkarsnk/commute2invoice/src/models"
	"github.com/rkarsnk/commute2invoice/src/pdf"
	"github.com/rkarsnk/commute2invoice/src/repository"
)

// InvoiceService は、請求書（Invoice）に関するビジネスロジックを実装します。
type InvoiceService struct {
	invoiceRepo repository.InvoiceRepository
	tripRepo    repository.TripRepository
	fareRepo    repository.FareRepository
}

// NewInvoiceService は、InvoiceService のコンストラクタです。
func NewInvoiceService(
	invoiceRepo repository.InvoiceRepository,
	tripRepo repository.TripRepository,
	fareRepo repository.FareRepository,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		tripRepo:    tripRepo,
		fareRepo:    fareRepo,
	}
}

// GetInvoicePreview は、請求書のプレビュー情報を生成します。
func (s *InvoiceService) GetInvoicePreview(ctx context.Context, year int, month int, billingTarget string) (*models.InvoicePreviewResponse, error) {
	// バリデーション
	if year < 2000 || year > 2100 {
		return nil, fmt.Errorf("validation error: invalid year: %d", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("validation error: invalid month: %d", month)
	}
	if billingTarget != "自社" && billingTarget != "客先" {
		return nil, fmt.Errorf("validation error: invalid billing_target: %s", billingTarget)
	}

	// 指定年月・精算先の Trip を全て取得
	trips, err := s.tripRepo.ListTripsByYearMonth(ctx, year, month, billingTarget)
	if err != nil {
		return nil, fmt.Errorf("failed to list trips: %w", err)
	}

	// 請求書行データを構築
	var invoiceLines []models.InvoiceLineItem
	var totalFare int
	usedFareEvidenceIDs := make(map[int]bool)

	for _, trip := range trips {
		fare, err := s.fareRepo.GetFareEvidenceByID(ctx, trip.FareEvidenceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get fare evidence: %w", err)
		}

		tripFare := trip.CalculateFare(fare)

		invoiceLines = append(invoiceLines, models.InvoiceLineItem{
			Date:        trip.Date,
			CompanyText: fare.CompanyText,
			FromText:    fare.FromText,
			ToText:      fare.ToText,
			FareOneway:  fare.FareOneway,
			Fare:        tripFare,
		})

		totalFare += tripFare
		usedFareEvidenceIDs[trip.FareEvidenceID] = true
	}

	// 証跡画像を集約
	var evidenceImages []models.EvidenceImageInfo

	for fareEvidenceID := range usedFareEvidenceIDs {
		imageBase64, mimeType, err := s.GetFareEvidenceImageAsBase64(ctx, fareEvidenceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get fare evidence image: %w", err)
		}

		evidenceImages = append(evidenceImages, models.EvidenceImageInfo{
			FareEvidenceID: fareEvidenceID,
			MimeType:       mimeType,
			ImageBase64:    imageBase64,
		})
	}

	// 既存の確定済み請求書があるか確認
	status := "draft"
	var confirmedAt *string

	existingInvoice, err := s.invoiceRepo.GetInvoiceByYearMonthAndBillingTarget(ctx, year, month, billingTarget)
	if err == nil {
		status = "confirmed"
		confirmedAt = &existingInvoice.GeneratedAt
	}

	return &models.InvoicePreviewResponse{
		Year:           year,
		Month:          month,
		BillingTarget:  billingTarget,
		Status:         status,
		ConfirmedAt:    confirmedAt,
		InvoiceLines:   invoiceLines,
		TotalFare:      totalFare,
		EvidenceImages: evidenceImages,
	}, nil
}

// ConfirmInvoice は、請求書を確定し、PDF を生成・保存します。
func (s *InvoiceService) ConfirmInvoice(ctx context.Context, year int, month int, billingTarget string) (int, string, error) {
	// バリデーション
	if year < 2000 || year > 2100 {
		return 0, "", fmt.Errorf("validation error: invalid year: %d", year)
	}
	if month < 1 || month > 12 {
		return 0, "", fmt.Errorf("validation error: invalid month: %d", month)
	}
	if billingTarget != "自社" && billingTarget != "客先" {
		return 0, "", fmt.Errorf("validation error: invalid billing_target: %s", billingTarget)
	}

	// 請求書プレビューを取得
	preview, err := s.GetInvoicePreview(ctx, year, month, billingTarget)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get invoice preview: %w", err)
	}

	if len(preview.InvoiceLines) == 0 {
		return 0, "", fmt.Errorf("no invoice data: no trips for year=%d, month=%d, billingTarget=%s", year, month, billingTarget)
	}

	pdfData, err := pdf.Generate(ctx, preview)
	if err != nil {
		return 0, "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	// 既存の確定済み請求書があるか確認
	existingInvoice, err := s.invoiceRepo.GetInvoiceByYearMonthAndBillingTarget(ctx, year, month, billingTarget)
	if err == nil {
		// 既存レコードがあれば上書き（UPDATE）
		existingInvoice.PDF = pdfData
		if err := s.invoiceRepo.UpdateInvoice(ctx, existingInvoice); err != nil {
			return 0, "", fmt.Errorf("failed to update invoice: %w", err)
		}
		return existingInvoice.ID, existingInvoice.GeneratedAt, nil
	}

	// 新規作成（INSERT）
	invoice := &models.Invoice{
		YearMonth:     fmt.Sprintf("%04d-%02d", year, month),
		BillingTarget: billingTarget,
		PDF:           pdfData,
		GeneratedAt:   time.Now().Format(time.RFC3339),
	}

	id, err := s.invoiceRepo.CreateInvoice(ctx, invoice)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create invoice: %w", err)
	}

	return id, invoice.GeneratedAt, nil
}

// GetInvoicePDF は、確定済みの請求書 PDF を取得します。
func (s *InvoiceService) GetInvoicePDF(ctx context.Context, year int, month int, billingTarget string) ([]byte, error) {
	if year < 2000 || year > 2100 {
		return nil, fmt.Errorf("validation error: invalid year: %d", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("validation error: invalid month: %d", month)
	}
	if billingTarget != "自社" && billingTarget != "客先" {
		return nil, fmt.Errorf("validation error: invalid billing_target: %s", billingTarget)
	}

	invoice, err := s.invoiceRepo.GetInvoiceByYearMonthAndBillingTarget(ctx, year, month, billingTarget)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice.PDF, nil
}

// IsInvoiceConfirmed は、指定年月・精算先の請求書が確定済みかを返します。
func (s *InvoiceService) IsInvoiceConfirmed(ctx context.Context, year, month int, billingTarget string) (bool, error) {
	if year < 2000 || year > 2100 {
		return false, fmt.Errorf("validation error: invalid year: %d", year)
	}
	if month < 1 || month > 12 {
		return false, fmt.Errorf("validation error: invalid month: %d", month)
	}
	if billingTarget != "自社" && billingTarget != "客先" {
		return false, fmt.Errorf("validation error: invalid billing_target: %s", billingTarget)
	}

	_, err := s.invoiceRepo.GetInvoiceByYearMonthAndBillingTarget(ctx, year, month, billingTarget)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetFareEvidenceImageAsBase64 は、区間の画像をBase64エンコードして返します。
func (s *InvoiceService) GetFareEvidenceImageAsBase64(ctx context.Context, id int) (string, string, error) {
	if id <= 0 {
		return "", "", fmt.Errorf("validation error: id must be > 0")
	}

	imageData, mimeType, err := s.fareRepo.GetFareEvidenceImage(ctx, id)
	if err != nil {
		return "", "", fmt.Errorf("failed to get fare evidence image: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	return base64Image, mimeType, nil
}
