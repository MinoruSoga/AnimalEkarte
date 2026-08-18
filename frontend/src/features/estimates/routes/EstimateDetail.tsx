import { ICON, C } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
import { formatDate } from "@/lib/format/date";
import { paths } from "@/config/paths";
import { LoadingFallback } from "@/components/shared/DataStates";
import { useNavigate, useParams } from 'react-router';
import { AlertTriangle, FileText, Pencil, Trash2, ArrowLeft, FilePlus2 } from 'lucide-react';
import { useState, useCallback, useTransition } from 'react';
import { Button } from '@/components/ui/button';
import { PageLayout } from '@/components/shared/PageLayout/PageLayout';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog/ConfirmDialog';
import { EstimateStatusBadge } from '../components/EstimateStatusBadge/EstimateStatusBadge';
import { EstimateLineItems } from '../components/EstimateLineItems/EstimateLineItems';
import { useGetEstimate } from '../api/get-estimate';
import { useDeleteEstimate } from '../api/delete-estimate';
import { useCreateEstimateSuccessor } from '../api/create-estimate-successor';
import { usePermission } from "@/hooks/use-permission";
import { ResourceEstimates } from "@/types/generated/models";
import { isEstimateLockedStatus } from "../lib/is-estimate-locked-status";
import {
  shouldOfferEstimateSuccessor,
  isValidSuccessorReason,
} from "../lib/should-offer-estimate-successor";

export function EstimateDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showSuccessorDialog, setShowSuccessorDialog] = useState(false);
  const [successorReason, setSuccessorReason] = useState("");
  const [successorReasonError, setSuccessorReasonError] = useState<string | null>(null);

  const { data: estimate, isLoading, isError } = useGetEstimate(id);
  const { mutate: deleteEstimate, isPending: isDeleting } = useDeleteEstimate();
  const {
    mutate: createSuccessor,
    isPending: isCreatingSuccessor,
  } = useCreateEstimateSuccessor();
  const { canEdit, canDelete, canCreate } = usePermission("estimates");
  const [isDeletePending, startDeleteTransition] = useTransition();
  const [isSuccessorPending, startSuccessorTransition] = useTransition();

  const handleDelete = useCallback(() => {
    if (!id) return;
    startDeleteTransition(() => {
      deleteEstimate(id, {
        onSuccess: () => navigate(paths.estimates.getHref()),
      });
    });
  }, [id, deleteEstimate, navigate]);

  const handleOpenSuccessorDialog = useCallback(() => {
    setSuccessorReason("");
    setSuccessorReasonError(null);
    setShowSuccessorDialog(true);
  }, []);

  const handleCloseSuccessorDialog = useCallback(() => {
    if (isCreatingSuccessor || isSuccessorPending) return;
    setShowSuccessorDialog(false);
    setSuccessorReason("");
    setSuccessorReasonError(null);
  }, [isCreatingSuccessor, isSuccessorPending]);

  const handleCreateSuccessor = useCallback(() => {
    if (!id) return;
    if (!isValidSuccessorReason(successorReason)) {
      setSuccessorReasonError("理由は1〜500文字で入力してください");
      return;
    }
    const trimmed = successorReason.trim();
    setSuccessorReasonError(null);
    startSuccessorTransition(() => {
      createSuccessor(
        { id, reason: trimmed },
        {
          onSuccess: (created) => {
            setShowSuccessorDialog(false);
            setSuccessorReason("");
            navigate(paths.estimates.detail.getHref(created.id));
          },
        },
      );
    });
  }, [id, successorReason, createSuccessor, navigate]);

  if (isLoading) {
    return <LoadingFallback />;
  }
  if (isError || !estimate) {
    return <div className={`p-4 ${C.danger}`}>データの取得に失敗しました</div>;
  }

  const isLocked = isEstimateLockedStatus(estimate.status);
  const isExpired = estimate.validUntil ? estimate.validUntil.slice(0, 10) < todayJSTISO() : false;
  const showEdit = canEdit && !isLocked;
  const showDelete = canDelete && !isLocked;
  const showSuccessor = shouldOfferEstimateSuccessor({
    canCreate,
    status: estimate.status,
  });
  const successorBusy = isCreatingSuccessor || isSuccessorPending;
  const canSubmitSuccessor = isValidSuccessorReason(successorReason);

  return (
    <PageLayout
      title={`見積書 ${estimate.estimateNo}`}
      resource={ResourceEstimates}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(paths.estimates.getHref())}
            className="gap-1.5 text-sm"
          >
            <ArrowLeft className={ICON.action} />
            一覧へ
          </Button>
          {showEdit ? (
            <Button
              variant="outline"
              size="sm"
              onClick={() => id ? navigate(paths.estimates.edit.getHref(id)) : null}
              className="gap-1.5 text-sm"
            >
              <Pencil className={ICON.action} />
              編集
            </Button>
          ) : null}
          {showDelete ? (
            <Button
              variant="ghost-danger"
              size="sm"
              onClick={() => setShowDeleteDialog(true)}
              disabled={isDeleting || isDeletePending}
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
              onClick={handleOpenSuccessorDialog}
              disabled={successorBusy}
              className="gap-1.5 text-sm"
            >
              <FilePlus2 className={ICON.action} />
              後継ドラフトを作成
            </Button>
          ) : null}
        </div>
      }
      maxWidth="max-w-4xl"
    >
      <div className="space-y-6">
        {/* 基本情報 */}
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

        {/* 見積明細 */}
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

        {/* 備考 */}
        {estimate.notes ? (
          <div className={`${C.bgWhite} border ${C.borderLight} rounded-md p-6`}>
            <h3 className={`text-sm font-medium ${C.text} mb-2`}>備考</h3>
            <p className={`text-sm ${C.text70} whitespace-pre-wrap`}>{estimate.notes}</p>
          </div>
        ) : null}
      </div>

      <ConfirmDialog
        open={showDeleteDialog}
        onClose={() => setShowDeleteDialog(false)}
        onConfirm={handleDelete}
        title="見積書を削除しますか?"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
        isPending={isDeletePending}
      />

      {showSuccessorDialog ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="successor-dialog-title"
        >
          <button
            type="button"
            className="absolute inset-0 bg-black/40"
            aria-label="閉じる"
            onClick={handleCloseSuccessorDialog}
            disabled={successorBusy}
          />
          <div
            className={`relative w-full max-w-md rounded-md border ${C.borderLight} ${C.bgWhite} p-6 shadow-lg space-y-4`}
          >
            <h2
              id="successor-dialog-title"
              className={`text-base font-semibold ${C.text}`}
            >
              後継ドラフトを作成
            </h2>
            <p className={`text-sm ${C.text70}`}>
              確定済み見積は変更できません。訂正理由を入力して後継ドラフトを作成します。元の見積は変更されません。
            </p>
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
                onChange={(e) => {
                  setSuccessorReason(e.target.value);
                  setSuccessorReasonError(null);
                }}
                rows={4}
                className={`w-full rounded-md border ${C.borderLight} p-2 text-sm ${C.text}`}
                placeholder="訂正理由を入力"
                disabled={successorBusy}
              />
              {successorReasonError ? (
                <p className={`mt-1 text-sm ${C.danger}`}>{successorReasonError}</p>
              ) : null}
            </div>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={handleCloseSuccessorDialog}
                disabled={successorBusy}
              >
                キャンセル
              </Button>
              <Button
                size="sm"
                onClick={handleCreateSuccessor}
                disabled={successorBusy || !canSubmitSuccessor}
              >
                作成
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </PageLayout>
  );
}
