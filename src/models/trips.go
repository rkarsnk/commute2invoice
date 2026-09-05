package models

// Trip は、ユーザーが登録した交通利用実績を表します。
// DBの trips テーブルと 1:1 対応します。
type Trip struct {
	// プライマリキー
	ID int `json:"id"`

	// 利用日付（'YYYY-MM-DD' 形式）
	Date string `json:"date"`

	// 参照する区間の ID（FareEvidence.ID）
	FareEvidenceID int `json:"fareEvidenceId"`

	// 往復利用かどうか（0 = 片道、1 = 往復）
	// JSON ではブール値として扱う
	RoundTrip bool `json:"roundTrip"`

	// 精算先（"自社" または "客先"）
	BillingTarget string `json:"billingTarget"`

	// レコード作成日時
	CreatedAt string `json:"createdAt"`
}

// CalculateFare は、この Trip に対応する交通費を計算するヘルパーメソッドです。
// FareEvidence の往復チェック情報と RoundTrip フラグを組み合わせます。
// 実装例: 往復フラグが true なら GetRoundtripFare() を使用
func (t *Trip) CalculateFare(fare *FareEvidence) int {
	if t.RoundTrip {
		return fare.GetRoundtripFare()
	}
	return fare.FareOneway
}
