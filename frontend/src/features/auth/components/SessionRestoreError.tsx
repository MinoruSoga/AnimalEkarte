import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";

const DEFAULT_MESSAGE = "ログイン状態を確認できませんでした";
const RESTRICTED_MESSAGE =
  "このアカウントはアクセスが制限されています。管理者に権限を確認してください。";

interface SessionRestoreErrorProps {
  kind?: "transport" | "restricted";
  message?: string;
  onRetry: () => void;
  onSwitchToLogin: () => void;
}

/** Hook-free restore failure shell. role=alert; retry and manual-login only. */
export function SessionRestoreError({
  kind = "transport",
  message,
  onRetry,
  onSwitchToLogin,
}: SessionRestoreErrorProps) {
  const copy = message ?? (kind === "restricted" ? RESTRICTED_MESSAGE : DEFAULT_MESSAGE);
  return (
    <div
      className={`min-h-screen flex flex-col items-center justify-center px-4 ${C.bgPage}`}
      role="alert"
      aria-live="assertive"
      aria-atomic="true"
    >
      <div
        className={`inline-flex items-center justify-center size-[48px] rounded-xl mb-4 ${C.bgBrandIdentity}`}
      >
        <Stethoscope className={`size-[26px] ${C.textWhite}`} aria-hidden="true" />
      </div>
      <h1 className={`text-heading-3 font-bold leading-tight ${C.text} mb-1`}>ノア動物病院</h1>
      <p className={`text-base ${C.text50} mb-6`}>{copy}</p>
      <div className="flex w-full max-w-[380px] flex-col gap-2">
        <Button type="button" onClick={onRetry}>
          再試行
        </Button>
        <Button type="button" variant="outline" onClick={onSwitchToLogin}>
          ログイン切替
        </Button>
      </div>
    </div>
  );
}
