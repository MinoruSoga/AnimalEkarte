import { paths } from "@/config/paths";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { useGetMedicalRecords } from "@/hooks/use-medical-records";

/**
 * EMR-168 案A: 診察カルテ ⇔ トリミング記録の双方向ショートカット解決 hook。
 *
 * 同日・同一ペットの相方レコード/appointment を解決して遷移先を返す。
 * appointments がライフサイクルの SoT で、診察カルテは general・トリミング記録は
 * trimming カテゴリの appointment に紐づく（docs/spec/reservation-to-record-flow.md）。
 * feature 境界（medical-records ↔ trimming の直接 import 禁止）を越えるため
 * 共有 layer の src/hooks に置く。
 *
 * 作成経路は既存の record_shortcut 機構に委譲する:
 * - 診察カルテ: /medical-records/new?petId&visitDate → auto-create が
 *   同日 general appointment を再利用、なければ record_shortcut で新規作成する。
 * - トリミング記録: /trimming/new?petId&appointmentId&visitDate → フォームが
 *   既存 appointment に紐付けて保存（detail 未作成なら upsert）。
 *   appointmentId が無ければ保存時に record_shortcut で新規作成する。
 */

export type PartnerRecordKind = "medical-record" | "trimming";

export interface PartnerRecordTarget {
  /** "open": 相方が既に存在 → 既存データへ。"create": 未存在 → record_shortcut 入力経路へ。 */
  mode: "open" | "create";
  /** 遷移先 URL。reception 既存導線と同じ query 契約（petId / appointmentId / visitDate）。 */
  href: string;
  /** 相方 appointment id。trimming の既存 appointment 再利用時のみ。 */
  appointmentId?: string;
}

export interface PartnerRecordLinkResult {
  isLoading: boolean;
  /** 解決済み遷移先。enabled=false / 日付不正 / 解決中は null。 */
  target: PartnerRecordTarget | null;
}

const ISO_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

/**
 * 終了済み/キャンセルの appointment は「相方なし」扱いにして作成導線を許す。
 * medical-records 側の selectReusableGeneralAppointment と同じ除外集合。
 */
const TERMINAL_APPOINTMENT_STATUSES: ReadonlySet<string> = new Set([
  "completed",
  "cancelled",
  "no_show",
]);

function buildVisitParams(petId: string | undefined, visitDate: string): URLSearchParams {
  const params = new URLSearchParams({ visitDate });
  if (petId) params.set("petId", petId);
  return params;
}

export function usePartnerRecordLink(input: {
  kind: PartnerRecordKind;
  petId: string | undefined;
  /** 相方解決の基準日（YYYY-MM-DD） */
  visitDate: string;
  enabled?: boolean;
}): PartnerRecordLinkResult {
  const { kind, petId, visitDate, enabled = true } = input;
  const active = enabled && Boolean(petId) && ISO_DATE_PATTERN.test(visitDate);

  const medicalRecords = useGetMedicalRecords(
    { petId, startDate: visitDate, endDate: visitDate, limit: 1 },
    { enabled: active && kind === "medical-record" },
  );
  const trimmingAppointments = useGetReservations({
    date: visitDate,
    petId,
    enabled: active && kind === "trimming",
  });

  if (!active) {
    return { isLoading: false, target: null };
  }

  if (kind === "medical-record") {
    if (medicalRecords.isLoading) {
      return { isLoading: true, target: null };
    }
    const record = medicalRecords.data?.data[0];
    if (record) {
      return {
        isLoading: false,
        target: { mode: "open", href: paths.medicalRecords.detail.getHref(record.id) },
      };
    }
    const params = buildVisitParams(petId, visitDate);
    return {
      isLoading: false,
      target: {
        mode: "create",
        href: `${paths.medicalRecords.new.getHref()}?${params.toString()}`,
      },
    };
  }

  if (trimmingAppointments.isLoading) {
    return { isLoading: true, target: null };
  }
  const appointment = (trimmingAppointments.data ?? []).find(
    (candidate) =>
      candidate.category === "trimming" &&
      !TERMINAL_APPOINTMENT_STATUSES.has(String(candidate.status)),
  );
  const params = buildVisitParams(petId, visitDate);
  if (appointment) {
    params.set("appointmentId", appointment.id);
  }
  return {
    isLoading: false,
    target: {
      mode: appointment ? "open" : "create",
      href: `${paths.trimming.new.getHref()}?${params.toString()}`,
      appointmentId: appointment?.id,
    },
  };
}
