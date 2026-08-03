package main

import "github.com/animal-ekarte/backend/internal/billing"

type billingServices struct {
	core       billingCoreServices
	documents  billingDocumentServices
	references billingReferenceServices
}

type billingCoreServices struct {
	accounting   billing.AccountingService
	billingItems billing.BillingItemService
	insurance    billing.InsuranceService
	cashRegister billing.CashRegisterService
}

func newBillingCoreServices(
	r billingRepositories,
	d billingCompositionDependencies,
	auditTx billingAuditTxBridge,
) billingCoreServices {
	// billingItems を先に構築し、BUG-013 write-time unbilled guard として accounting へ注入する。
	billingItems := billing.NewBillingItemServiceWithCampaign(
		r.billingItems,
		r.accounting,
		d.Treatments,
		d.Transactor,
		d.TrimmingCourses,
		d.TrimmingOptions,
		r.campaigns,
		d.Owners,
		billing.WithBillingItemAuditTx(auditTx),
		// W-013 HIGH-2: 明細締め後変更も adjustment 台帳へ追記
		billing.WithBillingItemCloseRepository(r.cashRegisterCloses),
	)
	return billingCoreServices{
		accounting: billing.NewAccountingService(
			r.accounting,
			d.MedicalRecords,
			d.Hospitalizations,
			d.Reservations,
			d.TagSync,
			d.Transactor,
			auditTx,
			r.paymentMethods,
			// W-013: 締め後編集の append-only adjustment 台帳書込（justified DI wiring）
			billing.WithCashRegisterCloseRepository(r.cashRegisterCloses),
			// BUG-013: blocking unbilled warning がある pet の会計作成を拒否
			billing.WithUnbilledWriteGuard(billingItems),
			// BUG-018: complete command の ambient-tx 明細・totals collaborator
			billing.WithCompleteItemWriter(billingItems),
			billing.WithCompleteTotalsWriter(billingItems),
		),
		billingItems: billingItems,
		insurance:    billing.NewInsuranceService(r.insurance),
		cashRegister: billing.NewCashRegisterService(
			r.cashRegisterCloses,
			r.accounting,
			d.ClosingSettings,
			r.paymentMethods,
			d.Clinics,
		),
	}
}

type billingDocumentServices struct {
	confirmations billing.BillingConfirmationService
	estimates     billing.EstimateService
	refunds       billing.RefundService
}

func newBillingDocumentServices(
	r billingRepositories,
	d billingCompositionDependencies,
	auditLogger billingAuditBridge,
	auditTx billingAuditTxBridge,
) billingDocumentServices {
	return billingDocumentServices{
		confirmations: billing.NewBillingConfirmationService(
			r.billingConfirmations,
			d.MedicalRecords,
			d.Transactor,
		),
		estimates: billing.NewEstimateService(
			r.estimates,
			d.MedicalRecords,
			d.Reservations,
			d.StaffAssignments,
			auditLogger,
			d.Transactor,
			billing.WithEstimateAuditTx(auditTx),
		),
		refunds: billing.NewRefundService(
			r.refunds,
			r.accounting,
			auditTx,
			d.Transactor,
		),
	}
}

type billingReferenceServices struct {
	campaigns         billing.CampaignService
	paymentMethods    billing.PaymentMethodMasterService
	accountingReports billing.AccountingReportService
}

func newBillingReferenceServices(
	r billingRepositories,
	d billingCompositionDependencies,
) billingReferenceServices {
	return billingReferenceServices{
		campaigns: billing.NewCampaignService(
			r.campaigns,
			d.MerchandiseItems,
			d.Transactor,
		),
		paymentMethods: billing.NewPaymentMethodMasterService(r.paymentMethods),
		accountingReports: billing.NewAccountingReportService(
			r.accounting,
			r.paymentMethods,
			d.ClinicHolidays,
			d.Clinics,
		),
	}
}
