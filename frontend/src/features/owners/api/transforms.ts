import type { Owner } from "@/types/owner";
import type { Owner as BackendOwner, Pet as BackendPet } from "@/types/generated/models";
import type { Pet } from "@/types";
import { PET_GENDER_MAP, PET_STATUS_MAP, ACQUISITION_TYPE_MAP, DANGER_LEVEL_MAP } from "@/lib/transforms/pet";

const MEMBERSHIP_TYPE_FROM_API: Record<string, string> = {
  "non_member": "非会員",
  "member": "会員",
  "deceased": "退亡者",
  "transferred": "他診/準",
};

/**
 * Owner配下のペット変換（petInOwnerResponse 対応）
 * animal_species オブジェクトが存在しない場合でも安全に動作する
 */
const transformPetInOwner = (pet: BackendPet, ownerName: string): Pet => ({
  id: String(pet.id ?? 0),
  ownerId: String(pet.owner_id ?? 0),
  ownerNumber: undefined,
  ownerName,
  phone: "",
  petNumber: pet.pet_number,
  name: pet.name ?? "",
  petNameKana: pet.pet_name_kana ?? undefined,
  species: pet.animal_species?.name ?? "",
  animalSpeciesId: pet.animal_species_id != null ? String(pet.animal_species_id) : undefined,
  breed: pet.breed,
  color: pet.color,
  gender: pet.gender ? (PET_GENDER_MAP[pet.gender] ?? pet.gender) : undefined,
  status: pet.status ? PET_STATUS_MAP[pet.status] : undefined,
  birthDate: pet.birth_date ? pet.birth_date.split("T")[0] : undefined,
  neuteredDate: pet.neutered_date ? pet.neutered_date.split("T")[0] : undefined,
  weight: pet.weight?.toString(),
  food: pet.food,
  environment: pet.environment,
  acquisitionType: pet.acquisition_type ? (ACQUISITION_TYPE_MAP[pet.acquisition_type] ?? pet.acquisition_type) : undefined,
  dangerLevel: pet.danger_level ? (DANGER_LEVEL_MAP[pet.danger_level] ?? pet.danger_level) : undefined,
  lastVisit: pet.last_visit ?? undefined,
  insuranceId: pet.insurance_id != null ? String(pet.insurance_id) : undefined,
  insuranceName: pet.insurance?.name,
  insuranceDetails:
    pet.insurance?.coverage_rate != null
      ? `${pet.insurance.coverage_rate}%補償`
      : undefined,
  remarks: pet.remarks,
});

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
  birthDate: owner.birth_date ? owner.birth_date.split("T")[0] : undefined,
  phone: owner.phone ?? "",
  companyPhone: owner.company_phone ?? "",
  email: owner.email ?? "",
  remarks: owner.remarks ?? "",
  isDangerous: owner.is_dangerous ?? false,
  discountRate: owner.discount_rate ?? 0,
  membershipType: MEMBERSHIP_TYPE_FROM_API[owner.membership_type ?? ""] ?? owner.membership_type ?? "",
  createdAt: owner.created_at ?? "",
  updatedAt: owner.updated_at ?? "",
  pets: owner.pets?.map((pet) => transformPetInOwner(pet, owner.owner_name ?? "")),
});
