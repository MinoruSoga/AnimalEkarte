import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";

import type { MedicalRecordAutoCreateFailurePhase } from "../hooks/use-medical-record-auto-create";

interface MedicalRecordAutoCreateFailureProps {
  failurePhase: MedicalRecordAutoCreateFailurePhase;
  isRetrying: boolean;
  onRetry: () => void;
}

export function MedicalRecordAutoCreateFailure({
  failurePhase,
  isRetrying,
  onRetry,
}: MedicalRecordAutoCreateFailureProps) {
  const message =
    failurePhase === "appointment"
      ? "予約の作成に失敗しました。"
      : "カルテの作成に失敗しました。作成済みの予約は保持されています。";

  return (
    <div
      role="alert"
      className={`mx-4 mt-3 flex items-center justify-between gap-3 rounded border px-3 py-2 ${C.bgRed50} ${C.borderRed300} ${C.textRed700}`}
    >
      <p className="text-sm">{message}</p>
      <Button
        type="button"
        variant="outline"
        aria-label="カルテ作成を再試行する"
        disabled={isRetrying}
        onClick={onRetry}
        className="shrink-0"
      >
        {isRetrying ? "再試行中..." : "再試行する"}
      </Button>
    </div>
  );
}
