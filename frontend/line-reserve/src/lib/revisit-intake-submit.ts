// EMR-474: 再診問診 submit の純粋ロジック（request 組立・fingerprint・入力検証・エラー分類）。
// 契約: backend/docs/api.yaml RevisitIntakeSubmitRequest / RevisitIntakeError、
// technical-spec#revisit-intake-api-selected-v3「submit・結果照会の境界」。
import axios from "axios";
import {
  revisitIntakeClosedDetailsSchema,
  revisitIntakeErrorBodySchema,
  revisitIntakeInvalidInputDetailsSchema,
} from "../api/schemas";
import type {
  RevisitIntakeFullAnswers,
  RevisitIntakeOwnerCurrentTarget,
  RevisitIntakeSubmitRequest,
} from "../types/models";
import { REVISIT_INTAKE_FIELD_ORDER, type RevisitIntakeFieldName } from "./revisit-intake-options";

export type RevisitIntakeFieldErrors = Partial<Record<RevisitIntakeFieldName, string>>;

export const VISIT_PURPOSE_REQUIRED_MESSAGE =
  "来院目的を入力してください（問診を送信する場合のみ必須です）。";
const NUL_CHARACTER_MESSAGE = "使用できない文字が含まれています。";
const SERVER_INVALID_FIELD_MESSAGE = "入力内容を確認してください。";

/** Unicode 空白のみ（全角空白・改行を含む）を空として扱う。原文は trim しない。 */
const WHITESPACE_ONLY = /^\p{White_Space}*$/u;
const NUL_CHARACTER = new RegExp(String.fromCharCode(0));

/**
 * 送信前の入力検証。来院目的だけ送信時必須、他は任意（空欄を補完しない）。
 * NUL は server が 400 にし、共通 interceptor が黙って除去して原文を変えてしまうため送信前に止める。
 */
export function validateRevisitIntakeAnswers(
  values: RevisitIntakeFullAnswers,
): RevisitIntakeFieldErrors {
  const errors: RevisitIntakeFieldErrors = {};
  if (WHITESPACE_ONLY.test(values.visit_purpose)) {
    errors.visit_purpose = VISIT_PURPOSE_REQUIRED_MESSAGE;
  }
  for (const field of REVISIT_INTAKE_FIELD_ORDER) {
    const value = values[field];
    if (typeof value === "string" && NUL_CHARACTER.test(value) && errors[field] === undefined) {
      errors[field] = NUL_CHARACTER_MESSAGE;
    }
  }
  return errors;
}

export function firstErrorField(errors: RevisitIntakeFieldErrors): RevisitIntakeFieldName | null {
  return REVISIT_INTAKE_FIELD_ORDER.find((f) => errors[f] !== undefined) ?? null;
}

/**
 * 現在 target に対する submit body。expected_revision は head_revision（未回答は明示 null。
 * 不可視 head でも head_revision を期待版にする）。form 版は read の現在版。
 * 自由文は null（未入力）と空文字（入力して消した）をそのまま区別して送る。
 */
export function buildRevisitIntakeSubmitRequest(
  current: RevisitIntakeOwnerCurrentTarget,
  formDefinitionVersion: number,
  values: RevisitIntakeFullAnswers,
): RevisitIntakeSubmitRequest {
  return {
    pet_id: current.ref.pet_id,
    target_generation: current.ref.target_generation,
    expected_revision: current.head_revision,
    form_definition_version: formDefinitionVersion,
    answers: { ...values },
  };
}

/** 同 key 同内容判定用。server の fingerprint と同じく expected_revision を除外する。 */
export function revisitIntakeFingerprint(body: RevisitIntakeSubmitRequest): string {
  return JSON.stringify([
    body.pet_id,
    body.target_generation,
    body.form_definition_version,
    REVISIT_INTAKE_FIELD_ORDER.map((f) => body.answers[f]),
  ]);
}

export type RevisitIntakeSubmitErrorKind =
  | { type: "invalid"; fieldErrors: RevisitIntakeFieldErrors }
  | { type: "revision_conflict" }
  | { type: "form_version_mismatch" }
  | { type: "target_changed" }
  | {
      type: "intake_closed";
      reason: "reception_started" | "cancelled" | "deleted" | "unsupported_lifecycle" | null;
    }
  | { type: "key_content_mismatch" }
  | { type: "unauthorized" }
  | { type: "access_denied" }
  | { type: "payload_too_large" }
  | { type: "unsupported_media_type" }
  | { type: "rate_limited"; retryAfterSeconds: number | null }
  /** Retry-After 付き 503: 結果不明のまま、期限まで同 key の再送を待つ */
  | { type: "service_unavailable"; retryAfterSeconds: number }
  /** 応答喪失・5xx・応答形不正・未知 409 code: 結果不明として同 key で照会する */
  | { type: "unknown" };

function fieldFromPath(path: string): RevisitIntakeFieldName | null {
  const name = path.replace(/^answers[./]/, "").replace(/^\//, "");
  return (REVISIT_INTAKE_FIELD_ORDER as readonly string[]).includes(name)
    ? (name as RevisitIntakeFieldName)
    : null;
}

function parseRetryAfter(value: unknown): number | null {
  if (typeof value !== "string" && typeof value !== "number") return null;
  const seconds = Number(value);
  if (Number.isFinite(seconds) && seconds >= 0) return Math.ceil(seconds);
  const date = Date.parse(String(value));
  if (Number.isNaN(date)) return null;
  return Math.max(0, Math.ceil((date - Date.now()) / 1000));
}

/**
 * 429 / 503 の Retry-After 秒（正の値のときだけ）。結果照会など submit 以外の route でも
 * 同じ期限を尊重するために使う。それ以外の失敗・Retry-After なしは null。
 */
export function revisitIntakeRetryAfterSeconds(err: unknown): number | null {
  if (!axios.isAxiosError(err) || !err.response) return null;
  const { status, headers } = err.response;
  if (status !== 429 && status !== 503) return null;
  const seconds = parseRetryAfter(headers?.["retry-after"]);
  return seconds !== null && seconds > 0 ? seconds : null;
}

/** HTTP status と code だけで分岐する（error 文言では分岐しない）。 */
export function classifyRevisitIntakeSubmitError(err: unknown): RevisitIntakeSubmitErrorKind {
  if (!axios.isAxiosError(err) || !err.response) return { type: "unknown" };
  const { status, data, headers } = err.response;
  const parsedBody = revisitIntakeErrorBodySchema.safeParse(data);
  const body = parsedBody.success ? parsedBody.data : {};

  switch (status) {
    case 400: {
      const fieldErrors: RevisitIntakeFieldErrors = {};
      if (body.code === "INVALID_INPUT") {
        const details = revisitIntakeInvalidInputDetailsSchema.safeParse(body.details ?? null);
        if (details.success && details.data) {
          for (const f of details.data.fields) {
            const field = fieldFromPath(f.path);
            if (field && fieldErrors[field] === undefined) {
              fieldErrors[field] =
                field === "visit_purpose"
                  ? VISIT_PURPOSE_REQUIRED_MESSAGE
                  : SERVER_INVALID_FIELD_MESSAGE;
            }
          }
        }
      }
      return { type: "invalid", fieldErrors };
    }
    case 401:
      return { type: "unauthorized" };
    case 403:
    case 404:
      return { type: "access_denied" };
    case 409:
      switch (body.code) {
        case "REVISION_CONFLICT":
          return { type: "revision_conflict" };
        case "FORM_VERSION_MISMATCH":
          return { type: "form_version_mismatch" };
        case "TARGET_CHANGED":
          return { type: "target_changed" };
        case "INTAKE_CLOSED": {
          const details = revisitIntakeClosedDetailsSchema.safeParse(body.details);
          return { type: "intake_closed", reason: details.success ? details.data.reason : null };
        }
        case "KEY_CONTENT_MISMATCH":
          return { type: "key_content_mismatch" };
        default:
          return { type: "unknown" };
      }
    case 413:
      return { type: "payload_too_large" };
    case 415:
      return { type: "unsupported_media_type" };
    case 429:
      return {
        type: "rate_limited",
        retryAfterSeconds: parseRetryAfter(headers?.["retry-after"]),
      };
    case 503: {
      // 契約: Retry-After があれば尊重する。無い 503 は他の 5xx と同じく結果不明として照会する。
      const retryAfterSeconds = parseRetryAfter(headers?.["retry-after"]);
      return retryAfterSeconds !== null && retryAfterSeconds > 0
        ? { type: "service_unavailable", retryAfterSeconds }
        : { type: "unknown" };
    }
    default:
      return { type: "unknown" };
  }
}
