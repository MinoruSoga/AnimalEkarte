import { useTransition, useSyncExternalStore } from "react";
import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import {
  getClinicSelectionBlockReason,
  retryClinicSelectionRecovery,
  subscribeClinicSelectionBlock,
} from "@/lib/clinic-selection-recovery";

interface ClinicSelectionBlockedScreenProps {
  onLogout: () => Promise<void>;
}

export function ClinicSelectionBlockedScreen({ onLogout }: ClinicSelectionBlockedScreenProps) {
  const reason = useSyncExternalStore(
    subscribeClinicSelectionBlock,
    getClinicSelectionBlockReason,
    getClinicSelectionBlockReason,
  );
  const [isRetrying, startRetry] = useTransition();
  const [isLoggingOut, startLogout] = useTransition();

  if (reason === "none") {
    return null;
  }

  const title =
    reason === "no-clinic" ? "利用できる医院がありません" : "医院情報を再取得できませんでした";
  const description =
    reason === "no-clinic"
      ? "管理者に所属医院を確認してください。書き込みは停止しています。"
      : "書き込みは停止したままです。再試行するか、ログアウトしてください。";

  return (
    <div
      role="alertdialog"
      aria-labelledby="clinic-selection-blocked-title"
      aria-describedby="clinic-selection-blocked-description"
      className={`fixed inset-0 z-[100] flex items-center justify-center ${C.bgPage} p-4`}
    >
      <div
        className={`w-full max-w-md rounded-lg ${C.bgWhite} ${C.borderLight} border p-6 shadow-sm`}
      >
        <h1 id="clinic-selection-blocked-title" className={`text-lg font-semibold ${C.text90}`}>
          {title}
        </h1>
        <p id="clinic-selection-blocked-description" className={`mt-2 text-sm ${C.text60}`}>
          {description}
        </p>
        <div className="mt-6 flex flex-col gap-2">
          {reason === "recovery-failed" ? (
            <Button
              disabled={isRetrying || isLoggingOut}
              onClick={() => {
                startRetry(async () => {
                  await retryClinicSelectionRecovery();
                });
              }}
            >
              再試行
            </Button>
          ) : null}
          <Button
            variant={reason === "recovery-failed" ? "outline" : "default"}
            disabled={isLoggingOut}
            onClick={() => {
              startLogout(async () => {
                await onLogout();
              });
            }}
          >
            ログアウト
          </Button>
        </div>
      </div>
    </div>
  );
}
