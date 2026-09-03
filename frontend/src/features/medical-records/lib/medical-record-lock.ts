/**
 * BUG-035: finalized medical-record edit lock.
 * UI labels and BE wire use distinct strings; lock only on the FE domain label
 * produced by transformMedicalRecord (STATUS_MAP finalized → 確定済).
 */
export function isMedicalRecordFinalizedStatus(status: string | null | undefined): boolean {
  return status === "確定済";
}
