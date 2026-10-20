// React/Framework
import { useInfiniteQuery } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";

// Relative
import {
  REVISIT_INTAKE_ENERGY_APPETITE_LABELS,
  REVISIT_INTAKE_EXAMINATION_PREFERENCE_LABELS,
  REVISIT_INTAKE_IMPROVEMENT_LABELS,
} from "../lib/revisit-intake-options";

// EMR-477: カルテ固定 target の飼主事前問診（再診問診）原文 reader。
// 契約: backend/docs/api.yaml GET /medical-records/{recordId}/revisit-intake
//   → RevisitIntakeRecordRead（binding: unbound | mismatch | bound）。
// 純 GET。binding / target / answer を作らず変えない。書き込み API はこのファイルに置かない。
// runtime は未実装（EMR-476 の固定 binding capture/reader、EMR-472 の認可に依存）。

// ── Types (api.yaml の手書き写像。generated 型は未生成) ─────────────────

export interface RevisitIntakeAnswers {
  visit_purpose: string;
  onset: string | null;
  energy: string | null;
  appetite: string | null;
  current_symptoms: string | null;
  improvement: string | null;
  examination_preference: string | null;
  main_concern: string | null;
}

export interface RevisitIntakeAnswer {
  answer_id: string;
  revision: string;
  form_definition_version: number;
  received_at: string;
  provenance: "owner_submission";
  answers: RevisitIntakeAnswers;
}

export interface RevisitIntakePersistedTargetRef {
  target_id: string;
  appointment_id: string;
  pet_id: string;
  target_generation: string;
}

export interface RevisitIntakeBoundTarget {
  ref: RevisitIntakePersistedTargetRef;
  head_revision: string | null;
  answer_state: "unanswered" | "answered";
  latest_answer: RevisitIntakeAnswer | null;
  content_available: boolean;
}

export interface RevisitIntakeHistoryEntry {
  target: RevisitIntakePersistedTargetRef;
  answer: RevisitIntakeAnswer;
  target_relation: "current" | "previous";
}

export type RevisitIntakeRecordRead =
  | { binding: "unbound"; target: null; history: []; next_cursor: null }
  | { binding: "mismatch"; target: null; history: []; next_cursor: null }
  | {
      binding: "bound";
      target: RevisitIntakeBoundTarget;
      history: RevisitIntakeHistoryEntry[];
      next_cursor: string | null;
    };

// ── Parse（想定外の形は取得失敗として扱い、本文を推測表示しない） ─────────

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

// 契約 RevisitIntakeDecimalId: 正の正規形十進文字列（先頭ゼロ・空・符号・小数不可）で
// BIGINT 範囲内（1..9223372036854775807）。JS Number へは変換しない。
const MAX_BIGINT_DECIMAL = "9223372036854775807";

export function isCanonicalDecimalId(value: unknown): value is string {
  if (typeof value !== "string" || !/^[1-9][0-9]*$/.test(value)) return false;
  if (value.length < MAX_BIGINT_DECIMAL.length) return true;
  return value.length === MAX_BIGINT_DECIMAL.length && value <= MAX_BIGINT_DECIMAL;
}

function compareDecimal(a: string, b: string): number {
  // 正規形十進文字列同士。JS Number へ変換せず、桁数 → 辞書順で比較する。
  if (a.length !== b.length) return a.length < b.length ? -1 : 1;
  if (a === b) return 0;
  return a < b ? -1 : 1;
}

function sameRef(a: RevisitIntakePersistedTargetRef, b: RevisitIntakePersistedTargetRef): boolean {
  return (
    a.target_id === b.target_id &&
    a.appointment_id === b.appointment_id &&
    a.pet_id === b.pet_id &&
    a.target_generation === b.target_generation
  );
}

function parseRef(value: unknown): RevisitIntakePersistedTargetRef {
  if (
    !isRecord(value) ||
    !isCanonicalDecimalId(value.target_id) ||
    !isCanonicalDecimalId(value.appointment_id) ||
    !isCanonicalDecimalId(value.pet_id) ||
    !isCanonicalDecimalId(value.target_generation)
  ) {
    throw new Error("revisit-intake: invalid target ref");
  }
  return {
    target_id: value.target_id,
    appointment_id: value.appointment_id,
    pet_id: value.pet_id,
    target_generation: value.target_generation,
  };
}

// 契約: 任意項目は key 必須で null か値。欠落(undefined)・型違いは「未回答」に見せず契約失敗にする。
function nullableText(value: unknown): string | null {
  if (value === null || typeof value === "string") return value;
  throw new Error("revisit-intake: invalid answer");
}

function nullableChoice(value: unknown, labels: Readonly<Record<string, string>>): string | null {
  if (value === null) return null;
  if (typeof value === "string" && Object.hasOwn(labels, value)) return value;
  throw new Error("revisit-intake: invalid answer");
}

// RFC3339（UTC offset 必須）。各成分を暦・時刻の範囲で検証する。
// UTC と local の暦日は比較しない（offset 付きの正当な値は UTC で別日になりうる）。
const RFC3339_WITH_OFFSET =
  /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d+)?(Z|[+-](\d{2}):(\d{2}))$/;

function daysInMonth(year: number, month: number): number {
  if (month === 2) {
    const leap = (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0;
    return leap ? 29 : 28;
  }
  return [4, 6, 9, 11].includes(month) ? 30 : 31;
}

function isRfc3339(value: unknown): value is string {
  if (typeof value !== "string") return false;
  const m = RFC3339_WITH_OFFSET.exec(value);
  if (!m) return false;
  const [year, month, day, hour, minute, second] = m.slice(1, 7).map(Number);
  if (month < 1 || month > 12) return false;
  if (day < 1 || day > daysInMonth(year, month)) return false;
  if (hour > 23 || minute > 59 || second > 60) return false;
  if (m[8] !== undefined && (Number(m[8]) > 23 || Number(m[9]) > 59)) return false;
  // 閏秒（:60）は Date.parse が解釈しない実装があるため 59 秒として解釈可否だけ確認する。
  const parseable = second === 60 ? value.replace(/:60(?=[.Z+-])/, ":59") : value;
  return !Number.isNaN(Date.parse(parseable));
}

function parseAnswer(value: unknown): RevisitIntakeAnswer {
  if (
    !isRecord(value) ||
    !isCanonicalDecimalId(value.answer_id) ||
    !isCanonicalDecimalId(value.revision) ||
    value.form_definition_version !== 1 ||
    !isRfc3339(value.received_at) ||
    value.provenance !== "owner_submission" ||
    !isRecord(value.answers) ||
    typeof value.answers.visit_purpose !== "string" ||
    value.answers.visit_purpose.trim() === ""
  ) {
    throw new Error("revisit-intake: invalid answer");
  }
  const a = value.answers;
  return {
    answer_id: value.answer_id,
    revision: value.revision,
    form_definition_version: value.form_definition_version,
    received_at: value.received_at,
    provenance: "owner_submission",
    answers: {
      visit_purpose: a.visit_purpose as string,
      onset: nullableText(a.onset),
      energy: nullableChoice(a.energy, REVISIT_INTAKE_ENERGY_APPETITE_LABELS),
      appetite: nullableChoice(a.appetite, REVISIT_INTAKE_ENERGY_APPETITE_LABELS),
      current_symptoms: nullableText(a.current_symptoms),
      improvement: nullableChoice(a.improvement, REVISIT_INTAKE_IMPROVEMENT_LABELS),
      examination_preference: nullableChoice(a.examination_preference, REVISIT_INTAKE_EXAMINATION_PREFERENCE_LABELS),
      main_concern: nullableText(a.main_concern),
    },
  };
}

export function parseRevisitIntakeRecordRead(value: unknown): RevisitIntakeRecordRead {
  if (!isRecord(value)) throw new Error("revisit-intake: invalid response");
  if (value.binding === "unbound" || value.binding === "mismatch") {
    // 契約上 target=null / history=[]。万一本文が付いていても採用しない。
    return { binding: value.binding, target: null, history: [], next_cursor: null };
  }
  if (value.binding !== "bound" || !isRecord(value.target) || !Array.isArray(value.history)) {
    throw new Error("revisit-intake: invalid binding");
  }
  const rawCursor = value.next_cursor;
  if (rawCursor !== null && (typeof rawCursor !== "string" || rawCursor === "")) {
    throw new Error("revisit-intake: invalid next_cursor");
  }
  const nextCursor: string | null = rawCursor;
  const t = value.target;
  const answerState = t.answer_state;
  const contentAvailable = t.content_available;
  const headRevision = t.head_revision;
  if (
    (answerState !== "unanswered" && answerState !== "answered") ||
    typeof contentAvailable !== "boolean" ||
    (headRevision !== null && !isCanonicalDecimalId(headRevision)) ||
    t.latest_answer === undefined
  ) {
    throw new Error("revisit-intake: invalid target state");
  }
  const latest = t.latest_answer === null ? null : parseAnswer(t.latest_answer);
  // 状態の組合せ不整合は契約失敗。「本文は閲覧できません」等へ正規化しない。
  if (answerState === "unanswered") {
    if (headRevision !== null || latest !== null || contentAvailable) {
      throw new Error("revisit-intake: inconsistent unanswered target");
    }
  } else if (headRevision === null || (contentAvailable ? latest === null : latest !== null)) {
    throw new Error("revisit-intake: inconsistent answered target");
  } else if (latest !== null && latest.revision !== headRevision) {
    // 可視本文は head revision の回答でなければならない。
    throw new Error("revisit-intake: latest answer is not head revision");
  }
  const targetRef = parseRef(t.ref);
  const history: RevisitIntakeHistoryEntry[] = value.history.map((entry) => {
    if (
      !isRecord(entry) ||
      (entry.target_relation !== "current" && entry.target_relation !== "previous")
    ) {
      throw new Error("revisit-intake: invalid history entry");
    }
    return {
      target: parseRef(entry.target),
      answer: parseAnswer(entry.answer),
      target_relation: entry.target_relation,
    };
  });
  // head は固定 target の最大 revision。同 ref の履歴が head を超える／未回答なのに存在する応答は矛盾。
  for (const entry of history) {
    if (!sameRef(entry.target, targetRef)) continue;
    if (headRevision === null || compareDecimal(entry.answer.revision, headRevision) > 0) {
      throw new Error("revisit-intake: history exceeds head revision");
    }
  }
  return {
    binding: "bound",
    target: {
      ref: targetRef,
      head_revision: headRevision,
      answer_state: answerState,
      latest_answer: latest,
      content_available: contentAvailable,
    },
    history,
    next_cursor: nextCursor,
  };
}

// ── Query key（feature 内定義。共通 queryKeys への追加は報告のみ） ────────

export const revisitIntakeQueryKeys = {
  all: () => ["medical-record-revisit-intake"] as const,
  // 別カルテ・別医院・別患者の原文キャッシュを使い回さない。
  record: (recordId: string, clinicId: string, petId: string) =>
    ["medical-record-revisit-intake", recordId, clinicId, petId] as const,
};

// ── Fetch ─────────────────────────────────────────────────────────────

export const getRecordRevisitIntake = async (
  recordId: string,
  clinicId: string,
  petId: string,
  cursor?: string,
): Promise<RevisitIntakeRecordRead> => {
  const { data } = await axios.get<unknown>(
    `/v1/medical-records/${encodeURIComponent(recordId)}/revisit-intake`,
    {
      // 拠点横断で開いたカルテは record 自身の clinicId を送る（treatments.ts と同流儀）。
      headers: { "X-Clinic-ID": clinicId },
      params: cursor ? { cursor } : undefined,
    },
  );
  const parsed = parseRevisitIntakeRecordRead(data);
  // 呼び出し時点の患者と異なる pet の原文は契約エラー（現在予約への追従・fallback はしない）。
  if (
    parsed.binding === "bound" &&
    (parsed.target.ref.pet_id !== petId || parsed.history.some((h) => h.target.pet_id !== petId))
  ) {
    throw new Error("revisit-intake: pet mismatch");
  }
  return parsed;
};

function errorStatus(error: unknown): number | undefined {
  if (isRecord(error) && isRecord(error.response) && typeof error.response.status === "number") {
    return error.response.status;
  }
  return undefined;
}

export function getRevisitIntakeErrorStatus(error: unknown): number | undefined {
  return errorStatus(error);
}

interface UseGetRecordRevisitIntakeParams {
  recordId: string | undefined;
  clinicId: string | undefined;
  petId: string | undefined;
  /** 新規カルテ等で false。ID が揃わない間は取得しない。 */
  enabled?: boolean;
}

export const useGetRecordRevisitIntake = ({
  recordId,
  clinicId,
  petId,
  enabled = true,
}: UseGetRecordRevisitIntakeParams) => {
  const ready = enabled && !!recordId && !!clinicId && !!petId;
  return useInfiniteQuery({
    queryKey: revisitIntakeQueryKeys.record(recordId ?? "", clinicId ?? "", petId ?? ""),
    queryFn: ({ pageParam }) => getRecordRevisitIntake(recordId ?? "", clinicId ?? "", petId ?? "", pageParam),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) =>
      last.binding === "bound" ? (last.next_cursor ?? undefined) : undefined,
    enabled: ready,
    // 認証 scope の原文は Cache-Control: no-store。常に再取得し、unmount 後に保持しない。
    staleTime: 0,
    gcTime: 0,
    // 502/504 は axios interceptor が GET を再試行済み。その他の失敗はカード内「再試行」で明示操作させる。
    retry: false,
  });
};
