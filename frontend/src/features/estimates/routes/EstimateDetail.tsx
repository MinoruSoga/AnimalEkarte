import { ICON, C } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
import { paths } from "@/config/paths";
import { useBackNavigation } from "@/hooks/use-back-navigation";
import { LoadingFallback, EmptyState } from "@/components/shared/DataStates";
import { Button } from "@/components/ui/button";
import { resolveEntityReadResult } from "@/lib/entity-read-result";
import { useNavigate, useParams } from "react-router";
import { FileText } from "lucide-react";
import { useState, useCallback, useActionState, useTransition, type ReactNode } from "react";
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

interface EstimateDetailStatusPageProps {
  onBack: () => void;
  children: ReactNode;
}

// 読み込み/エラー/未検出のページ状態を PageLayout 内に置き、説明と一覧への安全な戻りを提供する。
function EstimateDetailStatusPage({ onBack, children }: EstimateDetailStatusPageProps) {
  return (
    <PageLayout
      title="見積書"
      resource={ResourceEstimates}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      onBack={onBack}
      maxWidth="max-w-4xl"
    >
      {children}
    </PageLayout>
  );
}

export function EstimateDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showSuccessorDialog, setShowSuccessorDialog] = useState(false);
  const [successorReason, setSuccessorReason] = useState("");
  const [successorReasonError, setSuccessorReasonError] = useState<string | null>(null);

  const { data, isLoading, isError, error, refetch } = useGetEstimate(id);
  const handleBack = useBackNavigation(paths.estimates.getHref());
  const { mutate: deleteEstimate, isPending: isDeleting } = useDeleteEstimate();
  const { mutateAsync: createSuccessorAsync, isPending: isCreatingSuccessor } =
    useCreateEstimateSuccessor();
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

  const estimateRead = resolveEntityReadResult({
    id,
    data,
    isLoading,
    isError,
    error,
    refetch,
  });

  if (estimateRead.status === "loading") {
    return (
      <EstimateDetailStatusPage onBack={handleBack}>
        <LoadingFallback />
      </EstimateDetailStatusPage>
    );
  }
  if (estimateRead.status === "error") {
    return (
      <EstimateDetailStatusPage onBack={handleBack}>
        <EmptyState
          message="見積書の取得に失敗しました"
          description="通信状況を確認して、もう一度お試しください。"
        >
          <div className="flex flex-wrap justify-center gap-2">
            {estimateRead.retry ? (
              <Button type="button" variant="outline" size="sm" onClick={estimateRead.retry}>
                再試行
              </Button>
            ) : null}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => navigate(paths.estimates.getHref())}
            >
              一覧に戻る
            </Button>
          </div>
        </EmptyState>
      </EstimateDetailStatusPage>
    );
  }
  if (estimateRead.status !== "found") {
    return (
      <EstimateDetailStatusPage onBack={handleBack}>
        <EmptyState
          message="見積書が見つかりません"
          description="指定された見積書は存在しないか、表示する権限がありません。"
        >
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => navigate(paths.estimates.getHref())}
          >
            一覧に戻る
          </Button>
        </EmptyState>
      </EstimateDetailStatusPage>
    );
  }

  const estimate = estimateRead.data;

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
          onBack={handleBack}
          showEdit={showEdit}
          onEdit={() => (id ? navigate(paths.estimates.edit.getHref(id)) : undefined)}
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
