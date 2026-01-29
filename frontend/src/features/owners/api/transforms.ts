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
  ownerNumber: o.owner_number,
  name: o.name,
  name_kana: o.name_kana,
  phone: o.phone,
  email: o.email,
  address: o.address,
  notes: o.notes,
  created_at: o.created_at,
  updated_at: o.updated_at,
  pets: o.pets?.map((p) => {
    const pet = transformPet(p);
    pet.ownerName = o.name;
    pet.ownerNumber = o.owner_number;
    return pet;
  }),
});
