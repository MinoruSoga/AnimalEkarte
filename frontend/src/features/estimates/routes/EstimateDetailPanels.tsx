import { AlertTriangle, Pencil, Trash2, ArrowLeft, FilePlus2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { EstimateStatusBadge } from "../components/EstimateStatusBadge/EstimateStatusBadge";
import { EstimateLineItems } from "../components/EstimateLineItems/EstimateLineItems";
import type { Estimate } from "../types";

interface EstimateDetailHeaderActionsProps {
  onBack: () => void;
  showEdit: boolean;
  onEdit: () => void;
  showDelete: boolean;
  onDelete: () => void;
  deleteBusy: boolean;
  showSuccessor: boolean;
  onSuccessor: () => void;
  successorBusy: boolean;
}

export function EstimateDetailHeaderActions({
  onBack,
  showEdit,
  onEdit,
  showDelete,
  onDelete,
  deleteBusy,
  showSuccessor,
  onSuccessor,
  successorBusy,
}: EstimateDetailHeaderActionsProps) {
  return (
    <div className="flex gap-2">
      <Button variant="outline" size="sm" onClick={onBack} className="gap-1.5 text-sm">
        <ArrowLeft className={ICON.action} />
        一覧へ
      </Button>
      {showEdit ? (
        <Button variant="outline" size="sm" onClick={onEdit} className="gap-1.5 text-sm">
          <Pencil className={ICON.action} />
          編集
        </Button>
      ) : null}
      {showDelete ? (
        <Button
          variant="ghost-danger"
          size="sm"
          onClick={onDelete}
          disabled={deleteBusy}
          className={`gap-1.5 text-sm border ${C.borderDanger20}`}
        >
          <Trash2 className={ICON.action} />
          削除
        </Button>
      ) : null}
      {showSuccessor ? (
        <Button
          variant="outline"
          size="sm"
          onClick={onSuccessor}
          disabled={successorBusy}
          className="gap-1.5 text-sm"
        >
          <FilePlus2 className={ICON.action} />
          後継ドラフトを作成
        </Button>
      ) : null}
    </div>
  );
}

interface EstimateDetailInfoProps {
  estimate: Estimate;
  isExpired: boolean;
}

export function EstimateDetailInfo({ estimate, isExpired }: EstimateDetailInfoProps) {
  return (
    <div className="space-y-6">
      <div className={`${C.bgWhite} border ${C.borderLight} rounded-md p-6 space-y-4`}>
        <div className="flex items-center justify-between">
          <h2 className={`text-base font-semibold ${C.text}`}>{estimate.title}</h2>
          <EstimateStatusBadge status={estimate.status} />
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
          <div>
            <dt className={`${C.text50} mb-0.5`}>見積番号</dt>
            <dd className={`font-mono ${C.text}`}>{estimate.estimateNo}</dd>
          </div>
          {estimate.ownerName ? (
            <div>
              <dt className={`${C.text50} mb-0.5`}>飼主名</dt>
              <dd className={C.text}>{estimate.ownerName}</dd>
            </div>
          ) : null}
          {estimate.petName ? (
            <div>
              <dt className={`${C.text50} mb-0.5`}>ペット名</dt>
              <dd className={C.text}>{estimate.petName}</dd>
            </div>
          ) : null}
          {estimate.validUntil ? (
            <div>
              <dt className={`${C.text50} mb-0.5`}>有効期限</dt>
              <dd className={`flex flex-wrap items-center gap-2 ${C.text}`}>
                <span>{formatDate(estimate.validUntil)}</span>
                {isExpired ? (
                  <span
                    role="status"
                    aria-label="見積期限"
                    className={`inline-flex items-center gap-1 font-medium ${C.danger}`}
                  >
                    <AlertTriangle className="h-4 w-4" aria-hidden="true" />
                    期限切れ
                  </span>
                ) : null}
              </dd>
            </div>
          ) : null}
          <div>
            <dt className={`${C.text50} mb-0.5`}>作成日</dt>
            <dd className={C.text}>{formatDate(estimate.createdAt)}</dd>
          </div>
          <div>
            <dt className={`${C.text50} mb-0.5`}>更新日</dt>
            <dd className={C.text}>{formatDate(estimate.updatedAt)}</dd>
          </div>
        </div>

        {estimate.comment ? (
          <div>
            <dt className={`text-sm ${C.text50} mb-0.5`}>コメント</dt>
            <dd className={`text-sm ${C.text} whitespace-pre-wrap`}>{estimate.comment}</dd>
          </div>
        ) : null}
      </div>

      <div className={`${C.bgWhite} border ${C.borderLight} rounded-md p-6`}>
        <h3 className={`text-sm font-medium ${C.text} mb-4`}>見積明細</h3>
        <EstimateLineItems
          items={estimate.items}
          subtotal={estimate.subtotal}
          taxTotal={estimate.taxTotal}
          insuranceAmount={estimate.insuranceAmount}
          discountAmount={estimate.discountAmount}
          totalAmount={estimate.totalAmount}
        />
      </div>

      {estimate.notes ? (
        <div className={`${C.bgWhite} border ${C.borderLight} rounded-md p-6`}>
          <h3 className={`text-sm font-medium ${C.text} mb-2`}>備考</h3>
          <p className={`text-sm ${C.text70} whitespace-pre-wrap`}>{estimate.notes}</p>
        </div>
      ) : null}
    </div>
  );
}

interface EstimateSuccessorDialogProps {
  successorReason: string;
  successorReasonError: string | null;
  successorBusy: boolean;
  canSubmitSuccessor: boolean;
  onReasonChange: (value: string) => void;
  onClose: () => void;
  formAction: (formData: FormData) => void;
}

export function EstimateSuccessorDialog({
  successorReason,
  successorReasonError,
  successorBusy,
  canSubmitSuccessor,
  onReasonChange,
  onClose,
  formAction,
}: EstimateSuccessorDialogProps) {
  return (
    <Dialog
      open
      onOpenChange={(next) => {
        if (!next && !successorBusy) onClose();
      }}
    >
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>後継ドラフトを作成</DialogTitle>
          <DialogDescription>
            確定済み見積は変更できません。訂正理由を入力して後継ドラフトを作成します。元の見積は変更されません。
          </DialogDescription>
        </DialogHeader>
        <form action={formAction} noValidate>
          <div>
            <label
              htmlFor="successor-reason"
              className={`block text-sm font-medium ${C.text} mb-1`}
            >
              理由（必須）
            </label>
            <textarea
              id="successor-reason"
              value={successorReason}
              onChange={(e) => onReasonChange(e.target.value)}
              rows={4}
              className={`w-full rounded-md border ${C.borderLight} p-2 text-sm ${C.text}`}
              placeholder="訂正理由を入力"
              disabled={successorBusy}
            />
            {successorReasonError ? (
              <p className={`mt-1 text-sm ${C.danger}`}>{successorReasonError}</p>
            ) : null}
          </div>
          <DialogFooter className="mt-4">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClose}
              disabled={successorBusy}
            >
              キャンセル
            </Button>
            <SubmitButton size="sm" disabled={successorBusy || !canSubmitSuccessor}>
              作成
            </SubmitButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
