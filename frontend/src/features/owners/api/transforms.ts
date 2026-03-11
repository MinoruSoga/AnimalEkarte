import type { Owner } from "@/types/owner";
import type { Pet } from "@/types";
import type { BackendOwner, BackendPet } from "./types";

/**
 * Transform backend pet response to frontend Pet type
 */
export const transformPet = (pet: BackendPet): Pet => ({
  id: pet.id,
  ownerId: pet.owner_id,
  ownerName: "",
  phone: "",
  petNumber: pet.pet_number,
  name: pet.name,
  species: pet.species,
  breed: pet.breed,
  gender: pet.gender,
  status: pet.status,
  birthDate: pet.birth_date,
  weight: pet.weight?.toString(),
  environment: pet.environment,
  lastVisit: pet.last_visit,
  insuranceName: pet.insurance_name,
  insuranceDetails: pet.insurance_details,
  remarks: pet.notes,
});

/**
 * Transform backend owner response to frontend Owner type
 */
export const transformOwner = (owner: BackendOwner): Owner => ({
  id: owner.id,
  ownerName: owner.owner_name,
  ownerNameKana: owner.owner_name_kana,
  company: owner.company,
  postalCode: owner.postal_code,
  address1: owner.address1,
  address2: owner.address2,
  homePostalCode: owner.home_postal_code,
  homeAddress1: owner.home_address1,
  homeAddress2: owner.home_address2,
  birthDate: owner.birth_date,
  phone: owner.phone,
  companyPhone: owner.company_phone,
  email: owner.email,
  remarks: owner.remarks,
  isDangerous: owner.is_dangerous,
  discountRate: owner.discount_rate,
  membershipType: owner.membership_type,
  createdAt: owner.created_at,
  updatedAt: owner.updated_at,
  pets: owner.pets?.map((pet) => {
    const transformed = transformPet(pet);
    transformed.ownerName = owner.owner_name;
    return transformed;
  }),
});
