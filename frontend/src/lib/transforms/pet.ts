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

export const ACQUISITION_TYPE_REVERSE_MAP: Record<string, string> = {
  "購入": "purchased",
  "譲渡": "transferred",
  "保護": "rescued",
  "その他": "other",
};

export const ACQUISITION_TYPE_MAP: Record<string, string> = {
  purchased: "購入",
  transferred: "譲渡",
  rescued: "保護",
  other: "その他",
};

export const DANGER_LEVEL_REVERSE_MAP: Record<string, string> = {
  "低": "low",
  "中": "medium",
  "高": "high",
};

export const DANGER_LEVEL_MAP: Record<string, string> = {
  low: "低",
  medium: "中",
  high: "高",
};

/**
 * バックエンドペットレスポンスをフロントエンド Pet 型に変換
 * ReturnType<typeof transformBackendPetToFrontend> が Pet 型の正式定義
 */
export const transformBackendPetToFrontend = (p: BackendPet) => ({
  id: String(p.id ?? 0),
  ownerId: String(p.owner_id ?? 0),
  ownerNumber: p.owner?.id,
  ownerName: p.owner?.owner_name ?? "",
  phone: p.owner?.phone ?? "",
  petNumber: p.pet_number,
  name: p.name ?? "",
  petNameKana: p.pet_name_kana ?? undefined,
  species: p.animal_species?.name ?? "",
  animalSpeciesId: p.animal_species_id != null ? String(p.animal_species_id) : undefined,
  breed: p.breed,
  color: p.color,
  gender: p.gender ? (PET_GENDER_MAP[p.gender] ?? p.gender) : undefined,
  status: p.status ? PET_STATUS_MAP[p.status] : undefined,
  birthDate: p.birth_date ? p.birth_date.split("T")[0] : undefined,
  neuteredDate: p.neutered_date ? p.neutered_date.split("T")[0] : undefined,
  weight: p.weight?.toString(),
  food: p.food,
  environment: p.environment,
  acquisitionType: p.acquisition_type ? (ACQUISITION_TYPE_MAP[p.acquisition_type] ?? p.acquisition_type) : undefined,
  dangerLevel: p.danger_level ? (DANGER_LEVEL_MAP[p.danger_level] ?? p.danger_level) : undefined,
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
 * Pet フロントエンド型 — transformBackendPetToFrontend の戻り値から自動導出
 * 手動管理せず BackendPet（models.ts）と常に同期される
 */
export type Pet = ReturnType<typeof transformBackendPetToFrontend>;

/**
 * フロントエンドフォームデータの共通入力型
 * transformCreatePetRequest / transformUpdatePetRequest で共用
 */
export type PetFormInput = {
  ownerId?: string;
  name?: string;
  animalSpeciesId?: string;
  petNumber?: string;
  petNameKana?: string;
  breed?: string;
  color?: string;
  gender?: string;
  birthDate?: string;
  weight?: string;
  food?: string;
  environment?: string;
  neuteredDate?: string;
  acquisitionType?: string;
  dangerLevel?: string;
  status?: "alive" | "deceased";
  insuranceId?: string;
  remarks?: string;
};

/**
 * フロントエンドフォームデータからバックエンド CreatePetRequest に変換
 */
export const transformCreatePetRequest = (data: PetFormInput & {
  ownerId: string;
  name: string;
  animalSpeciesId: string;
}): CreatePetRequest => ({
  owner_id: Number(data.ownerId),
  name: data.name,
  animal_species_id: Number(data.animalSpeciesId),
  pet_number: data.petNumber,
  pet_name_kana: data.petNameKana,
  breed: data.breed,
  color: data.color,
  gender: data.gender ? (PET_GENDER_REVERSE_MAP[data.gender] ?? data.gender) : undefined,
  birth_date: data.birthDate ? `${data.birthDate}T00:00:00Z` : undefined,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  food: data.food,
  environment: data.environment,
  neutered_date: data.neuteredDate ? `${data.neuteredDate}T00:00:00Z` : undefined,
  acquisition_type: data.acquisitionType ? (ACQUISITION_TYPE_REVERSE_MAP[data.acquisitionType] ?? data.acquisitionType) : undefined,
  danger_level: data.dangerLevel ? (DANGER_LEVEL_REVERSE_MAP[data.dangerLevel] ?? data.dangerLevel) : undefined,
  status: data.status,
  insurance_id: data.insuranceId ? Number(data.insuranceId) : undefined,
  remarks: data.remarks,
});

/**
 * フロントエンドフォームデータからバックエンド UpdatePetRequest に変換
 */
export const transformUpdatePetRequest = (data: PetFormInput): UpdatePetRequest => ({
  owner_id: data.ownerId ? Number(data.ownerId) : undefined,
  name: data.name,
  animal_species_id: data.animalSpeciesId ? Number(data.animalSpeciesId) : undefined,
  pet_number: data.petNumber,
  pet_name_kana: data.petNameKana,
  breed: data.breed,
  color: data.color,
  gender: data.gender ? (PET_GENDER_REVERSE_MAP[data.gender] ?? data.gender) : undefined,
  birth_date: data.birthDate ? `${data.birthDate}T00:00:00Z` : undefined,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  food: data.food,
  environment: data.environment,
  neutered_date: data.neuteredDate ? `${data.neuteredDate}T00:00:00Z` : undefined,
  acquisition_type: data.acquisitionType ? (ACQUISITION_TYPE_REVERSE_MAP[data.acquisitionType] ?? data.acquisitionType) : undefined,
  danger_level: data.dangerLevel ? (DANGER_LEVEL_REVERSE_MAP[data.dangerLevel] ?? data.dangerLevel) : undefined,
  status: data.status,
  insurance_id: data.insuranceId ? Number(data.insuranceId) : undefined,
  remarks: data.remarks,
});
