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

	nameFontSize    = 24.0
	titleFontSize   = 11.0
	contactFontSize = 9.0
	bodyFontSize    = 10.0
	lineHeight      = 5.5
	minSectionSpace = 35.0
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

type Job struct {
	Company string        `json:"company"`
	Title   string        `json:"title"`
	Period  string        `json:"period"`
	Groups  []BulletGroup `json:"groups"`
	Bullets []string      `json:"bullets"`
}

type Degree struct {
	Title  string   `json:"title"`
	Period string   `json:"period"`
	Lines  []string `json:"lines"`
}

type Course struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type Certification struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type AboutSection struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type SkillsSection struct {
	Title  string     `json:"title"`
	Skills []SkillRow `json:"skills"`
}

type ExperienceSection struct {
	Title   string `json:"title"`
	Entries []Job  `json:"entries"`
}

type EducationSection struct {
	Title   string   `json:"title"`
	Entries []Degree `json:"entries"`
}

type CoursesSection struct {
	Title   string   `json:"title"`
	Entries []Course `json:"entries"`
}

type CertificationsSection struct {
	Title   string          `json:"title"`
	Entries []Certification `json:"entries"`
}

type Sections struct {
	About          *AboutSection          `json:"about"`
	Skills         *SkillsSection         `json:"skills"`
	Experience     *ExperienceSection     `json:"experience"`
	Education      *EducationSection      `json:"education"`
	Courses        *CoursesSection        `json:"courses"`
	Certifications *CertificationsSection `json:"certifications"`
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
			drawAbout(pdf, s.About.Text)
		})
	}
	if s.Skills != nil {
		drawSection(pdf, s.Skills.Title, func() {
			drawSkills(pdf, s.Skills.Skills)
		})
	}
	if s.Experience != nil {
		drawSection(pdf, s.Experience.Title, func() {
			for _, e := range s.Experience.Entries {
				drawJob(pdf, e)
			}
		})
	}
	if s.Education != nil {
		drawSection(pdf, s.Education.Title, func() {
			for _, e := range s.Education.Entries {
				drawDegree(pdf, e)
			}
		})
	}
	if s.Courses != nil {
		drawSection(pdf, s.Courses.Title, func() {
			for _, e := range s.Courses.Entries {
				drawCourse(pdf, e)
			}
		})
	}
	if s.Certifications != nil {
		drawSection(pdf, s.Certifications.Title, func() {
			for _, e := range s.Certifications.Entries {
				drawCertification(pdf, e)
			}
		})
	}

	return pdf.OutputFileAndClose(outputPath)
}

func drawHeader(pdf *fpdf.Fpdf, resume *Resume) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "B", nameFontSize)
	pdf.CellFormat(contentWidth, 11, resume.Name, "", 1, "C", false, 0, "")

	pdf.Ln(1)
	pdf.SetFont("Inter", "", titleFontSize)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(contentWidth, 6, resume.Title, "", 1, "C", false, 0, "")

	if len(resume.Contacts) > 0 {
		pdf.Ln(2)
		pdf.SetTextColor(80, 80, 80)

		const sep = "   |   "
		pdf.SetFont("Inter", "", contactFontSize)

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
			pdf.SetFont("Inter", "U", contactFontSize)
			w := pdf.GetStringWidth(text)
			pdf.SetXY(x, y)
			pdf.CellFormat(w, 5, text, "", 0, "L", false, 0, c.URL)
			x += w

			if i < len(resume.Contacts)-1 {
				pdf.SetFont("Inter", "", contactFontSize)
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
	pdf.SetFont("Inter", "B", titleFontSize)
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
	if remaining < minSectionSpace {
		pdf.AddPage()
	} else {
		drawSectionTitle(pdf, title)
	}

	draw()

	currentSectionTitle = ""
	pdf.Ln(3)
}

func drawAbout(pdf *fpdf.Fpdf, text string) {
	pdf.SetFont("Inter", "", bodyFontSize)
	pdf.SetTextColor(40, 40, 40)
	pdf.MultiCell(contentWidth, lineHeight, text, "", "J", false)
}

func drawSkills(pdf *fpdf.Fpdf, skills []SkillRow) {
	const labelWidth = 52.0

	for _, skill := range skills {
		pdf.SetTextColor(40, 40, 40)

		startY := pdf.GetY()

		pdf.SetFont("Inter", "B", bodyFontSize)
		pdf.SetLeftMargin(marginLeft)
		pdf.SetX(marginLeft)
		pdf.MultiCell(labelWidth, lineHeight, skill.Label+":", "", "L", false)
		labelEndY := pdf.GetY()

		pdf.SetFont("Inter", "", bodyFontSize)
		pdf.SetXY(marginLeft+labelWidth, startY)
		pdf.SetLeftMargin(marginLeft + labelWidth)
		pdf.MultiCell(contentWidth-labelWidth, lineHeight, skill.Value, "", "L", false)
		valueEndY := pdf.GetY()

		pdf.SetLeftMargin(marginLeft)

		pdf.SetY(max(labelEndY, valueEndY))
		pdf.Ln(2)
	}
}

func drawTitleWithPeriod(pdf *fpdf.Fpdf, title, period string) {
	if period != "" {
		periodWidth := pdf.GetStringWidth(period) + 2
		pdf.CellFormat(contentWidth-periodWidth, 6, title, "", 0, "L", false, 0, "")
		pdf.SetFont("Inter", "", bodyFontSize)
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(periodWidth, 6, period, "", 1, "R", false, 0, "")
	} else {
		pdf.CellFormat(contentWidth, 6, title, "", 1, "L", false, 0, "")
	}
}

func drawJob(pdf *fpdf.Fpdf, job Job) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "B", bodyFontSize)
	drawTitleWithPeriod(pdf, job.Company, job.Period)

	pdf.SetFont("Inter", "", bodyFontSize)
	pdf.SetTextColor(60, 60, 60)
	pdf.CellFormat(contentWidth, lineHeight, job.Title, "", 1, "L", false, 0, "")

	pdf.Ln(3)

	for _, group := range job.Groups {
		pdf.SetFont("Inter", "B", bodyFontSize)
		pdf.SetTextColor(40, 40, 40)
		pdf.CellFormat(contentWidth, lineHeight, group.Label, "", 1, "L", false, 0, "")
		pdf.Ln(1)

		for _, bullet := range group.Bullets {
			drawBullet(pdf, bullet)
		}

		pdf.Ln(2)
	}

	for _, bullet := range job.Bullets {
		drawBullet(pdf, bullet)
	}

	pdf.Ln(2)
}

func drawDegree(pdf *fpdf.Fpdf, degree Degree) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "B", bodyFontSize)
	drawTitleWithPeriod(pdf, degree.Title, degree.Period)

	pdf.SetFont("Inter", "", bodyFontSize)
	pdf.SetTextColor(60, 60, 60)
	for _, line := range degree.Lines {
		pdf.CellFormat(contentWidth, lineHeight, line, "", 1, "L", false, 0, "")
	}

	pdf.Ln(4)
}

func drawCourse(pdf *fpdf.Fpdf, course Course) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "", bodyFontSize)
	pdf.CellFormat(contentWidth, lineHeight, course.Title, "", 1, "L", false, 0, "")

	if course.Subtitle != "" {
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(contentWidth, lineHeight, course.Subtitle, "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)
}

func drawCertification(pdf *fpdf.Fpdf, cert Certification) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("Inter", "", bodyFontSize)
	pdf.CellFormat(contentWidth, lineHeight, cert.Title, "", 1, "L", false, 0, "")

	if cert.Subtitle != "" {
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(contentWidth, lineHeight, cert.Subtitle, "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)
}

func drawBullet(pdf *fpdf.Fpdf, text string) {
	pdf.SetFont("Inter", "", bodyFontSize)
	pdf.SetTextColor(40, 40, 40)
	const bulletLeft = 4.0
	const indent = 6.0
	pdf.SetX(marginLeft + bulletLeft)
	pdf.CellFormat(indent, lineHeight, "\u2022", "", 0, "L", false, 0, "")
	pdf.SetLeftMargin(marginLeft + bulletLeft + indent)
	pdf.SetX(marginLeft + bulletLeft + indent)
	pdf.MultiCell(contentWidth-bulletLeft-indent, lineHeight, text, "", "L", false)
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
