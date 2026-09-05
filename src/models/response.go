package models

// CreateFareEvidenceResponse は、区間登録完了レスポンス型です。
type CreateFareEvidenceResponse struct {
	// 作成された FareEvidence の ID
	ID int `json:"id"`

	// 作成日時
	CreatedAt string `json:"createdAt"`
}

// CreateTripResponse は、Trip 追加完了レスポンス型です。
type CreateTripResponse struct {
	// 作成された Trip の ID
	ID int `json:"id"`

	// 作成日時
	CreatedAt string `json:"createdAt"`
}

// UpdateTripResponse は、Trip 編集完了レスポンス型です。
type UpdateTripResponse struct {
	// 編集された Trip の ID
	ID int `json:"id"`

	// 更新日時
	UpdatedAt string `json:"updatedAt"`
}

// DeleteTripResponse は、Trip 削除完了レスポンス型です。
type DeleteTripResponse struct {
	// 削除された Trip の ID
	ID int `json:"id"`

	// 削除完了フラグ
	Deleted bool `json:"deleted"`
}

// CalendarDayInfo は、カレンダー表示用の1日の情報です。
type CalendarDayInfo struct {
	// 日付（'YYYY-MM-DD'）
	Date string `json:"date"`

	// その日の Trip 件数
	TripCount int `json:"tripCount"`

	// 請求書確定状態（キー："自社"/"客先"）
	InvoiceStatus map[string]bool `json:"invoiceStatus"`
}

// CalendarResponse は、月間カレンダー取得レスポンス型です。
type CalendarResponse struct {
	// 対象年
	Year int `json:"year"`

	// 対象月
	Month int `json:"month"`

	// 各日付の情報
	Days []CalendarDayInfo `json:"days"`
}

// DailyTripsResponse は、指定日の Trip 一覧レスポンス型です。
type DailyTripsResponse struct {
	// 対象日付
	Date string `json:"date"`

	// その日の Trip 一覧
	Trips []TripDetail `json:"trips"`
}

// TripDetail は、Trip の詳細情報（FareEvidence を含む）です。
type TripDetail struct {
	// Trip 情報
	Trip *Trip `json:"trip"`

	// 参照先の FareEvidence 情報
	FareEvidence *FareEvidence `json:"fareEvidence"`
}

// InvoiceLineItem は、請求書の1行分のデータです。
type InvoiceLineItem struct {
	// 利用日
	Date string `json:"date"`

	// 鉄道会社
	CompanyText string `json:"companyText"`

	// 乗車駅
	FromText string `json:"fromText"`

	// 降車駅
	ToText string `json:"toText"`

	// 片道運賃
	FareOneway int `json:"fareOneway"`

	// 実際の交通費（往復判定後）
	Fare int `json:"fare"`
}

// InvoicePreviewResponse は、請求書プレビューレスポンス型です。
type InvoicePreviewResponse struct {
	// 対象年
	Year int `json:"year"`

	// 対象月
	Month int `json:"month"`

	// 精算先
	BillingTarget string `json:"billingTarget"`

	// 状態（"draft" = 未確定、"confirmed" = 確定済み）
	Status string `json:"status"`

	// 確定日時（Status が "confirmed" の場合のみ）
	ConfirmedAt *string `json:"confirmedAt,omitempty"`

	// 請求書の行データ
	InvoiceLines []InvoiceLineItem `json:"invoiceLines"`

	// 合計交通費
	TotalFare int `json:"totalFare"`

	// 使用している証跡画像一覧
	EvidenceImages []EvidenceImageInfo `json:"evidenceImages"`
}

// EvidenceImageInfo は、請求書に添付する証跡画像情報です。
type EvidenceImageInfo struct {
	// FareEvidence の ID
	FareEvidenceID int `json:"fareEvidenceId"`

	// 画像のMIMEタイプ
	MimeType string `json:"mimeType"`

	// Base64 エンコードされた画像バイナリ
	ImageBase64 string `json:"imageBase64"`
}

// ConfirmInvoiceResponse は、請求書確定レスポンス型です。
type ConfirmInvoiceResponse struct {
	// 作成された Invoice の ID
	ID int `json:"id"`

	// 確定日時
	ConfirmedAt string `json:"confirmedAt"`

	// PDF ダウンロード URL
	PDFURL string `json:"pdfUrl"`
}

// ErrorResponse は、エラーレスポンス型です。
type ErrorResponse struct {
	// エラーメッセージ
	Error string `json:"error"`

	// エラーコード
	Code string `json:"code"`

	// 追加の詳細情報（オプション）
	Details map[string]interface{} `json:"details,omitempty"`
}
