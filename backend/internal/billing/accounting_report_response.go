package billing

// monthlyTaxBreakdownEntryResponse は税率別の課税対象額・税額
type monthlyTaxBreakdownEntryResponse struct {
	TaxableAmount int64 `json:"taxable_amount"`
	TaxAmount     int64 `json:"tax_amount"`
}

// monthlyTaxBreakdownSummaryResponse は標準・軽減税率別サマリー
type monthlyTaxBreakdownSummaryResponse struct {
	Standard monthlyTaxBreakdownEntryResponse `json:"standard"`
	Reduced  monthlyTaxBreakdownEntryResponse `json:"reduced"`
}

// monthlyReportSummaryResponse は月次サマリー情報
type monthlyReportSummaryResponse struct {
	WorkingDays     int                                `json:"working_days"`
	TotalBillings   int64                              `json:"total_billings"`
	TotalAmount     int64                              `json:"total_amount"`
	TotalRefund     int64                              `json:"total_refund"`
	NetAmount       int64                              `json:"net_amount"`
	ByPaymentMethod map[string]int64                   `json:"by_payment_method"`
	ByCategory      map[string]int64                   `json:"by_category"`
	TaxBreakdown    monthlyTaxBreakdownSummaryResponse `json:"tax_breakdown"`
}

// dailyReportDetailResponse は日別明細
type dailyReportDetailResponse struct {
	Date      string `json:"date"`
	Weekday   string `json:"weekday"`
	AMCount   int64  `json:"am_count"`
	AMNet     int64  `json:"am_net"`
	PMCount   int64  `json:"pm_count"`
	PMNet     int64  `json:"pm_net"`
	DayNet    int64  `json:"day_net"`
	Refund    int64  `json:"refund"`
	AMClosed  bool   `json:"am_closed"`
	PMClosed  bool   `json:"pm_closed"`
	IsHoliday bool   `json:"is_holiday"`
}

// monthlyReportResponse は月次売上レポートのハンドラー向けレスポンス
type monthlyReportResponse struct {
	Year         int                          `json:"year"`
	Month        int                          `json:"month"`
	StartDate    string                       `json:"start_date"`
	EndDate      string                       `json:"end_date"`
	Summary      monthlyReportSummaryResponse `json:"summary"`
	DailyDetails []dailyReportDetailResponse  `json:"daily_details"`
}

func toMonthlyReportResponse(r *MonthlyReportResponse) monthlyReportResponse {
	dailyDetails := make([]dailyReportDetailResponse, 0, len(r.DailyDetails))
	for _, d := range r.DailyDetails {
		dailyDetails = append(dailyDetails, dailyReportDetailResponse(d))
	}
	return monthlyReportResponse{
		Year:      r.Year,
		Month:     r.Month,
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
		Summary: monthlyReportSummaryResponse{
			WorkingDays:     r.Summary.WorkingDays,
			TotalBillings:   r.Summary.TotalBillings,
			TotalAmount:     r.Summary.TotalAmount,
			TotalRefund:     r.Summary.TotalRefund,
			NetAmount:       r.Summary.NetAmount,
			ByPaymentMethod: r.Summary.ByPaymentMethod,
			ByCategory:      r.Summary.ByCategory,
			TaxBreakdown: monthlyTaxBreakdownSummaryResponse{
				Standard: monthlyTaxBreakdownEntryResponse{
					TaxableAmount: r.Summary.TaxBreakdown.Standard.TaxableAmount,
					TaxAmount:     r.Summary.TaxBreakdown.Standard.TaxAmount,
				},
				Reduced: monthlyTaxBreakdownEntryResponse{
					TaxableAmount: r.Summary.TaxBreakdown.Reduced.TaxableAmount,
					TaxAmount:     r.Summary.TaxBreakdown.Reduced.TaxAmount,
				},
			},
		},
		DailyDetails: dailyDetails,
	}
}
