/**
 * トリミング API 型定義（BE-119: appointments ベース）
 *
 * FE7-3: BackendTrimming は tygo 生成の TrimmingResponse
 * （internal/trimming/trimming_response.go の TrimmingResponse に対応）
 * を継承する。pet/staff の2フィールドのみ手書き — 生成型はこの2フィールドが参照する
 * petSummaryResponse/staffSummaryResponse（11+18箇所で共有される非公開型）を
 * tygo が解決できないため tstype:"-" で生成対象から除外されている（BE の json タグ・
 * wire 形状自体は不変。フィールドは実際のレスポンスに引き続き含まれる）。
 */
import type { ReservationRoute } from "@/features/reservations";
import type { TrimmingResponse } from "@/types/generated/trimming-responses";

// -------------------------------------------------------
// Backend Response DTO (trimming_response.go に対応)
// -------------------------------------------------------

/** ペット情報 summary（関連レコード。生成型が tstype:"-" で除外しているため手書き維持） */
interface PetSummary {
  id: number;
  name: string;
  pet_number?: string;
  weight?: number | null;
  /** #231: 犬種等。空文字/未設定は breed 情報なしを示す。 */
  breed?: string;
  animal_species?: { id: number; name: string } | null;
  owner?: { id: number; name: string } | null;
}

/** スタッフ情報 summary（関連レコード。生成型が tstype:"-" で除外しているため手書き維持） */
interface StaffSummary {
  id: number;
  name: string;
}

/**
 * トリミング予約 API レスポンス DTO
 * internal/trimming/trimming_response.go の TrimmingResponse に対応。
 * id/clinic_id/status/course/options 等のフラットフィールドは生成型を継承するため
 * BE 側の形状変更（フィールド追加・optional 化等）は type-check で検知できる。
 */
export interface BackendTrimming extends TrimmingResponse {
  pet?: PetSummary | null;
  staff?: StaffSummary | null;
}

// -------------------------------------------------------
// Request Types
// -------------------------------------------------------

/**
 * トリミング作成リクエスト（POST /v1/trimmings）
 * internal/trimming/trimming_request.go の createTrimmingRequest に対応
 */
export interface CreateTrimmingRequest {
  appointment_id?: number;
  reservation_type_id: number;
  start_time?: string;
  end_time?: string;
  pet_id?: number | null;
  staff_id?: number | null;
  status?: string;
  reservation_route?: ReservationRoute;
  course_id?: number | null;
  style_request?: string;
  bw?: number | null;
  bw_unit?: string;
  bt?: number | null;
  used_shampoo?: string;
  used_ribbon?: string;
  remarks?: string;
  style_image?: string;
  completed_image?: string;
  option_ids?: number[];
}

/**
 * トリミング更新リクエスト（PATCH /v1/trimmings/:id — 全フィールドoptional）
 * FE3-6: CreateTrimmingRequest からフィールド集合が完全一致（3 フィールド除く）だったため
 * 手書き二重定義を廃し派生型化。値・フィールド集合は変更していない。
 * internal/trimming/trimming_request.go の updateTrimmingRequest に対応
 */
export type UpdateTrimmingRequest = Omit<CreateTrimmingRequest, "appointment_id" | "reservation_type_id" | "reservation_route">;

// -------------------------------------------------------
// List Response
// -------------------------------------------------------

/**
 * トリミング一覧APIレスポンス
 */
export interface TrimmingListResponse {
  data: BackendTrimming[];
  total: number;
  page: number;
  limit: number;
}

// -------------------------------------------------------
// UI Form Data (フォーム専用・手書きOK)
// -------------------------------------------------------

/**
 * トリミングフォーム入力データ（UI専用）
 */
export interface TrimmingFormData {
  reservationTypeId: string;
  startTime: string;
  endTime: string;
  styleRequest: string;
  styleImage: File | null;
  bw: string;
  bwUnit: "Kg" | "g";
  bt: string;
  usedShampoo: string;
  usedRibbon: string;
  remarks: string;
  completedImage: File | null;
  courseId: string;
  optionIds: string[];
  staffId: string;
  staffName: string;
  /** #233: カルテ画面から直接新規作成する場合のみ選択可能な登録時ステータス。デフォルトは in_consultation。 */
  initialStatus: "in_consultation" | "pending";
}
