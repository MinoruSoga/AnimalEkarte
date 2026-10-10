package lstep

import "github.com/animal-ekarte/backend/internal/model"

// This marker is written only when revalidation failed before the send callback
// started. Failures after an API attempt remain blocking: delivery may have
// succeeded even when its response or the subsequent log write failed.
const preSendRevalidationFailure = "pre_send_revalidation_failed"

const blockingDeliveryClaimPredicate = `NOT (
	status = 'failed' AND COALESCE(excluded_reason, '') = 'pre_send_revalidation_failed'
	AND fired_at IS NULL AND suppressed_by_priority = FALSE
)`

func retryablePreSendClaim(log *model.LstepDeliveryTriggerLog) bool {
	return log.Status == model.TriggerStatusFailed && log.FiredAt == nil &&
		!log.SuppressedByPriority && log.ExcludedReason != nil &&
		*log.ExcludedReason == preSendRevalidationFailure
}
