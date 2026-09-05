package models

// Invoice は、確定済みの請求書スナップショットを表します。
// DBの invoices テーブルと 1:1 対応します。
type Invoice struct {
	// プライマリキー
	ID int `json:"id"`

	// 対象年月（'YYYY-MM' 形式）
	YearMonth string `json:"yearMonth"`

	// 精算先（"自社" または "客先"）
	BillingTarget string `json:"billingTarget"`

	// 請求書 PDF バイナリ
	PDF []byte `json:"-"` // JSONレスポンスでは送らない

	// 生成日時
	GeneratedAt string `json:"generatedAt"`
}
