// Package service は、ビジネスロジック層を提供します。
// handler から呼び出され、repository を使用してデータアクセスを行います。
// バリデーション・集計・複雑な処理はこの層で実装されます。
package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/rkarsnk/commute2invoice/src/models"
	"github.com/rkarsnk/commute2invoice/src/repository"
)

// FareService は、運賃区間（FareEvidence）に関するビジネスロジックを実装します。
type FareService struct {
	repo repository.FareRepository
}

// NewFareService は、FareService のコンストラクタです。
func NewFareService(repo repository.FareRepository) *FareService {
	return &FareService{repo: repo}
}

// CreateFareEvidence は、運賃区間情報を登録します。
func (s *FareService) CreateFareEvidence(ctx context.Context, req *models.CreateFareEvidenceRequest) (int, error) {
	// バリデーション
	if req.Label == "" {
		return 0, fmt.Errorf("validation error: label is required")
	}
	if req.FareOneway <= 0 {
		return 0, fmt.Errorf("validation error: fare_oneway must be > 0")
	}
	if req.FromText == "" || req.ToText == "" {
		return 0, fmt.Errorf("validation error: station names are required")
	}

	if _, err := time.Parse("2006-01-02", req.ValidFrom); err != nil {
		return 0, fmt.Errorf("validation error: invalid date format: %w", err)
	}
	if req.ValidUntil != nil {
		if _, err := time.Parse("2006-01-02", *req.ValidUntil); err != nil {
			return 0, fmt.Errorf("validation error: invalid valid_until format: %w", err)
		}
		if *req.ValidUntil < req.ValidFrom {
			return 0, fmt.Errorf("validation error: valid_until must not be before valid_from")
		}
	}

	// Base64 画像をデコード
	imageData, err := base64.StdEncoding.DecodeString(req.ImageData)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64 image: %w", err)
	}

	if len(imageData) == 0 {
		return 0, fmt.Errorf("validation error: image data is empty")
	}

	fare := &models.FareEvidence{
		Label:         req.Label,
		CompanyText:   req.CompanyText,
		FromText:      req.FromText,
		ToText:        req.ToText,
		FareOneway:    req.FareOneway,
		FareRoundtrip: req.FareRoundtrip,
		Image:         imageData,
		ImageMime:     req.ImageMime,
		ValidFrom:     req.ValidFrom,
		ValidUntil:    req.ValidUntil,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	id, err := s.repo.CreateFareEvidence(ctx, fare)
	if err != nil {
		return 0, fmt.Errorf("failed to create fare evidence: %w", err)
	}

	return id, nil
}

// GetFareEvidenceByID は、IDで区間情報を取得します。
func (s *FareService) GetFareEvidenceByID(ctx context.Context, id int) (*models.FareEvidence, error) {
	if id <= 0 {
		return nil, fmt.Errorf("validation error: id must be > 0")
	}

	fare, err := s.repo.GetFareEvidenceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get fare evidence: %w", err)
	}

	return fare, nil
}

// ListValidFareEvidenceForDate は、指定日付で有効な区間一覧を取得します。
func (s *FareService) ListValidFareEvidenceForDate(ctx context.Context, date string) ([]*models.FareEvidence, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("validation error: invalid date format: %w", err)
	}

	fares, err := s.repo.ListValidFareEvidence(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("failed to list valid fare evidence: %w", err)
	}

	return fares, nil
}

// ListAllFareEvidence は、全ての区間情報を取得します。
func (s *FareService) ListAllFareEvidence(ctx context.Context) ([]*models.FareEvidence, error) {
	fares, err := s.repo.ListAllFareEvidence(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all fare evidence: %w", err)
	}

	return fares, nil
}

// DeleteFareEvidence は、運賃区間情報を削除します。
func (s *FareService) DeleteFareEvidence(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("validation error: id must be > 0")
	}
	if err := s.repo.DeleteFareEvidence(ctx, id); err != nil {
		return fmt.Errorf("failed to delete fare evidence: %w", err)
	}
	return nil
}

// UpdateFareEvidenceValidity は、使用済み区間を含む有効終了日を更新します。
func (s *FareService) UpdateFareEvidenceValidity(ctx context.Context, id int, validUntil *string) error {
	if id <= 0 {
		return fmt.Errorf("validation error: id must be > 0")
	}
	fare, err := s.repo.GetFareEvidenceByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get fare evidence: %w", err)
	}
	if validUntil != nil {
		if _, err := time.Parse("2006-01-02", *validUntil); err != nil {
			return fmt.Errorf("validation error: invalid date format: %w", err)
		}
		if *validUntil < fare.ValidFrom {
			return fmt.Errorf("validation error: valid_until must not be before valid_from")
		}
	}
	if err := s.repo.UpdateFareEvidenceValidity(ctx, id, validUntil); err != nil {
		return fmt.Errorf("failed to update fare evidence validity: %w", err)
	}
	return nil
}

// GetFareEvidenceImageAsBase64 は、区間の画像をBase64エンコードして返します。
func (s *FareService) GetFareEvidenceImageAsBase64(ctx context.Context, id int) (string, string, error) {
	if id <= 0 {
		return "", "", fmt.Errorf("validation error: id must be > 0")
	}

	imageData, mimeType, err := s.repo.GetFareEvidenceImage(ctx, id)
	if err != nil {
		return "", "", fmt.Errorf("failed to get fare evidence image: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	return base64Image, mimeType, nil
}
