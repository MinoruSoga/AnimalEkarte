import { useActionState, useEffect, useState } from "react";
import { CheckCircle2, Circle, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { useGetOwnerLineTags } from "../api/get-owner-line-tags";
import { useUpdateOwnerLine, useDeleteOwnerLine } from "../api/update-owner-line";
import { useUpdateOwnerLstepOptOut } from "../api/update-owner-lstep-opt-out";
import { LstepTagList } from "./LstepTagList";
import { LstepTagAddDialog } from "./LstepTagAddDialog";
import { LstepTagRemoveInline } from "./LstepTagRemoveInline";

interface LineIntegrationCardProps {
  ownerId: string;
  ownerName: string;
}

interface LineIdFormState {
  error: string | null;
  success: boolean;
}

const INITIAL_LINE_ID_STATE: LineIdFormState = { error: null, success: false };

export function LineIntegrationCard({
  ownerId,
  ownerName,
}: LineIntegrationCardProps) {
  const { canEdit } = usePermission("owners");
  const { data, isLoading, isError } = useGetOwnerLineTags(ownerId);
  const { mutateAsync: updateLine } = useUpdateOwnerLine(ownerId);
  const { mutate: deleteLine, isPending: isDeletingLine } = useDeleteOwnerLine(ownerId);
  const { mutate: updateOptOut, isPending: isUpdatingOptOut } =
    useUpdateOwnerLstepOptOut(ownerId);

  const [tagAddDialogOpen, setTagAddDialogOpen] = useState(false);
  const [removeTagName, setRemoveTagName] = useState<string | null>(null);
  const [confirmUnlinkOpen, setConfirmUnlinkOpen] = useState(false);
  const [confirmOptOutOpen, setConfirmOptOutOpen] = useState(false);

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

  // フォーム成功後に入力フィールドをリセットする
  useEffect(() => {
    // success フラグはクエリ invalidate 後に自然にリセットされる
  }, [lineIdState.success]);

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

  // 連携済み + 配信停止中
  if (is_linked && lstep_opt_out) {
    return (
      <div className={`rounded-lg border ${C.borderLight} p-4 flex flex-col gap-4`}>
        <h3 className={`text-sm font-medium ${C.text70} uppercase tracking-wide`}>
          LINE / Lステップ連携
        </h3>

        {/* 配信停止バナー */}
        <div
          className={`flex items-center justify-between gap-3 rounded-md border ${C.borderNotice} ${C.bgNotice} px-4 py-3`}
        >
          <div className="flex items-center gap-2">
            <AlertTriangle
              className={`${ICON.smXs} ${C.textNotice} shrink-0`}
              style={{ color: PALETTE.noticeText }}
            />
            <span className={`text-sm font-medium ${C.textNotice}`}>
              配信停止中
            </span>
            {line_user_id !== null ? (
              <span className={`text-xs ${C.text50} font-mono`}>
                ({line_user_id.slice(0, 8)}...{line_user_id.slice(-4)})
              </span>
            ) : null}
          </div>

          {canEdit ? (
            <Button
              type="button"
              size="sm"
              variant="outline"
              className="h-8 px-3 text-xs shrink-0"
              disabled={isUpdatingOptOut}
              onClick={() => updateOptOut({ opt_out: false })}
            >
              {isUpdatingOptOut ? "処理中..." : "配信を再開"}
            </Button>
          ) : null}
        </div>

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
            {line_user_id !== null ? (
              <span className={`text-xs ${C.text50} font-mono`}>
                ({line_user_id.slice(0, 8)}...{line_user_id.slice(-4)})
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
            updateOptOut({ opt_out: true });
            setConfirmOptOutOpen(false);
          }}
          title="配信を停止しますか？"
          description={`${ownerName} さんへのLstep配信を停止します。`}
          confirmLabel="停止する"
          cancelLabel="キャンセル"
          isPending={isUpdatingOptOut}
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
    </div>
  );
}
