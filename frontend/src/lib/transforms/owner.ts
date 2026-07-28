import type { Owner } from "@/types/owner";
import type { Owner as BackendOwner, Pet as BackendPet } from "@/types/generated/models";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";

/**
 * API レスポンス型 — ownerResponse Go struct に準拠（json:"owner_name"）
 * models.ts の Owner（json:"name"）とは異なる。
 * @see backend/internal/owner/http_response.go ownerResponse
 */
export interface OwnerApiResponse extends Omit<BackendOwner, "name" | "name_kana" | "dm_preference"> {
  owner_name: string;
  owner_name_kana?: string;
  dm_preference?: boolean | null;
}

const MEMBERSHIP_TYPE_FROM_API: Record<string, string> = {
  "non_member": "非会員",
  "member": "会員",
  "deceased": "退亡者",
  "transferred": "他診/準",
};

/**
 * Owner配下のペット変換
 * transformBackendPetToFrontendをベースにownerName/phoneをオーナー親から上書き
 */
const transformPetInOwner = (pet: BackendPet, ownerName: string) => ({
  ...transformBackendPetToFrontend(pet),
  ownerName,
  phone: "",
});

export const transformOwner = (owner: OwnerApiResponse): Owner => ({
  id: String(owner.id ?? 0),
  ownerName: owner.owner_name ?? "",
  ownerNameKana: owner.owner_name_kana,
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
  membershipType: MEMBERSHIP_TYPE_FROM_API[owner.membership_type ?? ""] ?? owner.membership_type ?? "",
  lineUserId: owner.line_user_id,
  lineIdConfirmedAt: owner.line_id_confirmed_at,
  deliveryExcluded: owner.delivery_excluded ?? false,
  deliveryExcludedReason: owner.delivery_excluded_reason,
  deliveryCaution: owner.delivery_caution ?? false,
  deliveryCautionReason: owner.delivery_caution_reason,
  isTransferred: owner.is_transferred ?? false,
  transferAt: owner.transfer_at,
  // #158: 未設定（undefined）と false（不要）を区別するため ?? false で潰さない。
  dmPreference: owner.dm_preference,
  lstepOptOut: owner.lstep_opt_out ?? false,
  lstepOptOutAt: owner.lstep_opt_out_at,
  lstepOptOutReason: owner.lstep_opt_out_reason,
  createdAt: owner.created_at ?? "",
  updatedAt: owner.updated_at ?? "",
  pets: owner.pets?.map((pet) => transformPetInOwner(pet, owner.owner_name ?? "")),
});
