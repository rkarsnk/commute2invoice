// Package models は、DBエンティティと API リクエスト/レスポンスのデータ型定義を含みます。
// Go初心者向けの構造体定義パターンを示しています。
package models

// FareEvidence は、運賃区間の証跡情報を表します。
// DBの fare_evidence テーブルと 1:1 対応します。
type FareEvidence struct {
	// プライマリキー
	ID int `json:"id"`

	// 表示名（例："上新庄→神戸三宮"）
	Label string `json:"label"`

	// 鉄道会社名（支払先）
	CompanyText string `json:"companyText"`

	// 乗車駅
	FromText string `json:"fromText"`

	// 降車駅
	ToText string `json:"toText"`

	// 片道運賃（円）
	FareOneway int `json:"fareOneway"`

	// 往復運賃（nil なら FareOneway * 2 を使用）
	// Go では nullable 型は *int で表現します
	FareRoundtrip *int `json:"fareRoundtrip,omitempty"`

	// 証跡画像バイナリ
	Image []byte `json:"-"` // JSONレスポンスでは送らない（フロントで別途取得）

	// 画像の MIME type（例："image/png"）
	ImageMime string `json:"imageMime"`

	// 有効開始日付（'YYYY-MM-DD' 形式）
	ValidFrom string `json:"validFrom"`

	// 有効終了日付（nil なら現在も有効）
	ValidUntil *string `json:"validUntil,omitempty"`

	// レコード作成日時
	CreatedAt string `json:"createdAt"`
}

// GetRoundtripFare は、実際の往復運賃を返すヘルパーメソッドです。
// 往復運賃が明示的に指定されていない場合、片道×2 を返します。
func (f *FareEvidence) GetRoundtripFare() int {
	if f.FareRoundtrip != nil {
		return *f.FareRoundtrip
	}
	return f.FareOneway * 2
}
