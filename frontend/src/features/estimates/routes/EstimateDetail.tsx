import { ICON, C } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
import { paths } from "@/config/paths";
import { LoadingFallback } from "@/components/shared/DataStates";
import { useNavigate, useParams } from "react-router";
import { FileText } from "lucide-react";
import { useState, useCallback, useActionState, useTransition } from "react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { useGetEstimate } from "../api/get-estimate";
import { useDeleteEstimate } from "../api/delete-estimate";
import { useCreateEstimateSuccessor } from "../api/create-estimate-successor";
import { usePermission } from "@/hooks/use-permission";
import { ResourceEstimates } from "@/types/generated/models";
import { isEstimateLockedStatus } from "../lib/is-estimate-locked-status";
import {
  shouldOfferEstimateSuccessor,
  isValidSuccessorReason,
} from "../lib/should-offer-estimate-successor";
import {
  EstimateDetailHeaderActions,
  EstimateDetailInfo,
  EstimateSuccessorDialog,
} from "./EstimateDetailPanels";

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
    mutateAsync: createSuccessorAsync,
    isPending: isCreatingSuccessor,
  } = useCreateEstimateSuccessor();
  const { canEdit, canDelete, canCreate } = usePermission("estimates");
  const [isDeletePending, startDeleteTransition] = useTransition();

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
    if (isCreatingSuccessor) return;
    setShowSuccessorDialog(false);
    setSuccessorReason("");
    setSuccessorReasonError(null);
  }, [isCreatingSuccessor]);

  const [, successorFormAction, isSuccessorPending] = useActionState<null, FormData>(
    async (_prev: null, _formData: FormData) => {
      if (!id) return null;
      if (!isValidSuccessorReason(successorReason)) {
        setSuccessorReasonError("理由は1〜500文字で入力してください");
        return null;
      }
      const trimmed = successorReason.trim();
      setSuccessorReasonError(null);
      try {
        const created = await createSuccessorAsync({ id, reason: trimmed });
        setShowSuccessorDialog(false);
        setSuccessorReason("");
        navigate(paths.estimates.detail.getHref(created.id));
      } catch {
        // useCreateEstimateSuccessor.onError が既に通知済み(FE-RC-005)
      }
      return null;
    },
    null,
  );

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
        <EstimateDetailHeaderActions
          onBack={() => navigate(paths.estimates.getHref())}
          showEdit={showEdit}
          onEdit={() => id ? navigate(paths.estimates.edit.getHref(id)) : undefined}
          showDelete={showDelete}
          onDelete={() => setShowDeleteDialog(true)}
          deleteBusy={isDeleting || isDeletePending}
          showSuccessor={showSuccessor}
          onSuccessor={handleOpenSuccessorDialog}
          successorBusy={successorBusy}
        />
      }
      maxWidth="max-w-4xl"
    >
      <EstimateDetailInfo estimate={estimate} isExpired={isExpired} />

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
        <EstimateSuccessorDialog
          successorReason={successorReason}
          successorReasonError={successorReasonError}
          successorBusy={successorBusy}
          canSubmitSuccessor={canSubmitSuccessor}
          onReasonChange={(value) => {
            setSuccessorReason(value);
            setSuccessorReasonError(null);
          }}
          onClose={handleCloseSuccessorDialog}
          formAction={successorFormAction}
        />
      ) : null}
    </PageLayout>
  );
}
