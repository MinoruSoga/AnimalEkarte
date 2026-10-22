/**
 * 予約・予約種別 API の移行 HTTP DTO（非永続）。
 *
 * 保存 model（生成 models.ts）ではなく、現行 handler が返す JSON の列をそのまま写す。
 * Go の json tag に合わせ、omitempty は省略可（?）、omitempty なしのポインタは null 可で表す。
 * lib/transforms・hooks・reception・master が同じ wire を読むため共有層に置く
 * （components/hooks/lib から features は import できない）。新しい予約契約の consumer を
 * 実装するときに、この adapter ごと置き換える（EMR-527）。
 */
import type { ReservationStatus, VisitType } from "./index";
import type {
  DangerLevel,
  PetStatus,
  ReservationDayOption,
  ReservationSource,
  ReservationTypeCategory,
} from "./reservation-type";

/** @see backend/internal/reservation/reservation_type_response.go groupSummary */
export interface LegacyReservationTypeGroupSummary {
  id: number;
  name: string;
  color: string;
}

/** @see backend/internal/reservation/reservation_type_response.go reservationTypeSummary */
export interface LegacyReservationTypeParentSummary {
  id: number;
  name: string;
}

/** @see backend/internal/reservation/reservation_type_response.go ReservationTypeResponse */
export interface LegacyReservationTypeResponse {
  id: number;
  clinic_id: number;
  name: string;
  color: string;
  category: ReservationTypeCategory;
  is_active: boolean;
  description: string;
  sort_order: number;
  group_id?: number;
  group?: LegacyReservationTypeGroupSummary;
  parent_id?: number;
  parent?: LegacyReservationTypeParentSummary;
  children?: LegacyReservationTypeResponse[];
  created_at: string;
  updated_at: string;
  // LINE予約用フィールド
  reservation_display_name: string;
  duration_minutes: number;
  max_concurrent?: number;
  short_name: string;
  show_short_name: boolean;
  reservation_visible: boolean;
  reservation_comment: string;
  reservation_image_url: string;
  reservation_day_option: ReservationDayOption;
  is_internal: boolean;
}

/** @see backend/internal/reservation/nested_summary_response.go staffSummaryResponse */
export interface LegacyReservationStaffSummary {
  id: number;
  name: string;
}

/**
 * is_dangerous はスタッフ向け予約応答のみ（EMR-173）。LIFF/飼主向け契約には出さない。
 * @see backend/internal/reservation/nested_summary_response.go ownerSummaryResponse
 */
export interface LegacyReservationOwnerSummary {
  id: number;
  name: string;
  is_dangerous: boolean;
}

/** @see backend/internal/reservation/nested_summary_response.go animalSpeciesSummaryResponse */
export interface LegacyReservationAnimalSpeciesSummary {
  id: number;
  name: string;
}

/**
 * danger_level はこの予約ドメインのペットサマリだけが持つ（受付の臨床安全バッジ用）。
 * @see backend/internal/reservation/nested_summary_response.go petSummaryResponse
 */
export interface LegacyReservationPetSummary {
  id: number;
  name: string;
  pet_number: string;
  weight?: number;
  status?: PetStatus;
  /** YYYY-MM-DD */
  birth_date?: string;
  deceased_at?: string;
  danger_level?: DangerLevel;
  danger_reason?: string;
  breed?: string;
  animal_species?: LegacyReservationAnimalSpeciesSummary;
  owner?: LegacyReservationOwnerSummary;
}

/**
 * status は "tentative"（EMR-308 次回予定）も届くが、生成 union は含まないため
 * consumer 側は String 比較で判別する。
 * @see backend/internal/reservation/reservation_response.go reservationResponse
 */
export interface LegacyReservationResponse {
  id: number;
  clinic_id: number;
  start_time: string;
  end_time: string;
  owner_id?: number;
  pet_id?: number;
  visit_type: VisitType;
  reservation_type_id: number;
  doctor_id?: number;
  is_designated: boolean;
  status: ReservationStatus;
  notes: string;
  checked_in_at?: string;
  source: ReservationSource;
  created_by?: number;
  line_customer_id?: number;
  is_staff_delegated: boolean;
  /** json.RawMessage（LINE予約のオーナー未紐付け時の入力値）。null もありうる */
  customer_fields: unknown;
  reservation_route?: string;
  actual_reservation_at?: string;
  /** EMR-308: tentative 次回予定の暦日（YYYY-MM-DD、JST） */
  intended_date?: string;
  source_medical_record_id?: number;
  /** EMR-308 A3: intended_date が設定されていたか（cancel 後も保持） */
  had_intended_date: boolean;
  owner?: LegacyReservationOwnerSummary;
  pet?: LegacyReservationPetSummary;
  reservation_type?: LegacyReservationTypeResponse;
  doctor?: LegacyReservationStaffSummary;
  created_by_staff?: LegacyReservationStaffSummary;
  created_at: string;
  updated_at: string;
}
