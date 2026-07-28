import { C } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { formatJSTWallDate, formatJSTWallTime } from "@/lib/jst-date";
import type { Reservation } from "@/lib/transforms/reservation";
import type { Pet } from "@/types";
import { RESERVATION_STATUS_LABELS } from "@/types";

import type { OwnerClinicalBriefingData } from "../hooks/use-owner-clinical-briefing-data";
import {
  normalizeClinicalDate,
  selectAppointmentBriefing,
} from "../lib/clinical-briefing";
import {
  countLatestExaminationAbnormalResults,
  selectLatestOverdueVaccination,
  selectUpcomingVaccination,
} from "../lib/report-summary";
import { ClinicalBriefingPanel } from "./ClinicalBriefingPanel";
import {
  BriefingField,
  DataStatus,
  DetailField,
} from "./ClinicalBriefingFields";

interface DataPanelProps {
  data: OwnerClinicalBriefingData;
}

function latestExamination(data: OwnerClinicalBriefingData) {
  const examinations = data.permissions.examination.canView
    ? (data.examinationsQuery.data?.items ?? [])
    : [];
  return [...examinations]
    .filter((item) => normalizeClinicalDate(item.date))
    .sort((left, right) => right.date.localeCompare(left.date))[0];
}

function latestMedicine(data: OwnerClinicalBriefingData) {
  const treatments = data.treatmentsQuery.data?.items ?? [];
  return [...treatments]
    .filter(
      (item) =>
        item.itemType === "medicine" && normalizeClinicalDate(item.date),
    )
    .sort((left, right) => {
      const rightDate = normalizeClinicalDate(right.date) ?? "";
      const leftDate = normalizeClinicalDate(left.date) ?? "";
      return rightDate.localeCompare(leftDate);
    })[0];
}

function vaccinationBriefing(data: OwnerClinicalBriefingData) {
  const vaccinations = data.permissions.vaccination.canView
    ? (data.vaccinationsQuery.data ?? [])
    : [];
  const upcoming = selectUpcomingVaccination(vaccinations, data.today);
  const overdue = selectLatestOverdueVaccination(vaccinations, data.today);
  return {
    plan: upcoming ?? overdue,
    isOverdue: !upcoming && Boolean(overdue),
  };
}

function PreVisitStatusFields({ data }: DataPanelProps) {
  const examinations = data.permissions.examination.canView
    ? (data.examinationsQuery.data?.items ?? [])
    : [];
  const examination = latestExamination(data);
  const abnormalCount = countLatestExaminationAbnormalResults(examinations);
  const vaccination = vaccinationBriefing(data);

  return (
    <div
      className={`mt-1 grid min-w-0 grid-cols-2 gap-1 border-t pt-1 ${C.borderDivider}`}
    >
      <DataStatus
        noPermission={!data.permissions.examination.canView}
        isLoading={data.examinationsQuery.isLoading}
        isError={data.examinationsQuery.isError}
        emptyMessage="検査記録なし"
      >
        {examination ? (
          <BriefingField
            label="最新検査"
            value={`${examination.testType || "検査"}・${abnormalCount > 0 ? `基準外 ${abnormalCount}件` : examination.items?.length ? "基準外なし" : "結果明細なし"}`}
            alert={abnormalCount > 0}
          />
        ) : undefined}
      </DataStatus>
      <DataStatus
        noPermission={!data.permissions.vaccination.canView}
        isLoading={data.vaccinationsQuery.isLoading}
        isError={data.vaccinationsQuery.isError}
        emptyMessage="予防予定なし"
      >
        {vaccination.plan ? (
          <BriefingField
            label="予防予定"
            value={`${vaccination.plan.name}・${formatDate(vaccination.plan.nextDate)}`}
            alert={vaccination.isOverdue}
          />
        ) : undefined}
      </DataStatus>
    </div>
  );
}

interface PreVisitPanelProps extends DataPanelProps {
  pet: Pet;
}

export function PreVisitPanel({ data, pet }: PreVisitPanelProps) {
  const medicine = latestMedicine(data);
  const medicineValue = data.treatmentsQuery.isLoading
    ? "読み込み中..."
    : data.treatmentsQuery.isError
      ? "取得失敗"
      : medicine?.name || "確定情報なし";

  return (
    <ClinicalBriefingPanel
      title="診療前の確認"
      description="要確認事項を先に確認"
      areaClassName="owner-report-area-attention"
      bodyClassName="p-1.5"
    >
      <div className="grid min-w-0 grid-cols-3 gap-1 max-[680px]:grid-cols-1">
        <BriefingField
          label="診療メモ"
          value={pet.remarks || "記載なし"}
          alert={Boolean(pet.remarks)}
        />
        <BriefingField label="血液型" value={pet.bloodType || "記録なし"} />
        <BriefingField
          label="直近の薬・処方"
          value={medicineValue}
          alert={data.treatmentsQuery.isError}
        />
      </div>
      <PreVisitStatusFields data={data} />
    </ClinicalBriefingPanel>
  );
}

function reservationVisitType(reservation: Reservation): string {
  return reservation.visitType === "first" ? "初診" : "再診";
}

function reservationStatus(reservation: Reservation): string {
  return RESERVATION_STATUS_LABELS[reservation.status] ?? reservation.status;
}

export function TodayVisitPanel({ data }: DataPanelProps) {
  const reservations = data.permissions.reservation.canView
    ? (data.reservationsQuery.data ?? [])
    : [];
  const appointment = selectAppointmentBriefing(reservations, data.today).today;

  return (
    <ClinicalBriefingPanel
      title="今日の来院"
      description="本日の予約・受付状況"
      areaClassName="owner-report-area-today"
      bodyClassName="p-1.5"
    >
      <DataStatus
        noPermission={!data.permissions.reservation.canView}
        isLoading={data.reservationsQuery.isLoading}
        isError={data.reservationsQuery.isError}
        emptyMessage="本日の予約なし"
      >
        {appointment ? (
          <div className="grid grid-cols-3 gap-1">
            <BriefingField
              label="時刻"
              value={formatJSTWallTime(appointment.start)}
            />
            <BriefingField
              label="区分"
              value={reservationVisitType(appointment)}
            />
            <BriefingField
              label="状態"
              value={reservationStatus(appointment)}
            />
            <BriefingField
              label="予約内容"
              value={appointment.type || "記録なし"}
            />
            <BriefingField label="担当" value={appointment.doctor || "未定"} />
          </div>
        ) : undefined}
      </DataStatus>
    </ClinicalBriefingPanel>
  );
}

function nextReservationValue(data: OwnerClinicalBriefingData): string {
  if (!data.permissions.reservation.canView) return "閲覧権限なし";
  if (data.reservationsQuery.isLoading) return "読み込み中...";
  if (data.reservationsQuery.isError) return "取得失敗";
  const appointment = selectAppointmentBriefing(
    data.reservationsQuery.data ?? [],
    data.today,
  ).next;
  return appointment
    ? `${formatJSTWallDate(appointment.start)} ${formatJSTWallTime(appointment.start)}`
    : "予約なし";
}

function recommendedVisitValue(data: OwnerClinicalBriefingData): string {
  if (data.medicalRecordsQuery.isLoading) return "読み込み中...";
  if (data.medicalRecordsQuery.isError) return "取得失敗";
  const recommendedDate =
    data.medicalRecordsQuery.data?.data[0]?.nextVisitRecommendedDate;
  return recommendedDate ? formatDate(recommendedDate) : "記録なし";
}

function vaccinationPlanValue(data: OwnerClinicalBriefingData): string {
  if (!data.permissions.vaccination.canView) return "閲覧権限なし";
  if (data.vaccinationsQuery.isLoading) return "読み込み中...";
  if (data.vaccinationsQuery.isError) return "取得失敗";
  const vaccination = vaccinationBriefing(data);
  return vaccination.plan
    ? `${formatDate(vaccination.plan.nextDate)}${vaccination.isOverdue ? "・期限超過" : ""}`
    : "予定なし";
}

export function NextActionPanel({ data }: DataPanelProps) {
  const vaccination = vaccinationBriefing(data);
  return (
    <ClinicalBriefingPanel
      title="次の行動"
      description="予約・再診・予防の予定"
      areaClassName="owner-report-area-next"
      bodyClassName="p-1.5"
    >
      <div className="grid grid-cols-3 gap-1 max-[680px]:grid-cols-1">
        <BriefingField label="次回予約" value={nextReservationValue(data)} />
        <BriefingField
          label="来院推奨"
          value={recommendedVisitValue(data)}
          alert={data.medicalRecordsQuery.isError}
        />
        <BriefingField
          label="予防予定"
          value={vaccinationPlanValue(data)}
          alert={data.vaccinationsQuery.isError || vaccination.isOverdue}
        />
      </div>
    </ClinicalBriefingPanel>
  );
}

export function PreviousVisitPanel({ data }: DataPanelProps) {
  const record = data.medicalRecordsQuery.data?.data[0];
  return (
    <ClinicalBriefingPanel
      title="前回診療"
      description="直近の確定カルテ"
      areaClassName="owner-report-area-last"
      bodyClassName="p-1.5"
    >
      <DataStatus
        isLoading={data.medicalRecordsQuery.isLoading}
        isError={data.medicalRecordsQuery.isError}
        emptyMessage="確定カルテなし"
      >
        {record ? (
          <dl className="grid grid-cols-2 gap-x-2">
            <DetailField label="診療日" value={record.date} />
            <DetailField label="担当" value={record.doctor} />
            <DetailField label="診療区分" value={record.visitType} />
            <DetailField label="主訴" value={record.chiefComplaint} />
            <DetailField label="評価" value={record.assessment} />
            <DetailField label="方針" value={record.plan} />
          </dl>
        ) : undefined}
      </DataStatus>
    </ClinicalBriefingPanel>
  );
}
