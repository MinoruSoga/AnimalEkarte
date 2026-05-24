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
}

/**
 * 飼主更新リクエスト — updateOwnerRequest Go struct に準拠（全フィールド optional）
 * @see backend/internal/handler/owner_request.go updateOwnerRequest
 */
export interface UpdateOwnerRequest {
  owner_name?: string;
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
}
