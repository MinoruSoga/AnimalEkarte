import { C } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { usePermission } from "@/hooks/use-permission";
import { useGetPetVaccinations } from "@/hooks/use-pet-vaccinations";
import { ResourceExaminations, ResourceVaccinations } from "@/types/generated/models";
import type { Pet } from "@/types";
import { useGetMedicalRecords, type MedicalRecord } from "@/features/medical-records";

import { useGetPetExaminations } from "../api/get-pet-examinations";
import {
  countLatestExaminationAbnormalResults,
  selectLatestOverdueVaccination,
  selectUpcomingVaccination,
} from "../lib/report-summary";

interface OwnerReportSummaryProps {
  pet: Pet;
  firstVisitDate?: string | null;
}

interface SummaryCellProps {
  label: string;
  value: string;
  detail?: string;
  isAlert?: boolean;
  announce?: boolean;
}

function SummaryCell({ label, value, detail, isAlert = false, announce = false }: SummaryCellProps) {
  return (
    <div
      className={`min-w-0 px-3 py-2.5 ${C.bgWhite} [@media(max-height:600px)]:py-1.5`}
    >
      <p className={`text-2xs font-medium tracking-wide ${C.text50}`}>{label}</p>
      <p
        className={`mt-0.5 break-words text-sm leading-snug font-semibold ${isAlert ? C.danger : C.text}`}
        role={announce ? "status" : undefined}
        aria-live={announce ? "polite" : undefined}
      >
        {value}
      </p>
      {detail ? (
        <p
          className={`mt-0.5 break-words text-2xs ${C.text50} [@media(max-height:600px)]:hidden`}
        >
          {detail}
        </p>
      ) : null}
    </div>
  );
}

function queryStateValue(
  canView: boolean,
  isLoading: boolean,
  isError: boolean,
  value: string,
): string {
  if (!canView) return "閲覧権限なし";
  if (isLoading) return "読み込み中...";
  if (isError) return "取得失敗";
  return value;
}

function RecentMedicalRecordsTimeline({ records }: { records: MedicalRecord[] }) {
  return (
    <ol
      aria-label="直近診療タイムライン"
      className={`grid gap-0 border-t ${C.borderDivider} md:grid-cols-3 md:divide-x ${C.divideDividerFaint} [@media(max-height:600px)]:hidden`}
    >
      {records.slice(0, 3).map((record, index) => (
        <li key={record.id} className="relative min-w-0 px-3 py-2 pl-7">
          <span
            aria-hidden="true"
            className={`absolute top-3 left-3 size-2 rounded-full ${
              index === 0 ? C.bgBrandDot : C.bgStatusGrayMedium
            }`}
          />
          <div className="flex min-w-0 items-baseline gap-2">
            <time
              dateTime={record.date ? record.date.replaceAll("/", "-") : undefined}
              className={`shrink-0 font-mono text-xs ${C.text70}`}
            >
              {record.date || "-"}
            </time>
            <span className={`break-words text-sm font-medium ${C.text}`}>
              {record.chiefComplaint || "主訴記録なし"}
            </span>
          </div>
          <p className={`mt-0.5 break-words text-2xs ${C.text50}`}>
            {[record.visitType, record.doctor].filter(Boolean).join(" ・ ") || "担当記録なし"}
          </p>
        </li>
      ))}
    </ol>
  );
}

export function OwnerReportSummary({ pet, firstVisitDate }: OwnerReportSummaryProps) {
  const examinationPermission = usePermission(ResourceExaminations);
  const vaccinationPermission = usePermission(ResourceVaccinations);
  const examinationsQuery = useGetPetExaminations(
    examinationPermission.canView ? pet.id : undefined,
  );
  const vaccinationsQuery = useGetPetVaccinations(
    vaccinationPermission.canView ? pet.id : undefined,
  );
  const medicalRecordsQuery = useGetMedicalRecords({
    petId: pet.id,
    status: "finalized",
    sort: "date",
    order: "desc",
    page: 1,
    limit: 5,
  });

  const abnormalCount = countLatestExaminationAbnormalResults(
    examinationsQuery.data?.items ?? [],
  );
  const upcomingVaccination = selectUpcomingVaccination(vaccinationsQuery.data ?? []);
  const overdueVaccination = upcomingVaccination
    ? undefined
    : selectLatestOverdueVaccination(vaccinationsQuery.data ?? []);
  const vaccinationSchedule = upcomingVaccination ?? overdueVaccination;
  const recentRecords = medicalRecordsQuery.data?.data ?? [];
  const latestRecord = recentRecords[0];

  const abnormalValue = queryStateValue(
    examinationPermission.canView,
    examinationsQuery.isLoading,
    examinationsQuery.isError,
    `${abnormalCount}件`,
  );
  const vaccinationValue = queryStateValue(
    vaccinationPermission.canView,
    vaccinationsQuery.isLoading,
    vaccinationsQuery.isError,
    vaccinationSchedule?.name ?? "予定なし",
  );
  const vaccinationDetail =
    vaccinationPermission.canView &&
    !vaccinationsQuery.isLoading &&
    !vaccinationsQuery.isError &&
    vaccinationSchedule
      ? `${vaccinationSchedule.next}${overdueVaccination ? "（予定日経過）" : ""}`
      : "予防接種";
  const recentValue = medicalRecordsQuery.isLoading
    ? "読み込み中..."
    : medicalRecordsQuery.isError
      ? "取得失敗"
      : latestRecord?.chiefComplaint || (latestRecord ? "主訴記録なし" : "診療記録なし");

  return (
    <section
      aria-label="診療サマリー"
      className={`mx-2 mt-2 shrink-0 overflow-hidden rounded-lg border ${C.borderLight} ${C.bgWhite}`}
    >
      <div className={`grid grid-cols-2 gap-px ${C.bgLight} md:grid-cols-4`}>
        <SummaryCell
          label="要確認"
          value={abnormalValue}
          detail="最新検査の異常値"
          announce
          isAlert={Boolean(
            examinationPermission.canView &&
              !examinationsQuery.isLoading &&
              !examinationsQuery.isError &&
              abnormalCount > 0,
          )}
        />
        <SummaryCell
          label="次回予定"
          value={vaccinationValue}
          detail={vaccinationDetail}
          announce
        />
        <SummaryCell
          label="初診 / 前回来院"
          value={`${formatDate(firstVisitDate)} / ${formatDate(pet.lastVisit)}`}
          detail="来院履歴"
        />
        <SummaryCell
          label="直近の診療"
          value={recentValue}
          detail={latestRecord?.date ?? "確定済みカルテ"}
          announce
        />
      </div>
      {medicalRecordsQuery.isLoading ? (
        <p
          className={`border-t px-3 py-2 text-xs ${C.borderDivider} ${C.text50}`}
          role="status"
          aria-live="polite"
        >
          直近診療を読み込み中...
        </p>
      ) : medicalRecordsQuery.isError ? (
        <p className={`border-t px-3 py-2 text-xs ${C.borderDivider} ${C.danger}`} role="alert">
          直近診療の読み込みに失敗しました
        </p>
      ) : recentRecords.length > 0 ? (
        <RecentMedicalRecordsTimeline records={recentRecords} />
      ) : null}
    </section>
  );
}
