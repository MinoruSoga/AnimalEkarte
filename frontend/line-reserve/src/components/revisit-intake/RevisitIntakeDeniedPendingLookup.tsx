// EMR-474: 予約の閲覧が 403 / 404 になった後も、未確定の送信の結果だけは照会できるようにする。
// 結果照会 route は現予約の閲覧権限と分離されている（technical-spec#revisit-intake-api-selected-v3）。
// ここでは回答本文・target の中身を出さず、送信が保存されたかの状態文言だけを表示する。
import { useState } from "react";
import { liffApi } from "../../api/liff-api";
import { resolveFetchError } from "@/shared-liff/handle-fetch-error";
import type {
  RevisitIntakeDraftStore,
  RevisitIntakePendingOperation,
} from "../../hooks/use-revisit-intake-drafts";
import { revisitIntakeRetryAfterSeconds } from "../../lib/revisit-intake-submit";

interface RevisitIntakeDeniedPendingLookupProps {
  clinicId: string;
  idToken: string;
  appointmentId: string;
  targetKey: string;
  pending: RevisitIntakePendingOperation;
  draftStore: RevisitIntakeDraftStore;
}

type LookupState =
  | { kind: "idle" }
  | { kind: "busy" }
  | { kind: "succeeded" }
  | { kind: "not_found" }
  | { kind: "message"; text: string };

const MISMATCH_TEXT = "送信結果を確認できませんでした。もう一度お試しください。";

const IDLE_TEXT =
  "この予約の問診は表示できませんが、直前の送信の結果が未確定です。結果だけ確認できます。";

export function RevisitIntakeDeniedPendingLookup({
  clinicId,
  idToken,
  appointmentId,
  targetKey,
  pending,
  draftStore,
}: RevisitIntakeDeniedPendingLookupProps) {
  const [state, setState] = useState<LookupState>({ kind: "idle" });
  const [availableAt, setAvailableAt] = useState<number | null>(pending.resendAvailableAt ?? null);

  async function lookup() {
    if (availableAt !== null && Date.now() < availableAt) return;
    setState({ kind: "busy" });
    try {
      const result = await liffApi.getRevisitIntakeOperation(clinicId, pending.key, idToken);
      const op = result.operation_result;
      if (op.operation_id !== pending.key) {
        setState({ kind: "message", text: MISMATCH_TEXT });
        return;
      }
      if (op.state === "succeeded") {
        const refs = op.result_refs;
        if (
          refs &&
          (refs.appointment_id !== appointmentId || refs.pet_id !== pending.body.pet_id)
        ) {
          setState({ kind: "message", text: MISMATCH_TEXT });
          return;
        }
        draftStore.clear(targetKey);
        setState({ kind: "succeeded" });
        return;
      }
      setState(
        op.state === "not_found"
          ? { kind: "not_found" }
          : {
              kind: "message",
              text: "送信の保存はまだ確認できていません。時間をおいてもう一度確認してください。",
            },
      );
    } catch (err: unknown) {
      const retryAfter = revisitIntakeRetryAfterSeconds(err);
      if (retryAfter !== null) {
        setAvailableAt(Date.now() + retryAfter * 1000);
        setState({
          kind: "message",
          text: `ただいま混み合っています。${retryAfter}秒ほど待ってから、送信結果を確認してください。`,
        });
        return;
      }
      setState({ kind: "message", text: resolveFetchError(err, "送信結果の確認").message });
    }
  }

  const text =
    state.kind === "succeeded"
      ? "回答の送信は完了しています。"
      : state.kind === "not_found"
        ? "送信の記録は見つかりませんでした。現在この予約を閲覧できないため、再送信はできません。医院へお問い合わせください。"
        : state.kind === "message"
          ? state.text
          : IDLE_TEXT;

  return (
    <div
      role="alert"
      className="space-y-3 rounded-lg border border-noah-warning-border bg-noah-warning-box-bg px-4 py-3"
    >
      <p className="text-sm text-noah-warning-text">{text}</p>
      {state.kind === "succeeded" ? null : (
        <button
          type="button"
          onClick={() => void lookup()}
          disabled={state.kind === "busy" || (availableAt !== null && Date.now() < availableAt)}
          className="min-h-11 w-full rounded-lg border border-noah-border-input bg-white px-4 text-sm font-semibold text-noah-teal-dark disabled:opacity-50"
        >
          {state.kind === "busy" ? "確認中..." : "送信結果を確認"}
        </button>
      )}
    </div>
  );
}
