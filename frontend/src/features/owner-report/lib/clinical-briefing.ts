import { formatJSTDate } from "@/lib/jst-date";

export const CLINICAL_HISTORY_KINDS = [
  "診療",
  "検査",
  "薬・処方",
  "予防接種",
  "処置",
  "ケア",
] as const;

export type ClinicalHistoryKind = (typeof CLINICAL_HISTORY_KINDS)[number];

interface MedicalRecordSource {
  id: string;
  date: string;
  chiefComplaint?: string;
  doctor?: string;
  assessment?: string;
  plan?: string;
}

interface ExaminationItemSource {
  name: string;
  inspectionValue?: string;
  result?: string;
  unit?: string;
  isAbnormal: boolean;
}

interface ExaminationSource {
  id: string;
  date: string;
  testType?: string;
  status?: string;
  items?: ReadonlyArray<ExaminationItemSource>;
  resultSummary?: string;
}

interface CheckupSource {
  id: number;
  checkupId: number;
  date: string;
  checkupTypeName: string;
  fieldName: string;
  isAbnormal: boolean;
}

interface TreatmentSource {
  id: string;
  date: string;
  itemType: string;
  name: string;
  adminRoute?: string;
  quantity: number;
  anesthesia?: string;
  isSurgery?: boolean;
}

interface VaccinationSource {
  id: string | number;
  date: string;
  name: string;
  next: string;
  remarks?: string;
}

interface TrimmingSource {
  id: string | number;
  date: string;
  courseName?: string;
  staff?: string;
  status?: string;
}

export interface ClinicalHistorySources {
  medicalRecords: ReadonlyArray<MedicalRecordSource>;
  examinations: ReadonlyArray<ExaminationSource>;
  checkups: ReadonlyArray<CheckupSource>;
  treatments: ReadonlyArray<TreatmentSource>;
  vaccinations: ReadonlyArray<VaccinationSource>;
  trimmings: ReadonlyArray<TrimmingSource>;
}

interface ClinicalHistoryEntry {
  id: string;
  kind: ClinicalHistoryKind;
  dateKey: string;
  source: string;
  title: string;
  detail?: string;
  status?: string;
  isAlert?: boolean;
}

interface ClinicalHistoryColumn {
  dateKey: string;
  label: string;
  entries: ClinicalHistoryEntry[];
}

interface ClinicalHistoryRow {
  kind: ClinicalHistoryKind;
  count: number;
}

export interface ClinicalHistoryMatrix {
  columns: ClinicalHistoryColumn[];
  rows: ClinicalHistoryRow[];
  total: number;
}

function isValidDateParts(year: number, month: number, day: number): boolean {
  const date = new Date(Date.UTC(year, month - 1, day));
  return (
    date.getUTCFullYear() === year &&
    date.getUTCMonth() === month - 1 &&
    date.getUTCDate() === day
  );
}

export function normalizeClinicalDate(value: string): string | undefined {
  const match = value.trim().match(/^(\d{2}|\d{4})[-/](\d{1,2})[-/](\d{1,2})/);
  if (!match) return undefined;

  const rawYear = Number(match[1]);
  const year = match[1].length === 2 ? 2000 + rawYear : rawYear;
  const month = Number(match[2]);
  const day = Number(match[3]);
  if (!isValidDateParts(year, month, day)) return undefined;

  return `${year}-${String(month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
}

function displayClinicalDate(dateKey: string): string {
  return dateKey.replaceAll("-", "/");
}

function joinDetail(parts: ReadonlyArray<string | undefined>): string | undefined {
  const detail = parts.filter((part): part is string => Boolean(part)).join(" / ");
  return detail || undefined;
}

function medicalRecordEntries(
  records: ReadonlyArray<MedicalRecordSource>,
): ClinicalHistoryEntry[] {
  return records.flatMap((record) => {
    const dateKey = normalizeClinicalDate(record.date);
    if (!dateKey) return [];
    return [
      {
        id: `record-${record.id}`,
        kind: "診療" as const,
        dateKey,
        source: record.doctor ? `カルテ・${record.doctor}` : "カルテ",
        title: record.chiefComplaint || "主訴記録なし",
        detail: joinDetail([record.assessment, record.plan]),
      },
    ];
  });
}

function examinationEntries(
  examinations: ReadonlyArray<ExaminationSource>,
): ClinicalHistoryEntry[] {
  return examinations.flatMap((examination) => {
    const dateKey = normalizeClinicalDate(examination.date);
    if (!dateKey) return [];
    const abnormalItems = (examination.items ?? []).filter((item) => item.isAbnormal);
    const abnormalDetail = abnormalItems
      .slice(0, 2)
      .map((item) => {
        const value = item.inspectionValue || item.result || "-";
        return `${item.name} ${value}${item.unit ? ` ${item.unit}` : ""}`;
      })
      .join(" / ");
    const itemDetail = (examination.items ?? [])
      .slice(0, 2)
      .map((item) => item.name)
      .join(" / ");
    return [
      {
        id: `examination-${examination.id}`,
        kind: "検査" as const,
        dateKey,
        source: "検査記録",
        title: examination.testType || "検査",
        detail: abnormalDetail || examination.resultSummary || itemDetail || undefined,
        status: abnormalItems.length > 0 ? "基準外" : examination.status,
        isAlert: abnormalItems.length > 0,
      },
    ];
  });
}

function checkupEntries(checkups: ReadonlyArray<CheckupSource>): ClinicalHistoryEntry[] {
  const groups = checkups.reduce<Record<string, CheckupSource[]>>((current, result) => {
    const key = String(result.checkupId);
    return { ...current, [key]: [...(current[key] ?? []), result] };
  }, {});

  return Object.values(groups).flatMap((results) => {
    const first = results[0];
    const dateKey = first ? normalizeClinicalDate(first.date) : undefined;
    if (!first || !dateKey) return [];
    const abnormalCount = results.filter((result) => result.isAbnormal).length;
    return [
      {
        id: `checkup-${first.checkupId}`,
        kind: "検査" as const,
        dateKey,
        source: "健診記録",
        title: first.checkupTypeName || "健診",
        detail: `${results.length}項目${abnormalCount > 0 ? ` / 要注意 ${abnormalCount}件` : ""}`,
        status: abnormalCount > 0 ? "要注意" : undefined,
        isAlert: abnormalCount > 0,
      },
    ];
  });
}

function treatmentEntries(
  treatments: ReadonlyArray<TreatmentSource>,
  kind: "薬・処方" | "処置",
): ClinicalHistoryEntry[] {
  const isMedicineKind = kind === "薬・処方";
  return treatments.flatMap((treatment) => {
    const isMedicine = treatment.itemType === "medicine";
    if (isMedicine !== isMedicineKind) return [];
    const dateKey = normalizeClinicalDate(treatment.date);
    if (!dateKey) return [];
    return [
      {
        id: `treatment-${kind}-${treatment.id}`,
        kind,
        dateKey,
        source: isMedicine ? "薬・処方" : "処置記録",
        title: treatment.name || "内容記録なし",
        detail: joinDetail([
          treatment.adminRoute,
          treatment.quantity ? `数量 ${treatment.quantity}` : undefined,
          treatment.anesthesia,
        ]),
        status: treatment.isSurgery ? "手術" : undefined,
      },
    ];
  });
}

function vaccinationEntries(
  vaccinations: ReadonlyArray<VaccinationSource>,
): ClinicalHistoryEntry[] {
  return vaccinations.flatMap((vaccination) => {
    const dateKey = normalizeClinicalDate(vaccination.date);
    if (!dateKey) return [];
    return [
      {
        id: `vaccination-${vaccination.id}`,
        kind: "予防接種" as const,
        dateKey,
        source: "接種記録",
        title: vaccination.name || "ワクチン名なし",
        detail: joinDetail([
          vaccination.next && vaccination.next !== "-" ? `次回 ${vaccination.next}` : undefined,
          vaccination.remarks,
        ]),
        status: "実施済",
      },
    ];
  });
}

function trimmingEntries(trimmings: ReadonlyArray<TrimmingSource>): ClinicalHistoryEntry[] {
  return trimmings.flatMap((trimming) => {
    const dateKey = normalizeClinicalDate(trimming.date);
    if (!dateKey) return [];
    return [
      {
        id: `trimming-${trimming.id}`,
        kind: "ケア" as const,
        dateKey,
        source: "ケア記録",
        title: trimming.courseName || "トリミング",
        detail: trimming.staff ? `担当 ${trimming.staff}` : undefined,
        status: trimming.status,
      },
    ];
  });
}

export function buildClinicalHistoryMatrix(
  sources: ClinicalHistorySources,
): ClinicalHistoryMatrix {
  const entries = [
    ...medicalRecordEntries(sources.medicalRecords),
    ...examinationEntries(sources.examinations),
    ...checkupEntries(sources.checkups),
    ...treatmentEntries(sources.treatments, "薬・処方"),
    ...vaccinationEntries(sources.vaccinations),
    ...treatmentEntries(sources.treatments, "処置"),
    ...trimmingEntries(sources.trimmings),
  ];
  const dateKeys = [...new Set(entries.map((entry) => entry.dateKey))]
    .sort((left, right) => right.localeCompare(left));

  return {
    columns: dateKeys.map((dateKey) => ({
      dateKey,
      label: displayClinicalDate(dateKey),
      entries: entries.filter((entry) => entry.dateKey === dateKey),
    })),
    rows: CLINICAL_HISTORY_KINDS.map((kind) => ({
      kind,
      count: entries.filter((entry) => entry.kind === kind).length,
    })),
    total: entries.length,
  };
}

interface AppointmentSource {
  id: string;
  start: Date;
  end: Date;
  status: string;
}

// FE-RC-027: ローカル getter (getFullYear/getMonth/getDate) はクライアント OS の
// タイムゾーンに依存するため、JST 固定の壁時計日付を保証する共通ヘルパーに委譲する
// （@/lib/jst-date と同じ UTC オフセット計算契約に統一）。
function wallDateISO(date: Date): string {
  return formatJSTDate(date);
}

const INACTIVE_FUTURE_STATUSES = new Set(["cancelled", "no_show", "completed"]);

export function selectAppointmentBriefing<T extends AppointmentSource>(
  reservations: ReadonlyArray<T>,
  today: string,
): { today?: T; next?: T } {
  const sorted = [...reservations].sort((left, right) => left.start.getTime() - right.start.getTime());
  const todayReservations = sorted.filter((reservation) => wallDateISO(reservation.start) === today);
  const todayActive = todayReservations.find(
    (reservation) => !INACTIVE_FUTURE_STATUSES.has(reservation.status),
  );
  const next = sorted.find(
    (reservation) =>
      wallDateISO(reservation.start) > today &&
      !INACTIVE_FUTURE_STATUSES.has(reservation.status),
  );

  return { today: todayActive ?? todayReservations[0], next };
}
