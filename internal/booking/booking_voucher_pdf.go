package booking

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"safari-quest-api/pkg/geo"

	"github.com/go-pdf/fpdf"
)

// ErrNoVoucher: only confirmed (or completed) bookings get a confirmation voucher.
var ErrNoVoucher = errors.New("The voucher is available once the company confirms the booking")

// Brand colours (match the website).
var (
	brand = [3]int{113, 55, 0}    // burnt orange-brown
	ink   = [3]int{28, 27, 24}    // body text
	muted = [3]int{110, 104, 96}  // labels
	rule  = [3]int{221, 214, 204} // lines
	soft  = [3]int{247, 243, 237} // table header fill
)

const (
	pageW   = 210.0
	margin  = 16.0
	content = pageW - 2*margin
)

type voucherPDF struct {
	pdf *fpdf.Fpdf
	tr  func(string) string
}

func (v voucherPDF) color(c [3]int) { v.pdf.SetTextColor(c[0], c[1], c[2]) }

func (v voucherPDF) text(size float64, style string, c [3]int) {
	v.pdf.SetFont("Helvetica", style, size)
	v.color(c)
}

// heading writes a small uppercase section title with a rule under it.
func (v voucherPDF) heading(title string) {
	v.pdf.Ln(5)
	v.text(8.5, "B", muted)
	v.pdf.CellFormat(content, 5, v.tr(strings.ToUpper(title)), "", 1, "L", false, 0, "")
	v.pdf.SetDrawColor(rule[0], rule[1], rule[2])
	y := v.pdf.GetY()
	v.pdf.Line(margin, y, margin+content, y)
	v.pdf.Ln(2)
}

// facts writes label/value pairs in two columns.
func (v voucherPDF) facts(pairs [][2]string) {
	col := content / 2
	for i := 0; i < len(pairs); i += 2 {
		y := v.pdf.GetY()
		h := 0.0
		for j := 0; j < 2 && i+j < len(pairs); j++ {
			x := margin + float64(j)*col
			v.pdf.SetXY(x, y)
			v.text(8, "", muted)
			v.pdf.CellFormat(col-4, 4.5, v.tr(pairs[i+j][0]), "", 2, "L", false, 0, "")
			v.text(10.5, "", ink)
			v.pdf.SetX(x)
			v.pdf.MultiCell(col-4, 5, v.tr(pairs[i+j][1]), "", "L", false)
			if d := v.pdf.GetY() - y; d > h {
				h = d
			}
		}
		v.pdf.SetXY(margin, y+h+2)
	}
}

// table writes rows with wrapped cells; the first row is the header.
func (v voucherPDF) table(widths []float64, rows [][]string) {
	const line = 5.0
	for r, row := range rows {
		header := r == 0
		style, size := "", 9.5
		if header {
			style, size = "B", 8.5
		}
		v.pdf.SetFont("Helvetica", style, size)
		lines := 1
		for i, cell := range row {
			// SplitLines measures the cp1252 bytes the PDF actually prints.
			if n := len(v.pdf.SplitLines([]byte(v.tr(cell)), widths[i]-3)); n > lines {
				lines = n
			}
		}
		h := float64(lines)*line + 2
		if v.pdf.GetY()+h > 297-22 {
			v.pdf.AddPage()
		}
		y := v.pdf.GetY()
		x := margin
		if header {
			v.pdf.SetFillColor(soft[0], soft[1], soft[2])
			v.pdf.Rect(margin, y, content, h, "F")
			v.color(muted)
		} else {
			v.color(ink)
		}
		for i, cell := range row {
			v.pdf.SetXY(x+1.5, y+1)
			v.pdf.MultiCell(widths[i]-3, line, v.tr(cell), "", "L", false)
			x += widths[i]
		}
		v.pdf.SetDrawColor(rule[0], rule[1], rule[2])
		v.pdf.Line(margin, y+h, margin+content, y+h)
		v.pdf.SetXY(margin, y+h)
	}
}

func (v voucherPDF) paragraph(s string) {
	v.text(10, "", ink)
	v.pdf.MultiCell(content, 5, v.tr(s), "", "L", false)
}

func countryName(code string) string {
	if n, ok := geo.Countries[code]; ok {
		return n
	}
	return code
}

func meals(m []string) string {
	if len(m) == 0 {
		return "—"
	}
	out := make([]string, len(m))
	for i, x := range m {
		out[i] = strings.ToUpper(x[:1]) + x[1:]
	}
	return strings.Join(out, ", ")
}

// VoucherPDF renders a print-ready A4 booking confirmation.
// shortcut: core PDF fonts cover Western European characters only (cp1252);
// names in other scripts print as "?". Embed a Unicode TTF if that matters.
func VoucherPDF(d Detail) ([]byte, error) {
	if d.Status != StatusConfirmed && d.Status != StatusCompleted {
		return nil, ErrNoVoucher
	}
	pdf := fpdf.New("P", "mm", "A4", "")
	v := voucherPDF{pdf: pdf, tr: pdf.UnicodeTranslatorFromDescriptor("")}
	pdf.SetTitle("Booking confirmation "+d.Code, true)
	pdf.SetAuthor("SafariQuest", true)
	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(true, 22)
	pdf.AliasNbPages("{nb}")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-14)
		v.text(7.5, "", muted)
		pdf.CellFormat(content/2, 4, v.tr("SafariQuest · Booking "+d.Code), "", 0, "L", false, 0, "")
		pdf.CellFormat(content/2, 4, fmt.Sprintf("Page %d of {nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
	})
	pdf.AddPage()

	// ── Header band ──
	pdf.SetFillColor(brand[0], brand[1], brand[2])
	pdf.Rect(0, 0, pageW, 30, "F")
	pdf.SetXY(margin, 9)
	v.text(18, "B", [3]int{255, 255, 255})
	pdf.CellFormat(content/2, 8, "SafariQuest", "", 0, "L", false, 0, "")
	v.text(11, "", [3]int{255, 240, 225})
	pdf.CellFormat(content/2, 8, v.tr("Booking confirmation"), "", 1, "R", false, 0, "")
	pdf.SetX(margin)
	v.text(8.5, "", [3]int{255, 232, 210})
	pdf.CellFormat(content, 5, v.tr("Ethical safaris and cultural immersion in Tanzania"), "", 1, "L", false, 0, "")

	// ── Reference ──
	pdf.SetXY(margin, 38)
	v.text(8, "", muted)
	pdf.CellFormat(content/2, 4, "REFERENCE", "", 0, "L", false, 0, "")
	pdf.CellFormat(content/2, 4, "STATUS", "", 1, "R", false, 0, "")
	v.text(20, "B", ink)
	pdf.CellFormat(content/2, 10, d.Code, "", 0, "L", false, 0, "")
	v.text(12, "B", [3]int{36, 80, 29})
	pdf.CellFormat(content/2, 10, strings.ToUpper(string(d.Status)), "", 1, "R", false, 0, "")
	issued := "Issued " + time.Now().Format("2 Jan 2006")
	if d.ConfirmedAt != nil {
		issued = "Confirmed " + d.ConfirmedAt.Format("2 Jan 2006, 15:04") + " · " + issued
	}
	v.text(8.5, "", muted)
	pdf.CellFormat(content, 5, v.tr(issued), "", 1, "L", false, 0, "")

	// ── Trip ──
	pdf.Ln(4)
	v.text(15, "B", ink)
	pdf.MultiCell(content, 7, v.tr(d.Trip.Name), "", "L", false)
	where := d.Trip.PackageType
	if d.Trip.Location != nil && *d.Trip.Location != "" {
		where = *d.Trip.Location + " · " + where
	}
	v.text(10, "", muted)
	pdf.CellFormat(content, 5, v.tr(where), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	start, _ := time.Parse("2006-01-02", d.Trip.StartDate)
	end, _ := time.Parse("2006-01-02", d.Trip.EndDate)
	days := int(end.Sub(start).Hours()/24) + 1
	kind := fmt.Sprintf("Shared seats · group %d", d.Group)
	if d.BookingType == TypePrivate {
		kind = fmt.Sprintf("Private group %d", d.Group)
	}
	travellers := fmt.Sprintf("%d", d.Pax)
	if d.Pricing == "per_person" && d.ChargedTravellers > d.Pax {
		travellers += fmt.Sprintf(" (charged for %d)", d.ChargedTravellers)
	}
	pairs := [][2]string{
		{"Dates", start.Format("Mon 2 Jan 2006") + " – " + end.Format("Mon 2 Jan 2006")},
		{"Length", fmt.Sprintf("%d days", days)},
		{"Booking", kind},
		{"Travellers", travellers},
		{"Total price", fmt.Sprintf("%s %s", d.Trip.Currency, money(d.TotalPrice))},
		{"Lead traveller", d.Traveler.Name},
	}
	if d.Trip.MeetingPoint != nil && *d.Trip.MeetingPoint != "" {
		pairs = append(pairs, [2]string{"Meeting point", *d.Trip.MeetingPoint})
	}
	v.facts(pairs)

	// ── Group ──
	if len(d.Travellers) > 0 {
		v.heading("Travelling group")
		rows := [][]string{{"#", "Name", "Age group", "Nationality"}}
		for i, t := range d.Travellers {
			name := t.FullName
			if t.IsLead {
				name += " (lead)"
			}
			rows = append(rows, []string{fmt.Sprint(i + 1), name, t.AgeGroup, countryName(t.Nationality)})
		}
		v.table([]float64{10, 82, 40, content - 132}, rows)
		if len(d.PreferredLanguages) > 0 {
			names := make([]string, len(d.PreferredLanguages))
			for i, l := range d.PreferredLanguages {
				names[i] = geo.Languages[l]
			}
			pdf.Ln(2)
			v.paragraph("Guide languages: " + strings.Join(names, ", "))
		}
	} else {
		v.heading("Travelling group")
		v.paragraph("Traveller details have not been added yet. Add them in My bookings before the trip.")
	}

	// ── Company ──
	v.heading("Your tour company")
	v.text(11, "B", ink)
	pdf.CellFormat(content, 6, v.tr(d.Company.Name), "", 1, "L", false, 0, "")
	contact := d.Company.Phone + "  ·  " + d.Company.Email
	if d.Company.Website != nil && *d.Company.Website != "" {
		contact += "  ·  " + *d.Company.Website
	}
	v.paragraph(contact)

	// ── Itinerary ──
	if len(d.Trip.Itinerary) > 0 {
		v.heading("Itinerary")
		rows := [][]string{{"Day", "Plan", "Overnight", "Meals"}}
		for _, day := range d.Trip.Itinerary {
			stay := "—"
			if day.Accommodation != nil && *day.Accommodation != "" {
				stay = *day.Accommodation
			}
			rows = append(rows, []string{fmt.Sprint(day.Day), day.Title, stay, meals(day.Meals)})
		}
		v.table([]float64{14, 84, 48, content - 146}, rows)
	}

	// ── Included ──
	if len(d.Trip.Inclusions) > 0 || len(d.Trip.Exclusions) > 0 {
		v.heading("What's included")
		if len(d.Trip.Inclusions) > 0 {
			v.paragraph("Included: " + strings.Join(d.Trip.Inclusions, "; ") + ".")
		}
		if len(d.Trip.Exclusions) > 0 {
			v.paragraph("Not included: " + strings.Join(d.Trip.Exclusions, "; ") + ".")
		}
	}

	// ── Note ──
	pdf.Ln(6)
	v.text(8.5, "I", muted)
	pdf.MultiCell(content, 4.5, v.tr("Show this confirmation, printed or on your phone, when you meet "+d.Company.Name+
		". Contact the company about the trip itself and SafariQuest about your booking."), "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// money formats 12345.5 as "12,345.50".
func money(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	whole, frac := s[:len(s)-3], s[len(s)-3:]
	var b strings.Builder
	for i, r := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteRune(',')
		}
		b.WriteRune(r)
	}
	return b.String() + frac
}
