// LIFF App 型定義

export interface LiffSettings {
  liff_id: string;
  header_text: string;
  phone_number: string;
  status: "running" | "stopped";
  request_example: string;
  reservation_notice: string;
  cancel_notice: string;
  privacy_policy: string;
  show_no_staff_option: boolean;
  booking_window: number; // 予約可能な日数（今日から何日後まで）
}

export interface OwnerPet {
  id: number;
  name: string;
  animal_species?: { name: string };
}

export interface LiffProfile {
  line_user_id: string;
  display_name: string;
  additional_fields: {
    name?: string;
    phone?: string;
    owner_name?: string;
    pets?: Array<{ name: string; type: string; is_new?: boolean }>;
  };
  // 自動紐付けされたオーナー情報（紐付け済みの場合のみ存在）
  owner?: {
    owner_name: string;
    phone: string;
    pets?: OwnerPet[];
  };
}

export interface Course {
  id: number;
  name: string;
  short_name: string;
  show_short_name: boolean;
  duration_minutes: number;
  reservation_comment: string;
  reservation_image_url: string;
  sort_order: number;
  category?: "general" | "trimming";
}

export interface TrimmingCourse {
  id: number;
  name: string;
  description: string;
  price: number | null;
  sort_order: number;
}

export interface TrimmingOption {
  id: number;
  name: string;
  description: string;
  price: number | null;
  sort_order: number;
}

export interface Staff {
  id: number;
  name: string;
  reservation_comment: string;
  reservation_image_url: string;
  sort_order: number;
}

export interface AvailableDate {
  date: string; // "YYYY-MM-DD"
  available: boolean;
  reason?: "closed" | "holiday" | "staff_off" | "no_slots"; // 予約不可の理由
}

export interface AvailableTime {
  start_time: string; // "HHMM"
  end_time: string; // "HHMM"
  display_time?: string; // Backend未実装のためオプショナル。formatTime でフォールバック
}

// status は backend ReservationStatus (backend/internal/model/reservation.go) と1:1一致。
// EMR-519: id / pet_id は正の int64 範囲の十進文字列。pet_id / pet_name は両方 non-null か
// 両方 null（schemas.ts で旧応答の両方省略も両方 null に正規化する）。
export interface Reservation {
  id: string;
  pet_id: string | null;
  pet_name: string | null;
  course_name: string;
  staff_name: string; // 指名なし予約は backend が省略するため schemas.ts 側で '' デフォルト
  date: string; // "YYYY-MM-DD"
  start_time: string; // "HH:MM"
  end_time: string; // "HH:MM"
  status:
    | "confirmed"
    | "pending"
    | "cancelled"
    | "checked_in"
    | "in_consultation"
    | "accounting"
    | "completed"
    | "no_show";
  notes: string; // 空文字時は backend が省略するため schemas.ts 側で '' デフォルト
  created_at: string;
}

/** POST /reservations のレスポンス（予約作成直後は id と notes のみ返す） */
export interface CreateReservationResponse {
  id: string; // 正の int64 範囲の十進文字列（EMR-519）
  notes: string;
}

export interface PetSelection {
  name: string;
  type: string;
  isNew?: boolean; // true = 自由入力で追加した新規ペット
}

interface CustomerFields {
  customer_name?: string;
  phone?: string;
  owner_name?: string;
  pets?: Array<{ name: string; type: string; is_new?: boolean }>;
}

export interface CreateReservationBody {
  course_id: number;
  staff_id: number;
  date: string; // "YYYY-MM-DD"
  start_time: string; // "HHMM"
  end_time: string; // "HHMM"
  customer_fields: CustomerFields;
  request_text: string;
  trimming_course_id?: number; // BE-120
  trimming_option_ids?: number[]; // BE-120
}

export interface CustomerInfo {
  name: string;
  phone: string;
  ownerName: string;
  pets: PetSelection[];
}

export interface ReservationFlow {
  customerInfo: CustomerInfo;
  courseId: number | null;
  courseName: string;
  courseCategory: "general" | "trimming";
  staffId: number; // 0 = 指名なし
  staffName: string;
  date: string; // "YYYY-MM-DD"
  startTime: string; // "HHMM"
  endTime: string; // "HHMM"
  requestText: string;
  trimmingCourseId: number | null; // BE-120
  trimmingCourseName: string; // BE-120
  trimmingOptionIds: number[]; // BE-120
}

export type PageType =
  | "loading"
  | "error"
  | "maintenance"
  | "top"
  | "my-reservations"
  | "step1"
  | "step2"
  | "step2b" // トリミングコース選択（category=trimming の場合）
  | "step2c" // トリミングオプション選択
  | "step3"
  | "step4"
  | "step5"
  | "step6"
  | "step7"
  | "step8"
  | "revisit-intake"; // EMR-474: 予約単位（1 予約 = 1 ペット）の再診事前問診

// ---------------------------------------------------------------------------
// EMR-474: LIFF 再診事前問診（api.yaml RevisitIntake* 契約の手書き型）
// ID・版・世代は JS Number へ変換しない正規形10進文字列（RevisitIntakeId 等）。
// ---------------------------------------------------------------------------

export type RevisitIntakeEnergyAppetite = "usual" | "slightly_reduced" | "markedly_reduced";
export type RevisitIntakeImprovement =
  "improved" | "slightly_improved" | "unchanged" | "worse" | "new_symptoms";
export type RevisitIntakeExaminationPreference =
  "consultation_only" | "medication_only" | "examination" | "decide_with_vet";

/** RevisitIntakeFullAnswers（応答。8保存値すべて present、任意値は null 可、空文字は null と区別） */
export interface RevisitIntakeFullAnswers {
  visit_purpose: string;
  onset: string | null;
  energy: RevisitIntakeEnergyAppetite | null;
  appetite: RevisitIntakeEnergyAppetite | null;
  current_symptoms: string | null;
  improvement: RevisitIntakeImprovement | null;
  examination_preference: RevisitIntakeExaminationPreference | null;
  main_concern: string | null;
}

/** RevisitIntakeInputAnswers（送信。visit_purpose 必須、任意値は null 可） */
export type RevisitIntakeInputAnswers = RevisitIntakeFullAnswers;

export interface RevisitIntakeTargetRef {
  target_id: string | null;
  appointment_id: string;
  pet_id: string;
  target_generation: string;
}

export interface RevisitIntakeAnswer {
  answer_id: string;
  revision: string;
  form_definition_version: number;
  received_at: string;
  provenance: "owner_submission";
  answers: RevisitIntakeFullAnswers;
}

export type RevisitIntakeEditBlockReason =
  | "reception_started"
  | "cancelled"
  | "deleted"
  | "target_unresolved"
  | "authorization_denied"
  | "unsupported_lifecycle";

export interface RevisitIntakeOwnerCurrentTarget {
  ref: RevisitIntakeTargetRef;
  head_revision: string | null;
  answer_state: "unanswered" | "answered";
  latest_answer: RevisitIntakeAnswer | null;
  content_available: boolean;
  editable: boolean;
  edit_block_reason: RevisitIntakeEditBlockReason | null;
}

export interface RevisitIntakeHistoryEntry {
  target: RevisitIntakeTargetRef & { target_id: string };
  answer: RevisitIntakeAnswer;
  target_relation: "current" | "previous";
}

export interface RevisitIntakeLIFFRead {
  availability: "available" | "target_unresolved";
  form_definition_version: number;
  current_target: RevisitIntakeOwnerCurrentTarget | null;
  history: RevisitIntakeHistoryEntry[];
  next_cursor: string | null;
}

export interface RevisitIntakeSubmitRequest {
  pet_id: string;
  target_generation: string;
  expected_revision: string | null;
  form_definition_version: number;
  answers: RevisitIntakeInputAnswers;
}

export interface RevisitIntakeResultRefs {
  appointment_id: string;
  pet_id: string;
  target_id: string;
  answer_id: string;
}

export interface RevisitIntakeCommittedVersions {
  target_generation: string;
  answer_revision: string;
}

/** RevisitIntakeSuccessResponse（visible / redacted の2形）。 */
export interface RevisitIntakeSucceededResult {
  operation_result: {
    operation_id: string;
    intent: "revisit-intake.submit";
    state: "succeeded";
    committed_at: string;
    result_refs: RevisitIntakeResultRefs | null;
    committed_versions: RevisitIntakeCommittedVersions | null;
  };
  current_state: RevisitIntakeOwnerCurrentTarget | null;
  content_available: boolean;
}

/** RevisitIntakeUnresolvedResponse（claimed_uncommitted / result_unavailable）。失敗確定ではない。 */
export interface RevisitIntakeUnresolvedResult {
  operation_result: {
    operation_id: string;
    intent: "revisit-intake.submit";
    state: "claimed_uncommitted" | "result_unavailable";
  };
  current_state: null;
  content_available: false;
}

export type RevisitIntakeResultResponse =
  RevisitIntakeSucceededResult | RevisitIntakeUnresolvedResult;

/** 結果照会の結果。404 RevisitIntakeLookupNotFound は not_found（失敗確定ではない）として返す。 */
export type RevisitIntakeLookupResult =
  | RevisitIntakeResultResponse
  | {
      operation_result: {
        operation_id: string;
        intent: "revisit-intake.submit";
        state: "not_found";
      };
      current_state: null;
      content_available: false;
    };
