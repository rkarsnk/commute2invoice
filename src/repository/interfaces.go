// Package repository は、データアクセス層の抽象化を提供します。
// ビジネスロジック層（service）は、具体的なDB実装に依存せず、
// インターフェース経由でDBアクセスを行います。
// これにより、テスト時に mock を挿入可能です。
package repository

import (
	"context"

	"github.com/rkarsnk/commute2invoice/src/models"
)

// FareRepository は、fare_evidence テーブルのCRUD操作を定義するインターフェースです。
type FareRepository interface {
	// CreateFareEvidence は、新しい区間情報を登録します。
	// 戻り値: 作成されたレコードのID、エラー
	CreateFareEvidence(ctx context.Context, fare *models.FareEvidence) (int, error)

	// GetFareEvidenceByID は、IDで区間情報を取得します。
	GetFareEvidenceByID(ctx context.Context, id int) (*models.FareEvidence, error)

	// ListValidFareEvidence は、指定日付で有効な区間一覧を取得します。
	// valid_from <= date AND (valid_until IS NULL OR valid_until >= date) の条件で絞ります。
	ListValidFareEvidence(ctx context.Context, date string) ([]*models.FareEvidence, error)

	// ListAllFareEvidence は、全ての区間情報を取得します。（管理画面用）
	ListAllFareEvidence(ctx context.Context) ([]*models.FareEvidence, error)

	// GetFareEvidenceImage は、区間の画像バイナリを取得します。
	GetFareEvidenceImage(ctx context.Context, id int) ([]byte, string, error) // []byte = image, string = MIME type

	// DeleteFareEvidence は、参照されていない区間情報を削除します。
	DeleteFareEvidence(ctx context.Context, id int) error

	// UpdateFareEvidenceValidity は、区間の有効終了日を更新します。
	UpdateFareEvidenceValidity(ctx context.Context, id int, validUntil *string) error
}

// TripRepository は、trips テーブルのCRUD操作を定義するインターフェースです。
type TripRepository interface {
	// CreateTrip は、新しい利用実績を登録します。
	CreateTrip(ctx context.Context, trip *models.Trip) (int, error)

	// GetTripByID は、IDで利用実績を取得します。
	GetTripByID(ctx context.Context, id int) (*models.Trip, error)

	// ListTripsByDate は、指定日付の利用実績一覧を取得します。
	ListTripsByDate(ctx context.Context, date string) ([]*models.Trip, error)

	// ListTripsByYearMonth は、指定年月の全利用実績を取得します。
	// 集計・請求書生成時に使用します。
	ListTripsByYearMonth(ctx context.Context, year int, month int, billingTarget string) ([]*models.Trip, error)

	// UpdateTrip は、利用実績を更新します。
	UpdateTrip(ctx context.Context, trip *models.Trip) error

	// DeleteTrip は、利用実績を削除します。
	DeleteTrip(ctx context.Context, id int) error

	// CountTripsByDateAndBillingTarget は、指定日付・精算先の Trip 件数をカウントします。
	// カレンダー表示時に使用します。
	CountTripsByDateAndBillingTarget(ctx context.Context, date string, billingTarget string) (int, error)
}

// InvoiceRepository は、invoices テーブルのCRUD操作を定義するインターフェースです。
type InvoiceRepository interface {
	// CreateInvoice は、新しい請求書を登録します。
	CreateInvoice(ctx context.Context, invoice *models.Invoice) (int, error)

	// GetInvoiceByYearMonthAndBillingTarget は、指定年月・精算先の確定済み請求書を取得します。
	GetInvoiceByYearMonthAndBillingTarget(ctx context.Context, year int, month int, billingTarget string) (*models.Invoice, error)

	// UpdateInvoice は、請求書を更新します。（再生成時に上書き）
	UpdateInvoice(ctx context.Context, invoice *models.Invoice) error

	// DeleteInvoice は、請求書を削除します。
	DeleteInvoice(ctx context.Context, id int) error
}

// Repository は、全てのリポジトリインターフェースを集約します。
// Service 層は、この1つのインターフェースに依存するだけで済みます。
type Repository interface {
	FareRepository
	TripRepository
	InvoiceRepository
}
