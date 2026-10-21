import { z } from "zod";
import { axios } from "@/lib/axios";

/**
 * EMR-523: 診療項目パッケージ master API の wire 型と zod 解析（手書き）。
 * 契約正本: backend/docs/api.yaml `TreatmentPackage*`（採用 SHA f1ded6ec367459d91baec08414ce780df39e9a3f）
 * と docs/work/emr/emr-518-package-api-contract-20261020.md。
 * `types/generated` は編集しない（`make codegen` はユーザー実行）。
 * 本ファイルから `@/types/generated/models` は import しない（TASK-444-S1 allowlist fail-closed）。
 *
 * ID / version / 円 / quantity / tax_rate は**文字列**（canonical signed64 十進・非負整数円・
 * ExactNumeric 十進）。number 化・float 演算・丸め・0 補完をしない。
 * 解析は fail-closed: 契約外の形は例外にして、推測で補完しない。
 */

const canonicalPositiveInt64 = z.string().regex(/^[1-9][0-9]*$/).max(19);
const canonicalDecimal = z
  .string()
  .regex(/^-?(0|[1-9][0-9]*)(\.[0-9]*[1-9])?$/)
  .refine((v) => v !== "-0", "-0 は許可されません");

export const treatmentPackageIdSchema = canonicalPositiveInt64;
export const treatmentPackageVersionSchema = canonicalPositiveInt64;
export const treatmentPackageQuantitySchema = z
  .string()
  .regex(/^([1-9][0-9]*(\.[0-9]*[1-9])?|0\.[0-9]*[1-9])$/);
export const treatmentPackageDecimalSchema = canonicalDecimal;
export const treatmentPackageYenSchema = z.string().regex(/^(0|[1-9][0-9]*)$/).max(19);
export const treatmentPackageCanonicalUuidSchema = z
  .string()
  .regex(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/)
  .refine(
    (v) => v !== "00000000-0000-0000-0000-000000000000",
    "nil UUID は許可されません",
  );

export const treatmentPackageManualDefinitionSchema = z
  .object({
    name: z.string().min(1).regex(/\S/),
    reason: z.string().min(1).regex(/\S/),
    unit_label: z.string().nullable(),
    unit_price: treatmentPackageYenSchema,
    tax_type: z.enum(["included", "excluded", "exempt"]),
    tax_rate: canonicalDecimal,
    is_non_insurance: z.boolean(),
  })
  .strict();
export type TreatmentPackageManualDefinition = z.infer<
  typeof treatmentPackageManualDefinitionSchema
>;

/** ItemRef: 7 typed master + 正規 manual_other の closed 8 分岐。他型 FK 混入は拒否。 */
export const treatmentPackageItemRefSchema = z.discriminatedUnion("kind", [
  z.object({ kind: z.literal("consultation"), consultation_id: treatmentPackageIdSchema }).strict(),
  z.object({ kind: z.literal("procedure"), procedure_id: treatmentPackageIdSchema }).strict(),
  z.object({ kind: z.literal("medicine"), medicine_id: treatmentPackageIdSchema }).strict(),
  z
    .object({
      kind: z.literal("examination_type"),
      examination_type_id: treatmentPackageIdSchema,
    })
    .strict(),
  z.object({ kind: z.literal("vaccine"), vaccine_id: treatmentPackageIdSchema }).strict(),
  z.object({ kind: z.literal("checkup_type"), checkup_type_id: treatmentPackageIdSchema }).strict(),
  z
    .object({
      kind: z.literal("merchandise_item"),
      merchandise_item_id: treatmentPackageIdSchema,
    })
    .strict(),
  z
    .object({ kind: z.literal("manual_other"), manual: treatmentPackageManualDefinitionSchema })
    .strict(),
]);
export type TreatmentPackageItemRef = z.infer<typeof treatmentPackageItemRefSchema>;
export type TreatmentPackageItemKind = TreatmentPackageItemRef["kind"];

export const TREATMENT_PACKAGE_ITEM_KINDS = [
  "consultation",
  "procedure",
  "medicine",
  "examination_type",
  "vaccine",
  "checkup_type",
  "merchandise_item",
  "manual_other",
] as const satisfies readonly TreatmentPackageItemKind[];

export const treatmentPackageComponentInputSchema = z
  .object({
    component_key: treatmentPackageCanonicalUuidSchema,
    item: treatmentPackageItemRefSchema,
    quantity: treatmentPackageQuantitySchema,
  })
  .strict();
export type TreatmentPackageComponentInput = z.infer<
  typeof treatmentPackageComponentInputSchema
>;

export const treatmentPackageCreateSchema = z
  .object({
    name: z.string().min(1).max(255).regex(/\S/),
    description: z.string(),
    is_active: z.boolean(),
    components: z.array(treatmentPackageComponentInputSchema).min(1),
  })
  .strict();
export type TreatmentPackageCreate = z.infer<typeof treatmentPackageCreateSchema>;

/** PATCH: 省略フィールド = 維持。明示 null は schema 上受理しない。 */
export const treatmentPackagePatchSchema = z
  .object({
    version: treatmentPackageVersionSchema,
    name: z.string().min(1).max(255).regex(/\S/).optional(),
    description: z.string().optional(),
    is_active: z.boolean().optional(),
    components: z.array(treatmentPackageComponentInputSchema).min(1).optional(),
  })
  .strict();
export type TreatmentPackagePatch = z.infer<typeof treatmentPackagePatchSchema>;

export const treatmentPackageSchema = z
  .object({
    id: treatmentPackageIdSchema,
    clinic_id: treatmentPackageIdSchema,
    version: treatmentPackageVersionSchema,
    name: z.string().min(1).max(255).regex(/\S/),
    description: z.string(),
    is_active: z.boolean(),
    sort_order: z.number().int().min(0),
    components: z.array(treatmentPackageComponentInputSchema).min(1),
    created_at: z.string(),
    updated_at: z.string(),
  })
  .strict();
export type TreatmentPackage = z.infer<typeof treatmentPackageSchema>;

export const treatmentPackageListSchema = z
  .object({
    items: z.array(treatmentPackageSchema),
    order_version: treatmentPackageVersionSchema,
  })
  .strict();
export type TreatmentPackageList = z.infer<typeof treatmentPackageListSchema>;

export const treatmentPackageReorderSchema = z
  .object({
    order_version: treatmentPackageVersionSchema,
    ids: z.array(treatmentPackageIdSchema),
  })
  .strict();
export type TreatmentPackageReorder = z.infer<typeof treatmentPackageReorderSchema>;

const fieldIssueSchema = z
  .object({ path: z.string(), code: z.string(), message: z.string() })
  .strict();
export const treatmentPackageErrorSchema = z
  .object({
    error: z.string(),
    code: z.enum([
      "invalid-input",
      "unauthorized",
      "not-authorized",
      "scope-not-found",
      "version-conflict",
      "dependency-changed",
      "edit-right-lost",
      "validation",
      "internal-error",
      "temporarily-unavailable",
    ]),
    issues: z.array(fieldIssueSchema).optional(),
  })
  .strict();
export type TreatmentPackageErrorBody = z.infer<typeof treatmentPackageErrorSchema>;

// ─────────────────────────────────────────────────
// API 接続（本番コードは fixture を成功データとして読まない。mock はテスト側に分離）
// ─────────────────────────────────────────────────

const BASE = "/v1/masters/treatment-packages";

export async function listTreatmentPackages(): Promise<TreatmentPackageList> {
  const { data } = await axios.get<unknown>(BASE);
  return treatmentPackageListSchema.parse(data);
}

export async function getTreatmentPackage(packageId: string): Promise<TreatmentPackage> {
  const { data } = await axios.get<unknown>(`${BASE}/${packageId}`);
  return treatmentPackageSchema.parse(data);
}

/**
 * POST create。応答が不明（timeout / 5xx / parse 失敗）でも、呼び出し側は
 * 同名検索や盲目の再 POST で救済しない（契約: "Ambiguous master create must not
 * be blindly retried or inferred by name"）。react-query mutation の既定 retry=0 を維持する。
 */
export async function createTreatmentPackage(
  req: TreatmentPackageCreate,
): Promise<TreatmentPackage> {
  const { data } = await axios.post<unknown>(BASE, req);
  return treatmentPackageSchema.parse(data);
}

export async function updateTreatmentPackage(
  packageId: string,
  req: TreatmentPackagePatch,
): Promise<TreatmentPackage> {
  const { data } = await axios.patch<unknown>(`${BASE}/${packageId}`, req);
  return treatmentPackageSchema.parse(data);
}

export async function deleteTreatmentPackage(
  packageId: string,
  packageVersion: string,
): Promise<void> {
  await axios.delete(`${BASE}/${packageId}`, { params: { version: packageVersion } });
}

/** 非削除全件（inactive 含む）の完全メンバーシップ reorder。204。 */
export async function reorderTreatmentPackages(req: TreatmentPackageReorder): Promise<void> {
  await axios.patch(`${BASE}/reorder`, req);
}
