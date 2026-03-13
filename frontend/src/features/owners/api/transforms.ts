import type { Owner } from "@/types/owner";
import type { Pet } from "@/types";
import type { BackendOwner, BackendPet } from "./types";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";

const MEMBERSHIP_TYPE_FROM_API: Record<string, string> = {
  "non_member": "非会員",
  "member": "会員",
  "deceased": "退亡者",
  "transferred": "他診/準",
};

// transformPet は transforms/pet の transformBackendPetToFrontend を使用
const transformPet = (pet: BackendPet): Pet => transformBackendPetToFrontend(pet);

export const transformOwner = (owner: BackendOwner): Owner => ({
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
  birthDate: owner.birth_date ?? undefined,
  phone: owner.phone ?? "",
  companyPhone: owner.company_phone ?? "",
  email: owner.email ?? "",
  remarks: owner.remarks ?? "",
  isDangerous: owner.is_dangerous ?? false,
  discountRate: owner.discount_rate ?? 0,
  membershipType: MEMBERSHIP_TYPE_FROM_API[owner.membership_type ?? ""] ?? owner.membership_type ?? "",
  createdAt: owner.created_at ?? "",
  updatedAt: owner.updated_at ?? "",
  pets: owner.pets?.map((pet) => ({
    ...transformPet(pet),
    ownerName: owner.owner_name ?? "",
  })),
});
