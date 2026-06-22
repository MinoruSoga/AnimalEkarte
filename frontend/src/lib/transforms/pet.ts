import type { Pet as BackendPet } from "@/types/generated/models";
import { jstDateStartISOString } from "@/lib/jst-date";
import type { CreatePetRequest, UpdatePetRequest } from "@/types/pet";

const PET_STATUS_MAP: Partial<Record<string, "生存" | "死亡">> = {
  alive: "生存",
  deceased: "死亡",
};

// 外部公開: useOwnerForm で使用
export const PET_STATUS_REVERSE_MAP: Record<string, "alive" | "deceased"> = {
  "生存": "alive",
  "死亡": "deceased",
};

const PET_GENDER_MAP: Record<string, string> = {
  male: "雄",
  female: "雌",
  unknown: "不明",
};

const PET_GENDER_REVERSE_MAP: Record<string, string> = {
  "雄": "male",
  "雌": "female",
  "不明": "unknown",
};

const ACQUISITION_TYPE_REVERSE_MAP: Record<string, string> = {
  "購入": "purchased",
  "譲渡": "transferred",
  "保護": "rescued",
  "その他": "other",
};

const ACQUISITION_TYPE_MAP: Record<string, string> = {
  purchased: "購入",
  transferred: "譲渡",
  rescued: "保護",
  other: "その他",
};

const DANGER_LEVEL_REVERSE_MAP: Record<string, string> = {
  "低": "low",
  "中": "medium",
  "高": "high",
};

const DANGER_LEVEL_MAP: Record<string, string> = {
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
  // #86: 拠点横断一覧での医院名表示用。飼主の所属医院を優先し、無ければペット自身の clinic_id
  clinicId:
    p.owner?.clinic_id != null
      ? String(p.owner.clinic_id)
      : p.clinic_id != null
        ? String(p.clinic_id)
        : undefined,
  ownerId: String(p.owner_id ?? 0),
  ownerNumber: p.owner?.id,
  ownerName: p.owner?.name ?? "",
  ownerNameKana: p.owner?.name_kana ?? undefined,
  address: [p.owner?.address1, p.owner?.address2].filter(Boolean).join(" ") || undefined,
  phone: p.owner?.phone ?? "",
  petNumber: p.pet_number,
  name: p.name ?? "",
  petNameKana: p.name_kana ?? undefined,
  species: p.animal_species?.name ?? "",
  animalSpeciesId: p.animal_species_id != null ? String(p.animal_species_id) : undefined,
  breed: p.breed,
  color: p.color,
  // #158 レガシー EMR 準拠: 血液型 / マイクロチップ番号（未記録は undefined → 表示側で "-"）。
  bloodType: p.blood_type ?? undefined,
  microchipNumber: p.microchip_number ?? undefined,
  gender: p.gender ? (PET_GENDER_MAP[p.gender] ?? p.gender) : undefined,
  status: p.status ? PET_STATUS_MAP[p.status] : undefined,
  birthDate: p.birth_date ? p.birth_date.split("T")[0] : undefined,
  neuteredDate: p.neutered_date ? p.neutered_date.split("T")[0] : undefined,
  weight: p.weight?.toString(),
  food: p.food,
  environment: p.environment,
  acquisitionType: p.acquisition_type ? (ACQUISITION_TYPE_MAP[p.acquisition_type] ?? p.acquisition_type) : undefined,
  dangerLevel: p.danger_level ? (DANGER_LEVEL_MAP[p.danger_level] ?? p.danger_level) : undefined,
  // last_visit は birth_date / neutered_date と同じ date 型。兄弟フィールドと同様に
  // 日付部分のみへ正規化し、変換層の非対称を解消する（全消費者は formatDate 経由で無回帰）。
  lastVisit: p.last_visit ? p.last_visit.split("T")[0] : undefined,
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
type PetFormInput = {
  ownerId?: string;
  name?: string;
  animalSpeciesId?: string;
  petNumber?: string;
  petNameKana?: string;
  breed?: string;
  color?: string;
  bloodType?: string;
  microchipNumber?: string;
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
  name_kana: data.petNameKana,
  breed: data.breed,
  color: data.color,
  blood_type: data.bloodType,
  microchip_number: data.microchipNumber,
  gender: data.gender ? (PET_GENDER_REVERSE_MAP[data.gender] ?? data.gender) : undefined,
  birth_date: data.birthDate ? jstDateStartISOString(data.birthDate) : undefined,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  food: data.food,
  environment: data.environment,
  neutered_date: data.neuteredDate ? jstDateStartISOString(data.neuteredDate) : undefined,
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
  name_kana: data.petNameKana,
  breed: data.breed,
  color: data.color,
  blood_type: data.bloodType,
  microchip_number: data.microchipNumber,
  gender: data.gender ? (PET_GENDER_REVERSE_MAP[data.gender] ?? data.gender) : undefined,
  birth_date: data.birthDate ? jstDateStartISOString(data.birthDate) : undefined,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  food: data.food,
  environment: data.environment,
  neutered_date: data.neuteredDate ? jstDateStartISOString(data.neuteredDate) : undefined,
  acquisition_type: data.acquisitionType ? (ACQUISITION_TYPE_REVERSE_MAP[data.acquisitionType] ?? data.acquisitionType) : undefined,
  danger_level: data.dangerLevel ? (DANGER_LEVEL_REVERSE_MAP[data.dangerLevel] ?? data.dangerLevel) : undefined,
  status: data.status,
  insurance_id: data.insuranceId ? Number(data.insuranceId) : undefined,
  remarks: data.remarks,
});
