import type { Pet } from "@/types";
import type { Pet as BackendPet } from "@/types/generated/models";
import type { CreatePetRequest, UpdatePetRequest } from "@/types/pet";

export const PET_STATUS_MAP: Partial<Record<string, "生存" | "死亡">> = {
  alive: "生存",
  deceased: "死亡",
};

export const PET_GENDER_MAP: Record<string, string> = {
  male: "雄",
  female: "雌",
  unknown: "不明",
};

export const PET_GENDER_REVERSE_MAP: Record<string, string> = {
  "雄": "male",
  "雌": "female",
  "不明": "unknown",
};

/**
 * バックエンドペットレスポンスをフロントエンド Pet 型に変換
 */
export const transformBackendPetToFrontend = (p: BackendPet): Pet => ({
  id: String(p.id ?? 0),
  ownerId: String(p.owner_id ?? 0),
  ownerName: "",
  phone: "",
  petNumber: p.pet_number,
  name: p.name ?? "",
  petNameKana: p.pet_name_kana ?? undefined,
  species: p.animal_species?.name ?? "",
  animalSpeciesId: p.animal_species_id != null ? String(p.animal_species_id) : undefined,
  breed: p.breed,
  gender: p.gender ? (PET_GENDER_MAP[p.gender] ?? p.gender) : undefined,
  status: p.status ? PET_STATUS_MAP[p.status] : undefined,
  birthDate: p.birth_date ? p.birth_date.split("T")[0] : undefined,
  weight: p.weight?.toString(),
  environment: p.environment,
  lastVisit: p.last_visit ?? undefined,
  insuranceId: p.insurance_id != null ? String(p.insurance_id) : undefined,
  insuranceName: p.insurance?.name,
  insuranceDetails:
    p.insurance?.coverage_rate != null
      ? `${p.insurance.coverage_rate}%補償`
      : undefined,
  remarks: p.remarks,
});

/**
 * フロントエンドフォームデータからバックエンド CreatePetRequest に変換
 */
export const transformCreatePetRequest = (data: {
  ownerId: string;
  name: string;
  animalSpeciesId: string;
  petNumber?: string;
  petNameKana?: string;
  breed?: string;
  gender?: string;
  birthDate?: string;
  weight?: string;
  microchipId?: string;
  environment?: string;
  status?: "alive" | "deceased";
  insuranceId?: string;
  remarks?: string;
}): CreatePetRequest => ({
  owner_id: data.ownerId,
  name: data.name,
  animal_species_id: data.animalSpeciesId,
  pet_number: data.petNumber,
  pet_name_kana: data.petNameKana,
  breed: data.breed,
  gender: data.gender ? (PET_GENDER_REVERSE_MAP[data.gender] ?? data.gender) : undefined,
  birth_date: data.birthDate,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  microchip_id: data.microchipId,
  environment: data.environment,
  status: data.status,
  insurance_id: data.insuranceId,
  remarks: data.remarks,
});

/**
 * フロントエンドフォームデータからバックエンド UpdatePetRequest に変換
 */
export const transformUpdatePetRequest = (data: {
  ownerId?: string;
  name?: string;
  animalSpeciesId?: string;
  petNumber?: string;
  petNameKana?: string;
  breed?: string;
  gender?: string;
  birthDate?: string;
  weight?: string;
  microchipId?: string;
  environment?: string;
  status?: "alive" | "deceased";
  insuranceId?: string;
  remarks?: string;
}): UpdatePetRequest => ({
  owner_id: data.ownerId,
  name: data.name,
  animal_species_id: data.animalSpeciesId,
  pet_number: data.petNumber,
  pet_name_kana: data.petNameKana,
  breed: data.breed,
  gender: data.gender ? (PET_GENDER_REVERSE_MAP[data.gender] ?? data.gender) : undefined,
  birth_date: data.birthDate,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  microchip_id: data.microchipId,
  environment: data.environment,
  status: data.status,
  insurance_id: data.insuranceId,
  remarks: data.remarks,
});
