// EMR-474: 再診事前問診の未送信 draft を target 別に memory だけで保持する。
// 技術契約（technical-spec#revisit-intake-api-selected-v3）: draft は target 別 memory のみ、
// URL / localStorage / sessionStorage / server へ保存しない。再読込・LINE 終了で失われる（OS 終了保証なし）。
// target key は clinic|appointment|pet|target_generation。別 target（取り直し・対象変更・A→B→A）の
// draft は別 key になり、旧 target の入力を新 target へコピーしない。
import { useMemo, useRef } from "react";
import type {
  RevisitIntakeFullAnswers,
  RevisitIntakeSubmitRequest,
  RevisitIntakeTargetRef,
} from "../types/models";

/**
 * 結果が確定していない送信。同じ key・同じ内容でのみ再送／照会する。
 * - unknown: 応答喪失・5xx・429 等で結果を確認できていない
 * - claimed_uncommitted / not_found: 同 key で再送してよい（失敗確定ではない）
 * - result_unavailable: 成功/失敗を判定できない。新しい送信を作らない
 */
export interface RevisitIntakePendingOperation {
  key: string;
  fingerprint: string;
  body: RevisitIntakeSubmitRequest;
  status: "unknown" | "claimed_uncommitted" | "not_found" | "result_unavailable";
  /**
   * 429 / Retry-After 付き 503 の再送期限（epoch ms）。draft と一緒に保持し、画面を開き直しても
   * 期限前に「同じ内容で再送信」を有効にしない。期限切れは null と同じ扱い。
   */
  resendAvailableAt?: number | null;
}

export interface RevisitIntakeDraftEntry {
  values: RevisitIntakeFullAnswers;
  /**
   * draft の元になった最新回答の識別（head_revision|content_available|answer_id）。再開時に現在の
   * 値と違えば、古い回答を元にした入力として明示確認まで送信させない。
   */
  baselineToken?: string;
  pending: RevisitIntakePendingOperation | null;
}

export interface RevisitIntakeDraftStore {
  get: (targetKey: string) => RevisitIntakeDraftEntry | undefined;
  set: (targetKey: string, entry: RevisitIntakeDraftEntry) => void;
  clear: (targetKey: string) => void;
  /**
   * 指定予約に未確定の送信があれば返す。予約の GET が 403 / 404 になった後でも、key を失わずに
   * 結果照会へ戻れるようにするための索引（target の ref を読めない状況で使う）。
   */
  findPending: (
    clinicId: string,
    appointmentId: string,
  ) => { targetKey: string; pending: RevisitIntakePendingOperation } | undefined;
}

export function revisitIntakeTargetKey(clinicId: string, ref: RevisitIntakeTargetRef): string {
  return [clinicId, ref.appointment_id, ref.pet_id, ref.target_generation].join("|");
}

/** target key（clinic|appointment|pet|generation）の先頭 2 要素で、予約単位の未確定送信を探す。 */
export function findPendingEntry(
  entries: ReadonlyMap<string, RevisitIntakeDraftEntry>,
  clinicId: string,
  appointmentId: string,
): { targetKey: string; pending: RevisitIntakePendingOperation } | undefined {
  const prefix = `${clinicId}|${appointmentId}|`;
  for (const [targetKey, entry] of entries) {
    if (targetKey.startsWith(prefix) && entry.pending !== null) {
      return { targetKey, pending: entry.pending };
    }
  }
  return undefined;
}

export const EMPTY_REVISIT_INTAKE_ANSWERS: RevisitIntakeFullAnswers = {
  visit_purpose: "",
  onset: null,
  energy: null,
  appetite: null,
  current_symptoms: null,
  improvement: null,
  examination_preference: null,
  main_concern: null,
};

export function areRevisitIntakeAnswersEqual(
  a: RevisitIntakeFullAnswers,
  b: RevisitIntakeFullAnswers,
): boolean {
  return (
    a.visit_purpose === b.visit_purpose &&
    a.onset === b.onset &&
    a.energy === b.energy &&
    a.appetite === b.appetite &&
    a.current_symptoms === b.current_symptoms &&
    a.improvement === b.improvement &&
    a.examination_preference === b.examination_preference &&
    a.main_concern === b.main_concern
  );
}

/**
 * App 単位で1つだけ持つ memory store。ページ遷移（予約一覧へ戻る等）をまたいで同じ target の
 * draft を復元するため、ページではなく App で生成して渡す。
 */
export function useRevisitIntakeDraftStore(): RevisitIntakeDraftStore {
  const mapRef = useRef<Map<string, RevisitIntakeDraftEntry> | null>(null);
  return useMemo<RevisitIntakeDraftStore>(() => {
    const map = (): Map<string, RevisitIntakeDraftEntry> => {
      if (mapRef.current === null) mapRef.current = new Map();
      return mapRef.current;
    };
    return {
      get: (targetKey) => map().get(targetKey),
      set: (targetKey, entry) => {
        map().set(targetKey, entry);
      },
      clear: (targetKey) => {
        map().delete(targetKey);
      },
      findPending: (clinicId, appointmentId) => findPendingEntry(map(), clinicId, appointmentId),
    };
  }, []);
}
