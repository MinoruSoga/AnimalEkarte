/**
 * トリミング API 型定義（BE-119: appointments ベース）
 *
 * NOTE: BackendTrimming はハンドラの trimmingResponse DTO であり、
 * models.ts の Go モデルとは別物。make codegen では更新されない。
 */

// -------------------------------------------------------
// Backend Response DTO (trimming_response.go に対応)
// -------------------------------------------------------

/** トリミング予約詳細の option summary */
export interface TrimmingOptionSummary {
  id: number;
  name: string;
}

/** トリミングコース summary */
export interface TrimmingCourseSummary {
  id: number;
  name: string;
  price: number;
}

/** ペット情報 summary（関連レコード） */
export interface PetSummary {
  id: number;
  name: string;
  pet_number?: string;
  weight?: number | null;
  animal_species?: { id: number; name: string } | null;
  owner?: { id: number; name: string } | null;
}

/** スタッフ情報 summary（関連レコード） */
export interface StaffSummary {
  id: number;
  name: string;
}

/**
 * トリミング予約 API レスポンス DTO
 * handler/trimming_response.go の trimmingResponse に対応
 */
export interface BackendTrimming {
  id: number;
  clinic_id: number;
  reservation_type_id: number;
  start_time: string;
  end_time: string;
  pet_id?: number | null;
  staff_id?: number | null;
  status: string;
  source: string;
  // トリミング詳細（appointment_trimming_details から flat 化）
  course_id?: number | null;
  style_request: string;
  bw?: number | null;
  bw_unit: string;
  bt?: number | null;
  used_shampoo: string;
  used_ribbon: string;
  remarks: string;
  style_image: string;
  completed_image: string;
  created_at: string;
  updated_at: string;
  // リレーション
  pet?: PetSummary | null;
  staff?: StaffSummary | null;
  course?: TrimmingCourseSummary | null;
  options: TrimmingOptionSummary[];
}

// -------------------------------------------------------
// Request Types
// -------------------------------------------------------

/**
 * トリミング作成リクエスト（POST /v1/trimmings）
 * handler/trimming_request.go の createTrimmingRequest に対応
 */
export interface CreateTrimmingRequest {
  reservation_type_id: number;
  start_time: string;
  end_time: string;
  pet_id?: number | null;
  staff_id?: number | null;
  status?: string;
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
 * handler/trimming_request.go の updateTrimmingRequest に対応
 */
export interface UpdateTrimmingRequest {
  start_time?: string;
  end_time?: string;
  pet_id?: number | null;
  staff_id?: number | null;
  status?: string;
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
  memo: string;
  eggs: string;
  parts: {
    nail: boolean;
    analGland: boolean;
    eye: boolean;
    ear: boolean;
    skin: boolean;
    oral: boolean;
  };
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
}
