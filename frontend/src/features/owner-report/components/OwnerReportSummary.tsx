import { C } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { usePermission } from "@/hooks/use-permission";
import { useGetPetVaccinations } from "@/hooks/use-pet-vaccinations";
import { ResourceExaminations, ResourceVaccinations } from "@/types/generated/models";
import type { Pet } from "@/types";
import { useGetMedicalRecords, type MedicalRecord } from "@/features/medical-records";
import { AlertTriangle, CalendarClock, History, Stethoscope, type LucideIcon } from "lucide-react";

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
  icon: LucideIcon;
  tone?: "danger" | "neutral";
  announce?: boolean;
}

const summaryToneClasses = {
  danger: {
    card: `${C.borderDanger20} ${C.bgDanger8}`,
    icon: `${C.bgDanger8} ${C.danger}`,
    value: C.danger,
  },
  neutral: {
    card: `${C.borderLight} ${C.bgWhite}`,
    icon: `${C.bgPage} ${C.text60}`,
    value: C.text,
  },
} as const;

const expandableTextClasses = `cursor-help rounded-sm focus:overflow-visible focus:whitespace-normal focus:text-clip focus:outline-none focus:ring-2 ${C.focusRingBrand}`;

function SummaryCell({
  label,
  value,
  detail,
  icon: Icon,
  tone = "neutral",
  announce = false,
}: SummaryCellProps) {
  const toneClasses = summaryToneClasses[tone];

  return (
    <div
      className={`min-w-0 rounded-lg border px-3 py-2.5 ${toneClasses.card} [@media(max-height:600px)]:py-1.5`}
    >
      <div className="flex min-w-0 items-center gap-2">
        <span
          aria-hidden="true"
          className={`flex size-7 shrink-0 items-center justify-center rounded-lg ${toneClasses.icon} [@media(max-height:600px)]:size-6`}
        >
          <Icon className="size-4" />
        </span>
        <p className={`truncate text-2xs font-semibold ${C.text60}`}>{label}</p>
      </div>
      <p
        className={`mt-1.5 truncate text-base leading-snug font-medium ${toneClasses.value} ${expandableTextClasses} [@media(max-height:600px)]:mt-0.5 [@media(max-height:600px)]:text-sm`}
        role={announce ? "status" : undefined}
        aria-live={announce ? "polite" : undefined}
        aria-label={`${value}。フォーカスすると全文を表示します`}
        tabIndex={0}
        title={value}
      >
        {value}
      </p>
      {detail ? (
        <p
          className={`mt-1 truncate text-xs ${C.text50} ${expandableTextClasses} [@media(max-height:600px)]:hidden`}
          aria-label={`${detail}。フォーカスすると全文を表示します`}
          tabIndex={0}
          title={detail}
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
    <div
      className={`mt-2 overflow-hidden rounded-lg border ${C.borderLight} ${C.bgWhite} [@media(max-height:600px)]:hidden`}
    >
      <div className={`flex items-center justify-between border-b px-3 py-2 ${C.borderDivider} ${C.bgPage30}`}>
        <div className="flex items-center gap-2">
          <History aria-hidden="true" className={`size-4 ${C.text70}`} />
          <h3 className={`text-sm font-medium ${C.text}`}>直近診療タイムライン</h3>
        </div>
        <span className={`rounded-full px-2 py-0.5 text-2xs font-semibold ${C.bgPage} ${C.text70}`}>
          最新{Math.min(records.length, 3)}件
        </span>
      </div>
      <ol
        aria-label="直近診療タイムライン"
        className={`grid gap-0 md:grid-cols-3 md:divide-x ${C.divideDividerFaint}`}
      >
        {records.slice(0, 3).map((record, index) => (
          <li key={record.id} className="relative min-w-0 px-3 py-2.5 pl-8">
            {index < Math.min(records.length, 3) - 1 ? (
              <span
                aria-hidden="true"
                className={`absolute top-5 bottom-0 left-[15px] w-px md:hidden ${C.bgLight}`}
              />
            ) : null}
            <span
              aria-hidden="true"
              className={`absolute top-3.5 left-3 size-2 rounded-full ring-4 ring-white ${C.bgStatusGrayMedium}`}
            />
            <div className="flex min-w-0 items-baseline gap-2">
              <time
                dateTime={record.date ? record.date.replaceAll("/", "-") : undefined}
                className={`shrink-0 font-mono text-xs font-semibold ${C.text70}`}
              >
                {record.date || "-"}
              </time>
              <span
                className={`min-w-0 truncate text-sm font-semibold ${C.text} ${expandableTextClasses}`}
                aria-label={`${record.chiefComplaint || "主訴記録なし"}。フォーカスすると全文を表示します`}
                tabIndex={0}
                title={record.chiefComplaint || "主訴記録なし"}
              >
                {record.chiefComplaint || "主訴記録なし"}
              </span>
            </div>
            <p
              className={`mt-1 truncate text-xs ${C.text50} ${expandableTextClasses}`}
              aria-label={`${
                [record.visitType, record.doctor].filter(Boolean).join(" ・ ") || "担当記録なし"
              }。フォーカスすると全文を表示します`}
              tabIndex={0}
              title={
                [record.visitType, record.doctor].filter(Boolean).join(" ・ ") || "担当記録なし"
              }
            >
              {[record.visitType, record.doctor].filter(Boolean).join(" ・ ") ||
                "担当記録なし"}
            </p>
          </li>
        ))}
      </ol>
    </div>
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
  const hasAbnormalAlert = Boolean(
    examinationPermission.canView &&
      !examinationsQuery.isLoading &&
      !examinationsQuery.isError &&
      abnormalCount > 0,
  );

  return (
    <section
      aria-label="診療サマリー"
      className="mx-3 mt-3 shrink-0 [@media(max-height:600px)]:mt-2"
    >
      <div className="mb-2 flex items-end justify-between [@media(max-height:600px)]:hidden">
        <h2 className={`text-xl font-semibold ${C.text}`}>診療サマリー</h2>
        <p className={`text-xs ${C.text50}`}>選択中ペットの最新情報</p>
      </div>
      <div className="grid grid-cols-2 gap-2 md:grid-cols-4">
        <SummaryCell
          label="要確認"
          value={abnormalValue}
          detail="最新検査の異常値"
          icon={AlertTriangle}
          tone={hasAbnormalAlert ? "danger" : "neutral"}
          announce
        />
        <SummaryCell
          label="次回予定"
          value={vaccinationValue}
          detail={vaccinationDetail}
          icon={CalendarClock}
          tone="neutral"
          announce
        />
        <SummaryCell
          label="初診 / 前回来院"
          value={`${formatDate(firstVisitDate)} / ${formatDate(pet.lastVisit)}`}
          detail="来院履歴"
          icon={History}
        />
        <SummaryCell
          label="直近の診療"
          value={recentValue}
          detail={latestRecord?.date ?? "確定済みカルテ"}
          icon={Stethoscope}
          announce
        />
      </div>
      {medicalRecordsQuery.isLoading ? (
        <p
          className={`mt-2 rounded-lg border px-3 py-2 text-xs ${C.borderLight} ${C.bgWhite} ${C.text50}`}
          role="status"
          aria-live="polite"
        >
          直近診療を読み込み中...
        </p>
      ) : medicalRecordsQuery.isError ? (
        <p
          className={`mt-2 rounded-lg border px-3 py-2 text-xs ${C.borderDanger20} ${C.bgDanger8} ${C.danger}`}
          role="alert"
        >
          直近診療の読み込みに失敗しました
        </p>
      ) : recentRecords.length > 0 ? (
        <RecentMedicalRecordsTimeline records={recentRecords} />
      ) : null}
    </section>
  );
}
