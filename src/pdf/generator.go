// Package pdf は、請求書プレビューからPDFを生成します。
package pdf

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"os/exec"

	"github.com/rkarsnk/commute2invoice/src/models"
)

const invoiceTemplate = `<!doctype html>
<html lang="ja"><head><meta charset="utf-8"><style>
@page { size: A4; margin: 15mm; }
body { font-family: "Noto Sans CJK JP", sans-serif; font-size: 10pt; color: #111; }
h1 { font-size: 20pt; margin-bottom: 4mm; } p { margin: 0 0 2mm; }
table { width: 100%; border-collapse: collapse; margin-top: 6mm; }
th, td { border: 1px solid #444; padding: 2mm; text-align: left; }
th { background: #eee; } .number { text-align: right; }
tfoot th { background: #ddd; } .evidence { page-break-before: always; }
.evidence img { display: block; max-width: 170mm; max-height: 235mm; margin: 4mm 0 8mm; }
</style></head><body>
<h1>交通費精算書</h1>
<p>対象月: {{.Year}}年{{.Month}}月</p><p>精算先: {{.BillingTarget}}</p>
<table><thead><tr><th>利用日</th><th>鉄道会社</th><th>乗車駅</th><th>降車駅</th><th>運賃</th><th>交通費(往復)</th></tr></thead>
<tbody>{{range .InvoiceLines}}<tr><td>{{.Date}}</td><td>{{.CompanyText}}</td><td>{{.FromText}}</td><td>{{.ToText}}</td><td class="number">{{.FareOneway}}円</td><td class="number">{{.Fare}}円</td></tr>{{end}}</tbody>
<tfoot><tr><th colspan="5">合計</th><th class="number">{{.TotalFare}}円</th></tr></tfoot></table>
{{if .EvidenceImages}}<section class="evidence"><h2>運賃証跡</h2>{{range .EvidenceImages}}<h3>証跡 #{{.FareEvidenceID}}</h3><img src="{{imageDataURI .}}" alt="運賃証跡 #{{.FareEvidenceID}}">{{end}}</section>{{end}}
</body></html>`

var parsedInvoiceTemplate = template.Must(template.New("invoice").Funcs(template.FuncMap{
	"imageDataURI": imageDataURI,
}).Parse(invoiceTemplate))

// Generate は、請求書プレビューをA4 PDFバイナリに変換します。
func Generate(ctx context.Context, preview *models.InvoicePreviewResponse) ([]byte, error) {
	if preview == nil {
		return nil, fmt.Errorf("invoice preview is required")
	}

	var html bytes.Buffer
	if err := parsedInvoiceTemplate.Execute(&html, preview); err != nil {
		return nil, fmt.Errorf("failed to render invoice HTML: %w", err)
	}

	output, err := os.CreateTemp("", "commute2invoice-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary PDF: %w", err)
	}
	outputPath := output.Name()
	if err := output.Close(); err != nil {
		os.Remove(outputPath)
		return nil, fmt.Errorf("failed to close temporary PDF: %w", err)
	}
	defer os.Remove(outputPath)

	command := exec.CommandContext(ctx, "wkhtmltopdf",
		"--encoding", "utf-8",
		"--page-size", "A4",
		"--margin-top", "15mm",
		"--margin-right", "15mm",
		"--margin-bottom", "15mm",
		"--margin-left", "15mm",
		"-", outputPath,
	)
	command.Stdin = &html
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w: %s", err, stderr.String())
	}

	document, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}
	if len(document) == 0 {
		return nil, fmt.Errorf("PDF generator produced an empty document")
	}
	return document, nil
}

func imageDataURI(image models.EvidenceImageInfo) (template.URL, error) {
	switch image.MimeType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return template.URL("data:" + image.MimeType + ";base64," + image.ImageBase64), nil
	default:
		return "", fmt.Errorf("unsupported evidence image MIME type: %s", image.MimeType)
	}
}
