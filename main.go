package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"log"
	"os"

	"codeberg.org/go-pdf/fpdf"
)

//go:embed fonts/Inter_18pt-Regular.ttf
var interRegular []byte

//go:embed fonts/Inter_18pt-Bold.ttf
var interBold []byte

const (
	marginLeft   = 20.0
	marginRight  = 20.0
	marginTop    = 22.0
	marginBottom = 20.0
	pageWidth    = 210.0
	contentWidth = pageWidth - marginLeft - marginRight
)

type Contact struct {
	Label string `json:"label"`
	Value string `json:"value"`
	URL   string `json:"url"`
}

type SkillRow struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type BulletGroup struct {
	Label   string   `json:"label"`
	Bullets []string `json:"bullets"`
}

type Entry struct {
	Company  string        `json:"company"`
	Title    string        `json:"title"`
	Subtitle string        `json:"subtitle"`
	Period   string        `json:"period"`
	Lines    []string      `json:"lines"`
	Groups   []BulletGroup `json:"groups"`
	Bullets  []string      `json:"bullets"`
}

type Section struct {
	Title   string     `json:"title"`
	Text    string     `json:"text"`
	Skills  []SkillRow `json:"skills"`
	Entries []Entry    `json:"entries"`
}

type Sections struct {
	About          *Section `json:"about"`
	Skills         *Section `json:"skills"`
	Experience     *Section `json:"experience"`
	Education      *Section `json:"education"`
	Courses        *Section `json:"courses"`
	Certifications *Section `json:"certifications"`
}

type Resume struct {
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Contacts []Contact `json:"contacts"`
	Sections Sections  `json:"sections"`
}

var currentSectionTitle string

func readResume(filePath string) (*Resume, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var resume Resume
	if err := json.Unmarshal(data, &resume); err != nil {
		return nil, err
	}
	return &resume, nil
}

func generatePDF(resume *Resume, outputPath string) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginLeft, marginTop, marginRight)
	pdf.SetAutoPageBreak(true, marginBottom)

	pdf.AddUTF8FontFromBytes("Inter", "", interRegular)
	pdf.AddUTF8FontFromBytes("Inter", "B", interBold)

	pdf.SetHeaderFunc(func() {
		if currentSectionTitle != "" {
			pdf.SetXY(marginLeft, marginTop)
			drawSectionTitle(pdf, currentSectionTitle)
		}
	})

	pdf.AddPage()

	drawHeader(pdf, resume)

	s := resume.Sections
	if s.About != nil {
		drawSection(pdf, s.About.Title, func() {
			drawTextSection(pdf, s.About.Text)
		})
	}
	if s.Skills != nil {
		drawSection(pdf, s.Skills.Title, func() {
			drawSkillsSection(pdf, s.Skills.Skills)
		})
	}
	if s.Experience != nil {
		drawSection(pdf, s.Experience.Title, func() {
			for _, e := range s.Experience.Entries {
				drawExperienceEntry(pdf, e)
			}
		})
	}
	if s.Education != nil {
		drawSection(pdf, s.Education.Title, func() {
			for _, e := range s.Education.Entries {
				drawEducationEntry(pdf, e)
			}
		})
	}
	if s.Courses != nil {
		drawSection(pdf, s.Courses.Title, func() {
			for _, e := range s.Courses.Entries {
				drawSimpleEntry(pdf, e)
			}
		})
	}
	if s.Certifications != nil {
		drawSection(pdf, s.Certifications.Title, func() {
			for _, e := range s.Certifications.Entries {
				drawSimpleEntry(pdf, e)
			}
		})
	}

	return pdf.OutputFileAndClose(outputPath)
}

func drawHeader(pdf *fpdf.Fpdf, resume *Resume) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "B", 24)
	pdf.CellFormat(contentWidth, 11, resume.Name, "", 1, "C", false, 0, "")

	pdf.Ln(1)
	pdf.SetFont("Inter", "", 11)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(contentWidth, 6, resume.Title, "", 1, "C", false, 0, "")

	if len(resume.Contacts) > 0 {
		pdf.Ln(2)
		pdf.SetTextColor(80, 80, 80)

		const sep = "   |   "
		const fontSize = 9.0

		pdf.SetFont("Inter", "", fontSize)

		totalWidth := 0.0
		for i, c := range resume.Contacts {
			text := c.Value
			if c.Label != "" {
				text = c.Label
			}
			totalWidth += pdf.GetStringWidth(text)
			if i < len(resume.Contacts)-1 {
				totalWidth += pdf.GetStringWidth(sep)
			}
		}

		x := marginLeft + (contentWidth-totalWidth)/2
		y := pdf.GetY()

		for i, c := range resume.Contacts {
			text := c.Value
			if c.Label != "" {
				text = c.Label
			}
			pdf.SetFont("Inter", "U", fontSize)
			w := pdf.GetStringWidth(text)
			pdf.SetXY(x, y)
			pdf.CellFormat(w, 5, text, "", 0, "L", false, 0, c.URL)
			x += w

			if i < len(resume.Contacts)-1 {
				pdf.SetFont("Inter", "", fontSize)
				sepW := pdf.GetStringWidth(sep)
				pdf.SetXY(x, y)
				pdf.CellFormat(sepW, 5, sep, "", 0, "L", false, 0, "")
				x += sepW
			}
		}

		pdf.SetY(y + 5)
	}

	pdf.Ln(6)
}

func drawSectionTitle(pdf *fpdf.Fpdf, title string) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "B", 11)
	pdf.CellFormat(contentWidth, 7, title, "", 1, "L", false, 0, "")

	pdf.Ln(2)
	pdf.SetDrawColor(20, 20, 20)
	pdf.SetLineWidth(0.4)
	y := pdf.GetY()
	pdf.Line(marginLeft, y, marginLeft+contentWidth, y)
	pdf.Ln(5)
}

func drawSection(pdf *fpdf.Fpdf, title string, draw func()) {
	currentSectionTitle = title

	_, pageH := pdf.GetPageSize()
	remaining := pageH - marginBottom - pdf.GetY()
	if remaining < 35 {
		pdf.AddPage()
	} else {
		drawSectionTitle(pdf, title)
	}

	draw()

	currentSectionTitle = ""
	pdf.Ln(3)
}

func drawTextSection(pdf *fpdf.Fpdf, text string) {
	pdf.SetFont("Inter", "", 10)
	pdf.SetTextColor(40, 40, 40)
	pdf.MultiCell(contentWidth, 5.5, text, "", "J", false)
}

func drawSkillsSection(pdf *fpdf.Fpdf, skills []SkillRow) {
	const labelWidth = 52.0

	for _, skill := range skills {
		pdf.SetTextColor(40, 40, 40)

		startY := pdf.GetY()

		pdf.SetFont("Inter", "B", 10)
		pdf.SetLeftMargin(marginLeft)
		pdf.SetX(marginLeft)
		pdf.MultiCell(labelWidth, 5.5, skill.Label+":", "", "L", false)
		labelEndY := pdf.GetY()

		pdf.SetFont("Inter", "", 10)
		pdf.SetXY(marginLeft+labelWidth, startY)
		pdf.SetLeftMargin(marginLeft + labelWidth)
		pdf.MultiCell(contentWidth-labelWidth, 5.5, skill.Value, "", "L", false)
		valueEndY := pdf.GetY()

		pdf.SetLeftMargin(marginLeft)

		maxY := labelEndY
		if valueEndY > maxY {
			maxY = valueEndY
		}
		pdf.SetY(maxY)
		pdf.Ln(2)
	}
}

func drawTitleWithPeriod(pdf *fpdf.Fpdf, title, period string) {
	if period != "" {
		periodWidth := pdf.GetStringWidth(period) + 2
		pdf.CellFormat(contentWidth-periodWidth, 6, title, "", 0, "L", false, 0, "")
		pdf.SetFont("Inter", "", 10)
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(periodWidth, 6, period, "", 1, "R", false, 0, "")
	} else {
		pdf.CellFormat(contentWidth, 6, title, "", 1, "L", false, 0, "")
	}
}

func drawExperienceEntry(pdf *fpdf.Fpdf, entry Entry) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "B", 10)
	drawTitleWithPeriod(pdf, entry.Company, entry.Period)

	pdf.SetFont("Inter", "", 10)
	pdf.SetTextColor(60, 60, 60)
	pdf.CellFormat(contentWidth, 5.5, entry.Title, "", 1, "L", false, 0, "")

	pdf.Ln(3)

	for _, group := range entry.Groups {
		pdf.SetFont("Inter", "B", 10)
		pdf.SetTextColor(40, 40, 40)
		pdf.CellFormat(contentWidth, 5.5, group.Label, "", 1, "L", false, 0, "")
		pdf.Ln(1)

		for _, bullet := range group.Bullets {
			drawBullet(pdf, bullet)
		}

		pdf.Ln(2)
	}

	for _, bullet := range entry.Bullets {
		drawBullet(pdf, bullet)
	}

	pdf.Ln(2)
}

func drawEducationEntry(pdf *fpdf.Fpdf, entry Entry) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "B", 10)
	drawTitleWithPeriod(pdf, entry.Title, entry.Period)

	pdf.SetFont("Inter", "", 10)
	pdf.SetTextColor(60, 60, 60)
	for _, line := range entry.Lines {
		pdf.CellFormat(contentWidth, 5.5, line, "", 1, "L", false, 0, "")
	}

	pdf.Ln(4)
}

func drawSimpleEntry(pdf *fpdf.Fpdf, entry Entry) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "", 10)
	pdf.CellFormat(contentWidth, 5.5, entry.Title, "", 1, "L", false, 0, "")

	if entry.Subtitle != "" {
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(contentWidth, 5.5, entry.Subtitle, "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)
}

func drawBullet(pdf *fpdf.Fpdf, text string) {
	pdf.SetFont("Inter", "", 10)
	pdf.SetTextColor(40, 40, 40)
	const bulletLeft = 4.0
	const indent = 6.0
	pdf.SetX(marginLeft + bulletLeft)
	pdf.CellFormat(indent, 5.5, "\u2022", "", 0, "L", false, 0, "")
	pdf.SetLeftMargin(marginLeft + bulletLeft + indent)
	pdf.SetX(marginLeft + bulletLeft + indent)
	pdf.MultiCell(contentWidth-bulletLeft-indent, 5.5, text, "", "L", false)
	pdf.SetLeftMargin(marginLeft)
}

func main() {
	input := flag.String("i", "resume.json", "input JSON file")
	output := flag.String("o", "resume.pdf", "output PDF file")
	flag.Parse()

	resume, err := readResume(*input)
	if err != nil {
		log.Fatal("Error reading resume: ", err)
	}

	if err := generatePDF(resume, *output); err != nil {
		log.Fatal("Error generating PDF: ", err)
	}

	log.Printf("Resume generated: %s", *output)
}
