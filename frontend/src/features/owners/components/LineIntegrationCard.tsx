import { useActionState, useRef, useState } from "react";
import { CheckCircle2, Circle, Ban, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import type { Owner } from "@/types/owner";
import { useGetOwnerLineTags } from "../api/get-owner-line-tags";
import { useUpdateOwnerLine, useDeleteOwnerLine } from "../api/update-owner-line";
import { useConfirmOwnerLineId } from "../api/confirm-owner-line-id";
import { useUpdateOwnerDeliveryExclusion } from "../api/update-owner-delivery-exclusion";
import { useUpdateOwnerDeliveryCaution } from "../api/update-owner-delivery-caution";
import { useUpdateOwnerTransferStatus } from "../api/update-owner-transfer-status";
import { LstepTagList } from "./LstepTagList";
import { LstepTagAddDialog } from "./LstepTagAddDialog";
import { LstepTagRemoveInline } from "./LstepTagRemoveInline";

interface LineIntegrationCardProps {
  ownerId: string;
  ownerName: string;
  owner?: Owner;
}

interface LineIdFormState {
  error: string | null;
  success: boolean;
}

const INITIAL_LINE_ID_STATE: LineIdFormState = { error: null, success: false };

export function LineIntegrationCard({
  ownerId,
  ownerName,
  owner,
}: LineIntegrationCardProps) {
  const { canEdit } = usePermission("owners");
  const { data, isLoading, isError } = useGetOwnerLineTags(ownerId);
  const { mutateAsync: updateLine } = useUpdateOwnerLine(ownerId);
  const { mutate: deleteLine, isPending: isDeletingLine } = useDeleteOwnerLine(ownerId);
  const { mutate: confirmLineId, isPending: isConfirmingLineId } =
    useConfirmOwnerLineId(ownerId);
  const { mutate: updateDeliveryExclusion, isPending: isUpdatingDeliveryExclusion } =
    useUpdateOwnerDeliveryExclusion(ownerId);
  const { mutate: updateDeliveryCaution, isPending: isUpdatingDeliveryCaution } =
    useUpdateOwnerDeliveryCaution(ownerId);
  const { mutate: updateTransferStatus, isPending: isUpdatingTransferStatus } =
    useUpdateOwnerTransferStatus(ownerId);

  const [tagAddDialogOpen, setTagAddDialogOpen] = useState(false);
  const [removeTagName, setRemoveTagName] = useState<string | null>(null);
  const [confirmUnlinkOpen, setConfirmUnlinkOpen] = useState(false);
  const [confirmOptOutOpen, setConfirmOptOutOpen] = useState(false);
  const [confirmTransferOpen, setConfirmTransferOpen] = useState(false);
  const deliveryReasonInputRef = useRef<HTMLInputElement>(null);
  const getDeliveryReasonInput = () =>
    deliveryReasonInputRef.current?.value.trim() || undefined;
  const deliveryCautionReasonInputRef = useRef<HTMLInputElement>(null);
  const getDeliveryCautionReasonInput = () =>
    deliveryCautionReasonInputRef.current?.value.trim() || undefined;

  const [lineIdState, lineIdFormAction] = useActionState(
    async (
      _prevState: LineIdFormState,
      formData: FormData
    ): Promise<LineIdFormState> => {
      const lineUserId = (formData.get("line_user_id") as string).trim();
      if (!lineUserId) {
        return { error: "LINE User IDを入力してください", success: false };
      }
      try {
        await updateLine({ line_user_id: lineUserId });
        return { error: null, success: true };
      } catch {
        return { error: null, success: false };
      }
    },
    INITIAL_LINE_ID_STATE
  );

  if (isLoading) {
    return (
      <div className={`rounded-lg border ${C.borderLight} p-4 ${C.bgPage}`}>
        <div className={`h-4 w-32 rounded ${C.bgSkeleton} animate-pulse`} />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className={`rounded-lg border ${C.borderLight} p-4`}>
        <p className={`text-sm ${C.danger}`}>
          LINE連携情報の取得に失敗しました
        </p>
      </div>
    );
  }

  const { is_linked, line_user_id, tags, lstep_opt_out } = data;
  const lineUserId = owner?.lineUserId ?? line_user_id ?? undefined;
  const lineIdConfirmedAt = owner?.lineIdConfirmedAt;
  const hasExclusionTag = tags.includes("EXCL_配信停止");
  const isDeliveryStopped = Boolean(
    owner?.deliveryExcluded ||
    owner?.lstepOptOut ||
    owner?.isTransferred ||
    owner?.membershipType === "他診/準" ||
    lstep_opt_out ||
    hasExclusionTag
  );
  const deliveryStopReason = owner?.deliveryExcludedReason
    ?? owner?.lstepOptOutReason
    ?? (owner?.isTransferred || owner?.membershipType === "他診/準" ? "転院済み" : undefined)
    ?? (hasExclusionTag ? "EXCL_配信停止" : undefined);

  /** 配信停止バナー / 配信注意バナー（優先: 停止 > 注意） */
  const bannerSection = isDeliveryStopped ? (
    <div
      className={`flex items-center gap-2 rounded-md border ${C.borderRedBadge} ${C.bgRedLight} px-4 py-3`}
    >
      <Ban className={`${ICON.smXs} ${C.textNotionRed} shrink-0`} />
      <span className={`text-sm font-medium ${C.textNotionRed}`}>配信停止中</span>
      <span className={`text-xs ${C.text50}`}>この飼い主はLステップ配信対象外です</span>
      {deliveryStopReason ? (
        <span className={`text-xs ${C.text50}`}>— {deliveryStopReason}</span>
      ) : null}
    </div>
  ) : owner?.deliveryCaution ? (
    <div
      className={`flex items-center gap-2 rounded-md border ${C.borderNotice} ${C.bgNotice} px-4 py-3`}
      data-testid="delivery-caution-banner"
    >
      <AlertTriangle className={`${ICON.smXs} ${C.textNotice} shrink-0`} />
      <span className={`text-sm font-medium ${C.textNotice}`}>配信注意</span>
      <span className={`text-xs ${C.text50}`}>この飼い主は配信注意対象です</span>
      {owner.deliveryCautionReason ? (
        <span className={`text-xs ${C.text50}`}>— {owner.deliveryCautionReason}</span>
      ) : null}
    </div>
  ) : null;

  /** LINE ID 確認セクション（is_linked 時） */
  const lineIdConfirmSection =
    is_linked && lineUserId ? (
      <div
        className={`flex items-center justify-between gap-3 rounded-md border px-4 py-3 ${
          lineIdConfirmedAt
            ? C.borderLight
            : `${C.borderNotice} ${C.bgNotice}`
        }`}
      >
        <div className="flex items-center gap-2">
          <span
            className={`text-sm font-medium ${
              lineIdConfirmedAt ? C.textStatusGreen : C.textNotice
            }`}
          >
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
            disabled={isConfirmingLineId}
            onClick={() => confirmLineId()}
          >
            {isConfirmingLineId ? "処理中..." : "確認する"}
          </Button>
        ) : null}
      </div>
    ) : null;

  /** 配信除外スイッチ + 理由入力 */
  const deliveryExclusionSection = canEdit ? (
    <div className={`border-t ${C.borderLight} pt-3 flex flex-col gap-2`}>
      <div className="flex items-center justify-between gap-3">
        <span className={`text-sm ${C.text60}`}>配信除外</span>
        <Switch
          checked={owner?.deliveryExcluded ?? false}
          disabled={isUpdatingDeliveryExclusion || !owner}
          onCheckedChange={(checked) => {
            if (checked) {
              updateDeliveryExclusion({
                excluded: true,
                reason: getDeliveryReasonInput(),
              });
            } else {
              updateDeliveryExclusion({ excluded: false, reason: null });
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
          defaultValue={owner?.deliveryExcludedReason ?? ""}
        />
        <Button
          type="button"
          size="sm"
          variant="outline"
          className="h-8 px-3 text-xs shrink-0"
          disabled={isUpdatingDeliveryExclusion || !owner?.deliveryExcluded}
          onClick={() =>
            updateDeliveryExclusion({
              excluded: true,
              reason: getDeliveryReasonInput(),
            })
          }
        >
          理由を保存
        </Button>
      </div>
    </div>
  ) : null;

  /** 配信注意スイッチ + 理由入力 */
  const deliveryCautionSection = canEdit ? (
    <div className={`border-t ${C.borderLight} pt-3 flex flex-col gap-2`}>
      <div className="flex items-center justify-between gap-3">
        <span className={`text-sm ${C.text60}`}>配信注意</span>
        <Switch
          checked={owner?.deliveryCaution ?? false}
          disabled={isUpdatingDeliveryCaution || !owner}
          onCheckedChange={(checked) => {
            if (checked) {
              updateDeliveryCaution({
                caution: true,
                reason: getDeliveryCautionReasonInput(),
              });
            } else {
              updateDeliveryCaution({ caution: false, reason: null });
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
            defaultValue={owner?.deliveryCautionReason ?? ""}
          />
          <Button
            type="button"
            size="sm"
            variant="outline"
            className="h-8 px-3 text-xs shrink-0"
            disabled={isUpdatingDeliveryCaution || !owner?.deliveryCaution}
            onClick={() =>
              updateDeliveryCaution({
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
  ) : null;

  /** 転院ステータススイッチ */
  const transferStatusSection = canEdit ? (
    <div className={`border-t ${C.borderLight} pt-3`}>
      <div className="flex items-center justify-between gap-3">
        <span className={`text-sm ${C.text60}`}>転院済み</span>
        <Switch
          checked={owner?.isTransferred ?? false}
          disabled={isUpdatingTransferStatus || !owner}
          onCheckedChange={(checked) => {
            if (checked) {
              setConfirmTransferOpen(true);
            } else {
              updateTransferStatus({ is_transferred: false });
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
  ) : null;

  // 連携済み + 配信停止中
  if (is_linked && lstep_opt_out) {
    return (
      <div className={`rounded-lg border ${C.borderLight} p-4 flex flex-col gap-4`}>
        <h3 className={`text-sm font-medium ${C.text70} uppercase tracking-wide`}>
          LINE / Lステップ連携
        </h3>

        {bannerSection}

        {canEdit ? (
          <div className="flex justify-end">
            <Button
              type="button"
              size="sm"
              variant="outline"
              className="h-8 px-3 text-xs shrink-0"
              disabled={isUpdatingDeliveryExclusion}
              onClick={() => {
                updateDeliveryExclusion({ excluded: false, reason: null });
                if (deliveryReasonInputRef.current) {
                  deliveryReasonInputRef.current.value = "";
                }
              }}
            >
              {isUpdatingDeliveryExclusion ? "処理中..." : "配信を再開"}
            </Button>
          </div>
        ) : null}

        {lineIdConfirmSection}

        {/* タグ一覧（グレーアウト） */}
        <div className="flex flex-col gap-2">
          <span className={`text-xs ${C.text55} uppercase tracking-wide`}>
            Lステップタグ
          </span>
          <LstepTagList
            tags={tags}
            onRemove={() => undefined}
            disabled
            canEdit={false}
          />
        </div>

        {deliveryExclusionSection}
        {deliveryCautionSection}
        {transferStatusSection}

        <ConfirmDialog
          open={confirmTransferOpen}
          onClose={() => setConfirmTransferOpen(false)}
          onConfirm={() => {
            updateTransferStatus({ is_transferred: true });
            setConfirmTransferOpen(false);
          }}
          title="転院済みに設定しますか？"
          description="転院フラグを設定すると Lステップへの配信が停止されます。よろしいですか？"
          confirmLabel="転院済みに設定"
          cancelLabel="キャンセル"
          isPending={isUpdatingTransferStatus}
        />
      </div>
    );
  }

  // 連携済み + 配信中
  if (is_linked) {
    return (
      <div className={`rounded-lg border ${C.borderLight} p-4 flex flex-col gap-4`}>
        <h3 className={`text-sm font-medium ${C.text70} uppercase tracking-wide`}>
          LINE / Lステップ連携
        </h3>

        {bannerSection}

        {/* 連携ステータス */}
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <CheckCircle2
              className={`${ICON.smXs} shrink-0`}
              style={{ color: PALETTE.lineGreen }}
            />
            <span className={`text-sm font-medium ${C.textStatusGreen}`}>
              連携済み
            </span>
            {lineUserId ? (
              <span className={`text-xs ${C.text50} font-mono`}>
                ({lineUserId.slice(0, 8)}...{lineUserId.slice(-4)})
              </span>
            ) : null}
          </div>

          {canEdit ? (
            <Button
              type="button"
              size="sm"
              variant="outline"
              className={`h-8 px-3 text-xs shrink-0 ${C.danger} ${C.borderDanger}`}
              onClick={() => setConfirmUnlinkOpen(true)}
            >
              解除
            </Button>
          ) : null}
        </div>

        {lineIdConfirmSection}

        {/* タグ一覧 */}
        <div className="flex flex-col gap-2">
          <span className={`text-xs ${C.text55} uppercase tracking-wide`}>
            Lステップタグ
          </span>

          {removeTagName !== null ? (
            <LstepTagRemoveInline
              tagName={removeTagName}
              ownerId={ownerId}
              onCancel={() => setRemoveTagName(null)}
            />
          ) : null}

          <LstepTagList
            tags={tags}
            onRemove={(tagName) => setRemoveTagName(tagName)}
            canEdit={canEdit}
          />

          {canEdit ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className={`self-start h-8 px-2 text-xs ${C.text60} ${C.hoverText}`}
              onClick={() => setTagAddDialogOpen(true)}
            >
              タグを追加 ＋
            </Button>
          ) : null}
        </div>

        {/* 配信停止 */}
        {canEdit ? (
          <div className={`border-t ${C.borderLight} pt-3`}>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className={`h-8 px-2 text-xs ${C.text40} ${C.hoverText60}`}
              onClick={() => setConfirmOptOutOpen(true)}
            >
              配信を停止する
            </Button>
          </div>
        ) : null}

        {deliveryExclusionSection}
        {deliveryCautionSection}
        {transferStatusSection}

        {/* 連携解除ConfirmDialog */}
        <ConfirmDialog
          open={confirmUnlinkOpen}
          onClose={() => setConfirmUnlinkOpen(false)}
          onConfirm={() => {
            deleteLine();
            setConfirmUnlinkOpen(false);
          }}
          title="LINE連携を解除しますか？"
          description={`${ownerName} さんのLINE連携を解除します。この操作はLstepのタグも削除される場合があります。`}
          confirmLabel="解除する"
          cancelLabel="キャンセル"
          variant="destructive"
          isPending={isDeletingLine}
        />

        {/* 配信停止ConfirmDialog */}
        <ConfirmDialog
          open={confirmOptOutOpen}
          onClose={() => setConfirmOptOutOpen(false)}
          onConfirm={() => {
            updateDeliveryExclusion({
              excluded: true,
              reason: getDeliveryReasonInput(),
            });
            setConfirmOptOutOpen(false);
          }}
          title="配信を停止しますか？"
          description={`${ownerName} さんへのLstep配信を停止します。`}
          confirmLabel="停止する"
          cancelLabel="キャンセル"
          isPending={isUpdatingDeliveryExclusion}
        />

        {/* 転院ConfirmDialog */}
        <ConfirmDialog
          open={confirmTransferOpen}
          onClose={() => setConfirmTransferOpen(false)}
          onConfirm={() => {
            updateTransferStatus({ is_transferred: true });
            setConfirmTransferOpen(false);
          }}
          title="転院済みに設定しますか？"
          description="転院フラグを設定すると Lステップへの配信が停止されます。よろしいですか？"
          confirmLabel="転院済みに設定"
          cancelLabel="キャンセル"
          isPending={isUpdatingTransferStatus}
        />

        {/* タグ追加ダイアログ */}
        <LstepTagAddDialog
          open={tagAddDialogOpen}
          onOpenChange={setTagAddDialogOpen}
          ownerId={ownerId}
        />
      </div>
    );
  }

  // 未連携
  return (
    <div className={`rounded-lg border ${C.borderLight} p-4 flex flex-col gap-4`}>
      <h3 className={`text-sm font-medium ${C.text70} uppercase tracking-wide`}>
        LINE / Lステップ連携
      </h3>

      {bannerSection}

      <div className="flex items-center gap-2">
        <Circle className={`${ICON.smXs} ${C.text40} shrink-0`} />
        <span className={`text-sm ${C.text50}`}>未連携</span>
      </div>

      {canEdit ? (
        <form action={lineIdFormAction} className="flex flex-col gap-2">
          <label htmlFor="line_user_id" className={STYLE.formLabel}>
            LINE User ID
          </label>
          <div className="flex gap-2">
            <input
              id="line_user_id"
              name="line_user_id"
              type="text"
              placeholder="Uxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
              className={`${STYLE.formInput} flex-1 rounded-md px-3`}
            />
            <SubmitButton loadingText="設定中...">設定</SubmitButton>
          </div>
          {lineIdState.error !== null ? (
            <p className={`text-sm ${C.danger}`}>{lineIdState.error}</p>
          ) : null}
        </form>
      ) : null}

      {deliveryExclusionSection}
      {deliveryCautionSection}
      {transferStatusSection}

      <ConfirmDialog
        open={confirmTransferOpen}
        onClose={() => setConfirmTransferOpen(false)}
        onConfirm={() => {
          updateTransferStatus({ is_transferred: true });
          setConfirmTransferOpen(false);
        }}
        title="転院済みに設定しますか？"
        description="転院フラグを設定すると Lステップへの配信が停止されます。よろしいですか？"
        confirmLabel="転院済みに設定"
        cancelLabel="キャンセル"
        isPending={isUpdatingTransferStatus}
      />
    </div>
  );
}
