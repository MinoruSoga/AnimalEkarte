import { useCallback, useEffect, useRef } from "react";
import { useParams, useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { EyeOff, FileText } from "lucide-react";

import { Button } from "@/components/ui/button";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { paths } from "@/config/paths";
import { useGetPet } from "@/hooks/use-pet";
import { usePermission } from "@/hooks/use-permission";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { C, ICON } from "@/lib/design-tokens";
import {
  isNonDisclosureReadStatus,
  resolveEntityReadResult,
} from "@/lib/entity-read-result";
import { ResourceEstimates } from "@/types/generated/models";

import { useGetEstimate } from "../api/get-estimate";
import { useEstimateForm } from "../hooks/use-estimate-form";
import {
  ESTIMATE_LOCKED_EDIT_MESSAGE,
  isEstimateLockedStatus,
} from "../lib/is-estimate-locked-status";

import { AmountSection, BasicInfoSection, TextSection } from "../components/EstimateFormSections";

function EstimateFormContent({ id }: { id?: string }) {
  const navigate = useNavigate();
  // BUG-019: mode from route param; found state from classified read result
  const isEdit = !!id;
  const {
    data: estimateData,
    isLoading,
    isError,
    error: estimateError,
    refetch: refetchEstimate,
  } = useGetEstimate(id);
  const entityRead = resolveEntityReadResult({
    id,
    data: estimateData,
    isLoading,
    isError,
    error: estimateError,
    refetch: refetchEstimate,
  });
  const foundEstimate = entityRead.status === "found" ? entityRead.data : undefined;

  const [searchParams] = useSearchParams();
  const petIdFromQuery = searchParams.get("petId") ?? "";
  const {
    data: petFromQuery,
    isPending: petQueryPending,
    isSuccess: petQuerySuccess,
    isError: petQueryError,
  } = useGetPet(petIdFromQuery);

  const { canEdit, canCreate } = usePermission("estimates");
  // BUG-372: 割引権限（割引額制御）
  const { canEdit: canEditDiscount } = usePermission("discount");
  const canSubmit = isEdit ? canEdit : canCreate;

  // FE-RC-002/004: `?petId=` から採用したペットが死亡・不明のとき新規見積書作成を fail-closed で拒否する。
  // 選択 UI と同様、生存が明示されるまでブロックする（pending / error / 不明 status も含む）。
  const hasPetIdFromQuery = Boolean(petIdFromQuery);
  const isNewEstimatePetDeceased = Boolean(
    !isEdit && hasPetIdFromQuery && petQuerySuccess && petFromQuery?.status === "死亡",
  );
  const blocksNewEstimatePet = Boolean(
    !isEdit &&
      hasPetIdFromQuery &&
      (petQueryPending ||
        petQueryError ||
        !petQuerySuccess ||
        !petFromQuery ||
        petFromQuery.status !== "生存"),
  );
  // 表示メッセージは settle 後のみ（pending 中は fieldset disabled のみ）。
  const deceasedPetBlockMessage = (() => {
    if (isEdit || !hasPetIdFromQuery || petQueryPending) return undefined;
    if (isNewEstimatePetDeceased) return "死亡したペットの見積書は作成できません";
    if (petQueryError || (petQuerySuccess && petFromQuery?.status !== "生存")) {
      return "ペットの生死状態を確認できないため、新規見積書を作成できません";
    }
    return undefined;
  })();

  const { form, handleChange, formAction, formState, handleCancel, isPending } = useEstimateForm({
    mode: isEdit ? "edit" : "create",
    estimate: foundEstimate,
    initialOwnerId: petFromQuery?.ownerId,
    initialPetId: petFromQuery?.id,
    // FE-RC-001: action 別の最新権限値を mutation 直前に再検査するため hook へ渡す。
    permissions: { canCreate, canEdit },
    // FE-RC-002/004: callback 側の二重防壁（render 側の fieldset/banner と同じ判定）。
    blockCreateReason: deceasedPetBlockMessage,
  });

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // React 19 Action の成功を検知して遷移
  // rerender-dependencies: estimate（オブジェクト）の代わりに estimate?.id（primitive）を deps に使用
  const estimateId = foundEstimate?.id;
  const estimateStatus = foundEstimate?.status;
  const isLockedEdit =
    isEdit && estimateId != null && estimateStatus != null
      ? isEstimateLockedStatus(estimateStatus)
      : false;
  const lockedRedirectRef = useRef(false);

  useEffect(() => {
    if (!isLockedEdit || estimateId == null || lockedRedirectRef.current) return;
    lockedRedirectRef.current = true;
    toast.info(ESTIMATE_LOCKED_EDIT_MESSAGE);
    navigate(paths.estimates.detail.getHref(estimateId), { replace: true });
  }, [isLockedEdit, estimateId, navigate]);

  useEffect(() => {
    if (formState.success) {
      markClean();
      if (isEdit && estimateId != null) {
        navigate(paths.estimates.detail.getHref(estimateId));
      } else {
        // Since we don't have the new ID easily here, we might need to handle it in hook
        // but for now redirect to list
        navigate(paths.estimates.getHref());
      }
    }
  }, [formState.success, formState.timestamp, navigate, markClean, isEdit, estimateId]);

  // rerender-memo: memo'd セクションに渡すハンドラを useCallback で安定化
  const handleChangeWithDirty = useCallback(
    (key: string, value: unknown) => {
      markDirty();
      (handleChange as (k: string, v: unknown) => void)(key, value);
    },
    [markDirty, handleChange]
  );

  if (isEdit && entityRead.status === "loading") {
    return <LoadingFallback />;
  }

  // BUG-019: missing / other-clinic / forbidden → Not Found (non-disclosure), never blank form
  if (isEdit && isNonDisclosureReadStatus(entityRead.status)) {
    return (
      <PageLayout
        title="見積書"
        resource={ResourceEstimates}
        icon={<FileText className={`${ICON.page} ${C.text}`} />}
        maxWidth="max-w-2xl"
      >
        <ErrorFallback message="見積書が見つかりません" />
      </PageLayout>
    );
  }

  if (isEdit && entityRead.status === "error") {
    return (
      <PageLayout
        title="見積書"
        resource={ResourceEstimates}
        icon={<FileText className={`${ICON.page} ${C.text}`} />}
        maxWidth="max-w-2xl"
      >
        <div className="space-y-3">
          <ErrorFallback message="見積書の取得に失敗しました" />
          {entityRead.retry ? (
            <Button type="button" variant="outline" size="sm" onClick={entityRead.retry}>
              再試行
            </Button>
          ) : null}
        </div>
      </PageLayout>
    );
  }

  // locked 見積の edit 直アクセス: 更新 UI を出さず detail へリダイレクト中
  if (isLockedEdit) {
    return <LoadingFallback />;
  }

  return (
    <form action={formAction}>
    <PageLayout
      title={isEdit ? "見積書編集" : "新規見積書作成"}
      resource={ResourceEstimates}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex gap-2">
          <Button variant="outline" type="button" size="sm" onClick={handleCancel} className="h-11 text-sm">
            キャンセル
          </Button>
          {canSubmit && !blocksNewEstimatePet ? (
            <SubmitButton
              size="sm"
              colorVariant="primary"
              disabled={!form.title.trim()}
              className="h-11 text-sm"
            >
              {isEdit ? "更新" : "作成"}
            </SubmitButton>
          ) : null}
        </div>
      }
      maxWidth="max-w-2xl"
    >
      <NavigationBlocker when={isDirty ? !isPending : false} />
      {/* FE-RC-002/004: 死亡・生死不明ペットの新規見積書作成を render 側で拒否する（callback 側は use-estimate-form.ts）。 */}
      {deceasedPetBlockMessage ? (
        <div
          className={`flex items-center gap-2 px-4 py-2.5 rounded-md border mb-4 ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
          role="status"
          aria-label="作成不可"
        >
          <EyeOff className={`shrink-0 h-4 w-4 ${C.textWarningIcon}`} aria-hidden="true" />
          <span className="text-sm font-medium">{deceasedPetBlockMessage}</span>
        </div>
      ) : null}
      <fieldset disabled={blocksNewEstimatePet} className="border-0 p-0 m-0 min-w-0">
        <div className={`${C.bgWhite} border ${C.borderLight} rounded-md p-6 space-y-6`}>
          {/* rerender-memo: BasicInfoSection — 金額/テキスト変更では再レンダーしない */}
          <BasicInfoSection
            title={form.title}
            status={form.status}
            validUntil={form.validUntil}
            isEdit={isEdit}
            onChange={handleChangeWithDirty}
            titleError={formState.fieldErrors?.title}
            statusError={formState.fieldErrors?.status}
          />

          {/* rerender-memo: AmountSection — 基本情報/テキスト変更では再レンダーしない */}
          <AmountSection
            subtotal={form.subtotal}
            taxTotal={form.taxTotal}
            insuranceAmount={form.insuranceAmount}
            discountAmount={form.discountAmount}
            totalAmount={form.totalAmount}
            canEditDiscount={canEditDiscount}
            onChange={handleChangeWithDirty}
          />

          {/* rerender-memo: TextSection — 基本情報/金額変更では再レンダーしない */}
          <TextSection
            comment={form.comment}
            notes={form.notes}
            onChange={handleChangeWithDirty}
          />
        </div>
      </fieldset>
    </PageLayout>
    </form>
  );
}

export function EstimateForm() {
  const { id } = useParams<{ id: string }>();
  return <EstimateFormContent id={id} />;
}
