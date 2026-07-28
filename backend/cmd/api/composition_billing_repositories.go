package main

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

type billingRepositories struct {
	accounting           billing.AccountingRepository
	billingConfirmations billing.BillingConfirmationRepository
	billingItems         billing.BillingItemRepository
	campaigns            billing.CampaignRepository
	cashRegisterCloses   billing.CashRegisterCloseRepository
	estimates            billing.EstimateRepository
	insurance            billing.InsuranceRepository
	paymentMethods       billing.PaymentMethodMasterRepository
	refunds              billing.RefundRepository
}

func newBillingRepositories(db *gorm.DB) billingRepositories {
	return billingRepositories{
		accounting:           billing.NewAccountingRepository(db),
		billingConfirmations: billing.NewBillingConfirmationRepository(db),
		billingItems:         billing.NewBillingItemRepository(db),
		campaigns:            billing.NewCampaignRepository(db),
		cashRegisterCloses:   billing.NewCashRegisterCloseRepository(db),
		estimates:            billing.NewEstimateRepository(db),
		insurance:            billing.NewInsuranceRepository(db),
		paymentMethods:       billing.NewPaymentMethodMasterRepository(db),
		refunds:              billing.NewRefundRepository(db),
	}
}
