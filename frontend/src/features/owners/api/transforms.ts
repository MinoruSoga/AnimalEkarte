import type { Owner } from "@/types/owner";
import type { Pet } from "@/types";
import type { BackendOwner, BackendPet } from "./types";

/**
 * Transform backend pet response to frontend Pet type
 */
export const transformPet = (p: BackendPet): Pet => ({
  id: p.id,
  ownerId: p.owner_id,
  ownerName: "",
  phone: "",
  petNumber: p.pet_number,
  name: p.name,
  species: p.species,
  breed: p.breed,
  gender: p.gender,
  status: p.status,
  birthDate: p.birth_date,
  weight: p.weight?.toString(),
  environment: p.environment,
  lastVisit: p.last_visit,
  insuranceName: p.insurance_name,
  insuranceDetails: p.insurance_details,
  remarks: p.notes,
});

/**
 * Transform backend owner response to frontend Owner type
 */
export const transformOwner = (o: BackendOwner): Owner => ({
  id: o.id,
  ownerName: o.owner_name,
  ownerNameKana: o.owner_name_kana,
  company: o.company,
  postalCode: o.postal_code,
  address1: o.address1,
  address2: o.address2,
  homePostalCode: o.home_postal_code,
  homeAddress1: o.home_address1,
  homeAddress2: o.home_address2,
  birthDate: o.birth_date,
  phone: o.phone,
  companyPhone: o.company_phone,
  email: o.email,
  remarks: o.remarks,
  isDangerous: o.is_dangerous,
  discountRate: o.discount_rate,
  membershipType: o.membership_type,
  createdAt: o.created_at,
  updatedAt: o.updated_at,
  pets: o.pets?.map((p) => {
    const pet = transformPet(p);
    pet.ownerName = o.owner_name;
    return pet;
  }),
});
