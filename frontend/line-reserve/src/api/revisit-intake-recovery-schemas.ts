// EMR-474 operation-recovery Slice A: LIFF 再診問診の「再読込後の未確定操作の発見」wire schema。
// 契約正本: docs/work/emr/emr-474-487-operation-recovery-contract-20261021.md §3/§4/§11 と
// backend/docs/api.yaml の OperationRecovery* / LiffOperationRecovery* schema。
//
// 【未接続】route（POST /api/liff/{clinicId}/revisit-intake/operation-recovery/query）は BE 未実装
// （x-runtime-status: not-implemented）。このファイルは契約どおりの検証だけを持ち、利用開始は
// LIFF_REVISIT_INTAKE_ACTIVE（lib/revisit-intake-activation.ts）に従う。
//
// 方針:
// - 新しい envelope（request / response / item / error）は closed（未知 field 拒否）。
//   既存 lookup response（revisitIntakeResultResponseSchema）は既存 wire のまま再利用し混ぜない。
// - ID・世代は正規形 10 進文字列のまま扱い Number へ変換しない。
// - complete は「その上限の走査終了」だけを表す。新 claim 不存在・次操作許可の根拠として扱う field は作らない。
import { z } from "zod";
import { isRfc3339DateTime, revisitIntakeResultResponseSchema } from "./schemas";

const BIGINT_MAX = BigInt("9223372036854775807");

/** 正の 10 進文字列（1..9223372036854775807・先頭ゼロ不可）。 */
const canonicalPositiveDecimal = z
  .string()
  .regex(/^[1-9][0-9]*$/)
  .max(19)
  .refine((v) => BigInt(v) <= BIGINT_MAX);

/** lowercase hyphenated UUID・nil UUID 拒否。 */
const recoveryOperationKeySchema = z
  .string()
  .regex(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/)
  .refine((v) => v !== "00000000-0000-0000-0000-000000000000");

const recoveryGenerationSchema = canonicalPositiveDecimal;

/** receipt_digest: succeeded 時だけの SHA-256 lowerhex64。 */
const receiptDigestSchema = z.string().regex(/^[0-9a-f]{64}$/);

/** opaque base64url cursor。server が発行した値をそのまま返す（client は中身を解釈・構築しない）。 */
const RECOVERY_CURSOR_MAX_LENGTH = 1024;
export const recoveryCursorSchema = z
  .string()
  .min(1)
  .max(RECOVERY_CURSOR_MAX_LENGTH)
  .regex(/^[A-Za-z0-9_-]+$/);

// ---------------------------------------------------------------------------
// Request（oneOf: target / self）。cursor は required nullable。
// ---------------------------------------------------------------------------

const recoveryTargetRefSchema = z.strictObject({
  appointment_id: canonicalPositiveDecimal,
  pet_id: canonicalPositiveDecimal,
  target_generation: canonicalPositiveDecimal,
});

export const recoveryQueryRequestSchema = z.union([
  z.strictObject({
    scope: z.literal("target"),
    target_ref: recoveryTargetRefSchema,
    cursor: recoveryCursorSchema.nullable(),
  }),
  z.strictObject({
    scope: z.literal("self"),
    cursor: recoveryCursorSchema.nullable(),
  }),
]);

export type RevisitIntakeRecoveryQueryRequest = z.infer<typeof recoveryQueryRequestSchema>;
export type RevisitIntakeRecoveryTargetRef = z.infer<typeof recoveryTargetRefSchema>;

// ---------------------------------------------------------------------------
// Response item
// ---------------------------------------------------------------------------

const itemShape = {
  operation_key: recoveryOperationKeySchema,
  intent: z.literal("revisit-intake.submit"),
  claimed_at: z.string().refine(isRfc3339DateTime),
  discovery_state: z.enum(["result", "unavailable"]),
  execution_state: z.enum(["succeeded", "unresolved", "closed_without_effect"]),
  receipt_digest: receiptDigestSchema.nullable(),
  result: revisitIntakeResultResponseSchema.nullable(),
};

export const recoveryItemSchema = z.strictObject(itemShape).superRefine((item, ctx) => {
  const fail = (message: string) => ctx.addIssue({ code: "custom", message });
  const resultState = item.result?.operation_result.state ?? null;
  if (item.result !== null && item.result.operation_result.operation_id !== item.operation_key) {
    fail("result.operation_id must equal operation_key");
  }
  if (item.discovery_state === "unavailable") {
    // lookup 不調・metadata 不明は unavailable + unresolved + result=null として保留する。
    if (item.execution_state !== "unresolved" || item.receipt_digest !== null || item.result !== null) {
      fail("unavailable item must be unresolved with null digest and null result");
    }
    return;
  }
  if (item.execution_state === "succeeded") {
    // 成功の権威ある読取だけが succeeded。digest 必須・result は成功 response。
    if (item.receipt_digest === null) fail("succeeded item requires receipt_digest");
    if (resultState !== "succeeded") fail("succeeded item requires a succeeded result");
    return;
  }
  // unresolved / closed_without_effect: digest は null、result は null か未確定 response。
  if (item.receipt_digest !== null) fail("non-succeeded item must have null receipt_digest");
  if (resultState !== null && resultState !== "claimed_uncommitted" && resultState !== "result_unavailable") {
    fail("non-succeeded item result must be null or an unresolved response");
  }
});

export type RevisitIntakeRecoveryItem = z.infer<typeof recoveryItemSchema>;

// ---------------------------------------------------------------------------
// Response page
// ---------------------------------------------------------------------------

const RECOVERY_PAGE_MAX_ITEMS = 50;

export const recoveryPageSchema = z
  .strictObject({
    items: z.array(recoveryItemSchema).max(RECOVERY_PAGE_MAX_ITEMS),
    next_cursor: recoveryCursorSchema.nullable(),
    complete: z.boolean(),
    recovery_generation: recoveryGenerationSchema.nullable(),
    active_operation_key: recoveryOperationKeySchema.nullable(),
  })
  .superRefine((page, ctx) => {
    const fail = (message: string) => ctx.addIssue({ code: "custom", message });
    if (page.complete && page.next_cursor !== null) fail("complete=true must not carry next_cursor");
    if (page.active_operation_key !== null && page.recovery_generation === null) {
      fail("active_operation_key requires recovery_generation");
    }
    const keys = new Set<string>();
    for (const item of page.items) {
      if (keys.has(item.operation_key)) fail("duplicate operation_key in page");
      keys.add(item.operation_key);
    }
  });

export type RevisitIntakeRecoveryPage = z.infer<typeof recoveryPageSchema>;

/** scope 別の追加検査: self は generation / active が両方 null、target は generation 必須。 */
export function isPageConsistentWithScope(
  scope: "target" | "self",
  page: RevisitIntakeRecoveryPage,
): boolean {
  if (scope === "self") {
    return page.recovery_generation === null && page.active_operation_key === null;
  }
  return page.recovery_generation !== null;
}

// ---------------------------------------------------------------------------
// Error（この新 API だけの closed {error, code}）
// ---------------------------------------------------------------------------

export const recoveryErrorBodySchema = z.strictObject({
  error: z.string(),
  code: z.string(),
});
