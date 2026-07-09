import { CheckCircle2, Circle } from "lucide-react";

import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";
import { C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";

import { LstepTagAddDialog } from "./LstepTagAddDialog";
import { LstepTagList } from "./LstepTagList";
import { LstepTagRemoveInline } from "./LstepTagRemoveInline";
import type { LineIdFormState } from "../hooks/use-line-integration-card-state";

export function LineIntegrationCardFrame({ children }: { children: React.ReactNode }) {
  return (
    <div className={`rounded-lg border ${C.borderLight} p-4 flex flex-col gap-4`}>
      <h3 className={`text-sm font-medium ${C.text70} uppercase tracking-wide`}>
        LINE / Lステップ連携
      </h3>
      {children}
    </div>
  );
}

export function LineIntegrationLoading() {
  return (
    <div className={`rounded-lg border ${C.borderLight} p-4 ${C.bgPage}`}>
      <div className={`h-4 w-32 rounded ${C.bgSkeleton} animate-pulse`} />
    </div>
  );
}

export function LineIntegrationError() {
  return (
    <div className={`rounded-lg border ${C.borderLight} p-4`}>
      <p className={`text-sm ${C.danger}`}>
        LINE連携情報の取得に失敗しました
      </p>
    </div>
  );
}

interface LinkedStatusRowProps {
  lineUserId?: string;
  canEdit: boolean;
  onUnlinkClick: () => void;
}

export function LinkedStatusRow({
  lineUserId,
  canEdit,
  onUnlinkClick,
}: LinkedStatusRowProps) {
  return (
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
          onClick={onUnlinkClick}
        >
          解除
        </Button>
      ) : null}
    </div>
  );
}

interface LstepTagsSectionProps {
  ownerId: string;
  tags: string[];
  canEdit: boolean;
  removeTagName: string | null;
  disabled?: boolean;
  onRemoveTagNameChange: (tagName: string | null) => void;
  onAddClick: () => void;
}

export function LstepTagsSection({
  ownerId,
  tags,
  canEdit,
  removeTagName,
  disabled = false,
  onRemoveTagNameChange,
  onAddClick,
}: LstepTagsSectionProps) {
  return (
    <div className="flex flex-col gap-2">
      <span className={`text-xs ${C.text55} uppercase tracking-wide`}>
        Lステップタグ
      </span>

      {removeTagName !== null ? (
        <LstepTagRemoveInline
          tagName={removeTagName}
          ownerId={ownerId}
          onCancel={() => onRemoveTagNameChange(null)}
        />
      ) : null}

      <LstepTagList
        tags={tags}
        onRemove={(tagName) => onRemoveTagNameChange(tagName)}
        disabled={disabled}
        canEdit={disabled ? false : canEdit}
      />

      {canEdit && !disabled ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className={`self-start h-8 px-2 text-xs ${C.text60} ${C.hoverText}`}
          onClick={onAddClick}
        >
          タグを追加 ＋
        </Button>
      ) : null}
    </div>
  );
}

interface UnlinkedLineIdFormProps {
  canEdit: boolean;
  lineIdFormAction: (payload: FormData) => void;
  lineIdState: LineIdFormState;
}

export function UnlinkedLineIdForm({
  canEdit,
  lineIdFormAction,
  lineIdState,
}: UnlinkedLineIdFormProps) {
  if (!canEdit) return null;

  return (
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
        <SubmitButton loadingText="設定中..." colorVariant="brand">設定</SubmitButton>
      </div>
      {lineIdState.error !== null ? (
        <p className={`text-sm ${C.danger}`}>{lineIdState.error}</p>
      ) : null}
    </form>
  );
}

export function UnlinkedStatusRow() {
  return (
    <div className="flex items-center gap-2">
      <Circle className={`${ICON.smXs} ${C.text40} shrink-0`} />
      <span className={`text-sm ${C.text50}`}>未連携</span>
    </div>
  );
}

interface LineIntegrationDialogsProps {
  ownerId: string;
  ownerName: string;
  tagAddDialogOpen: boolean;
  confirmUnlinkOpen: boolean;
  confirmOptOutOpen: boolean;
  confirmTransferOpen: boolean;
  isDeletingLine: boolean;
  isUpdatingDeliveryExclusion: boolean;
  isUpdatingTransferStatus: boolean;
  onTagAddOpenChange: (open: boolean) => void;
  onUnlinkClose: () => void;
  onUnlinkConfirm: () => void;
  onOptOutClose: () => void;
  onOptOutConfirm: () => void;
  onTransferClose: () => void;
  onTransferConfirm: () => void;
}

export function LineIntegrationDialogs({
  ownerId,
  ownerName,
  tagAddDialogOpen,
  confirmUnlinkOpen,
  confirmOptOutOpen,
  confirmTransferOpen,
  isDeletingLine,
  isUpdatingDeliveryExclusion,
  isUpdatingTransferStatus,
  onTagAddOpenChange,
  onUnlinkClose,
  onUnlinkConfirm,
  onOptOutClose,
  onOptOutConfirm,
  onTransferClose,
  onTransferConfirm,
}: LineIntegrationDialogsProps) {
  return (
    <>
      <ConfirmDialog
        open={confirmUnlinkOpen}
        onClose={onUnlinkClose}
        onConfirm={onUnlinkConfirm}
        title="LINE連携を解除しますか？"
        description={`${ownerName} さんのLINE連携を解除します。この操作はLstepのタグも削除される場合があります。`}
        confirmLabel="解除する"
        cancelLabel="キャンセル"
        variant="destructive"
        isPending={isDeletingLine}
      />

      <ConfirmDialog
        open={confirmOptOutOpen}
        onClose={onOptOutClose}
        onConfirm={onOptOutConfirm}
        title="配信を停止しますか？"
        description={`${ownerName} さんへのLstep配信を停止します。`}
        confirmLabel="停止する"
        cancelLabel="キャンセル"
        isPending={isUpdatingDeliveryExclusion}
      />

      <ConfirmDialog
        open={confirmTransferOpen}
        onClose={onTransferClose}
        onConfirm={onTransferConfirm}
        title="転院済みに設定しますか？"
        description="転院フラグを設定すると Lステップへの配信が停止されます。よろしいですか？"
        confirmLabel="転院済みに設定"
        cancelLabel="キャンセル"
        isPending={isUpdatingTransferStatus}
      />

      <LstepTagAddDialog
        open={tagAddDialogOpen}
        onOpenChange={onTagAddOpenChange}
        ownerId={ownerId}
      />
    </>
  );
}
