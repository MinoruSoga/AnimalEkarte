import { useActionState, useState, useRef } from "react";
import { toast } from "sonner";
import { C, STYLE } from "@/lib/design-tokens";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { INITIAL_ACTION_STATE } from "@/types/form";
import type { ActionState } from "@/types/form";
import {
  useGetLstepSettings,
  useUpdateLstepSettings,
  useTestLstepConnection,
  useTestLineMessagingConnection,
  useDeleteLstepSettings,
} from "../hooks/use-lstep-settings";
import { buildLstepSettingsRequest } from "../lib/lstep-settings-form-request";
import {
  LstepActionFooter,
  LstepConfiguredSummary,
  LstepConnectionFields,
  LstepCpmFields,
  LstepNoticeSection,
  LstepPreventionFields,
  LstepSettingsStatusHeader,
  LstepSyncSection,
} from "./LstepSettingsFormSections";

// ─────────────────────────────────────────────────
// Main component
// ─────────────────────────────────────────────────

export function LstepSettingsForm() {
  const { data: settings, isLoading } = useGetLstepSettings();
  const updateMutation = useUpdateLstepSettings();
  const testMutation = useTestLstepConnection();
  const lineTestMutation = useTestLineMessagingConnection();
  const deleteMutation = useDeleteLstepSettings();

  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [testResult, setTestResult] = useState<"success" | "error" | null>(null);
  const [lineTestResult, setLineTestResult] = useState<"success" | "error" | null>(null);
  const deleteButtonRef = useRef<HTMLButtonElement>(null);

  // undefined = サーバー値に追従。ユーザーがトグルを変更したら上書きする。
  const [syncEnabledOverride, setSyncEnabledOverride] = useState<boolean | undefined>(undefined);
  const isSyncEnabled = syncEnabledOverride ?? settings?.is_sync_enabled ?? false;
  const [showDisableSyncConfirm, setShowDisableSyncConfirm] = useState(false);

  const [_formState, formAction] = useActionState(
    async (_prevState: ActionState, formData: FormData): Promise<ActionState> => {
      const req = buildLstepSettingsRequest(formData, isSyncEnabled);

      try {
        await updateMutation.mutateAsync(req);
        setSyncEnabledOverride(undefined); // 保存後はサーバー値に戻す
        toast.success("Lステップ設定を保存しました");
        return { success: true, timestamp: Date.now() };
      } catch {
        // FE-RC-005: API エラーは useUpdateLstepSettings の onError
        // （use-lstep-settings.ts）が handleApiError 済み。ここでは再通知しない。
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE,
  );

  const handleTest = async () => {
    setTestResult(null);
    try {
      await testMutation.mutateAsync();
      setTestResult("success");
      toast.success("Lステップ接続テスト成功");
    } catch {
      // FE-RC-005: API エラーは useTestLstepConnection の onError が handleApiError 済み。
      setTestResult("error");
    }
  };

  const handleLineTest = async () => {
    setLineTestResult(null);
    try {
      await lineTestMutation.mutateAsync();
      setLineTestResult("success");
      toast.success("LINE Messaging API接続テスト成功");
    } catch {
      // FE-RC-005: API エラーは useTestLineMessagingConnection の onError が handleApiError 済み。
      setLineTestResult("error");
    }
  };

  const handleDeleteConfirm = () => {
    deleteMutation.mutate(undefined, {
      onSuccess: () => {
        toast.success("Lステップ設定を削除しました");
        setIsDeleteOpen(false);
      },
    });
  };

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center py-16 text-sm ${C.text50}`}>
        読み込み中...
      </div>
    );
  }

  const isConfigured = settings?.is_configured ?? false;

  return (
    <div className={`${STYLE.formCard} max-w-2xl`}>
      <LstepSettingsStatusHeader isConfigured={isConfigured} isSyncEnabled={isSyncEnabled} />
      <LstepConfiguredSummary settings={settings} isConfigured={isConfigured} />
      <form action={formAction} className="flex flex-col gap-4">
        <LstepConnectionFields settings={settings} />
        <LstepCpmFields settings={settings} />
        <LstepPreventionFields settings={settings} />
        <LstepSyncSection
          isConfigured={isConfigured}
          isSyncEnabled={isSyncEnabled}
          isPending={updateMutation.isPending}
          onDisableRequest={() => setShowDisableSyncConfirm(true)}
          onSyncEnabledChange={setSyncEnabledOverride}
        />
        <LstepNoticeSection
          hasLineAccessToken={Boolean(settings?.line_channel_access_token_masked)}
        />
        <LstepActionFooter
          isConfigured={isConfigured}
          hasLineAccessToken={Boolean(settings?.line_channel_access_token_masked)}
          testResult={testResult}
          lineTestResult={lineTestResult}
          isTesting={testMutation.isPending}
          isLineTesting={lineTestMutation.isPending}
          deleteButtonRef={deleteButtonRef}
          onTest={handleTest}
          onLineTest={handleLineTest}
          onDeleteOpen={() => setIsDeleteOpen(true)}
        />
      </form>

      {/* 削除確認ダイアログ */}
      <ConfirmDialog
        open={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        onConfirm={handleDeleteConfirm}
        title="Lステップ設定を削除しますか？"
        description="Lステップ連携設定をすべて削除します。LINE連携機能が無効になります。この操作は元に戻せません。"
        confirmLabel={deleteMutation.isPending ? "削除中..." : "削除する"}
        cancelLabel="キャンセル"
        variant="destructive"
        isPending={deleteMutation.isPending}
        triggerRef={deleteButtonRef}
      />

      {/* 同期無効化確認ダイアログ */}
      {/* data-testid="lstep-disable-sync-confirm-dialog" */}
      <ConfirmDialog
        open={showDisableSyncConfirm}
        onClose={() => setShowDisableSyncConfirm(false)}
        onConfirm={() => {
          setSyncEnabledOverride(false);
          setShowDisableSyncConfirm(false);
        }}
        title="同期を無効にしますか？"
        description="同期を無効にすると Lステップへのタグ付与が停止します。よろしいですか？"
        confirmLabel="無効にする"
        cancelLabel="キャンセル"
      />
    </div>
  );
}
