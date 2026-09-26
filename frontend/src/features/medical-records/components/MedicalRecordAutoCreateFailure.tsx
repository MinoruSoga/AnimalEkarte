import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";

import type { MedicalRecordAutoCreateFailurePhase } from "../hooks/use-medical-record-auto-create";

interface MedicalRecordAutoCreateFailureProps {
  failurePhase: MedicalRecordAutoCreateFailurePhase;
  isRetrying: boolean;
  onRetry: () => void;
}

const MESSAGES: Record<MedicalRecordAutoCreateFailurePhase, string> = {
  // BUG-MR-DRAFT-AUTOPOST-FAILED: master 欠落は再試行では解決しない（reservation-types は
  // staleTime STATIC のため追加後も同一セッションでは再解決されない）。復旧導線を明示する。
  "appointment-master-missing":
    "予約区分マスタに診察系の予約区分が登録されていません。マスタ設定 → 予約区分 で診察区分を追加した後、ページを再読み込みしてください。",
  appointment: "予約の作成に失敗しました。",
  "medical-record": "カルテの作成に失敗しました。作成済みの予約は保持されています。",
};

export function MedicalRecordAutoCreateFailure({
  failurePhase,
  isRetrying,
  onRetry,
}: MedicalRecordAutoCreateFailureProps) {
  const message = MESSAGES[failurePhase];
  const canRetry = failurePhase !== "appointment-master-missing";

  return (
    <div
      role="alert"
      className={`mx-4 mt-3 flex items-center justify-between gap-3 rounded border px-3 py-2 ${C.bgRed50} ${C.borderRed300} ${C.textRed700}`}
    >
      <p className="text-sm">{message}</p>
      {canRetry ? (
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
      ) : null}
    </div>
  );
}
