import type { TrimmingRecord as BackendTrimmingRecord } from "@/types/generated/models";

export type BackendTrimming = BackendTrimmingRecord;

/**
 * サーバー側で自動生成されるフィールド（リクエストに含めない）
 */
type ServerFields =
  | "id"
  | "clinic_id"
  | "weight"
  | "created_at"
  | "updated_at"
  | "pet"
  | "staff"
  | "course"
  | "options";

/**
 * リクエストで送信可能なトリミングフィールド
 * Goモデル変更 → make codegen → models.ts 更新で自動追従
 */
type TrimmingWritable = Omit<BackendTrimmingRecord, ServerFields>;

/**
 * トリミング作成リクエスト
 * pet_id / staff_id / course_id のみ必須、残りはoptional
 */
export type CreateTrimmingRequest =
  Required<Pick<TrimmingWritable, "pet_id" | "staff_id" | "course_id">> &
  Partial<Omit<TrimmingWritable, "pet_id" | "staff_id" | "course_id">> & {
    option_ids?: number[];
  };

/**
 * トリミング更新リクエスト（PATCH: 全フィールドoptional）
 */
export type UpdateTrimmingRequest = Partial<TrimmingWritable> & {
  option_ids?: number[];
};

/**
 * トリミング一覧APIレスポンス
 */
export interface TrimmingListResponse {
  data: BackendTrimming[];
  total: number;
  page: number;
  limit: number;
}

/**
 * トリミングフォーム入力データ（UI専用）
 */
export interface TrimmingFormData {
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
