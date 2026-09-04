import type { Owner } from "@/types/owner";
import type { OwnerResponse, PetInOwnerResponse } from "@/types/generated/owner-responses";
import type { PetResponse } from "@/types/generated/pet-responses";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";

/**
 * API レスポンス型 — OwnerResponse Go struct に準拠（json:"owner_name"）。
 * models.ts の Owner（json:"name"）は使わない（BUG-433 / TASK-444-S2）。
 * @see backend/internal/owner/http_response.go OwnerResponse
 *
 * OwnerApiResponse は既存 import 互換の alias。新規コードは OwnerResponse を使う。
 */
export type OwnerApiResponse = OwnerResponse;

const MEMBERSHIP_TYPE_FROM_API: Record<string, string> = {
  non_member: "非会員",
  member: "会員",
  deceased: "退亡者",
  transferred: "他診/準",
};

/**
 * Owner配下のペット変換。
 * owner 埋め込み wire は PetInOwnerResponse（pet_name_kana）。
 * detail 専用 PetResponse との共通フィールドへアサートして変換を再利用する。
 * LINE / lstep 系フィールドは owner detail DTO に無い — LINE 専用 API 経由で取得する。
 */
const transformPetInOwner = (pet: PetInOwnerResponse, ownerName: string) => ({
  // PetInOwnerResponse は clinic_id/phone/owner ネストを持たない subset。
  // transformBackendPetToFrontend は欠落フィールドを optional として扱う。
  ...transformBackendPetToFrontend(pet as unknown as PetResponse),
  ownerName,
  phone: "",
});

export const transformOwner = (owner: OwnerResponse): Owner => ({
  id: String(owner.id ?? 0),
  clinicId: owner.clinic_id != null ? String(owner.clinic_id) : undefined,
  ownerName: owner.owner_name ?? "",
  ownerNameKana: owner.owner_name_kana || undefined,
  company: owner.company ?? "",
  postalCode: owner.postal_code ?? "",
  address1: owner.address1 ?? "",
  address2: owner.address2 ?? "",
  homePostalCode: owner.home_postal_code ?? "",
  homeAddress1: owner.home_address1 ?? "",
  homeAddress2: owner.home_address2 ?? "",
  birthDate: owner.birth_date ? owner.birth_date.split("T")[0] : undefined,
  phone: owner.phone ?? "",
  companyPhone: owner.company_phone ?? "",
  email: owner.email ?? "",
  remarks: owner.remarks ?? "",
  isDangerous: owner.is_dangerous ?? false,
  discountRate: owner.discount_rate ?? 0,
  membershipType:
    MEMBERSHIP_TYPE_FROM_API[owner.membership_type ?? ""] ?? owner.membership_type ?? "",
  // line_user_id / lstep_* は OwnerResponse に存在しない（models.Owner のみ）。
  // LINE 連携 UI は owner-line-tags 等の専用 API を参照する（use-line-integration-card-state）。
  lineUserId: undefined,
  lineIdConfirmedAt: owner.line_id_confirmed_at,
  deliveryExcluded: owner.delivery_excluded ?? false,
  deliveryExcludedReason: owner.delivery_excluded_reason,
  deliveryCaution: owner.delivery_caution ?? false,
  deliveryCautionReason: owner.delivery_caution_reason,
  isTransferred: owner.is_transferred ?? false,
  transferAt: owner.transfer_at,
  // #158: 未設定（undefined）と false（不要）を区別するため ?? false で潰さない。
  dmPreference: owner.dm_preference,
  lstepOptOut: false,
  lstepOptOutAt: undefined,
  lstepOptOutReason: undefined,
  createdAt: owner.created_at ?? "",
  updatedAt: owner.updated_at ?? "",
  pets: owner.pets?.map((pet) => transformPetInOwner(pet, owner.owner_name ?? "")),
});
