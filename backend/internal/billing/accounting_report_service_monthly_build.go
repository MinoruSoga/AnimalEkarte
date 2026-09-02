package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/timeutil"
)

type monthlyDailyAgg struct {
	amNet    int64
	pmNet    int64
	amCount  int64
	pmCount  int64
	refund   int64
	amClosed bool
	pmClosed bool
}

func accumulateMonthlyDailyMap(raw *MonthlyReportResult, byPaymentMethod map[string]int64, payMethodNames map[uint64]string) (map[string]*monthlyDailyAgg, int64) {
	dailyMap := make(map[string]*monthlyDailyAgg)
	var totalAmount int64
	for _, row := range raw.PaymentRows {
		d, ok := dailyMap[row.Date]
		if !ok {
			d = &monthlyDailyAgg{
				amClosed: raw.ClosedAM[row.Date],
				pmClosed: raw.ClosedPM[row.Date],
			}
			dailyMap[row.Date] = d
		}
		d.pmNet += row.Amount
		totalAmount += row.Amount
		pmName := resolvePaymentMethodName(row.PaymentMethodID, payMethodNames)
		byPaymentMethod[pmName] += row.Amount
	}
	for date, count := range raw.DailyBillingCount {
		d, ok := dailyMap[date]
		if !ok {
			d = &monthlyDailyAgg{
				amClosed: raw.ClosedAM[date],
				pmClosed: raw.ClosedPM[date],
			}
			dailyMap[date] = d
		}
		d.pmCount = count
	}
	for date, closed := range raw.ClosedAM {
		if _, ok := dailyMap[date]; !ok {
			dailyMap[date] = &monthlyDailyAgg{}
		}
		if closed {
			dailyMap[date].amClosed = true
		}
	}
	for date, closed := range raw.ClosedPM {
		if _, ok := dailyMap[date]; !ok {
			dailyMap[date] = &monthlyDailyAgg{}
		}
		if closed {
			dailyMap[date].pmClosed = true
		}
	}
	return dailyMap, totalAmount
}

func summarizeMonthlyTax(raw *MonthlyReportResult, taxRates accountingReportTaxRates) TaxBreakdownSummary {
	var taxSummary TaxBreakdownSummary
	for _, tr := range raw.TaxBreakdown {
		if isReducedTaxRate(tr.TaxRate, taxRates) {
			taxSummary.Reduced.TaxableAmount += tr.TaxableAmount
			taxSummary.Reduced.TaxAmount += tr.TaxAmount
			continue
		}
		taxSummary.Standard.TaxableAmount += tr.TaxableAmount
		taxSummary.Standard.TaxAmount += tr.TaxAmount
	}
	return taxSummary
}

func buildMonthlyDailyDetails(startDate, endDate time.Time, dailyMap map[string]*monthlyDailyAgg, holidaySet map[string]bool) (dailyDetails []DailyReportDetail, workingDays int) {
	days := int(endDate.Sub(startDate).Hours()/24) + 1
	if days < 0 {
		days = 0
	}
	dailyDetails = make([]DailyReportDetail, 0, days)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format(time.DateOnly)
		isHoliday := holidaySet[dateStr]
		agg := dailyMap[dateStr]
		detail := DailyReportDetail{
			Date:      dateStr,
			Weekday:   timeutil.WeekdayJP(d),
			IsHoliday: isHoliday,
		}
		if agg != nil {
			detail.AMCount = agg.amCount
			detail.AMNet = agg.amNet
			detail.PMCount = agg.pmCount
			detail.PMNet = agg.pmNet
			detail.DayNet = agg.amNet + agg.pmNet
			detail.Refund = agg.refund
			detail.AMClosed = agg.amClosed
			detail.PMClosed = agg.pmClosed
		}
		if !isHoliday {
			workingDays++
		}
		dailyDetails = append(dailyDetails, detail)
	}
	return dailyDetails, workingDays
}
