import type { ReservationDayOption } from "@/types/generated/models";
import type { ReorderRequest } from "@/types/form";

// EMR-527: 予約の移行 HTTP DTO（./reservation-legacy-wire.ts）が使う enum union の
// type-only 再 export。新規ファイルは generated/models を直接 import できない
// （TASK-444-S1）ため、既存の import 箇所を経由する。
export type {
  DangerLevel,
  PetStatus,
  ReservationDayOption,
  ReservationSource,
  ReservationTypeCategory,
} from "@/types/generated/models";

// ─────────────────────────────────────────────────
// ReservationType request types（手書き・EMR-527）
// 保存 model（生成 ReservationType）から Pick / Omit せず、binding struct が受け取る列のうち
// buildReservationTypeCreateRequest / buildReservationTypeUpdateRequest が送るキーだけを
// 明示列挙する（TASK-444）。サーバー管理列・関係配列は入力型へ混入しない。
// ─────────────────────────────────────────────────

/**
 * 予約区分作成リクエスト。name のみ必須。
 * @see backend/internal/reservation/reservation_type_request.go createReservationTypeRequest
 */
export interface CreateReservationTypeRequest {
  name: string;
  description?: string;
  is_active?: boolean;
  group_id?: number;
  reservation_display_name?: string;
  duration_minutes?: number;
  short_name?: string;
  show_short_name?: boolean;
  reservation_visible?: boolean;
  reservation_comment?: string;
  reservation_image_url?: string;
  reservation_day_option?: ReservationDayOption;
  is_internal?: boolean;
}

/**
 * 予約区分更新リクエスト（PATCH・全キー optional）。送るキーは作成と同じ。
 * @see backend/internal/reservation/reservation_type_request.go updateReservationTypeRequest
 */
export type UpdateReservationTypeRequest = Partial<CreateReservationTypeRequest>;

// ─────────────────────────────────────────────────
// Reorder request type (手書き)
// ─────────────────────────────────────────────────

export type ReorderReservationTypeRequest = ReorderRequest;
