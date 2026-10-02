import { useActionState, useRef, useState } from "react";

import { LSTEP_EXCL_DELIVERY_STOP } from "@/constants/lstep-tag-names";
import { getFormString } from "@/lib/form-data";
import { isPetDeceasedForClinicalWrite } from "@/lib/transforms/pet";
import { usePermission } from "@/hooks/use-permission";
import type { Owner } from "@/types/owner";

import { useConfirmOwnerLineId } from "../api/confirm-owner-line-id";
import { useGenerateLineLinkToken } from "../api/generate-line-link-token";
import { useGetOwnerLineTags } from "../api/get-owner-line-tags";
import { useUpdateOwnerDeliveryCaution } from "../api/update-owner-delivery-caution";
import { useUpdateOwnerDeliveryExclusion } from "../api/update-owner-delivery-exclusion";
import { useUpdateOwnerLine, useDeleteOwnerLine } from "../api/update-owner-line";
import { useUpdateOwnerTransferStatus } from "../api/update-owner-transfer-status";

export interface LineIdFormState {
  error: string | null;
  success: boolean;
}

const INITIAL_LINE_ID_STATE: LineIdFormState = { error: null, success: false };

interface UseLineIntegrationCardStateArgs {
  ownerId: string;
  owner?: Owner;
}

export function useLineIntegrationCardState({ ownerId, owner }: UseLineIntegrationCardStateArgs) {
  const { canEdit } = usePermission("owners");
  const { data, isLoading, isError } = useGetOwnerLineTags(ownerId);
  const { mutateAsync: updateLine } = useUpdateOwnerLine(ownerId);
  const { mutate: deleteLine, isPending: isDeletingLine } = useDeleteOwnerLine(ownerId);
  const { mutate: confirmLineId, isPending: isConfirmingLineId } = useConfirmOwnerLineId(ownerId);
  const { mutate: updateDeliveryExclusion, isPending: isUpdatingDeliveryExclusion } =
    useUpdateOwnerDeliveryExclusion(ownerId);
  const { mutate: updateDeliveryCaution, isPending: isUpdatingDeliveryCaution } =
    useUpdateOwnerDeliveryCaution(ownerId);
  const { mutate: updateTransferStatus, isPending: isUpdatingTransferStatus } =
    useUpdateOwnerTransferStatus(ownerId);
  const {
    mutate: generateLinkToken,
    data: linkTokenResult,
    isPending: isGeneratingLinkToken,
  } = useGenerateLineLinkToken(ownerId);

  const [tagAddDialogOpen, setTagAddDialogOpen] = useState(false);
  const [removeTagName, setRemoveTagName] = useState<string | null>(null);
  const [confirmUnlinkOpen, setConfirmUnlinkOpen] = useState(false);
  const [confirmOptOutOpen, setConfirmOptOutOpen] = useState(false);
  const [confirmTransferOpen, setConfirmTransferOpen] = useState(false);
  const deliveryReasonInputRef = useRef<HTMLInputElement>(null);
  const deliveryCautionReasonInputRef = useRef<HTMLInputElement>(null);

  const [lineIdState, lineIdFormAction] = useActionState(
    async (_prevState: LineIdFormState, formData: FormData): Promise<LineIdFormState> => {
      const lineUserId = getFormString(formData, "line_user_id").trim();
      if (!lineUserId) {
        return { error: "LINE User IDを入力してください", success: false };
      }
      try {
        await updateLine({ line_user_id: lineUserId });
        return { error: null, success: true };
      } catch {
        // useUpdateOwnerLine の onError が handleApiError 済み。ここでは再通知しない。
        return { error: "LINE User ID の紐付けに失敗しました", success: false };
      }
    },
    INITIAL_LINE_ID_STATE,
  );

  const tags = data?.tags ?? [];
  const hasExclusionTag = tags.includes(LSTEP_EXCL_DELIVERY_STOP);
  const lineUserId = owner?.lineUserId ?? data?.line_user_id ?? undefined;
  const lineIdConfirmedAt = owner?.lineIdConfirmedAt;
  // EMR-251: 表示の正本は永続化済みフラグ — owner.deliveryExcluded（detail DTO）と
  // data.lstep_opt_out（LINE tags API）。owner.lstepOptOut / lstepOptOutReason は
  // detail DTO に存在せず transform で固定 false/undefined になるため参照しない。
  // 独立停止条件は BE SyncExclusionTags（internal/lstep/lstep_tag_sync_pet_exclusion.go）
  // と同じ集合で判定し、転院・退亡・全ペット死亡が残る限り「配信可能」と誤表示しない。
  const allPetsDeceased =
    owner?.pets !== undefined &&
    owner.pets.length > 0 &&
    owner.pets.every((pet) => isPetDeceasedForClinicalWrite(pet));
  const isTransferredStop = Boolean(owner?.isTransferred || owner?.membershipType === "他診/準");
  const isDeliveryStopped = Boolean(
    owner?.deliveryExcluded ||
    isTransferredStop ||
    owner?.membershipType === "退亡者" ||
    allPetsDeceased ||
    data?.lstep_opt_out ||
    hasExclusionTag,
  );
  const deliveryStopReason =
    owner?.deliveryExcludedReason ??
    (isTransferredStop ? "転院済み" : undefined) ??
    (owner?.membershipType === "退亡者" ? "退亡者会員" : undefined) ??
    (allPetsDeceased ? "全ペット死亡" : undefined) ??
    (hasExclusionTag ? LSTEP_EXCL_DELIVERY_STOP : undefined);

  const resumeDelivery = () => {
    updateDeliveryExclusion({ excluded: false, reason: null });
    if (deliveryReasonInputRef.current) {
      deliveryReasonInputRef.current.value = "";
    }
  };

  const confirmUnlink = () => {
    deleteLine();
    setConfirmUnlinkOpen(false);
  };

  const confirmOptOut = () => {
    updateDeliveryExclusion({
      excluded: true,
      reason: deliveryReasonInputRef.current?.value.trim() || undefined,
    });
    setConfirmOptOutOpen(false);
  };

  const confirmTransfer = () => {
    updateTransferStatus({ is_transferred: true });
    setConfirmTransferOpen(false);
  };

  return {
    canEdit,
    data,
    isLoading,
    isError,
    tags,
    lineUserId,
    lineIdConfirmedAt,
    isDeliveryStopped,
    deliveryStopReason,
    lineIdState,
    lineIdFormAction,
    deliveryReasonInputRef,
    deliveryCautionReasonInputRef,
    tagAddDialogOpen,
    setTagAddDialogOpen,
    removeTagName,
    setRemoveTagName,
    confirmUnlinkOpen,
    setConfirmUnlinkOpen,
    confirmOptOutOpen,
    setConfirmOptOutOpen,
    confirmTransferOpen,
    setConfirmTransferOpen,
    confirmLineId,
    updateDeliveryExclusion,
    updateDeliveryCaution,
    updateTransferStatus,
    resumeDelivery,
    confirmUnlink,
    confirmOptOut,
    confirmTransfer,
    isDeletingLine,
    isConfirmingLineId,
    isUpdatingDeliveryExclusion,
    isUpdatingDeliveryCaution,
    isUpdatingTransferStatus,
    generateLinkToken,
    linkTokenResult,
    isGeneratingLinkToken,
  };
}
