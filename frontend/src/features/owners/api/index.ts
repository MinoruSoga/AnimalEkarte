import { Owner, CreateOwnerRequest, UpdateOwnerRequest } from "@/types/owner";
import { Pet } from "@/types";

const API_URL = "/api/v1/owners";

// Raw Backend Types reflecting JSON response
interface BackendPet {
  id: string;
  owner_id: string;
  pet_number: string;
  name: string;
  species: string;
  breed?: string;
  gender?: string;
  birth_date?: string;
  weight?: number;
  microchip_id?: string;
  environment?: string;
  status?: "生存" | "死亡";
  insurance_name?: string;
  insurance_details?: string;
  last_visit?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

interface BackendOwner {
  id: string;
  name: string;
  name_kana: string;
  phone: string;
  email: string;
  address: string;
  notes: string;
  created_at: string;
  updated_at: string;
  pets?: BackendPet[];
}

const transformPet = (p: BackendPet): Pet => ({
  id: p.id,
  ownerId: p.owner_id,
  ownerName: "", // Populated by parent if needed, or left empty
  phone: "", // Not in pet model
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

const transformOwner = (o: BackendOwner): Owner => ({
  id: o.id,
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
    pet.ownerName = o.name; // Enrich with owner name
    return pet;
  }),
});

export const getOwners = async (): Promise<Owner[]> => {
  const response = await fetch(API_URL);
  if (!response.ok) {
    throw new Error("Failed to fetch owners");
  }
  const data: BackendOwner[] = await response.json();
  return data.map(transformOwner);
};

export const getOwner = async (id: string): Promise<Owner> => {
  const response = await fetch(`${API_URL}/${id}`);
  if (!response.ok) {
    throw new Error("Failed to fetch owner");
  }
  const data: BackendOwner = await response.json();
  return transformOwner(data);
};

export const createOwner = async (data: CreateOwnerRequest): Promise<Owner> => {
  const response = await fetch(API_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error("Failed to create owner");
  }
  const responseData: BackendOwner = await response.json();
  return transformOwner(responseData);
};

export const updateOwner = async (
  id: string,
  data: UpdateOwnerRequest
): Promise<Owner> => {
  const response = await fetch(`${API_URL}/${id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error("Failed to update owner");
  }
  const responseData: BackendOwner = await response.json();
  return transformOwner(responseData);
};

export const deleteOwner = async (id: string): Promise<void> => {
  const response = await fetch(`${API_URL}/${id}`, {
    method: "DELETE",
  });
  if (!response.ok) {
    throw new Error("Failed to delete owner");
  }
};