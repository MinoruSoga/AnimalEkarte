import type { Owner } from "@/types/owner";
import type { Pet } from "@/types";
import type { BackendOwner, BackendPet } from "./types";

/**
 * Transform backend pet response to frontend Pet type
 */
const PET_STATUS_MAP: Record<string, "生存" | "死亡"> = {
  alive: "生存",
  deceased: "死亡",
};

const PET_GENDER_MAP: Record<string, string> = {
  male: "雄",
  female: "雌",
  unknown: "不明",
};

export const transformPet = (pet: BackendPet): Pet => ({
  id: String(pet.id ?? 0),
  ownerId: String(pet.owner_id ?? 0),
  ownerName: "",
  phone: "",
  petNumber: pet.pet_number,
  name: pet.name ?? "",
  species: pet.animal_species?.name ?? "",
  animalSpeciesId: pet.animal_species_id != null ? String(pet.animal_species_id) : undefined,
  breed: pet.breed,
  gender: pet.gender ? (PET_GENDER_MAP[pet.gender] ?? pet.gender) : undefined,
  status: pet.status ? PET_STATUS_MAP[pet.status] : undefined,
  birthDate: pet.birth_date ?? undefined,
  weight: pet.weight?.toString(),
  environment: pet.environment,
  lastVisit: pet.last_visit ?? undefined,
  insuranceId: pet.insurance_id != null ? String(pet.insurance_id) : undefined,
  insuranceName: pet.insurance?.name,
  insuranceDetails:
    pet.insurance?.coverage_rate != null
      ? `${pet.insurance.coverage_rate}%補償`
      : undefined,
  remarks: pet.remarks,
});

/**
 * Transform backend owner response to frontend Owner type
 */
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
  membershipType: owner.membership_type ?? "",
  createdAt: owner.created_at ?? "",
  updatedAt: owner.updated_at ?? "",
  pets: owner.pets?.map((pet) => {
    const transformed = transformPet(pet);
    transformed.ownerName = owner.owner_name ?? "";
    return transformed;
  }),
});
