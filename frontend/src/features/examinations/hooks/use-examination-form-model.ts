import { jstDateStartISOString, todayJSTISO } from "@/lib/jst-date";
import { EXAM_STATUS_EN_TO_JA } from "@/lib/transforms/examination";
import type { ExaminationRecord } from "../api/transforms";
import type { ExamResult } from "../api/transforms";
import type {
  CreateExaminationRequest,
  UpdateExaminationRequest,
  UpsertExamItemRequest,
} from "../types";
import type { ExamItemRow } from "../components/ExamItemsTable";
import type { ExamTypeFieldRow } from "../api/get-exam-type-fields";
import {
  isPersistedCompletedSeal,
  isPersistedConfirmedStatus,
  isPersistedResultsLocked,
} from "../lib/examination-lock";

/** EXAM_STATUS_EN_TO_JA（正本）の逆写像を導出する（FE5-10）。両写像は完全対称であることを確認済み。 */
const EXAM_STATUS_JA_TO_EN = Object.fromEntries(
  Object.entries(EXAM_STATUS_EN_TO_JA).map(([en, ja]) => [ja, en]),
) as Record<string, "pending" | "in_progress" | "result_entered" | "completed" | "confirmed">;

export interface ExaminationMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  canUnconfirm: boolean;
}

export const DENIED_MUTATION_PERMISSIONS: Readonly<ExaminationMutationPermissions> = {
  canCreate: false,
  canEdit: false,
  canDelete: false,
  canUnconfirm: false,
};

// テンプレ（exam_type_fields）から ExamItemRow の初期行を組み立てる。
// status/isAbnormal は backend が保存後に導出するため未設定で開始する。
export function buildRowsFromTemplate(fields: ExamTypeFieldRow[]): ExamItemRow[] {
  return fields.map((f, idx) => ({
    key: `tmpl-${f.id}`,
    examTypeFieldId: f.id,
    name: f.name,
    inspectionValue: "",
    unit: f.unit,
    normalValue: f.normalValue,
    referenceValue: f.normalValue,
    sortOrder: f.sortOrder !== 0 ? f.sortOrder : idx,
  }));
}

// formItems → PUT リクエスト形式へ変換。空の項目（name 空 & 値空）は送信しない。
export function rowsToRequest(items: readonly ExamItemRow[]): UpsertExamItemRequest[] {
  return items
    .filter((it) => it.name.trim() !== "")
    .map((it, idx) => ({
      exam_type_field_id: it.examTypeFieldId ?? null,
      name: it.name,
      inspection_value: it.inspectionValue,
      normal_value: it.normalValue,
      unit: it.unit,
      reference_value: it.referenceValue,
      sort_order: it.sortOrder !== 0 ? it.sortOrder : idx,
    }));
}

export function mapExamResultsToFormRows(existingItems: readonly ExamResult[]): ExamItemRow[] {
  return existingItems.map((it) => ({
    key: `srv-${it.id}`,
    examTypeFieldId: it.examTypeFieldId,
    name: it.name,
    inspectionValue: it.inspectionValue,
    unit: it.unit,
    normalValue: it.normalValue,
    referenceValue: it.referenceValue,
    refMin: it.refMin,
    refMax: it.refMax,
    isAssessed: it.isAssessed,
    sortOrder: it.sortOrder,
    status: it.status,
    isAbnormal: it.isAbnormal,
  }));
}

export function createBlankExaminationForm(
  doctorId: string | null,
  localOverrides: Partial<ExaminationRecord>,
): Partial<ExaminationRecord> {
  return {
    status: "依頼中" as const,
    ownerName: "",
    petName: "",
    ...(doctorId && { doctorId }),
    ...localOverrides,
  };
}

export function omitCorrectedExaminationFieldErrors(
  previous: Record<string, string>,
  next: Partial<ExaminationRecord>,
): Record<string, string> {
  if (!("testTypeId" in next) && !("doctorId" in next)) {
    return previous;
  }
  let changed = false;
  const updated = { ...previous };
  if ("testTypeId" in next && updated.testTypeId) {
    delete updated.testTypeId;
    changed = true;
  }
  if ("doctorId" in next && updated.doctorId) {
    delete updated.doctorId;
    changed = true;
  }
  return changed ? updated : previous;
}

export function validateExaminationSave(input: {
  current: Partial<ExaminationRecord>;
  isEdit: boolean;
  isCurrentEditTarget: boolean;
  resultsLocked: boolean;
  areCurrentItemsReady: boolean;
  formItems: readonly ExamItemRow[];
}): Record<string, string> {
  const errors: Record<string, string> = {};
  if (!input.current.testTypeId) errors.testTypeId = "検査種別を選択してください";
  if (!input.current.doctorId) errors.doctorId = "担当医を選択してください";
  if (
    input.isEdit &&
    (!input.isCurrentEditTarget || (!input.resultsLocked && !input.areCurrentItemsReady))
  ) {
    errors.examItems = "検査項目の読み込み完了後に保存してください";
  }
  if (
    input.formItems.some((item) => item.name.trim() === "" && item.inspectionValue.trim() !== "")
  ) {
    errors.examItems = "結果値を入力した手動項目には項目名が必要です";
  }
  return errors;
}

export function isExaminationPatientChangeLocked(
  isEdit: boolean,
  canEdit: boolean,
  existingExam: ExaminationRecord | undefined,
): boolean {
  return (
    !isEdit ||
    !canEdit ||
    !existingExam ||
    isPersistedConfirmedStatus(existingExam.status) ||
    existingExam.currentRevisionVersion !== undefined
  );
}

export type ExaminationPatientChangeDecision =
  { kind: "unchanged" } | { kind: "apply"; petId: number } | { kind: "blocked" };

export function decideExaminationPatientChange(input: {
  currentPetId: string | undefined;
  existingPetId: string | undefined;
  isPatientChangeLocked: boolean;
  changedPatient: { id: string; status: string } | undefined;
}): ExaminationPatientChangeDecision {
  const patientChanged =
    input.currentPetId !== undefined && input.currentPetId !== input.existingPetId;
  if (!patientChanged) return { kind: "unchanged" };
  const changedPetID = Number(input.currentPetId);
  const canApply =
    !input.isPatientChangeLocked &&
    input.changedPatient?.id === input.currentPetId &&
    input.changedPatient?.status === "生存" &&
    Number.isSafeInteger(changedPetID) &&
    changedPetID > 0;
  if (!canApply) return { kind: "blocked" };
  return { kind: "apply", petId: changedPetID };
}

export function buildUpdateExaminationRequest(input: {
  current: Partial<ExaminationRecord>;
  items: UpsertExamItemRequest[];
  resultsLocked: boolean;
  patientChange: ExaminationPatientChangeDecision;
}): UpdateExaminationRequest {
  return {
    status: input.current.status ? EXAM_STATUS_JA_TO_EN[input.current.status] : undefined,
    result_summary: input.current.resultSummary,
    machine: input.current.machine,
    doctor_id: input.current.doctorId ? Number(input.current.doctorId) : null,
    date: input.current.date
      ? input.current.date.includes("T")
        ? input.current.date
        : jstDateStartISOString(input.current.date)
      : undefined,
    ...(!input.resultsLocked ? { items: input.items } : {}),
    ...(input.patientChange.kind === "apply" ? { pet_id: input.patientChange.petId } : {}),
  };
}

export function buildCreateExaminationRequest(input: {
  current: Partial<ExaminationRecord>;
  medicalRecordId: string;
  petId: string;
  items: UpsertExamItemRequest[];
}): CreateExaminationRequest {
  return {
    medical_record_id: input.medicalRecordId ? Number(input.medicalRecordId) : null,
    pet_id: Number(input.petId) || null,
    exam_type_id: Number(input.current.testTypeId) || 0,
    doctor_id: input.current.doctorId ? Number(input.current.doctorId) : null,
    date: input.current.date ?? jstDateStartISOString(todayJSTISO()),
    result_summary: input.current.resultSummary,
    machine: input.current.machine,
    // 新規は常に items を送る（作成時点で「確定」を選んでもロック対象はサーバ確定後のみ）
    items: input.items,
  };
}

export function deriveExaminationLockFlags(
  isEdit: boolean,
  existingExam: ExaminationRecord | undefined,
) {
  return {
    isPersistedConfirmed: isEdit && isPersistedConfirmedStatus(existingExam?.status),
    isPersistedCompletedLocked:
      isEdit &&
      isPersistedCompletedSeal(existingExam?.status, existingExam?.currentRevisionVersion),
    isPersistedResultsLocked:
      isEdit &&
      isPersistedResultsLocked(existingExam?.status, existingExam?.currentRevisionVersion),
  };
}
