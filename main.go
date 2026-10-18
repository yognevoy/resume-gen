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
var fontRegular []byte

//go:embed fonts/Inter_18pt-Bold.ttf
var fontBold []byte

const (
	fontName = "Inter"

	marginLeft   = 20.0
	marginRight  = 20.0
	marginTop    = 22.0
	marginBottom = 20.0
	pageWidth    = 210.0
	contentWidth = pageWidth - marginLeft - marginRight

	nameFontSize     = 24.0
	positionFontSize = 11.0
	titleFontSize    = 11.0
	contactFontSize  = 9.0
	bodyFontSize     = 10.0
	lineHeight       = 5.5
	minSectionSpace  = 35.0

	bulletLeft      = 4.0
	bulletIndent    = 6.0
	bulletTextWidth = contentWidth - bulletLeft - bulletIndent

	photoSize = 30.0
	photoGap  = 8.0

	nameLineHeight     = 11.0
	nameGap            = 1.0
	positionLineHeight = 6.0
	contactsGap        = 2.0
	contactsLineHeight = 5.0
	contactsWrapGap    = 1.5
	contactsSeparator  = "   |   "
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

type Language struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type Project struct {
	Title       string `json:"title"`
	Description string `json:"description"`
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

type ProjectsSection struct {
	Title   string    `json:"title"`
	Entries []Project `json:"entries"`
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

type LanguagesSection struct {
	Title   string     `json:"title"`
	Entries []Language `json:"entries"`
}

type Sections struct {
	About          *AboutSection          `json:"about"`
	Skills         *SkillsSection         `json:"skills"`
	Experience     *ExperienceSection     `json:"experience"`
	Projects       *ProjectsSection       `json:"projects"`
	Education      *EducationSection      `json:"education"`
	Courses        *CoursesSection        `json:"courses"`
	Certifications *CertificationsSection `json:"certifications"`
	Languages      *LanguagesSection      `json:"languages"`
}

type Resume struct {
	Name     string    `json:"name"`
	Position string    `json:"position"`
	Location string    `json:"location"`
	Photo    string    `json:"photo"`
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

	pdf.AddUTF8FontFromBytes(fontName, "", fontRegular)
	pdf.AddUTF8FontFromBytes(fontName, "B", fontBold)

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
	if s.Projects != nil {
		drawSection(pdf, s.Projects.Title, func() {
			for _, p := range s.Projects.Entries {
				drawProject(pdf, p)
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
	if s.Languages != nil {
		drawSection(pdf, s.Languages.Title, func() {
			for _, e := range s.Languages.Entries {
				drawLanguage(pdf, e)
			}
		})
	}

	return pdf.OutputFileAndClose(outputPath)
}

func drawHeader(pdf *fpdf.Fpdf, resume *Resume) {
	textWidth := contentWidth
	textX := marginLeft
	if resume.Photo != "" {
		textWidth = contentWidth - 2*(photoSize+photoGap)
		textX = marginLeft + photoSize + photoGap
	}

	pdf.SetFont(fontName, "", contactFontSize)
	contactLines := wrapContacts(pdf, resume.Contacts, textWidth)

	if resume.Photo != "" {
		drawPhoto(pdf, resume.Photo, marginLeft, marginTop, photoSize)

		textBlockHeight := nameLineHeight + nameGap + positionLineHeight
		if resume.Location != "" {
			textBlockHeight += contactsGap + contactsLineHeight
		}
		if len(contactLines) > 0 {
			textBlockHeight += contactsGap + float64(len(contactLines))*contactsLineHeight
			textBlockHeight += float64(len(contactLines)-1) * contactsWrapGap
		}
		if textBlockHeight < photoSize {
			pdf.SetY(marginTop + (photoSize-textBlockHeight)/2)
		}
	}

	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont(fontName, "B", nameFontSize)
	pdf.SetX(textX)
	pdf.CellFormat(textWidth, nameLineHeight, resume.Name, "", 1, "C", false, 0, "")

	pdf.Ln(nameGap)
	pdf.SetFont(fontName, "", positionFontSize)
	pdf.SetTextColor(80, 80, 80)
	pdf.SetX(textX)
	pdf.CellFormat(textWidth, positionLineHeight, resume.Position, "", 1, "C", false, 0, "")

	if resume.Location != "" {
		pdf.Ln(contactsGap)
		drawLocation(pdf, resume.Location, textX, textWidth)
	}

	if len(contactLines) > 0 {
		pdf.Ln(contactsGap)
		drawContacts(pdf, contactLines, textX, textWidth)
	}

	if resume.Photo != "" {
		if photoBottom := marginTop + photoSize; pdf.GetY() < photoBottom {
			pdf.SetY(photoBottom)
		}
	}

	pdf.Ln(6)
}

func wrapContacts(pdf *fpdf.Fpdf, contacts []Contact, maxWidth float64) [][]Contact {
	if len(contacts) == 0 {
		return nil
	}

	sepWidth := pdf.GetStringWidth(contactsSeparator)

	var lines [][]Contact
	var current []Contact
	currentWidth := 0.0

	for _, c := range contacts {
		text := c.Value
		if c.Label != "" {
			text = c.Label
		}

		w := pdf.GetStringWidth(text)
		addWidth := w
		if len(current) > 0 {
			addWidth += sepWidth
		}

		if len(current) > 0 && currentWidth+addWidth > maxWidth {
			lines = append(lines, current)
			current = nil
			currentWidth = 0
			addWidth = w
		}

		current = append(current, c)
		currentWidth += addWidth
	}

	return append(lines, current)
}

func drawLocation(pdf *fpdf.Fpdf, location string, textX, textWidth float64) {
	pdf.SetFont(fontName, "", contactFontSize)
	pdf.SetTextColor(80, 80, 80)
	pdf.SetX(textX)
	pdf.CellFormat(textWidth, contactsLineHeight, location, "", 1, "C", false, 0, "")
}

func drawContacts(pdf *fpdf.Fpdf, lines [][]Contact, textX, textWidth float64) {
	pdf.SetTextColor(80, 80, 80)
	pdf.SetFont(fontName, "", contactFontSize)
	sepWidth := pdf.GetStringWidth(contactsSeparator)

	y := pdf.GetY()

	for idx, line := range lines {
		pdf.SetFont(fontName, "", contactFontSize)
		lineWidth := 0.0
		for i, c := range line {
			text := c.Value
			if c.Label != "" {
				text = c.Label
			}
			lineWidth += pdf.GetStringWidth(text)
			if i < len(line)-1 {
				lineWidth += sepWidth
			}
		}

		x := textX + (textWidth-lineWidth)/2

		for i, c := range line {
			text := c.Value
			if c.Label != "" {
				text = c.Label
			}
			pdf.SetFont(fontName, "U", contactFontSize)
			w := pdf.GetStringWidth(text)
			pdf.SetXY(x, y)
			pdf.CellFormat(w, contactsLineHeight, text, "", 0, "L", false, 0, c.URL)
			x += w

			if i < len(line)-1 {
				pdf.SetFont(fontName, "", contactFontSize)
				pdf.SetXY(x, y)
				pdf.CellFormat(sepWidth, contactsLineHeight, contactsSeparator, "", 0, "L", false, 0, "")
				x += sepWidth
			}
		}

		y += contactsLineHeight
		if idx < len(lines)-1 {
			y += contactsWrapGap
		}
	}

	pdf.SetY(y)
}

func drawPhoto(pdf *fpdf.Fpdf, path string, x, y, diameter float64) {
	if _, err := os.Stat(path); err != nil {
		return
	}

	r := diameter / 2
	pdf.ClipCircle(x+r, y+r, r, false)
	pdf.ImageOptions(path, x, y, diameter, diameter, false, fpdf.ImageOptions{}, 0, "")
	pdf.ClipEnd()
}

func drawSectionTitle(pdf *fpdf.Fpdf, title string) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont(fontName, "B", titleFontSize)
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
	pdf.SetFont(fontName, "", bodyFontSize)
	pdf.SetTextColor(40, 40, 40)
	pdf.MultiCell(contentWidth, lineHeight, text, "", "J", false)
}

func drawSkills(pdf *fpdf.Fpdf, skills []SkillRow) {
	const labelWidth = 52.0

	for _, skill := range skills {
		pdf.SetTextColor(40, 40, 40)

		startY := pdf.GetY()

		pdf.SetFont(fontName, "B", bodyFontSize)
		pdf.SetLeftMargin(marginLeft)
		pdf.SetX(marginLeft)
		pdf.MultiCell(labelWidth, lineHeight, skill.Label+":", "", "L", false)
		labelEndY := pdf.GetY()

		pdf.SetFont(fontName, "", bodyFontSize)
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
		pdf.SetFont(fontName, "", bodyFontSize)
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(periodWidth, 6, period, "", 1, "R", false, 0, "")
	} else {
		pdf.CellFormat(contentWidth, 6, title, "", 1, "L", false, 0, "")
	}
}

func drawJob(pdf *fpdf.Fpdf, job Job) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont(fontName, "B", bodyFontSize)
	drawTitleWithPeriod(pdf, job.Company, job.Period)

	pdf.SetFont(fontName, "", bodyFontSize)
	pdf.SetTextColor(60, 60, 60)
	pdf.CellFormat(contentWidth, lineHeight, job.Title, "", 1, "L", false, 0, "")

	pdf.Ln(3)

	for _, group := range job.Groups {
		labelHeight := lineHeight + 1
		if len(group.Bullets) > 0 {
			pdf.SetFont(fontName, "", bodyFontSize)
			lines := pdf.SplitText(group.Bullets[0], bulletTextWidth)
			labelHeight += float64(len(lines)) * lineHeight
		}
		ensureSpace(pdf, labelHeight)

		pdf.SetFont(fontName, "B", bodyFontSize)
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
	pdf.SetFont(fontName, "B", bodyFontSize)
	drawTitleWithPeriod(pdf, degree.Title, degree.Period)

	pdf.SetFont(fontName, "", bodyFontSize)
	pdf.SetTextColor(60, 60, 60)
	for _, line := range degree.Lines {
		pdf.CellFormat(contentWidth, lineHeight, line, "", 1, "L", false, 0, "")
	}

	pdf.Ln(4)
}

func drawCourse(pdf *fpdf.Fpdf, course Course) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont(fontName, "", bodyFontSize)
	pdf.CellFormat(contentWidth, lineHeight, course.Title, "", 1, "L", false, 0, "")

	if course.Subtitle != "" {
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(contentWidth, lineHeight, course.Subtitle, "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)
}

func drawCertification(pdf *fpdf.Fpdf, cert Certification) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont(fontName, "", bodyFontSize)
	pdf.CellFormat(contentWidth, lineHeight, cert.Title, "", 1, "L", false, 0, "")

	if cert.Subtitle != "" {
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(contentWidth, lineHeight, cert.Subtitle, "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)
}

func drawLanguage(pdf *fpdf.Fpdf, language Language) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont(fontName, "", bodyFontSize)
	pdf.CellFormat(contentWidth, lineHeight, language.Title, "", 1, "L", false, 0, "")

	if language.Subtitle != "" {
		pdf.SetTextColor(60, 60, 60)
		pdf.CellFormat(contentWidth, lineHeight, language.Subtitle, "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)
}

func drawProject(pdf *fpdf.Fpdf, project Project) {
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont(fontName, "", bodyFontSize)
	pdf.CellFormat(contentWidth, lineHeight, project.Title, "", 1, "L", false, 0, "")

	if project.Description != "" {
		pdf.SetTextColor(60, 60, 60)
		pdf.MultiCell(contentWidth, lineHeight, project.Description, "", "L", false)
	}

	pdf.Ln(3)
}

func ensureSpace(pdf *fpdf.Fpdf, height float64) {
	_, pageH := pdf.GetPageSize()
	usableHeight := pageH - marginTop - marginBottom
	remaining := pageH - marginBottom - pdf.GetY()
	if height <= usableHeight && height > remaining {
		pdf.AddPage()
	}
}

func drawBullet(pdf *fpdf.Fpdf, text string) {
	pdf.SetFont(fontName, "", bodyFontSize)
	pdf.SetTextColor(40, 40, 40)

	lines := pdf.SplitText(text, bulletTextWidth)
	ensureSpace(pdf, float64(len(lines))*lineHeight)

	pdf.SetX(marginLeft + bulletLeft)
	pdf.CellFormat(bulletIndent, lineHeight, "\u2022", "", 0, "L", false, 0, "")
	pdf.SetLeftMargin(marginLeft + bulletLeft + bulletIndent)
	pdf.SetX(marginLeft + bulletLeft + bulletIndent)
	pdf.MultiCell(bulletTextWidth, lineHeight, text, "", "L", false)
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
