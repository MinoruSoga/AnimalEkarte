import type React from "react";
import { AlertTriangle, Info, Trash2, Wifi } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON } from "@/lib/design-tokens";

type TestResult = "success" | "error" | null;

interface LstepSyncSectionProps {
  isConfigured: boolean;
  isSyncEnabled: boolean;
  isPending: boolean;
  onDisableRequest: () => void;
  onSyncEnabledChange: (enabled: boolean) => void;
}

export function LstepSyncSection({
  isConfigured,
  isSyncEnabled,
  isPending,
  onDisableRequest,
  onSyncEnabledChange,
}: LstepSyncSectionProps) {
  return (
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
            onDisableRequest();
          } else {
            onSyncEnabledChange(next);
          }
        }}
        disabled={isPending}
        aria-label="同期を有効にする"
      />
    </div>
  );
}

interface LstepNoticeSectionProps {
  hasLineAccessToken: boolean;
}

export function LstepNoticeSection({ hasLineAccessToken }: LstepNoticeSectionProps) {
  return (
    <>
      <div className={`flex items-start gap-2 p-3 rounded-xs ${C.bgWarning50} border ${C.borderWarning20} text-sm ${C.textWarning}`}>
        <AlertTriangle className="shrink-0 mt-0.5 w-4 h-4" />
        <span>LINE Messaging API送信は LINE 公式アカウントの月間配信数にカウントされます。プランの上限に注意してください。</span>
      </div>

      {!hasLineAccessToken ? (
        <div className={`flex items-start gap-2 p-3 rounded-xs ${C.bgPage} border ${C.borderLight} text-sm ${C.text60}`}>
          <Info className="shrink-0 mt-0.5 w-4 h-4" />
          <span>LINE Messaging API トークンが未設定です。LINE公式アカウントのMessaging APIオプションを契約し、チャネルアクセストークンを入力してください。未設定の場合、LINE連携機能は動作しません。</span>
        </div>
      ) : null}
    </>
  );
}

interface LstepActionFooterProps {
  isConfigured: boolean;
  hasLineAccessToken: boolean;
  testResult: TestResult;
  lineTestResult: TestResult;
  isTesting: boolean;
  isLineTesting: boolean;
  deleteButtonRef: React.RefObject<HTMLButtonElement | null>;
  onTest: () => void;
  onLineTest: () => void;
  onDeleteOpen: () => void;
}

export function LstepActionFooter({
  isConfigured,
  hasLineAccessToken,
  testResult,
  lineTestResult,
  isTesting,
  isLineTesting,
  deleteButtonRef,
  onTest,
  onLineTest,
  onDeleteOpen,
}: LstepActionFooterProps) {
  return (
    <div className={`flex items-center justify-between pt-4 border-t ${C.borderLight}`}>
      <div className="flex flex-wrap items-center gap-2">
        {isConfigured ? (
          <Button
            type="button"
            variant="outline"
            onClick={onTest}
            disabled={isTesting}
            className="h-10 text-sm px-4"
          >
            <Wifi className={ICON.action} />
            {isTesting ? "テスト中..." : "Lステップ接続テスト"}
          </Button>
        ) : null}
        <TestResultLabel result={testResult} successLabel="L接続成功" errorLabel="L接続失敗" />

        {hasLineAccessToken ? (
          <Button
            type="button"
            variant="outline"
            onClick={onLineTest}
            disabled={isLineTesting}
            className="h-10 text-sm px-4"
          >
            <Wifi className={ICON.action} />
            {isLineTesting ? "テスト中..." : "LINE接続テスト"}
          </Button>
        ) : null}
        <TestResultLabel result={lineTestResult} successLabel="LINE接続成功" errorLabel="LINE接続失敗" />
      </div>

      <div className="flex items-center gap-2">
        {isConfigured ? (
          <Button
            ref={deleteButtonRef}
            type="button"
            variant="ghost-danger"
            onClick={onDeleteOpen}
            className={`h-10 text-sm px-4 border ${C.borderDanger}`}
          >
            <Trash2 className={ICON.action} />
            設定削除
          </Button>
        ) : null}
        <SubmitButton className="h-10 text-sm px-4">
          保存
        </SubmitButton>
      </div>
    </div>
  );
}

interface TestResultLabelProps {
  result: TestResult;
  successLabel: string;
  errorLabel: string;
}

function TestResultLabel({
  result,
  successLabel,
  errorLabel,
}: TestResultLabelProps) {
  if (result === "success") {
    return <span className={`text-sm ${C.textStatusGreen}`}>{successLabel}</span>;
  }
  if (result === "error") {
    return <span className={`text-sm ${C.danger}`}>{errorLabel}</span>;
  }
  return null;
}
