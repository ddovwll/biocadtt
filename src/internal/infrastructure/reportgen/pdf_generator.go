package reportgen

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
	"github.com/phpdave11/gofpdf"
)

//go:embed fonts/DejaVuSans.ttf
var dejavuSans []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var dejavuSansBold []byte

type PdfGenerator struct {
	outputDir string
}

type tableColumn struct {
	header string
	align  string
	minW   float64
	maxW   float64
	get    func(models.DeviceData) string
}

func NewService(outputDir string) (*PdfGenerator, error) {
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		return nil, fmt.Errorf("output directory is empty")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	return &PdfGenerator{outputDir: outputDir}, nil
}

func (s *PdfGenerator) Generate(ctx context.Context, unitGUID string, data []models.DeviceData) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	unitGUID = strings.TrimSpace(unitGUID)
	if unitGUID == "" {
		return fmt.Errorf("unitGUID is empty")
	}

	filename := sanitizeFileName(unitGUID) + "_" + time.Now().UTC().Format("20060102_150405") + ".pdf"
	outPath := filepath.Join(s.outputDir, filename)

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(8, 8, 8)
	pdf.SetAutoPageBreak(true, 8)
	pdf.AddUTF8FontFromBytes("DejaVuSans", "", dejavuSans)
	pdf.AddUTF8FontFromBytes("DejaVuSans", "B", dejavuSansBold)

	pdf.AddPage()
	pdf.SetFont("DejaVuSans", "B", 14)
	pdf.CellFormat(0, 8, unitGUID, "", 1, "L", false, 0, "")
	pdf.Ln(1)

	columns := deviceDataColumns()
	widths := calculateColumnWidths(pdf, columns, data)
	headers := make([]string, len(columns))
	aligns := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = c.header
		aligns[i] = c.align
	}

	pdf.SetFont("DejaVuSans", "B", 8)
	drawTableHeader(pdf, headers, widths)

	pdf.SetFont("DejaVuSans", "", 7)
	leftMargin, _, _, bottomMargin := pdf.GetMargins()
	_, pageH := pdf.GetPageSize()
	const lineHeight = 4.0
	const cellPadX = 0.6
	const cellPadY = 0.4

	for _, row := range data {
		if err := ctx.Err(); err != nil {
			return err
		}

		values := make([]string, len(columns))
		for i, c := range columns {
			values[i] = c.get(row)
		}

		linesByCol, rowHeight := splitRow(pdf, values, widths, lineHeight, cellPadX, cellPadY)
		if pdf.GetY()+rowHeight > pageH-bottomMargin {
			pdf.AddPage()
			pdf.SetFont("DejaVuSans", "B", 8)
			drawTableHeader(pdf, headers, widths)
			pdf.SetFont("DejaVuSans", "", 7)
		}

		rowX := leftMargin
		rowY := pdf.GetY()
		for i, lines := range linesByCol {
			pdf.Rect(rowX, rowY, widths[i], rowHeight, "")
			pdf.SetXY(rowX+cellPadX, rowY+cellPadY)
			pdf.MultiCell(widths[i]-2*cellPadX, lineHeight, strings.Join(lines, "\n"), "", aligns[i], false)
			rowX += widths[i]
		}
		pdf.SetXY(leftMargin, rowY+rowHeight)
	}

	if err := pdf.OutputFileAndClose(outPath); err != nil {
		return fmt.Errorf("write pdf: %w", err)
	}

	return nil
}

func drawTableHeader(pdf *gofpdf.Fpdf, headers []string, widths []float64) {
	for i, h := range headers {
		pdf.CellFormat(widths[i], 8, h, "1", 0, "CM", false, 0, "")
	}
	pdf.Ln(-1)
}

func splitRow(
	pdf *gofpdf.Fpdf,
	values []string,
	widths []float64,
	lineHeight float64,
	cellPadX float64,
	cellPadY float64,
) ([][]string, float64) {
	linesByCol := make([][]string, len(values))
	maxLines := 1
	for i, value := range values {
		contentWidth := widths[i] - 2*cellPadX
		if contentWidth < 1 {
			contentWidth = 1
		}
		lines := pdf.SplitText(value, contentWidth)
		if len(lines) == 0 {
			lines = []string{""}
		}
		linesByCol[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	rowHeight := float64(maxLines)*lineHeight + 2*cellPadY
	return linesByCol, rowHeight
}

func calculateColumnWidths(pdf *gofpdf.Fpdf, columns []tableColumn, rows []models.DeviceData) []float64 {
	widths := make([]float64, len(columns))

	pdf.SetFont("DejaVuSans", "B", 8)
	for i, c := range columns {
		widths[i] = clamp(pdf.GetStringWidth(c.header)+3, c.minW, c.maxW)
	}

	pdf.SetFont("DejaVuSans", "", 7)
	for _, row := range rows {
		for i, c := range columns {
			w := pdf.GetStringWidth(c.get(row)) + 3
			if w > widths[i] {
				widths[i] = clamp(w, c.minW, c.maxW)
			}
		}
	}

	scaleToPage(pdf, widths)
	return widths
}

func scaleToPage(pdf *gofpdf.Fpdf, widths []float64) {
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	pageW, _ := pdf.GetPageSize()
	usableW := pageW - leftMargin - rightMargin

	sum := 0.0
	for _, w := range widths {
		sum += w
	}
	if sum <= 0 || sum <= usableW {
		return
	}

	k := usableW / sum
	for i := range widths {
		widths[i] *= k
	}
}

func clamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

var invalidFileNameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)

func sanitizeFileName(name string) string {
	clean := invalidFileNameChars.ReplaceAllString(name, "_")
	clean = strings.TrimSpace(clean)
	clean = strings.Trim(clean, ". ")
	if clean == "" {
		return "report"
	}
	return clean
}

func deviceDataColumns() []tableColumn {
	return []tableColumn{
		{
			header: "N",
			align:  "R",
			minW:   8,
			maxW:   12,
			get: func(d models.DeviceData) string {
				return fmt.Sprintf("%d", d.N)
			},
		},
		{
			header: "MQTT",
			align:  "L",
			minW:   14,
			maxW:   26,
			get: func(d models.DeviceData) string {
				return d.MQTT
			},
		},
		{
			header: "Invid",
			align:  "L",
			minW:   14,
			maxW:   26,
			get: func(d models.DeviceData) string {
				return d.Invid
			},
		},
		{
			header: "UnitGuid",
			align:  "L",
			minW:   20,
			maxW:   32,
			get: func(d models.DeviceData) string {
				return d.UnitGuid
			},
		},
		{
			header: "MsgID",
			align:  "L",
			minW:   14,
			maxW:   24,
			get: func(d models.DeviceData) string {
				return d.MsgID
			},
		},
		{
			header: "Text",
			align:  "L",
			minW:   24,
			maxW:   50,
			get: func(d models.DeviceData) string {
				return d.Text
			},
		},
		{
			header: "Context",
			align:  "L",
			minW:   20,
			maxW:   36,
			get: func(d models.DeviceData) string {
				return d.Context
			},
		},
		{
			header: "Class",
			align:  "L",
			minW:   10,
			maxW:   16,
			get: func(d models.DeviceData) string {
				return d.Class
			},
		},
		{
			header: "Level",
			align:  "R",
			minW:   10,
			maxW:   14,
			get: func(d models.DeviceData) string {
				return fmt.Sprintf("%d", d.Level)
			},
		},
		{
			header: "Area",
			align:  "L",
			minW:   12,
			maxW:   18,
			get: func(d models.DeviceData) string {
				return d.Area
			},
		},
		{
			header: "Addr",
			align:  "L",
			minW:   12,
			maxW:   18,
			get: func(d models.DeviceData) string {
				return d.Addr
			},
		},
		{
			header: "Block",
			align:  "L",
			minW:   12,
			maxW:   18,
			get: func(d models.DeviceData) string {
				return d.Block
			},
		},
		{
			header: "Type",
			align:  "L",
			minW:   12,
			maxW:   18,
			get: func(d models.DeviceData) string {
				return d.Type
			},
		},
		{
			header: "Bit",
			align:  "L",
			minW:   10,
			maxW:   16,
			get: func(d models.DeviceData) string {
				return d.Bit
			},
		},
		{
			header: "InvertBit",
			align:  "L",
			minW:   12,
			maxW:   22,
			get: func(d models.DeviceData) string {
				return d.InvertBit
			},
		},
	}
}
