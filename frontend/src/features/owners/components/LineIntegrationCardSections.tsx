import type { RefObject } from "react";
import { AlertTriangle, Ban } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { Owner } from "@/types/owner";

interface DeliveryStatusBannerProps {
  isStopped: boolean;
  stopReason?: string;
  caution?: boolean;
  cautionReason?: string;
}

export function DeliveryStatusBanner({
  isStopped,
  stopReason,
  caution,
  cautionReason,
}: DeliveryStatusBannerProps) {
  if (isStopped) {
    return (
      <div className={`flex items-center gap-2 rounded-md border ${C.borderRedBadge} ${C.bgRedLight} px-4 py-3`}>
        <Ban className={`${ICON.smXs} ${C.textNotionRed} shrink-0`} />
        <span className={`text-sm font-medium ${C.textNotionRed}`}>配信停止中</span>
        <span className={`text-xs ${C.text50}`}>この飼い主はLステップ配信対象外です</span>
        {stopReason ? (
          <span className={`text-xs ${C.text50}`}>— {stopReason}</span>
        ) : null}
      </div>
    );
  }

  if (!caution) return null;

  return (
    <div
      className={`flex items-center gap-2 rounded-md border ${C.borderNotice} ${C.bgNotice} px-4 py-3`}
      data-testid="delivery-caution-banner"
    >
      <AlertTriangle className={`${ICON.smXs} ${C.textNotice} shrink-0`} />
      <span className={`text-sm font-medium ${C.textNotice}`}>配信注意</span>
      <span className={`text-xs ${C.text50}`}>この飼い主は配信注意対象です</span>
      {cautionReason ? (
        <span className={`text-xs ${C.text50}`}>— {cautionReason}</span>
      ) : null}
    </div>
  );
}

interface LineIdConfirmSectionProps {
  isLinked: boolean;
  lineUserId?: string;
  lineIdConfirmedAt?: string;
  canEdit: boolean;
  isPending: boolean;
  onConfirm: () => void;
}

export function LineIdConfirmSection({
  isLinked,
  lineUserId,
  lineIdConfirmedAt,
  canEdit,
  isPending,
  onConfirm,
}: LineIdConfirmSectionProps) {
  if (!isLinked || !lineUserId) return null;

  return (
    <div
      className={`flex items-center justify-between gap-3 rounded-md border px-4 py-3 ${
        lineIdConfirmedAt ? C.borderLight : `${C.borderNotice} ${C.bgNotice}`
      }`}
    >
      <div className="flex items-center gap-2">
        <span className={`text-sm font-medium ${lineIdConfirmedAt ? C.textStatusGreen : C.textNotice}`}>
          {lineIdConfirmedAt ? "LINE ID 確認済み" : "LINE ID 未確認"}
        </span>
        {lineIdConfirmedAt ? (
          <span className={`text-xs ${C.text50}`}>
            {lineIdConfirmedAt.split("T")[0]}
          </span>
        ) : null}
      </div>
      {canEdit && !lineIdConfirmedAt ? (
        <Button
          type="button"
          size="sm"
          variant="outline"
          className="h-8 px-3 text-xs shrink-0"
          disabled={isPending}
          onClick={onConfirm}
        >
          {isPending ? "処理中..." : "確認する"}
        </Button>
      ) : null}
    </div>
  );
}

interface LineDeliveryControlsProps {
  canEdit: boolean;
  owner?: Owner;
  deliveryReasonInputRef: RefObject<HTMLInputElement | null>;
  deliveryCautionReasonInputRef: RefObject<HTMLInputElement | null>;
  isUpdatingDeliveryExclusion: boolean;
  isUpdatingDeliveryCaution: boolean;
  isUpdatingTransferStatus: boolean;
  onUpdateDeliveryExclusion: (input: { excluded: boolean; reason?: string | null }) => void;
  onUpdateDeliveryCaution: (input: { caution: boolean; reason?: string | null }) => void;
  onUpdateTransferStatus: (input: { is_transferred: boolean }) => void;
  onTransferEnableRequest: () => void;
}

export function LineDeliveryControls({
  canEdit,
  owner,
  deliveryReasonInputRef,
  deliveryCautionReasonInputRef,
  isUpdatingDeliveryExclusion,
  isUpdatingDeliveryCaution,
  isUpdatingTransferStatus,
  onUpdateDeliveryExclusion,
  onUpdateDeliveryCaution,
  onUpdateTransferStatus,
  onTransferEnableRequest,
}: LineDeliveryControlsProps) {
  if (!canEdit) return null;

  const getDeliveryReasonInput = () =>
    deliveryReasonInputRef.current?.value.trim() || undefined;
  const getDeliveryCautionReasonInput = () =>
    deliveryCautionReasonInputRef.current?.value.trim() || undefined;

  return (
    <>
      <div className={`border-t ${C.borderLight} pt-3 flex flex-col gap-2`}>
        <div className="flex items-center justify-between gap-3">
          <span className={`text-sm ${C.text60}`}>配信除外</span>
          <Switch
            aria-label="配信除外"
            checked={owner?.deliveryExcluded ?? false}
            disabled={isUpdatingDeliveryExclusion || !owner}
            onCheckedChange={(checked) => {
              if (checked) {
                onUpdateDeliveryExclusion({
                  excluded: true,
                  reason: getDeliveryReasonInput(),
                });
              } else {
                onUpdateDeliveryExclusion({ excluded: false, reason: null });
                if (deliveryReasonInputRef.current) {
                  deliveryReasonInputRef.current.value = "";
                }
              }
            }}
          />
        </div>
        <div className="flex gap-2">
          <input
            key={owner?.deliveryExcludedReason ?? "no-delivery-exclusion-reason"}
            ref={deliveryReasonInputRef}
            type="text"
            maxLength={100}
            disabled={!owner || isUpdatingDeliveryExclusion}
            className={`${STYLE.formInput} flex-1 rounded-md px-3`}
            placeholder="除外理由（任意・100文字以内）"
            aria-label="配信除外理由"
            defaultValue={owner?.deliveryExcludedReason ?? ""}
          />
          <Button
            type="button"
            size="sm"
            variant="outline"
            className="h-8 px-3 text-xs shrink-0"
            disabled={isUpdatingDeliveryExclusion || !owner?.deliveryExcluded}
            onClick={() =>
              onUpdateDeliveryExclusion({
                excluded: true,
                reason: getDeliveryReasonInput(),
              })
            }
          >
            理由を保存
          </Button>
        </div>
      </div>

      <div className={`border-t ${C.borderLight} pt-3 flex flex-col gap-2`}>
        <div className="flex items-center justify-between gap-3">
          <span className={`text-sm ${C.text60}`}>配信注意</span>
          <Switch
            aria-label="配信注意"
            checked={owner?.deliveryCaution ?? false}
            disabled={isUpdatingDeliveryCaution || !owner}
            onCheckedChange={(checked) => {
              if (checked) {
                onUpdateDeliveryCaution({
                  caution: true,
                  reason: getDeliveryCautionReasonInput(),
                });
              } else {
                onUpdateDeliveryCaution({ caution: false, reason: null });
                if (deliveryCautionReasonInputRef.current) {
                  deliveryCautionReasonInputRef.current.value = "";
                }
              }
            }}
          />
        </div>
        {owner?.deliveryCaution ? (
          <div className="flex gap-2">
            <input
              key={owner?.deliveryCautionReason ?? "no-delivery-caution-reason"}
              ref={deliveryCautionReasonInputRef}
              type="text"
              maxLength={100}
              disabled={!owner || isUpdatingDeliveryCaution}
              className={`${STYLE.formInput} flex-1 rounded-md px-3`}
              placeholder="注意理由（任意・100文字以内）"
              aria-label="配信注意事項の理由"
              defaultValue={owner?.deliveryCautionReason ?? ""}
            />
            <Button
              type="button"
              size="sm"
              variant="outline"
              className="h-8 px-3 text-xs shrink-0"
              data-testid="delivery-caution-save-btn"
              disabled={isUpdatingDeliveryCaution || !owner?.deliveryCaution}
              onClick={() =>
                onUpdateDeliveryCaution({
                  caution: true,
                  reason: getDeliveryCautionReasonInput(),
                })
              }
            >
              理由を保存
            </Button>
          </div>
        ) : null}
      </div>

      <div className={`border-t ${C.borderLight} pt-3`}>
        <div className="flex items-center justify-between gap-3">
          <span className={`text-sm ${C.text60}`}>転院済み</span>
          <Switch
            aria-label="転院済み"
            checked={owner?.isTransferred ?? false}
            disabled={isUpdatingTransferStatus || !owner}
            onCheckedChange={(checked) => {
              if (checked) {
                onTransferEnableRequest();
              } else {
                onUpdateTransferStatus({ is_transferred: false });
              }
            }}
          />
        </div>
        {owner?.transferAt ? (
          <p className={`text-xs mt-1 ${C.text50}`}>
            転院日: {owner.transferAt.split("T")[0]}
          </p>
        ) : null}
      </div>
    </>
  );
}
