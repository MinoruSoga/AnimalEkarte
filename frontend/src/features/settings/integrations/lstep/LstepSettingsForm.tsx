import { useActionState, useState, useRef } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { INITIAL_ACTION_STATE } from "@/types/form";
import type { ActionState } from "@/types/form";
import {
  useGetLstepSettings,
  useUpdateLstepSettings,
  useTestLstepConnection,
  useTestLineMessagingConnection,
  useDeleteLstepSettings,
} from "./hooks/useLstepSettings";
import type { LstepSettingsRequest } from "./hooks/useLstepSettings";
import { Trash2, Wifi, AlertTriangle, Info } from "lucide-react";

// ─────────────────────────────────────────────────
// Subcomponents
// ─────────────────────────────────────────────────

function PasswordField({
  id,
  name,
  label,
  placeholder,
}: {
  id: string;
  name: string;
  label: string;
  placeholder?: string;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={STYLE.formLabel}>
        {label}
      </label>
      <input
        id={id}
        type="password"
        name={name}
        placeholder={placeholder ?? "新しい値を入力（空欄で変更なし）"}
        autoComplete="new-password"
        className={`${STYLE.formInput} rounded-[4px] border px-3 w-full outline-none focus:ring-2 focus:ring-[#2383E2]/30`}
      />
    </div>
  );
}

function TextField({
  id,
  name,
  label,
  defaultValue,
  placeholder,
}: {
  id: string;
  name: string;
  label: string;
  defaultValue?: string;
  placeholder?: string;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={STYLE.formLabel}>
        {label}
      </label>
      <input
        id={id}
        type="text"
        name={name}
        defaultValue={defaultValue ?? ""}
        placeholder={placeholder ?? ""}
        className={`${STYLE.formInput} rounded-[4px] border px-3 w-full outline-none focus:ring-2 focus:ring-[#2383E2]/30`}
      />
    </div>
  );
}

function CPMVersionSelectField({
  id,
  name,
  label,
  defaultValue,
}: {
  id: string;
  name: string;
  label: string;
  defaultValue: string;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={STYLE.formLabel}>
        {label}
      </label>
      <select
        id={id}
        name={name}
        defaultValue={defaultValue}
        className={`${STYLE.formInput} rounded-[4px] border px-3 w-full outline-none focus:ring-2 focus:ring-[#2383E2]/30`}
      >
        <option value="v1">V1</option>
        <option value="v2">V2</option>
      </select>
    </div>
  );
}

function NumberInputField({
  id,
  name,
  label,
  defaultValue,
}: {
  id: string;
  name: string;
  label: string;
  defaultValue: number;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={STYLE.formLabel}>
        {label}
      </label>
      <input
        id={id}
        type="number"
        name={name}
        defaultValue={defaultValue}
        min={1}
        className={`${STYLE.formInput} rounded-[4px] border px-3 w-full outline-none focus:ring-2 focus:ring-[#2383E2]/30`}
      />
    </div>
  );
}

function formatSyncDate(iso: string): string {
  return new Date(iso).toLocaleDateString("ja-JP", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

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
      const req: LstepSettingsRequest = {};

      const apiKey = formData.get("lstep_api_key");
      if (typeof apiKey === "string" && apiKey.trim() !== "") {
        req.lstep_api_key = apiKey.trim();
      }
      const token = formData.get("line_channel_access_token");
      if (typeof token === "string" && token.trim() !== "") {
        req.line_channel_access_token = token.trim();
      }
      const secret = formData.get("line_channel_secret");
      if (typeof secret === "string" && secret.trim() !== "") {
        req.line_channel_secret = secret.trim();
      }
      const liffId = formData.get("liff_id");
      if (typeof liffId === "string") {
        req.liff_id = liffId.trim();
      }
      const lstepBaseUrl = formData.get("lstep_base_url");
      if (typeof lstepBaseUrl === "string" && lstepBaseUrl.trim() !== "") {
        req.lstep_base_url = lstepBaseUrl.trim();
      }

      const cpmVersion = formData.get("cpm_version");
      if (cpmVersion === "v1" || cpmVersion === "v2") {
        req.cpm_version = cpmVersion;
      }

      const dormant180 = formData.get("dormant_prevention_180_days");
      if (typeof dormant180 === "string" && dormant180 !== "") {
        const v = parseInt(dormant180, 10);
        if (!isNaN(v) && v >= 1) req.dormant_prevention_180_days = v;
      }
      const dormant210 = formData.get("dormant_prevention_210_days");
      if (typeof dormant210 === "string" && dormant210 !== "") {
        const v = parseInt(dormant210, 10);
        if (!isNaN(v) && v >= 1) req.dormant_prevention_210_days = v;
      }
      const dormant240 = formData.get("dormant_prevention_240_days");
      if (typeof dormant240 === "string" && dormant240 !== "") {
        const v = parseInt(dormant240, 10);
        if (!isNaN(v) && v >= 1) req.dormant_prevention_240_days = v;
      }
      const dormant365 = formData.get("dormant_prevention_365_days");
      if (typeof dormant365 === "string" && dormant365 !== "") {
        const v = parseInt(dormant365, 10);
        if (!isNaN(v) && v >= 1) req.dormant_prevention_365_days = v;
      }

      // 同期ON/OFFは常に送信（boolean は空欄という概念がないため）
      req.is_sync_enabled = isSyncEnabled;

      try {
        await updateMutation.mutateAsync(req);
        setSyncEnabledOverride(undefined); // 保存後はサーバー値に戻す
        toast.success("Lステップ設定を保存しました");
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "Lステップ設定の保存");
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
    } catch (error) {
      setTestResult("error");
      handleApiError(error, "Lステップ接続テスト");
    }
  };

  const handleLineTest = async () => {
    setLineTestResult(null);
    try {
      await lineTestMutation.mutateAsync();
      setLineTestResult("success");
      toast.success("LINE Messaging API接続テスト成功");
    } catch (error) {
      setLineTestResult("error");
      handleApiError(error, "LINE Messaging API接続テスト");
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
      {/* 設定状態・同期状態バッジ */}
      <div className="flex items-center justify-between mb-6">
        <h2 className={`text-base font-semibold ${C.text}`}>
          Lステップ連携設定
        </h2>
        <div className="flex items-center gap-2">
          {isConfigured ? (
            <span className="inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full bg-[#DDEDEA] text-[#0F7B6C] border border-[#C3DFC3]">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-[#0F7B6C]" />
              設定済み
            </span>
          ) : (
            <span className={`inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full ${C.bgStatusGray} ${C.textStatusGray} border ${C.borderMuted}`}>
              <span className={`inline-block w-1.5 h-1.5 rounded-full ${C.bgStatusGrayMedium}`} />
              未設定
            </span>
          )}
          {isConfigured ? (
            isSyncEnabled ? (
              <span className="inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full bg-[#DDEDEA] text-[#0F7B6C] border border-[#C3DFC3]">
                <span className="inline-block w-1.5 h-1.5 rounded-full bg-[#0F7B6C]" />
                同期中
              </span>
            ) : (
              <span className={`inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full ${C.bgStatusGray} ${C.textStatusGray} border ${C.borderMuted}`}>
                <span className={`inline-block w-1.5 h-1.5 rounded-full ${C.bgStatusGrayMedium}`} />
                同期停止中
              </span>
            )
          ) : null}
        </div>
      </div>

      {/* 既存マスク値 + 同期有効化日時の表示 */}
      {isConfigured ? (
        <div className={`mb-6 p-3 rounded-[4px] ${C.bgPage} border ${C.borderLight} text-sm ${C.text60} flex flex-col gap-1`}>
          {settings?.lstep_api_key_masked ? (
            <span>Lステップ APIキー: <span className={`font-mono ${C.text}`}>••••{settings.lstep_api_key_masked}</span></span>
          ) : null}
          {settings?.line_channel_access_token_masked ? (
            <span>LINE Channel Access Token: <span className={`font-mono ${C.text}`}>••••{settings.line_channel_access_token_masked}</span></span>
          ) : null}
          {settings?.line_channel_secret_masked ? (
            <span>LINE Channel Secret: <span className={`font-mono ${C.text}`}>••••{settings.line_channel_secret_masked}</span></span>
          ) : null}
          {settings?.liff_id ? (
            <span>LIFF ID: <span className={`font-mono ${C.text}`}>{settings.liff_id}</span></span>
          ) : null}
          {settings?.lstep_base_url ? (
            <span>LステップベースURL: <span className={`font-mono ${C.text}`}>{settings.lstep_base_url}</span></span>
          ) : null}
          {settings?.sync_enabled_at ? (
            <span>同期有効化日時: <span className={`font-mono ${C.text}`}>{formatSyncDate(settings.sync_enabled_at)}</span></span>
          ) : null}
        </div>
      ) : null}

      {/* フォーム */}
      <form action={formAction} className="flex flex-col gap-4">
        <PasswordField
          id="lstep_api_key"
          name="lstep_api_key"
          label="Lステップ APIキー"
          placeholder={settings?.lstep_api_key_masked ? `現在: ••••${settings.lstep_api_key_masked}（変更する場合のみ入力）` : "Lステップ APIキーを入力"}
        />
        <PasswordField
          id="line_channel_access_token"
          name="line_channel_access_token"
          label="LINE Channel Access Token"
          placeholder={settings?.line_channel_access_token_masked ? `現在: ••••${settings.line_channel_access_token_masked}（変更する場合のみ入力）` : "LINE Channel Access Tokenを入力"}
        />
        <PasswordField
          id="line_channel_secret"
          name="line_channel_secret"
          label="LINE Channel Secret"
          placeholder={settings?.line_channel_secret_masked ? `現在: ••••${settings.line_channel_secret_masked}（変更する場合のみ入力）` : "LINE Channel Secretを入力"}
        />
        <TextField
          id="liff_id"
          name="liff_id"
          label="LIFF ID"
          defaultValue={settings?.liff_id ?? ""}
          placeholder="例: 1234567890-xxxxxxxx"
        />
        <TextField
          id="lstep_base_url"
          name="lstep_base_url"
          label="LステップベースURL"
          defaultValue={settings?.lstep_base_url ?? ""}
          placeholder="例: https://app.lstep.jp"
        />

        <CPMVersionSelectField
          id="cpm_version"
          name="cpm_version"
          label="CPMバージョン"
          defaultValue={settings?.cpm_version ?? "v1"}
        />

        <NumberInputField
          id="dormant_prevention_180_days"
          name="dormant_prevention_180_days"
          label="休眠予防閾値 (180日)"
          defaultValue={settings?.dormant_prevention_180_days ?? 180}
        />
        <NumberInputField
          id="dormant_prevention_210_days"
          name="dormant_prevention_210_days"
          label="休眠予防閾値 (210日)"
          defaultValue={settings?.dormant_prevention_210_days ?? 210}
        />
        <NumberInputField
          id="dormant_prevention_240_days"
          name="dormant_prevention_240_days"
          label="休眠予防閾値 (240日)"
          defaultValue={settings?.dormant_prevention_240_days ?? 240}
        />
        <NumberInputField
          id="dormant_prevention_365_days"
          name="dormant_prevention_365_days"
          label="休眠予防閾値 (365日)"
          defaultValue={settings?.dormant_prevention_365_days ?? 365}
        />

        {/* 同期 ON/OFF トグル */}
        <div className={`flex items-center justify-between gap-4 py-3 border-t ${C.borderLight}`}>
          <div className="flex flex-col gap-0.5">
            <span className={`text-sm font-medium ${C.text}`}>同期を有効にする</span>
            <span className={`text-xs ${C.text50}`}>
              {isConfigured
                ? "OFFにすると連携処理を一時停止します。APIキー設定は保持されます。"
                : "APIキー保存後に同期を開始する場合はONにします。OFFの場合もAPIキー設定は保持されます。"}
            </span>
          </div>
          <Switch
            checked={isSyncEnabled}
            onCheckedChange={(next: boolean) => {
              if (isSyncEnabled && !next) {
                setShowDisableSyncConfirm(true);
              } else {
                setSyncEnabledOverride(next);
              }
            }}
            disabled={updateMutation.isPending}
            aria-label="同期を有効にする"
          />
        </div>

        {/* 月間配信数の注意 */}
        <div className={`flex items-start gap-2 p-3 rounded-[4px] bg-[#FFF9E6] border border-[#F0D070] text-sm text-[#7A5C00]`}>
          <AlertTriangle className="shrink-0 mt-0.5 w-4 h-4" />
          <span>LINE Messaging API送信は LINE 公式アカウントの月間配信数にカウントされます。プランの上限に注意してください。</span>
        </div>

        {/* APIオプション未加入注意 */}
        {!settings?.line_channel_access_token_masked ? (
          <div className={`flex items-start gap-2 p-3 rounded-[4px] bg-[#F4F4F4] border ${C.borderLight} text-sm ${C.text60}`}>
            <Info className="shrink-0 mt-0.5 w-4 h-4" />
            <span>LINE Messaging API トークンが未設定です。LINE公式アカウントのMessaging APIオプションを契約し、チャネルアクセストークンを入力してください。未設定の場合、LINE連携機能は動作しません。</span>
          </div>
        ) : null}

        <div className={`flex items-center justify-between pt-4 border-t ${C.borderLight}`}>
          <div className="flex flex-wrap items-center gap-2">
            {/* Lステップ接続テスト */}
            {isConfigured ? (
              <Button
                type="button"
                variant="outline"
                onClick={handleTest}
                disabled={testMutation.isPending}
                className="h-10 text-sm px-4"
              >
                <Wifi className={ICON.action} />
                {testMutation.isPending ? "テスト中..." : "Lステップ接続テスト"}
              </Button>
            ) : null}
            {testResult === "success" ? (
              <span className="text-sm text-[#0F7B6C]">L接続成功</span>
            ) : null}
            {testResult === "error" ? (
              <span className={`text-sm ${C.danger}`}>L接続失敗</span>
            ) : null}

            {/* LINE Messaging API 接続テスト */}
            {settings?.line_channel_access_token_masked ? (
              <Button
                type="button"
                variant="outline"
                onClick={handleLineTest}
                disabled={lineTestMutation.isPending}
                className="h-10 text-sm px-4"
              >
                <Wifi className={ICON.action} />
                {lineTestMutation.isPending ? "テスト中..." : "LINE接続テスト"}
              </Button>
            ) : null}
            {lineTestResult === "success" ? (
              <span className="text-sm text-[#0F7B6C]">LINE接続成功</span>
            ) : null}
            {lineTestResult === "error" ? (
              <span className={`text-sm ${C.danger}`}>LINE接続失敗</span>
            ) : null}
          </div>

          <div className="flex items-center gap-2">
            {/* 削除ボタン */}
            {isConfigured ? (
              <Button
                ref={deleteButtonRef}
                type="button"
                variant="ghost-danger"
                onClick={() => setIsDeleteOpen(true)}
                className={`h-10 text-sm px-4 border ${C.borderDanger}`}
              >
                <Trash2 className={ICON.action} />
                設定削除
              </Button>
            ) : null}
            <SubmitButton className="h-10 text-sm px-5">
              保存
            </SubmitButton>
          </div>
        </div>
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
