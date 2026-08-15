import type { ExaminationPrintSnapshot } from "../api/get-examination-print-snapshot";

export interface ExaminationPrintRow {
  id: string;
  name: string;
  inspectionValue: string;
  unit: string;
  referenceValue: string;
  statusLabel: string;
  isAssessed: boolean;
  isAbnormal: boolean;
}

export interface ExaminationPrintModel {
  title: string;
  watermark: string;
  isDraft: boolean;
  version: number;
  date: string;
  examTypeName: string;
  petName: string;
  ownerName: string;
  speciesName: string;
  doctorName: string;
  medicalRecordNo: string;
  resultSummary: string;
  machine: string;
  rows: ExaminationPrintRow[];
}

function statusLabel(status: string, isAssessed: boolean): string {
  if (!isAssessed) {
    return "—";
  }
  switch (status) {
    case "high":
      return "H";
    case "low":
      return "L";
    case "normal":
      return "N";
    default:
      return status || "—";
  }
}

/**
 * Pure print view-model builder. Uses only the server snapshot DTO —
 * never formItems, live masters, or FE range recalculation.
 */
export function buildExaminationPrintModel(
  snapshot: ExaminationPrintSnapshot,
): ExaminationPrintModel {
  const isDraft = snapshot.printBoundary === "draft";
  return {
    title: "検査結果",
    watermark: isDraft
      ? snapshot.watermark || "DRAFT / 未確定"
      : "",
    isDraft,
    version: snapshot.version,
    date: snapshot.date,
    examTypeName: snapshot.display.examTypeName,
    petName: snapshot.display.petName,
    ownerName:
      snapshot.display.petOwnerName ||
      snapshot.display.medicalRecordOwnerName,
    speciesName: snapshot.display.speciesName,
    doctorName: snapshot.display.doctorName,
    medicalRecordNo: snapshot.display.medicalRecordNo,
    resultSummary: snapshot.resultSummary,
    machine: snapshot.machine,
    rows: snapshot.items
      .slice()
      .sort((a, b) => a.sortOrder - b.sortOrder || Number(a.id) - Number(b.id))
      .map((item) => ({
        id: item.id,
        name: item.name,
        inspectionValue: item.inspectionValue || "—",
        unit: item.unit,
        referenceValue: item.referenceValue || "—",
        statusLabel: statusLabel(item.status, item.isAssessed),
        isAssessed: item.isAssessed,
        isAbnormal: item.isAbnormal,
      })),
  };
}
