package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rkarsnk/commute2invoice/src/models"
)

// SQLiteRepository は、SQLite を使った Repository インターフェースの実装です。
// Dependency Injection (DI) パターンにより、*sql.DB を注入されます。
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository は、SQLiteRepository のコンストラクタです。
// service層 から Repository インターフェース経由で使用します。
func NewSQLiteRepository(db *sql.DB) Repository {
	return &SQLiteRepository{db: db}
}

// ===== FareRepository 実装 =====

// CreateFareEvidence は、新しい区間情報を fare_evidence テーブルに登録します。
func (r *SQLiteRepository) CreateFareEvidence(ctx context.Context, fare *models.FareEvidence) (int, error) {
	// SQL INSERT 文。? はプレースホルダ（SQLインジェクション対策）
	query := `
		INSERT INTO fare_evidence 
		(label, company_text, from_text, to_text, fare_oneway, fare_roundtrip, image, image_mime, valid_from, valid_until)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// ExecContext は、コンテキストをサポートした実行メソッド。
	// ctx でキャンセルやタイムアウト処理が可能です。
	result, err := r.db.ExecContext(ctx, query,
		fare.Label,
		fare.CompanyText,
		fare.FromText,
		fare.ToText,
		fare.FareOneway,
		fare.FareRoundtrip, // nil の場合も OK（SQLの NULL になる）
		fare.Image,         // BLOB はそのままバイナリで保存
		fare.ImageMime,
		fare.ValidFrom,
		fare.ValidUntil,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to insert fare_evidence: %w", err)
	}

	// LastInsertRowid で、自動採番されたIDを取得
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert row id: %w", err)
	}

	return int(id), nil
}

// GetFareEvidenceByID は、IDで区間情報を取得します。
func (r *SQLiteRepository) GetFareEvidenceByID(ctx context.Context, id int) (*models.FareEvidence, error) {
	query := `
		SELECT id, label, company_text, from_text, to_text, fare_oneway, fare_roundtrip, 
		       image, image_mime, valid_from, valid_until, created_at
		FROM fare_evidence
		WHERE id = ?
	`

	fare := &models.FareEvidence{}

	// QueryRowContext は1行取得。Scan で各カラムを構造体にマップします。
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&fare.ID,
		&fare.Label,
		&fare.CompanyText,
		&fare.FromText,
		&fare.ToText,
		&fare.FareOneway,
		&fare.FareRoundtrip, // nil の場合、SCAN時に自動的に nil になります
		&fare.Image,
		&fare.ImageMime,
		&fare.ValidFrom,
		&fare.ValidUntil,
		&fare.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("fare evidence not found: id=%d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query fare_evidence: %w", err)
	}

	return fare, nil
}

// ListValidFareEvidence は、指定日付で有効な区間一覧を取得します。
func (r *SQLiteRepository) ListValidFareEvidence(ctx context.Context, date string) ([]*models.FareEvidence, error) {
	query := `
		SELECT id, label, company_text, from_text, to_text, fare_oneway, fare_roundtrip, 
		       image, image_mime, valid_from, valid_until, created_at
		FROM fare_evidence
		WHERE valid_from <= ? AND (valid_until IS NULL OR valid_until >= ?)
		ORDER BY id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, date, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query valid fare_evidence: %w", err)
	}
	defer rows.Close()

	var fares []*models.FareEvidence

	for rows.Next() {
		fare := &models.FareEvidence{}
		if err := rows.Scan(
			&fare.ID,
			&fare.Label,
			&fare.CompanyText,
			&fare.FromText,
			&fare.ToText,
			&fare.FareOneway,
			&fare.FareRoundtrip,
			&fare.Image,
			&fare.ImageMime,
			&fare.ValidFrom,
			&fare.ValidUntil,
			&fare.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan fare_evidence row: %w", err)
		}
		fares = append(fares, fare)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return fares, nil
}

// ListAllFareEvidence は、全ての区間情報を取得します。
func (r *SQLiteRepository) ListAllFareEvidence(ctx context.Context) ([]*models.FareEvidence, error) {
	query := `
		SELECT id, label, company_text, from_text, to_text, fare_oneway, fare_roundtrip, 
		       image, image_mime, valid_from, valid_until, created_at
		FROM fare_evidence
		ORDER BY id DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all fare_evidence: %w", err)
	}
	defer rows.Close()

	var fares []*models.FareEvidence

	for rows.Next() {
		fare := &models.FareEvidence{}
		if err := rows.Scan(
			&fare.ID,
			&fare.Label,
			&fare.CompanyText,
			&fare.FromText,
			&fare.ToText,
			&fare.FareOneway,
			&fare.FareRoundtrip,
			&fare.Image,
			&fare.ImageMime,
			&fare.ValidFrom,
			&fare.ValidUntil,
			&fare.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan fare_evidence row: %w", err)
		}
		fares = append(fares, fare)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return fares, nil
}

// GetFareEvidenceImage は、区間の画像バイナリを取得します。
func (r *SQLiteRepository) GetFareEvidenceImage(ctx context.Context, id int) ([]byte, string, error) {
	query := `
		SELECT image, image_mime FROM fare_evidence WHERE id = ?
	`

	var imageData []byte
	var imageMime string

	err := r.db.QueryRowContext(ctx, query, id).Scan(&imageData, &imageMime)
	if err == sql.ErrNoRows {
		return nil, "", fmt.Errorf("fare evidence image not found: id=%d", id)
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to query fare_evidence image: %w", err)
	}

	return imageData, imageMime, nil
}

// DeleteFareEvidence は、利用実績から参照されていない区間情報を削除します。
func (r *SQLiteRepository) DeleteFareEvidence(ctx context.Context, id int) error {
	var tripCount int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM trips WHERE fare_evidence_id = ?", id).Scan(&tripCount); err != nil {
		return fmt.Errorf("failed to check fare evidence references: %w", err)
	}
	if tripCount > 0 {
		return fmt.Errorf("fare evidence is used by %d trip(s) and cannot be deleted", tripCount)
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM fare_evidence WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete fare_evidence: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("fare evidence not found: id=%d", id)
	}
	return nil
}

// UpdateFareEvidenceValidity は、区間の有効終了日を更新します。
func (r *SQLiteRepository) UpdateFareEvidenceValidity(ctx context.Context, id int, validUntil *string) error {
	result, err := r.db.ExecContext(ctx, "UPDATE fare_evidence SET valid_until = ? WHERE id = ?", validUntil, id)
	if err != nil {
		return fmt.Errorf("failed to update fare_evidence validity: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("fare evidence not found: id=%d", id)
	}
	return nil
}

// ===== TripRepository 実装 =====

// CreateTrip は、新しい利用実績を trips テーブルに登録します。
func (r *SQLiteRepository) CreateTrip(ctx context.Context, trip *models.Trip) (int, error) {
	query := `
		INSERT INTO trips (date, fare_evidence_id, round_trip, billing_target)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		trip.Date,
		trip.FareEvidenceID,
		boolToInt(trip.RoundTrip),
		trip.BillingTarget,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to insert trip: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert row id: %w", err)
	}

	return int(id), nil
}

// GetTripByID は、IDで利用実績を取得します。
func (r *SQLiteRepository) GetTripByID(ctx context.Context, id int) (*models.Trip, error) {
	query := `
		SELECT id, date, fare_evidence_id, round_trip, billing_target, created_at
		FROM trips
		WHERE id = ?
	`

	trip := &models.Trip{}
	var roundTripInt int

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&trip.ID,
		&trip.Date,
		&trip.FareEvidenceID,
		&roundTripInt,
		&trip.BillingTarget,
		&trip.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("trip not found: id=%d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query trip: %w", err)
	}

	trip.RoundTrip = intToBool(roundTripInt)

	return trip, nil
}

// ListTripsByDate は、指定日付の利用実績一覧を取得します。
func (r *SQLiteRepository) ListTripsByDate(ctx context.Context, date string) ([]*models.Trip, error) {
	query := `
		SELECT id, date, fare_evidence_id, round_trip, billing_target, created_at
		FROM trips
		WHERE date = ?
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query trips by date: %w", err)
	}
	defer rows.Close()

	var trips []*models.Trip

	for rows.Next() {
		trip := &models.Trip{}
		var roundTripInt int

		if err := rows.Scan(
			&trip.ID,
			&trip.Date,
			&trip.FareEvidenceID,
			&roundTripInt,
			&trip.BillingTarget,
			&trip.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trip row: %w", err)
		}

		trip.RoundTrip = intToBool(roundTripInt)
		trips = append(trips, trip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return trips, nil
}

// ListTripsByYearMonth は、指定年月・精算先の全利用実績を取得します。
func (r *SQLiteRepository) ListTripsByYearMonth(ctx context.Context, year int, month int, billingTarget string) ([]*models.Trip, error) {
	datePattern := fmt.Sprintf("%04d-%02d%%", year, month)

	query := `
		SELECT id, date, fare_evidence_id, round_trip, billing_target, created_at
		FROM trips
		WHERE date LIKE ? AND billing_target = ?
		ORDER BY date
	`

	rows, err := r.db.QueryContext(ctx, query, datePattern, billingTarget)
	if err != nil {
		return nil, fmt.Errorf("failed to query trips by year-month: %w", err)
	}
	defer rows.Close()

	var trips []*models.Trip

	for rows.Next() {
		trip := &models.Trip{}
		var roundTripInt int

		if err := rows.Scan(
			&trip.ID,
			&trip.Date,
			&trip.FareEvidenceID,
			&roundTripInt,
			&trip.BillingTarget,
			&trip.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trip row: %w", err)
		}

		trip.RoundTrip = intToBool(roundTripInt)
		trips = append(trips, trip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return trips, nil
}

// UpdateTrip は、利用実績を更新します。
func (r *SQLiteRepository) UpdateTrip(ctx context.Context, trip *models.Trip) error {
	query := `
		UPDATE trips
		SET round_trip = ?, billing_target = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		boolToInt(trip.RoundTrip),
		trip.BillingTarget,
		trip.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update trip: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("trip not found: id=%d", trip.ID)
	}

	return nil
}

// DeleteTrip は、利用実績を削除します。
func (r *SQLiteRepository) DeleteTrip(ctx context.Context, id int) error {
	query := `DELETE FROM trips WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete trip: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("trip not found: id=%d", id)
	}

	return nil
}

// CountTripsByDateAndBillingTarget は、指定日付・精算先の Trip 件数をカウントします。
func (r *SQLiteRepository) CountTripsByDateAndBillingTarget(ctx context.Context, date string, billingTarget string) (int, error) {
	query := `
		SELECT COUNT(*) FROM trips
		WHERE date = ? AND billing_target = ?
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, date, billingTarget).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count trips: %w", err)
	}

	return count, nil
}

// ===== InvoiceRepository 実装 =====

// CreateInvoice は、新しい請求書を invoices テーブルに登録します。
func (r *SQLiteRepository) CreateInvoice(ctx context.Context, invoice *models.Invoice) (int, error) {
	query := `
		INSERT INTO invoices (year_month, billing_target, pdf, generated_at)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		invoice.YearMonth,
		invoice.BillingTarget,
		invoice.PDF,
		time.Now().Format(time.RFC3339),
	)

	if err != nil {
		return 0, fmt.Errorf("failed to insert invoice: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert row id: %w", err)
	}

	return int(id), nil
}

// GetInvoiceByYearMonthAndBillingTarget は、指定年月・精算先の確定済み請求書を取得します。
func (r *SQLiteRepository) GetInvoiceByYearMonthAndBillingTarget(ctx context.Context, year int, month int, billingTarget string) (*models.Invoice, error) {
	yearMonth := fmt.Sprintf("%04d-%02d", year, month)

	query := `
		SELECT id, year_month, billing_target, pdf, generated_at
		FROM invoices
		WHERE year_month = ? AND billing_target = ?
	`

	invoice := &models.Invoice{}

	err := r.db.QueryRowContext(ctx, query, yearMonth, billingTarget).Scan(
		&invoice.ID,
		&invoice.YearMonth,
		&invoice.BillingTarget,
		&invoice.PDF,
		&invoice.GeneratedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invoice not found: yearMonth=%s, billingTarget=%s", yearMonth, billingTarget)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice: %w", err)
	}

	return invoice, nil
}

// UpdateInvoice は、請求書を更新します。
func (r *SQLiteRepository) UpdateInvoice(ctx context.Context, invoice *models.Invoice) error {
	query := `
		UPDATE invoices
		SET pdf = ?, generated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		invoice.PDF,
		time.Now().Format(time.RFC3339),
		invoice.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice not found: id=%d", invoice.ID)
	}

	return nil
}

// DeleteInvoice は、請求書を削除します。
func (r *SQLiteRepository) DeleteInvoice(ctx context.Context, id int) error {
	query := `DELETE FROM invoices WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice not found: id=%d", id)
	}

	return nil
}

// ===== ヘルパー関数 =====

// boolToInt は、Go の bool を SQLite の INTEGER に変換します。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool は、SQLite の INTEGER を Go の bool に変換します。
func intToBool(i int) bool {
	return i != 0
}
