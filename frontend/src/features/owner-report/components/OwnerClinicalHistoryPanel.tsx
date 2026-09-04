import { C } from "@/lib/design-tokens";

import type { OwnerClinicalBriefingData } from "../hooks/use-owner-clinical-briefing-data";
import { buildClinicalHistoryMatrix, type ClinicalHistoryKind } from "../lib/clinical-briefing";
import { ClinicalBriefingPanel } from "./ClinicalBriefingPanel";
import { ClinicalHistoryMatrix, type HistoryRowState } from "./ClinicalHistoryMatrix";

function rowState(
  permitted: boolean,
  query: { isLoading: boolean; isError: boolean },
): HistoryRowState {
  if (!permitted) return "no-permission";
  if (query.isLoading) return "loading";
  if (query.isError) return "error";
  return "ready";
}

function examinationRowState(data: OwnerClinicalBriefingData): HistoryRowState {
  const accessCount =
    Number(data.permissions.examination.canView) + Number(data.permissions.checkup.canView);
  const state = rowState(accessCount > 0, {
    isLoading:
      (data.permissions.examination.canView && data.examinationsQuery.isLoading) ||
      (data.permissions.checkup.canView && data.checkupsQuery.isLoading),
    isError:
      (data.permissions.examination.canView && data.examinationsQuery.isError) ||
      (data.permissions.checkup.canView && data.checkupsQuery.isError),
  });
  return state === "ready" && accessCount === 1 ? "partial-permission" : state;
}

function historyRowStates(
  data: OwnerClinicalBriefingData,
): Record<ClinicalHistoryKind, HistoryRowState> {
  return {
    診療: rowState(true, data.medicalRecordsQuery),
    検査: examinationRowState(data),
    "薬・処方": rowState(true, data.treatmentsQuery),
    予防接種: rowState(data.permissions.vaccination.canView, data.vaccinationsQuery),
    処置: rowState(true, data.treatmentsQuery),
    ケア: rowState(data.permissions.trimming.canView, data.trimmingQuery),
  };
}

function historyMatrix(data: OwnerClinicalBriefingData) {
  return buildClinicalHistoryMatrix({
    medicalRecords: data.medicalRecordsQuery.data?.data ?? [],
    examinations: data.permissions.examination.canView
      ? (data.examinationsQuery.data?.items ?? [])
      : [],
    checkups: data.permissions.checkup.canView ? (data.checkupsQuery.data ?? []) : [],
    treatments: data.treatmentsQuery.data?.items ?? [],
    vaccinations: data.permissions.vaccination.canView ? (data.vaccinationsQuery.data ?? []) : [],
    trimmings: data.permissions.trimming.canView ? (data.trimmingQuery.data?.items ?? []) : [],
  });
}

function historyIsTruncated(data: OwnerClinicalBriefingData): boolean {
  const medicalRecords = data.medicalRecordsQuery.data;
  return [
    Boolean(medicalRecords && medicalRecords.total > medicalRecords.data.length),
    data.examinationsQuery.data?.isTruncated,
    data.treatmentsQuery.data?.isTruncated,
    data.trimmingQuery.data?.isTruncated,
  ].some(Boolean);
}

export function ClinicalHistoryPanel({ data }: { data: OwnerClinicalBriefingData }) {
  const matrix = historyMatrix(data);
  const truncated = historyIsTruncated(data);
  return (
    <ClinicalBriefingPanel
      title="種類別履歴"
      description="縦=種類、横=日付（新しい順）"
      count={`${matrix.total}件${truncated ? "+" : ""}`}
      areaClassName="owner-report-area-history"
      bodyClassName="p-0"
      bodyTestId="owner-report-history-scroll"
    >
      <ClinicalHistoryMatrix matrix={matrix} rowStates={historyRowStates(data)} />
      {truncated ? (
        <p className={`sticky left-0 px-2 py-1 text-2xs ${C.text50}`}>
          取得上限を超える履歴があります。件数は表示中の範囲です。
        </p>
      ) : null}
    </ClinicalBriefingPanel>
  );
}
