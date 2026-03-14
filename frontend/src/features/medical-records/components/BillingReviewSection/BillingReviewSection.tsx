// React/Framework
import { lazy, Suspense, useState } from "react";

// External
import { CheckCircle, AlertCircle, Clock } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { BADGE, C } from "@/lib/design-tokens";

// Relative
import {
  useBillingReview,
  useConfirmBillingReview,
  useReturnBillingReview,
} from "../../api/billing-review";
import type { BillingReviewStatus } from "../../types";

const ReturnReasonDialog = lazy(() =>
  import("./ReturnReasonDialog").then((m) => ({ default: m.ReturnReasonDialog }))
);

interface BillingReviewSectionProps {
  medicalRecordId: string;
}

const STATUS_LABELS: Record<BillingReviewStatus, string> = {
  pending: "確認待ち",
  confirmed: "確認済み",
  returned: "差し戻し",
};

const STATUS_BADGE_CLASS: Record<BillingReviewStatus, string> = {
  pending: BADGE.yellow,
  confirmed: BADGE.green,
  returned: BADGE.red,
};

type StatusIconComponent = typeof Clock;

const STATUS_ICON: Record<BillingReviewStatus, StatusIconComponent> = {
  pending: Clock,
  confirmed: CheckCircle,
  returned: AlertCircle,
};

export function BillingReviewSection({
  medicalRecordId,
}: BillingReviewSectionProps) {
  const [isReturnDialogOpen, setIsReturnDialogOpen] = useState(false);

  const { data: review, isLoading } = useBillingReview(medicalRecordId);
  const confirmMutation = useConfirmBillingReview(medicalRecordId);
  const returnMutation = useReturnBillingReview(medicalRecordId);

  if (isLoading) {
    return (
      <div
        className={`flex items-center gap-2 px-3 py-1.5 text-sm ${C.text60}`}
      >
        読み込み中...
      </div>
    );
  }

  if (!review) {
    return null;
  }

  const status = review.status as BillingReviewStatus;
  const badgeClass = STATUS_BADGE_CLASS[status];
  const StatusIcon = STATUS_ICON[status];
  const label = STATUS_LABELS[status];
  const isConfirmDisabled =
    status === "confirmed" ||
    confirmMutation.isPending ||
    returnMutation.isPending;
  const isReturnDisabled =
    returnMutation.isPending || confirmMutation.isPending;

  const handleConfirm = () => {
    confirmMutation.mutate(undefined, {
      onSuccess: () => {
        toast.success("会計を確認しました");
      },
    });
  };

  const handleReturnSubmit = (reason: string) => {
    returnMutation.mutate(
      { return_reason: reason },
      { onSuccess: () => setIsReturnDialogOpen(false) }
    );
  };

  return (
    <>
      <div
        className={`flex items-center gap-3 px-4 py-2.5 rounded-lg border ${C.borderMedium} bg-white`}
      >
        <span
          className={`inline-flex items-center gap-1.5 px-2 py-0.5 text-xs font-medium rounded border ${badgeClass}`}
        >
          <StatusIcon className="h-3 w-3" />
          {label}
        </span>

        {review.return_reason ? (
          <span
            className={`text-xs ${C.text60} truncate max-w-[240px]`}
            title={review.return_reason}
          >
            差し戻し理由: {review.return_reason}
          </span>
        ) : null}

        <div className="ml-auto flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={handleConfirm}
            disabled={isConfirmDisabled}
            className="h-7 px-3 text-xs"
          >
            確認
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => setIsReturnDialogOpen(true)}
            disabled={isReturnDisabled}
            className={`h-7 px-3 text-xs ${C.borderDanger} ${C.danger} ${C.hoverBgDanger5}`}
          >
            差し戻し
          </Button>
        </div>
      </div>

      <Suspense fallback={null}>
        {isReturnDialogOpen ? (
          <ReturnReasonDialog
            open={isReturnDialogOpen}
            onOpenChange={setIsReturnDialogOpen}
            onSubmit={handleReturnSubmit}
            isPending={returnMutation.isPending}
          />
        ) : null}
      </Suspense>
    </>
  );
}
