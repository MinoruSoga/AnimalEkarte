import { z } from "zod";
import {
  certificateNotSucceededResultSchema,
  certificateOperationKeySchema,
  certificateSucceededResultSchema,
} from "./certificate-schemas";

/**
 * EMR-487 Slice A: 操作回復 discovery（POST /v1/vaccination-certificate-operation-recovery/query）の wire 型。
 * 契約正本: backend/docs/api.yaml `CertificateOperationRecovery*` / `OperationRecovery*`（selected・runtime 未実装）と
 * docs/work/emr/emr-474-487-operation-recovery-contract-20261021.md §3 / §4。
 *
 * すべて closed（未知 field は拒否）。zod v4 の `z.object` は未知 field を捨てるだけなので `strictObject` を使う。
 * 証明書 ID はこの feature の既存 wire どおり JSON number（安全整数のみ。丸めない）。
 * recovery_generation は 10 進文字列で、Number へ変換しない。
 */

/** 1 ページの件数上限（サーバ固定）。 */
export const RECOVERY_PAGE_LIMIT = 50;

const INT64_MAX = "9223372036854775807";

/** 正規十進文字列 1..int64 最大。19 桁のときだけ文字列比較（Number に丸めない）。 */
export function isCanonicalRecoveryGeneration(value: string): boolean {
  if (!/^[1-9][0-9]*$/.test(value) || value.length > INT64_MAX.length) return false;
  return value.length < INT64_MAX.length || value <= INT64_MAX;
}

export const recoveryGenerationSchema = z
  .string()
  .refine(isCanonicalRecoveryGeneration, "recovery_generation は 1..int64 の正規十進文字列");

/** opaque な base64url cursor。内容は解釈しない（認可証明にもしない）。 */
const cursorString = z
  .string()
  .min(1)
  .max(2048)
  .regex(/^[A-Za-z0-9_-]+={0,2}$/);

const safeId = z
  .number()
  .refine((v) => Number.isSafeInteger(v), "int64 を JS number で安全に表せません")
  .refine((v) => v >= 1, "ID は 1 以上");

/**
 * 要求。issue は edition_id = null、reprint は正の edition_id（両方向を強制）。
 * reprint + null / issue + 非 null は、送信前にここで拒否する。
 */
export const certificateRecoveryQueryRequestSchema = z.union([
  z.strictObject({
    target_ref: z.strictObject({ vaccination_id: safeId, edition_id: z.null() }),
    intent: z.literal("issue"),
    cursor: cursorString.nullable(),
  }),
  z.strictObject({
    target_ref: z.strictObject({ vaccination_id: safeId, edition_id: safeId }),
    intent: z.literal("reprint"),
    cursor: cursorString.nullable(),
  }),
]);
export type CertificateRecoveryQueryRequest = z.infer<typeof certificateRecoveryQueryRequestSchema>;

const ISO_DATE_TIME = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$/;
const claimedAt = z
  .string()
  .regex(ISO_DATE_TIME)
  .refine((v) => !Number.isNaN(Date.parse(v)), "claimed_at が日時ではありません");

const receiptDigest = z.string().regex(/^[0-9a-f]{64}$/);

// 既存 lookup response（C06）をそのまま保持する。ここでは未知 field を拒否する閉じた版を使う。
const succeededResult = certificateSucceededResultSchema.strict();
const notSucceededResult = certificateNotSucceededResultSchema.strict();

const itemBase = {
  operation_key: certificateOperationKeySchema,
  intent: z.enum(["issue", "reprint"]),
  claimed_at: claimedAt,
};

/** api.yaml の oneOf 3 通りだけを受理する（フィールド単独ではなく組み合わせで検証）。 */
const itemUnion = z.union([
  // 成功 receipt の権威ある読取だけが succeeded。digest 必須。
  z.strictObject({
    ...itemBase,
    discovery_state: z.literal("result"),
    execution_state: z.literal("succeeded"),
    receipt_digest: receiptDigest,
    result: succeededResult,
  }),
  // 未成功。closed_without_effect は closure 証拠がある key のみ。digest は null。
  z.strictObject({
    ...itemBase,
    discovery_state: z.literal("result"),
    execution_state: z.enum(["unresolved", "closed_without_effect"]),
    receipt_digest: z.null(),
    result: notSucceededResult.nullable(),
  }),
  // lookup 不調・metadata 不明。必ず unresolved + result null。
  z.strictObject({
    ...itemBase,
    discovery_state: z.literal("unavailable"),
    execution_state: z.literal("unresolved"),
    receipt_digest: z.null(),
    result: z.null(),
  }),
]);

export const certificateRecoveryItemSchema = itemUnion.refine(
  (item) => item.result === null || item.result.operation_key === item.operation_key,
  "item.result.operation_key が item.operation_key と一致しません",
);
export type CertificateRecoveryItem = z.infer<typeof certificateRecoveryItemSchema>;

export const certificateRecoveryQueryResponseSchema = z
  .strictObject({
    items: z.array(certificateRecoveryItemSchema).max(RECOVERY_PAGE_LIMIT),
    next_cursor: cursorString.nullable(),
    complete: z.boolean(),
    recovery_generation: recoveryGenerationSchema,
    active_operation_key: certificateOperationKeySchema.nullable(),
  })
  // complete は「この走査が上限まで終わった」ことだけを表す。次ページがあるのに complete は矛盾。
  .refine((page) => page.complete === (page.next_cursor === null), "complete と next_cursor が矛盾しています");
export type CertificateRecoveryQueryResponse = z.infer<typeof certificateRecoveryQueryResponseSchema>;
