// EMR-474: 予約の閲覧が 403 / 404 になった後も、未確定の送信の結果だけは照会できるようにする。
// 結果照会 route は現予約の閲覧権限と分離されている（technical-spec#revisit-intake-api-selected-v3）。
// ここでは回答本文・target の中身を出さず、送信が保存されたかの状態文言だけを表示する。
import { useEffect, useState } from "react";
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
  /**
   * 結果照会だけを出す理由。denied: 予約の閲覧が 403/404。read_only: 閲覧はできるが
   * current_target.editable が false（受付後など）。照会 route は編集可否と独立している。
   */
  reason?: "denied" | "read_only";
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
const IDLE_TEXT_READ_ONLY =
  "この問診は現在編集できませんが、直前の送信の結果が未確定です。結果だけ確認できます。";

// 401 でも未確定の key は store に残る。LINE の再起動は memory-only の key を失わせるため案内しない。
const UNAUTHORIZED_TEXT =
  "ログインの有効期限が切れました。送信結果が未確定のため、この画面を閉じずに、時間をおいてもう一度「送信結果を確認」してください。";

export function RevisitIntakeDeniedPendingLookup({
  clinicId,
  idToken,
  appointmentId,
  targetKey,
  pending,
  draftStore,
  reason = "denied",
}: RevisitIntakeDeniedPendingLookupProps) {
  const [state, setState] = useState<LookupState>({ kind: "idle" });
  const [availableAt, setAvailableAt] = useState<number | null>(
    pending.resendAvailableAt !== undefined &&
      pending.resendAvailableAt !== null &&
      pending.resendAvailableAt > Date.now()
      ? pending.resendAvailableAt
      : null,
  );

  // Retry-After の期限が来たら再び押せるようにする（Date.now() は reactive でないため）。
  useEffect(() => {
    if (availableAt === null) return;
    const timer = setTimeout(() => setAvailableAt(null), Math.max(0, availableAt - Date.now()));
    return () => clearTimeout(timer);
  }, [availableAt]);

  // editor と同じく store の pending を更新する（アクセスが戻ったとき editor の fail-closed 規則が効く）。
  function updatePending(patch: Partial<RevisitIntakePendingOperation>) {
    const entry = draftStore.get(targetKey);
    if (entry === undefined || entry.pending === null) return;
    draftStore.set(targetKey, { ...entry, pending: { ...entry.pending, ...patch } });
  }

  async function lookup() {
    if (availableAt !== null) return;
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
        const versions = op.committed_versions;
        if (
          (refs &&
            (refs.appointment_id !== appointmentId || refs.pet_id !== pending.body.pet_id)) ||
          (versions && versions.target_generation !== pending.body.target_generation)
        ) {
          setState({ kind: "message", text: MISMATCH_TEXT });
          return;
        }
        draftStore.clear(targetKey);
        setState({ kind: "succeeded" });
        return;
      }
      // not_found / claimed_uncommitted なら照会は済んだ（editor と同じく照会必須を解く）。
      updatePending(
        op.state === "result_unavailable"
          ? { status: op.state }
          : { status: op.state, lookupRequired: false },
      );
      setState(
        op.state === "not_found"
          ? { kind: "not_found" }
          : {
              kind: "message",
              text: "送信の保存はまだ確認できていません。時間をおいてもう一度確認してください。",
            },
      );
    } catch (err: unknown) {
      if (resolveFetchError(err, "送信結果の確認").status === 401) {
        // pending は store に残したまま、再起動を促さず待って再確認させる。
        setState({ kind: "message", text: UNAUTHORIZED_TEXT });
        return;
      }
      const retryAfter = revisitIntakeRetryAfterSeconds(err);
      if (retryAfter !== null) {
        const deadline = Date.now() + retryAfter * 1000;
        // 照会自体が 500 + Retry-After なら、期限後も成功する照会まで再送を許さない。
        updatePending(
          resolveFetchError(err, "送信結果の確認").status === 500
            ? { resendAvailableAt: deadline, lookupRequired: true }
            : { resendAvailableAt: deadline },
        );
        setAvailableAt(deadline);
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
        ? reason === "read_only"
          ? "送信の記録は見つかりませんでした。現在この問診は編集できないため、再送信はできません。医院へお問い合わせください。"
          : "送信の記録は見つかりませんでした。現在この予約を閲覧できないため、再送信はできません。医院へお問い合わせください。"
        : state.kind === "message"
          ? state.text
          : reason === "read_only"
            ? IDLE_TEXT_READ_ONLY
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
          disabled={state.kind === "busy" || availableAt !== null}
          className="min-h-11 w-full rounded-lg border border-noah-border-input bg-white px-4 text-sm font-semibold text-noah-teal-dark disabled:opacity-50"
        >
          {state.kind === "busy" ? "確認中..." : "送信結果を確認"}
        </button>
      )}
    </div>
  );
}
