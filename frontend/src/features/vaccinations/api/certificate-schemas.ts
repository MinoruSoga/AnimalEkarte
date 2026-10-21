import { z } from "zod";

/**
 * EMR-487: 予防接種証明書 API の wire 型と zod 解析（手書き）。
 * 契約正本: backend/docs/api.yaml `Certificate*`（EMR-486 selected・runtime 未実装）と
 * docs/work/emr/emr-486-certificate-api-contract-20261020.md。
 * `types/generated` は編集しない（`make codegen` はユーザー実行）。
 *
 * 解析は fail-closed: 契約外の形は例外にして、推測で補完しない。
 * int64 ID は JSON number で届く。`Number.isSafeInteger` を満たさない値は丸めずに拒否する
 * （2^53 超の扱いは open question。BE 契約は int64 文字列を採用していない）。
 */

const safeInt = z.number().refine((v) => Number.isSafeInteger(v), "int64 を JS number で安全に表せません");
const id = safeInt.refine((v) => v >= 1, "ID は 1 以上");
const hash = z.string().regex(/^[0-9a-f]{64}$/);
const dateStr = z.string().regex(/^\d{4}-\d{2}-\d{2}$/);
const nullableStr = z.string().nullable();

export const certificateKindSchema = z.enum(["normal", "rabies"]);
export const certificateLayoutSchema = z.enum(["portrait", "landscape", "rabies"]);
export const certificateBodySizeSchema = z.enum(["small", "medium", "large"]);
export const certificateValiditySchema = z.enum([
  "current",
  "superseded",
  "source_changed",
  "source_deleted",
]);

export type CertificateKind = z.infer<typeof certificateKindSchema>;
export type CertificateLayout = z.infer<typeof certificateLayoutSchema>;
export type CertificateBodySize = z.infer<typeof certificateBodySizeSchema>;
export type CertificateValidity = z.infer<typeof certificateValiditySchema>;

/** 狂犬病様式専用の職員入力 10 欄。全欄 optional/nullable。null と明示 0 は別物。 */
export interface CertificateInputs {
  certificate_number: string | null;
  injection_tag_number: string | null;
  dog_registration_tag_number: string | null;
  body_size: CertificateBodySize | null;
  distinguishing_features: string | null;
  injection_amount_text: string | null;
  registration_fee: number | null;
  injection_fee: number | null;
  tag_fee: number | null;
  total: number | null;
}

export const CERTIFICATE_INPUT_KEYS = [
  "certificate_number",
  "injection_tag_number",
  "dog_registration_tag_number",
  "body_size",
  "distinguishing_features",
  "injection_amount_text",
  "registration_fee",
  "injection_fee",
  "tag_fee",
  "total",
] as const satisfies readonly (keyof CertificateInputs)[];

export function emptyCertificateInputs(): CertificateInputs {
  return {
    certificate_number: null,
    injection_tag_number: null,
    dog_registration_tag_number: null,
    body_size: null,
    distinguishing_features: null,
    injection_amount_text: null,
    registration_fee: null,
    injection_fee: null,
    tag_fee: null,
    total: null,
  };
}

const fee = safeInt.nullable();
const storedInputsSchema = z.object({
  certificate_number: nullableStr,
  injection_tag_number: nullableStr,
  dog_registration_tag_number: nullableStr,
  body_size: certificateBodySizeSchema.nullable(),
  distinguishing_features: nullableStr,
  injection_amount_text: nullableStr,
  registration_fee: fee,
  injection_fee: fee,
  tag_fee: fee,
  total: fee,
});

/** context.initial_inputs: 料金 4 欄のみ必須。省略は null（空欄）へ正規化する。 */
const initialInputsSchema = z
  .object({
    certificate_number: nullableStr.optional(),
    injection_tag_number: nullableStr.optional(),
    dog_registration_tag_number: nullableStr.optional(),
    body_size: certificateBodySizeSchema.nullable().optional(),
    distinguishing_features: nullableStr.optional(),
    injection_amount_text: nullableStr.optional(),
    registration_fee: fee,
    injection_fee: fee,
    tag_fee: fee,
    total: fee,
  })
  .transform((v): CertificateInputs => ({ ...emptyCertificateInputs(), ...stripUndefined(v) }));

function stripUndefined<T extends object>(value: T): Partial<T> {
  return Object.fromEntries(Object.entries(value).filter(([, v]) => v !== undefined)) as Partial<T>;
}

const snapshotSchema = z.object({
  schema_version: z.literal(1),
  kind: certificateKindSchema,
  source: z.object({
    vaccination_id: id,
    version: id,
    vaccine_id: id,
    vaccine_name: nullableStr,
    administered_on: dateStr,
    next_scheduled_on: dateStr.nullable(),
    lots: z.array(nullableStr).length(4),
    doctor_name: nullableStr,
  }),
  owner: z.object({
    name: nullableStr,
    name_kana: nullableStr,
    postal_code: nullableStr,
    address: nullableStr,
    phone: nullableStr,
  }),
  pet: z.object({
    name: nullableStr,
    species: nullableStr,
    breed: nullableStr,
    sex: nullableStr,
    coat_color: nullableStr,
    birth_date: dateStr.nullable(),
    age_text: nullableStr,
  }),
  clinic: z.object({
    name: nullableStr,
    director_name: nullableStr,
    postal_code: nullableStr,
    address: nullableStr,
    phone: nullableStr,
    fax: nullableStr,
    email: nullableStr,
    url: nullableStr,
    veterinarian_association: nullableStr,
    designated_veterinarian: nullableStr,
    veterinarian_name: nullableStr,
  }),
  display_issue_date: dateStr,
  inputs: storedInputsSchema,
  eligibility: z.object({
    rabies_eligible: z.boolean().nullable(),
    vaccine_id: id,
    vaccine_version: safeInt.nullable(),
    provenance: z.enum(["master_snapshot", "legacy_unresolved"]),
  }),
  render_manifest: z.object({
    renderer_version: z.string(),
    font_asset_version: z.string(),
    templates: z.object({
      portrait: z.string().optional(),
      landscape: z.string().optional(),
      rabies: z.string().optional(),
    }),
  }),
});
export type CertificateSnapshot = z.infer<typeof snapshotSchema>;

export const certificateEditionSchema = z.object({
  id,
  series_id: id,
  vaccination_id: id,
  kind: certificateKindSchema,
  edition: id,
  source_version: id,
  previous_edition_id: id.nullable(),
  issued_at: z.string(),
  issued_by: id,
  validity: certificateValiditySchema,
  snapshot: snapshotSchema,
  snapshot_sha256: hash,
  available_layouts: z.array(certificateLayoutSchema),
});
export type CertificateEdition = z.infer<typeof certificateEditionSchema>;

export const certificateContextSchema = z.object({
  vaccination_id: id,
  source_version: id,
  eligibility: z.object({
    state: z.enum(["eligible", "ineligible", "unknown"]),
    allowed_kinds: z.array(certificateKindSchema),
    provenance: z.enum(["master_snapshot", "legacy_unresolved"]),
  }),
  can_issue: z.boolean(),
  latest: z.object({
    normal: certificateEditionSchema.nullable(),
    rabies: certificateEditionSchema.nullable(),
  }),
  initial_inputs: initialInputsSchema,
});
export type CertificateContext = z.infer<typeof certificateContextSchema>;

/**
 * preview 応答の PDF。空でないだけでは足りない（壊れた base64・PDF 以外を受理すると、
 * 確認できない内容の発行ボタンが有効になる）。base64 として正しく、先頭が %PDF- のものだけ受理する。
 */
export function isPdfBase64(value: string): boolean {
  if (value.length < 8 || value.length % 4 !== 0 || !/^[A-Za-z0-9+/]+={0,2}$/.test(value)) return false;
  try {
    return atob(value.slice(0, 8)).startsWith("%PDF-");
  } catch {
    return false;
  }
}
const pdfBase64 = z.string().min(1).refine(isPdfBase64, "pdf_base64 が PDF ではありません");

export const certificatePreviewResponseSchema = z.object({
  snapshot: snapshotSchema,
  snapshot_sha256: hash,
  preview_fingerprint: hash,
  source_version: id,
  latest_edition_id: id.nullable(),
  pdf_base64: pdfBase64,
  media_type: z.literal("application/pdf"),
  template_version: z.string(),
  renderer_version: z.string(),
});
export type CertificatePreviewResponse = z.infer<typeof certificatePreviewResponseSchema>;

export const certificateSavedPreviewSchema = z.object({
  edition_id: id,
  layout: certificateLayoutSchema,
  pdf_base64: pdfBase64,
  media_type: z.literal("application/pdf"),
  template_version: z.string(),
  renderer_version: z.string(),
  content_sha256: hash,
});
export type CertificateSavedPreview = z.infer<typeof certificateSavedPreviewSchema>;

export const certificateEditionPageSchema = z.object({
  data: z.array(certificateEditionSchema),
  total: safeInt.refine((v) => v >= 0),
  page: safeInt.refine((v) => v >= 1),
  limit: safeInt.refine((v) => v >= 1 && v <= 100),
});
export type CertificateEditionPage = z.infer<typeof certificateEditionPageSchema>;

const operationKey = z
  .string()
  .regex(/^(?!00000000-0000-0000-0000-000000000000$)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/);

const succeededSchema = z.object({
  status: z.enum(["succeeded", "replayed"]),
  operation_key: operationKey,
  // ID は string 値のまま保持（Number へ変換しない）。
  result_refs: z.record(z.string(), z.string()),
  committed_versions: z.record(z.string(), z.string()),
  recorded_at: z.string(),
});
const notSucceededSchema = z.object({
  status: z.literal("not-succeeded"),
  operation_key: operationKey,
});

export const certificateResultSchema = z.union([succeededSchema, notSucceededSchema]);
export type CertificateResult = z.infer<typeof certificateResultSchema>;
export type CertificateSucceededResult = z.infer<typeof succeededSchema>;

/** 成功 result から edition / artifact の ID(string) を取り出す。欠落は契約違反として null。 */
export function readSucceededRefs(
  result: CertificateSucceededResult,
): { editionId: string; artifactId: string } | null {
  const editionId = result.result_refs.edition_id;
  const artifactId = result.result_refs.artifact_id;
  const positive = /^[1-9][0-9]*$/;
  if (editionId === undefined || artifactId === undefined) return null;
  if (!positive.test(editionId) || !positive.test(artifactId)) return null;
  return { editionId, artifactId };
}
