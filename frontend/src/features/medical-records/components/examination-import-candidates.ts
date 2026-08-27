import type { ExaminationRecord } from "@/lib/transforms/examination";

/** Statuses BE still accepts for parent-field updates (e.g. medical_record_id link). */
const IMPORTABLE_STATUSES = new Set([
  "依頼中",
  "検査中",
  "結果入力済み",
  "完了",
  "pending",
  "in_progress",
  "result_entered",
  "completed",
]);

export type ExaminationImportCandidate = Pick<
  ExaminationRecord,
  "id" | "medicalRecordId" | "status" | "currentRevisionVersion"
>;

/**
 * BUG-014: import targets must be editable for medical_record_id attach.
 * - confirmed (確定): examinationFullyLocked → 409
 * - any revision history: patient relation change rejected → 409
 * - already linked to another chart: not a candidate for this dialog
 * - unknown status: fail closed
 */
export function isExaminationImportable(
  exam: ExaminationImportCandidate,
  medicalRecordId?: string,
): boolean {
  if (exam.medicalRecordId && exam.medicalRecordId !== medicalRecordId) {
    return false;
  }
  if (exam.currentRevisionVersion != null) {
    return false;
  }
  if (!exam.status || !IMPORTABLE_STATUSES.has(exam.status)) {
    return false;
  }
  return true;
}

export function filterImportableExaminations<T extends ExaminationImportCandidate>(
  examinations: T[],
  medicalRecordId?: string,
): T[] {
  return examinations.filter((exam) => isExaminationImportable(exam, medicalRecordId));
}
