import { z } from "zod";
import { axios } from "@/lib/axios";

/**
 * EMR-524: カルテ側 package preview/apply/result照会/共通plan read の wire 型と
 * zod 解析（手書き・契約mock段階）。
 * 契約正本: backend/docs/api.yaml `TreatmentPackage*` / `MedicalRecordPlan*` /
 * `MedicalRecordPlanOperationResultResponse`（採用 SHA f1ded6ec367459d91baec08414ce780df39e9a3f）
 * と docs/work/emr/emr-518-package-api-contract-20261020.md。
 * runtime apply 有効化は本票の非対象（BE未登録）。production は fixture を成功
 * データとして読まない。ID/version/円/quantity/tax_rate は**文字列**のまま。
 * number 化・float 演算・丸め・0 補完をしない。解析は fail-closed。
 */

const int64 = z
  .string()
  .regex(/^[1-9][0-9]*$/)
  .max(19);
const decimal = z
  .string()
  .regex(/^-?(0|[1-9][0-9]*)(\.[0-9]*[1-9])?$/)
  .refine((v) => v !== "-0");
const uuid = z
  .string()
  .regex(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/)
  .refine((v) => v !== "00000000-0000-0000-0000-000000000000");

export const planOperationKeySchema = uuid;
export const previewTokenSchema = z.string().regex(/^pkg-preview-v1:[0-9a-f]{64}$/);
export const sourceTokenSchema = z.string().regex(/^pkg-source-v1:[0-9a-f]{64}$/);

const manualSchema = z
  .object({
    name: z.string().min(1).regex(/\S/),
    reason: z.string().min(1).regex(/\S/),
    unit_label: z.string().nullable(),
    unit_price: z
      .string()
      .regex(/^(0|[1-9][0-9]*)$/)
      .max(19),
    tax_type: z.enum(["included", "excluded", "exempt"]),
    tax_rate: decimal,
    is_non_insurance: z.boolean(),
  })
  .strict();

export const planItemRefSchema = z.discriminatedUnion("kind", [
  z.object({ kind: z.literal("consultation"), consultation_id: int64 }).strict(),
  z.object({ kind: z.literal("procedure"), procedure_id: int64 }).strict(),
  z.object({ kind: z.literal("medicine"), medicine_id: int64 }).strict(),
  z.object({ kind: z.literal("examination_type"), examination_type_id: int64 }).strict(),
  z.object({ kind: z.literal("vaccine"), vaccine_id: int64 }).strict(),
  z.object({ kind: z.literal("checkup_type"), checkup_type_id: int64 }).strict(),
  z.object({ kind: z.literal("merchandise_item"), merchandise_item_id: int64 }).strict(),
  z.object({ kind: z.literal("manual_other"), manual: manualSchema }).strict(),
]);
export type PlanItemRef = z.infer<typeof planItemRefSchema>;

const uncertaintySchema = z
  .object({
    field: z.enum(["unit_price", "tax", "insurance"]),
    reason: z.enum(["legacy_unset", "not_configured_on_master", "pending_owner_confirmation"]),
    required_at: z.enum(["selection", "execution", "billing"]),
  })
  .strict();

export const priceSnapshotSchema = z
  .object({
    unit_price: z
      .string()
      .regex(/^(0|[1-9][0-9]*)$/)
      .max(19)
      .nullable(),
    tax_type: z.enum(["included", "excluded", "exempt"]).nullable(),
    tax_rate: decimal.nullable(),
    is_non_insurance: z.boolean().nullable(),
    insurance_setting_state: z.enum(["explicit", "legacy_unset", "not_configured_on_master"]),
    price_source: z.enum(["master", "owner-policy", "manual", "unresolved"]),
    tax_source: z.enum(["master", "owner-policy", "manual", "unresolved"]),
    insurance_source: z.enum(["master", "owner-policy", "manual", "unresolved"]),
    amount_basis: z.enum(["included", "excluded", "exempt", "unresolved"]),
    uncertainty: z.array(uncertaintySchema),
  })
  .strict();
export type TreatmentPackagePriceSnapshot = z.infer<typeof priceSnapshotSchema>;

const componentIssueSchema = z
  .object({
    component_key: uuid,
    severity: z.enum(["blocking", "warning"]),
    required_at: z.enum(["selection", "execution", "billing"]),
    code: z.enum([
      "inactive-item",
      "invalid-quantity",
      "price-unresolved",
      "tax-unresolved",
      "insurance-unresolved",
      "owner-contract-unresolved",
      "invalid-reference",
      "owner-selection-rule-rejected",
    ]),
    message: z.string(),
  })
  .strict();
export type TreatmentPackageComponentIssue = z.infer<typeof componentIssueSchema>;

const sourceStampSchema = z
  .object({
    native_version: int64.nullable(),
    dependency_token: sourceTokenSchema,
  })
  .strict();

export const previewItemSchema = z
  .object({
    component_key: uuid,
    item: planItemRefSchema,
    quantity: z.string().regex(/^([1-9][0-9]*(\.[0-9]*[1-9])?|0\.[0-9]*[1-9])$/),
    name: z.string(),
    unit_label: z.string().nullable(),
    source_stamp: sourceStampSchema,
    price: priceSnapshotSchema,
    classification: z.enum(["clinical", "commercial", "unresolved"]),
    issues: z.array(componentIssueSchema),
  })
  .strict();
export type TreatmentPackagePreviewItem = z.infer<typeof previewItemSchema>;

const previewTotalSchema = z
  .object({
    status: z.enum(["available", "unresolved"]),
    amount: z
      .string()
      .regex(/^(0|[1-9][0-9]*)$/)
      .max(19)
      .nullable(),
  })
  .strict();
export type TreatmentPackagePreviewTotal = z.infer<typeof previewTotalSchema>;

export const treatmentPackagePreviewSchema = z
  .object({
    record_id: int64,
    clinic_id: int64,
    pet_id: int64,
    package_id: int64,
    package_version: int64,
    plan_collection_version: int64.nullable(),
    version_state: z.enum(["ready", "uninitialized"]),
    preview_token: previewTokenSchema,
    applicable: z.boolean(),
    items: z.array(previewItemSchema).min(1),
    total: previewTotalSchema,
  })
  .strict();
export type TreatmentPackagePreview = z.infer<typeof treatmentPackagePreviewSchema>;

export const previewRequestSchema = z
  .object({ package_id: int64, package_version: int64 })
  .strict();
export type TreatmentPackagePreviewRequest = z.infer<typeof previewRequestSchema>;

/** apply body: 平坦 EMR494 形。clinic/price/items override なし。 */
export const treatmentPackageApplySchema = z
  .object({
    operation_key: uuid,
    edit_session_ref: uuid,
    edit_epoch: int64,
    owner_expected_version: int64,
    package_id: int64,
    package_version: int64,
    preview_token: previewTokenSchema,
  })
  .strict();
export type TreatmentPackageApply = z.infer<typeof treatmentPackageApplySchema>;

const successSchema = z
  .object({
    status: z.enum(["succeeded", "replayed"]),
    operation_key: uuid,
    result_refs: z.object({ record_id: int64, application_id: int64 }).strict(),
    committed_versions: z.object({ application: int64, plan_items: int64 }).strict(),
    recorded_at: z.string(),
  })
  .strict();
const nonSuccessSchema = z
  .object({
    status: z.enum([
      "not-succeeded",
      "operation-key-conflict",
      "operation-in-progress",
      "result-unavailable",
      "scope-not-found",
      "not-authorized",
      "validation",
      "version-conflict",
      "dependency-changed",
      "edit-right-lost",
    ]),
    operation_key: uuid,
  })
  .strict();
export const operationResultSchema = z.discriminatedUnion("status", [
  successSchema.extend({ status: z.literal("succeeded") }),
  successSchema.extend({ status: z.literal("replayed") }),
  nonSuccessSchema,
]);
export type MedicalRecordPlanOperationResult = z.infer<typeof operationResultSchema>;
export type OperationSuccess = Extract<
  MedicalRecordPlanOperationResult,
  { status: "succeeded" | "replayed" }
>;
export type OperationNonSuccessStatus = Exclude<
  MedicalRecordPlanOperationResult["status"],
  "succeeded" | "replayed"
>;

const previewItemSnapshotSchema = z
  .object({
    component_key: uuid,
    item: planItemRefSchema,
    quantity: previewItemSchema.shape.quantity,
    name: z.string(),
    unit_label: z.string().nullable(),
    source_stamp: sourceStampSchema,
    price: priceSnapshotSchema,
    classification: z.enum(["clinical", "commercial"]),
  })
  .strict();

export const applicationSummarySchema = z
  .object({
    id: int64,
    record_id: int64,
    package_id: int64,
    package_version: int64,
    package_name_snapshot: z.string(),
    applied_at: z.string(),
    item_count: z.number().int().min(1),
  })
  .strict();
export type TreatmentPackageApplicationSummary = z.infer<typeof applicationSummarySchema>;

export const applicationSchema = z
  .object({
    id: int64,
    record_id: int64,
    package_id: int64,
    package_version: int64,
    package_name_snapshot: z.string(),
    applied_at: z.string(),
    item_count: z.number().int().min(1),
    clinic_id: int64,
    pet_id: int64,
    version: int64,
    operation_key: uuid,
    items: z
      .array(
        z
          .object({
            component_key: uuid,
            plan_item_id: int64,
            applied_item: previewItemSnapshotSchema,
          })
          .strict(),
      )
      .min(1),
  })
  .strict();
export type TreatmentPackageApplication = z.infer<typeof applicationSchema>;

export const applicationListSchema = z
  .object({ items: z.array(applicationSummarySchema) })
  .strict();
export type TreatmentPackageApplicationList = z.infer<typeof applicationListSchema>;

const originSchema = z.discriminatedUnion("kind", [
  z.object({ kind: z.literal("single") }).strict(),
  z
    .object({
      kind: z.literal("package"),
      application_id: int64,
      package_id: int64,
      package_version: int64,
      component_key: uuid,
    })
    .strict(),
]);

const executionSchema = z.discriminatedUnion("kind", [
  z
    .object({
      kind: z.literal("clinical"),
      state: z.enum(["not_performed", "performed", "deferred", "cancelled"]),
      revision_id: int64,
      result_state: z.enum(["none", "awaiting", "linked"]),
    })
    .strict(),
  z
    .object({
      kind: z.literal("commercial"),
      state: z.literal("selected"),
      revision_id: int64,
    })
    .strict(),
]);

const ownerLinkSchema = z
  .object({
    id: int64,
    owner_kind: z.enum(["treatment", "vaccination", "examination", "checkup", "billing_item"]),
    owner_id: int64,
    owner_version: int64.nullable(),
    fixed_revision_id: int64.nullable(),
    status: z.enum(["active", "revoked"]),
  })
  .strict();

export const planItemSchema = z
  .object({
    id: int64,
    version: int64,
    record_id: int64,
    clinic_id: int64,
    pet_id: int64,
    item: planItemRefSchema,
    name_snapshot: z.string(),
    unit_label_snapshot: z.string().nullable(),
    quantity: previewItemSchema.shape.quantity,
    price_snapshot: priceSnapshotSchema,
    origin: originSchema,
    execution: executionSchema,
    owner_links: z.array(ownerLinkSchema),
  })
  .strict();
export type MedicalRecordPlanItem = z.infer<typeof planItemSchema>;

export const planListSchema = z
  .object({
    collection_version: int64.nullable(),
    version_state: z.enum(["ready", "uninitialized"]),
    items: z.array(planItemSchema),
  })
  .strict();
export type MedicalRecordPlanList = z.infer<typeof planListSchema>;

// ── API 接続（record 自身の clinicId を X-Clinic-ID で送る点は clinical-plan と同じ）──
function clinicHeaderConfig(clinicId?: string) {
  return clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined;
}
const base = (recordId: string) => `/v1/medical-records/${recordId}`;

/** read-only POST preview。選択だけで実施/結果/在庫/請求/証明書は起動しない。 */
export async function previewTreatmentPackage(
  recordId: string,
  req: TreatmentPackagePreviewRequest,
  clinicId?: string,
): Promise<TreatmentPackagePreview> {
  const { data } = await axios.post<unknown>(
    `${base(recordId)}/treatment-package-previews`,
    req,
    clinicHeaderConfig(clinicId),
  );
  return treatmentPackagePreviewSchema.parse(data);
}

/** C06 wire 応答の三分類。ResultResponse body は HTTP status を問わず確定結果として解釈し、
 *  前段 error envelope と通信断/解析失敗を区別する。 */
export type OperationCallOutcome =
  | { kind: "result"; result: MedicalRecordPlanOperationResult }
  | { kind: "error-envelope"; httpStatus: number; code: string | null }
  | { kind: "unavailable" };

function asErrorEnvelope(data: unknown): { code: string | null } | null {
  if (typeof data !== "object" || data === null) return null;
  if (typeof (data as { error?: unknown }).error !== "string") return null;
  const code = (data as { code?: unknown }).code;
  return { code: typeof code === "string" ? code : null };
}

function operationCallOutcome(data: unknown, httpStatus: number): OperationCallOutcome {
  const parsed = operationResultSchema.safeParse(data);
  if (parsed.success) return { kind: "result", result: parsed.data };
  const envelope = asErrorEnvelope(data);
  if (envelope !== null) {
    return { kind: "error-envelope", httpStatus, code: envelope.code };
  }
  return { kind: "unavailable" };
}

/** apply は C06 wire。同 key 照会のみ許し、盲目再送しない。
 *  非2xxでも ResultResponse body を解析し、確定非成功と通信断・結果不明を分ける。 */
export async function applyTreatmentPackage(
  recordId: string,
  req: TreatmentPackageApply,
  clinicId?: string,
): Promise<OperationCallOutcome> {
  try {
    const res = await axios.post<unknown>(`${base(recordId)}/treatment-package-applications`, req, {
      ...clinicHeaderConfig(clinicId),
      validateStatus: () => true,
    });
    return operationCallOutcome(res.data, res.status);
  } catch {
    return { kind: "unavailable" };
  }
}

export async function listTreatmentPackageApplications(
  recordId: string,
  clinicId?: string,
): Promise<TreatmentPackageApplicationList> {
  const { data } = await axios.get<unknown>(
    `${base(recordId)}/treatment-package-applications`,
    clinicHeaderConfig(clinicId),
  );
  return applicationListSchema.parse(data);
}

export async function getTreatmentPackageApplication(
  recordId: string,
  applicationId: string,
  clinicId?: string,
): Promise<TreatmentPackageApplication> {
  const { data } = await axios.get<unknown>(
    `${base(recordId)}/treatment-package-applications/${applicationId}`,
    clinicHeaderConfig(clinicId),
  );
  return applicationSchema.parse(data);
}

/** 結果不明は元 key 照会のみ。intent は medical_record.package_apply 固定。
 *  非2xxでも ResultResponse body を解析し、確定非成功と通信断・結果不明を分ける。 */
export async function getMedicalRecordOperationResult(
  recordId: string,
  operationKey: string,
  clinicId?: string,
): Promise<OperationCallOutcome> {
  try {
    const res = await axios.get<unknown>(`${base(recordId)}/operations/${operationKey}`, {
      ...clinicHeaderConfig(clinicId),
      params: { intent: "medical_record.package_apply" },
      validateStatus: () => true,
    });
    return operationCallOutcome(res.data, res.status);
  } catch {
    return { kind: "unavailable" };
  }
}

export async function listMedicalRecordPlanItems(
  recordId: string,
  clinicId?: string,
): Promise<MedicalRecordPlanList> {
  const { data } = await axios.get<unknown>(
    `${base(recordId)}/plan-items`,
    clinicHeaderConfig(clinicId),
  );
  return planListSchema.parse(data);
}

export async function getMedicalRecordPlanItem(
  recordId: string,
  planItemId: string,
  clinicId?: string,
): Promise<MedicalRecordPlanItem> {
  const { data } = await axios.get<unknown>(
    `${base(recordId)}/plan-items/${planItemId}`,
    clinicHeaderConfig(clinicId),
  );
  return planItemSchema.parse(data);
}

// ── 候補一覧（EMR-523 master と同じ wire。cross-feature import を避けるため
//    同一契約を本ファイルへ手書きで複写する）──
const packageComponentSchema = z
  .object({
    component_key: uuid,
    item: planItemRefSchema,
    quantity: previewItemSchema.shape.quantity,
  })
  .strict();
export const treatmentPackageListItemSchema = z
  .object({
    id: int64,
    clinic_id: int64,
    version: int64,
    name: z.string().min(1),
    description: z.string(),
    is_active: z.boolean(),
    sort_order: z.number().int().min(0),
    components: z.array(packageComponentSchema).min(1),
    created_at: z.string(),
    updated_at: z.string(),
  })
  .strict();
export type TreatmentPackageCandidate = z.infer<typeof treatmentPackageListItemSchema>;
export const treatmentPackageListResponseSchema = z
  .object({ items: z.array(treatmentPackageListItemSchema), order_version: int64 })
  .strict();

export async function listTreatmentPackageCandidates(
  clinicId?: string,
): Promise<TreatmentPackageCandidate[]> {
  const { data } = await axios.get<unknown>(
    "/v1/masters/treatment-packages",
    clinicHeaderConfig(clinicId),
  );
  return treatmentPackageListResponseSchema.parse(data).items;
}
