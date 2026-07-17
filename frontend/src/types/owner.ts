/**
 * Owner UI types (camelCase, string IDs) + API Request types (snake_case)
 * NOTE: API request DTOs use owner_name / owner_name_kana (not name / name_kana from BackendOwner).
 * See backend/internal/handler/owner_request.go for the authoritative field names.
 */
import type { Pet } from "./index";

/** UI-facing Owner type (camelCase, string IDs — post-transform) */
export interface Owner {
  id: string;
  ownerName: string;
  ownerNameKana?: string;
  company: string;
  postalCode: string;
  address1: string;
  address2: string;
  homePostalCode: string;
  homeAddress1: string;
  homeAddress2: string;
  birthDate?: string;
  phone: string;
  companyPhone: string;
  email: string;
  remarks: string;
  isDangerous: boolean;
  discountRate: number;
  membershipType: string;
  lineUserId?: string;
  lineIdConfirmedAt?: string;
  deliveryExcluded: boolean;
  deliveryExcludedReason?: string;
  deliveryCaution: boolean;
  deliveryCautionReason?: string | null;
  isTransferred: boolean;
  transferAt?: string;
  /** #158 レガシー EMR 準拠: DM（ダイレクトメール）送付希望。undefined=未設定 / true=必要 / false=不要。 */
  dmPreference?: boolean | null;
  lstepOptOut: boolean;
  lstepOptOutAt?: string;
  lstepOptOutReason?: string;
  createdAt: string;
  updatedAt: string;
  pets?: Pet[];
}

/**
 * 飼主作成リクエスト — createOwnerRequest Go struct に準拠（json:"owner_name"）
 * NOTE: BackendOwner.name (json:"name") とは異なる。API DTO は owner_name を使用する。
 * @see backend/internal/handler/owner_request.go createOwnerRequest
 */
export interface CreateOwnerRequest {
  /** #84: 登録先医院の任意指定（所属医院のみ）。未指定時はサーバ側で現在の医院になる */
  clinic_id?: number;
  owner_name: string;
  owner_name_kana?: string;
  birth_date?: string;
  company?: string;
  postal_code?: string;
  address1?: string;
  address2?: string;
  home_postal_code?: string;
  home_address1?: string;
  home_address2?: string;
  phone?: string;
  company_phone?: string;
  email?: string;
  remarks?: string;
  is_dangerous?: boolean;
  discount_rate?: number;
  membership_type?: string;
  dm_preference?: boolean | null;
}

/**
 * 飼主更新リクエスト — updateOwnerRequest Go struct に準拠（全フィールド optional）
 * FE3-6: CreateOwnerRequest とフィールド集合が完全一致（clinic_id 除く全フィールドが optional）だったため
 * 手書き二重定義を廃し派生型化。値・フィールド集合は変更していない。
 * @see backend/internal/handler/owner_request.go updateOwnerRequest
 */
export type UpdateOwnerRequest = Partial<Omit<CreateOwnerRequest, "clinic_id">>;
