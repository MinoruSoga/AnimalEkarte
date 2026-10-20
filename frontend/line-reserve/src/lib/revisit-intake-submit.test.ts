import { AxiosError, AxiosHeaders, type AxiosResponse } from "axios";
import { describe, expect, it } from "vitest";
import {
  buildRevisitIntakeSubmitRequest,
  classifyRevisitIntakeSubmitError,
  revisitIntakeFingerprint,
  revisitIntakeRetryAfterSeconds,
  validateRevisitIntakeAnswers,
} from "./revisit-intake-submit";
import { createRevisitIntakeOperationKey } from "./revisit-intake-operation-key";
import type { RevisitIntakeFullAnswers, RevisitIntakeOwnerCurrentTarget } from "../types/models";

const ANSWERS: RevisitIntakeFullAnswers = {
  visit_purpose: "合成の来院目的",
  onset: null,
  energy: "usual",
  appetite: null,
  current_symptoms: "",
  improvement: null,
  examination_preference: null,
  main_concern: null,
};

const CURRENT: RevisitIntakeOwnerCurrentTarget = {
  ref: { target_id: "5401", appointment_id: "12001", pet_id: "801", target_generation: "3" },
  head_revision: "2",
  answer_state: "answered",
  latest_answer: null,
  content_available: false,
  editable: true,
  edit_block_reason: null,
};

function axiosError(status: number, data: unknown, headers: Record<string, string> = {}) {
  const response = {
    status,
    statusText: "",
    data,
    headers: new AxiosHeaders(headers),
    config: { headers: new AxiosHeaders() },
  } as AxiosResponse;
  return new AxiosError("synthetic", "ERR_BAD_RESPONSE", undefined, undefined, response);
}

describe("revisit-intake-submit（EMR-474）", () => {
  it("expected_revision は head_revision（不可視 head でも）、null/空文字は区別したまま組み立てる", () => {
    const body = buildRevisitIntakeSubmitRequest(CURRENT, 1, ANSWERS);
    expect(body).toEqual({
      pet_id: "801",
      target_generation: "3",
      expected_revision: "2",
      form_definition_version: 1,
      answers: ANSWERS,
    });
    expect(body.answers.onset).toBeNull();
    expect(body.answers.current_symptoms).toBe("");
  });

  it("fingerprint は expected_revision を含めず、回答の null と空文字を区別する", () => {
    const a = buildRevisitIntakeSubmitRequest(CURRENT, 1, ANSWERS);
    const b = { ...a, expected_revision: "9" };
    const c = buildRevisitIntakeSubmitRequest(CURRENT, 1, { ...ANSWERS, current_symptoms: null });
    expect(revisitIntakeFingerprint(a)).toBe(revisitIntakeFingerprint(b));
    expect(revisitIntakeFingerprint(a)).not.toBe(revisitIntakeFingerprint(c));
  });

  it("来院目的だけ送信時必須（Unicode 空白のみも不可）。任意欄の空欄はエラーにしない", () => {
    expect(
      validateRevisitIntakeAnswers({ ...ANSWERS, visit_purpose: "" }).visit_purpose,
    ).toBeDefined();
    expect(
      validateRevisitIntakeAnswers({ ...ANSWERS, visit_purpose: "　\n " }).visit_purpose,
    ).toBeDefined();
    expect(validateRevisitIntakeAnswers(ANSWERS)).toEqual({});
  });

  it("NUL を含む自由文は送信前に止める（共通 sanitizer による原文改変を避ける）", () => {
    const nul = String.fromCharCode(0);
    expect(
      validateRevisitIntakeAnswers({ ...ANSWERS, main_concern: `a${nul}b` }).main_concern,
    ).toBe("使用できない文字が含まれています。");
  });

  it.each([
    [
      409,
      { error: "x", code: "REVISION_CONFLICT", details: { head_revision: "3" } },
      "revision_conflict",
    ],
    [
      409,
      {
        error: "x",
        code: "FORM_VERSION_MISMATCH",
        details: { supported_form_definition_version: 1 },
      },
      "form_version_mismatch",
    ],
    [409, { error: "x", code: "TARGET_CHANGED", details: null }, "target_changed"],
    [409, { error: "x", code: "INTAKE_CLOSED", details: { reason: "cancelled" } }, "intake_closed"],
    [409, { error: "x", code: "KEY_CONTENT_MISMATCH", details: null }, "key_content_mismatch"],
    [409, { error: "x", code: "SOMETHING_NEW", details: null }, "unknown"],
    [401, { error: "unauthorized" }, "unauthorized"],
    [403, { error: "forbidden" }, "access_denied"],
    [404, { error: "not found" }, "access_denied"],
    [413, { error: "too large" }, "payload_too_large"],
    [415, { error: "unsupported" }, "unsupported_media_type"],
    [500, { error: "internal" }, "unknown"],
    [503, { error: "unavailable" }, "unknown"],
  ])("HTTP %s は code で分類する（error 文言では分岐しない）", (status, data, expected) => {
    expect(classifyRevisitIntakeSubmitError(axiosError(status, data)).type).toBe(expected);
  });

  it("応答なし（通信断）は unknown（失敗確定にしない）", () => {
    expect(
      classifyRevisitIntakeSubmitError(new AxiosError("Network Error", "ERR_NETWORK")).type,
    ).toBe("unknown");
    expect(classifyRevisitIntakeSubmitError(new Error("response validation failed")).type).toBe(
      "unknown",
    );
  });

  it("400 INVALID_INPUT の details.fields を欄へ対応づける", () => {
    const kind = classifyRevisitIntakeSubmitError(
      axiosError(400, {
        error: "invalid",
        code: "INVALID_INPUT",
        details: { fields: [{ path: "answers.main_concern", reason: "invalid" }] },
      }),
    );
    expect(kind).toEqual({
      type: "invalid",
      fieldErrors: { main_concern: "入力内容を確認してください。" },
    });
  });

  it("429 は Retry-After 秒を読む", () => {
    expect(
      classifyRevisitIntakeSubmitError(
        axiosError(429, { error: "too many" }, { "retry-after": "30" }),
      ),
    ).toEqual({ type: "rate_limited", retryAfterSeconds: 30 });
  });

  it("Retry-After 付き 503 は service_unavailable、無い 503 は unknown（結果照会）", () => {
    expect(
      classifyRevisitIntakeSubmitError(
        axiosError(503, { error: "unavailable" }, { "retry-after": "20" }),
      ),
    ).toEqual({ type: "service_unavailable", retryAfterSeconds: 20 });
    expect(classifyRevisitIntakeSubmitError(axiosError(503, { error: "unavailable" })).type).toBe(
      "unknown",
    );
  });

  it("Retry-After 付き 500 は service_unavailable（期限まで照会・再送しない）、無い 500 は unknown", () => {
    expect(
      classifyRevisitIntakeSubmitError(
        axiosError(500, { error: "internal" }, { "retry-after": "15" }),
      ),
    ).toEqual({ type: "service_unavailable", retryAfterSeconds: 15 });
    expect(classifyRevisitIntakeSubmitError(axiosError(500, { error: "internal" })).type).toBe(
      "unknown",
    );
    expect(
      revisitIntakeRetryAfterSeconds(axiosError(500, { error: "x" }, { "retry-after": "15" })),
    ).toBe(15);
  });

  it("操作 key は lowercase hyphenated UUID", () => {
    const key = createRevisitIntakeOperationKey();
    expect(key).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/);
    expect(createRevisitIntakeOperationKey()).not.toBe(key);
  });
});
