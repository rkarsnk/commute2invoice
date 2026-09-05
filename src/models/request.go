package models

// CreateFareEvidenceRequest は、区間登録ウィザード Step3 の確認・登録リクエスト型です。
type CreateFareEvidenceRequest struct {
	// Step1 でアップロードした一時的な画像ID
	ImageID string `json:"imageId" binding:"required"`

	// 画像のMIMEタイプ（例："image/png"）
	ImageMime string `json:"imageMime" binding:"required"`

	// Base64 エンコードされた画像バイナリ
	ImageData string `json:"imageData" binding:"required"`

	// 表示名
	Label string `json:"label" binding:"required"`

	// 鉄道会社名
	CompanyText string `json:"companyText" binding:"required"`

	// 乗車駅
	FromText string `json:"fromText" binding:"required"`

	// 降車駅
	ToText string `json:"toText" binding:"required"`

	// 片道運賃
	FareOneway int `json:"fareOneway" binding:"required,gt=0"`

	// 往復運賃（オプション。nil なら FareOneway * 2 を使用）
	FareRoundtrip *int `json:"fareRoundtrip,omitempty"`

	// 有効開始日付
	ValidFrom string `json:"validFrom" binding:"required"`

	// 有効終了日付（省略時は現在も有効）
	ValidUntil *string `json:"validUntil,omitempty"`
}

// UpdateFareEvidenceValidityRequest は、区間の有効終了日を更新するリクエスト型です。
type UpdateFareEvidenceValidityRequest struct {
	ValidUntil *string `json:"validUntil"`
}

// CreateTripRequest は、Trip 追加リクエスト型です。
type CreateTripRequest struct {
	// 参照する FareEvidence の ID
	FareEvidenceID int `json:"fareEvidenceId" binding:"required,gt=0"`

	// 往復利用かどうか
	RoundTrip bool `json:"roundTrip"`

	// 精算先（"自社" または "客先"）
	BillingTarget string `json:"billingTarget" binding:"required,oneof=自社 客先"`
}

// UpdateTripRequest は、Trip 編集リクエスト型です。
type UpdateTripRequest struct {
	// 往復利用かどうか
	RoundTrip bool `json:"roundTrip"`

	// 精算先（"自社" または "客先"）
	BillingTarget string `json:"billingTarget" binding:"required,oneof=自社 客先"`
}

// ConfirmInvoiceRequest は、請求書確定リクエスト型です。
type ConfirmInvoiceRequest struct {
	// 対象年
	Year int `json:"year" binding:"required,gt=0"`

	// 対象月
	Month int `json:"month" binding:"required,min=1,max=12"`

	// 精算先
	BillingTarget string `json:"billingTarget" binding:"required,oneof=自社 客先"`
}
