import type { PetGender, AcquisitionType, DangerLevel } from "@/types/generated/models";
import type { PetResponse } from "@/types/generated/pet-responses";
import { jstDateStartISOString } from "@/lib/jst-date";
import type { CreatePetRequest, UpdatePetRequest } from "@/types/pet";

// UpdatePetRequest は danger_reason / name_origin / meeting_story を string|null で保持する
// （types/pet.ts の tri-state 上書き）。key不在=変更なし / null=クリア / 値=更新。

export type PetStatusLabel = "生存" | "死亡" | "不明";

/** API境界の未知・欠損statusを生存へ推測せず、臨床操作をfail-closedにする。 */
export function mapPetStatusLabel(status: string | null | undefined): PetStatusLabel {
  if (status === "alive") return "生存";
  if (status === "deceased") return "死亡";
  return "不明";
}

/**
 * 臨床 write / 予約選択の死亡契約（BE ValidatePetNotDeceased と同義）。
 * status ラベルが「死亡」または deceasedAt が非 null なら死亡。日時は捏造しない。
 */
export function isPetDeceasedForClinicalWrite(pet: {
  status?: PetStatusLabel | string | null;
  deceasedAt?: string | null;
}): boolean {
  return pet.status === "死亡" || pet.deceasedAt != null;
}

// 外部公開: useOwnerForm で使用
export const PET_STATUS_REVERSE_MAP: Record<string, "alive" | "deceased"> = {
  生存: "alive",
  死亡: "deceased",
};

export const PET_GENDER_MAP: Record<string, string> = {
  male: "雄",
  female: "雌",
  unknown: "不明",
};

// FE6-2: REVERSE_MAP の値は生成型の値域に収まる文字列のみ（PET_GENDER_MAP のキー = 逆写像元）。
const PET_GENDER_REVERSE_MAP: Record<string, PetGender> = {
  雄: "male",
  雌: "female",
  不明: "unknown",
};

const ACQUISITION_TYPE_REVERSE_MAP: Record<string, AcquisitionType> = {
  購入: "purchased",
  譲渡: "transferred",
  保護: "rescued",
  その他: "other",
};

export const ACQUISITION_TYPE_MAP: Record<string, string> = {
  purchased: "購入",
  transferred: "譲渡",
  rescued: "保護",
  other: "その他",
};

const DANGER_LEVEL_REVERSE_MAP: Record<string, DangerLevel> = {
  低: "low",
  中: "medium",
  高: "high",
};

export const DANGER_LEVEL_MAP: Record<string, string> = {
  low: "低",
  medium: "中",
  high: "高",
};

/**
 * バックエンドペット詳細レスポンス（PetResponse）をフロントエンド Pet 型に変換。
 * ReturnType<typeof transformBackendPetToFrontend> が Pet 型の正式定義。
 * wire 正本は pet domain の PetResponse（pet_name_kana 等）。models.Pet は使わない（BUG-433）。
 */
export const transformBackendPetToFrontend = (p: PetResponse) => ({
  id: String(p.id ?? 0),
  // #86: 医院名表示用。detail wire の PetOwnerNested に clinic_id は無いため pet.clinic_id を使う。
  clinicId: p.clinic_id != null ? String(p.clinic_id) : undefined,
  ownerId: String(p.owner_id ?? 0),
  ownerNumber: p.owner?.owner_number ?? p.owner?.id,
  ownerName: p.owner?.name ?? "",
  ownerNameKana: p.owner?.name_kana ?? undefined,
  // スタッフ向け飼主危険マーク (EMR-173)。PetOwnerNested.is_dangerous は
  // staff GET /v1/pets 系契約のみに存在し、LIFF/owner 契約へは出さない。
  // true の時だけキーを出して optional プロパティにする（false/欠落は同じく非表示、
  // owner-report 等 transform 以外の経路が組み立てる Pet とも型互換を保つ）。
  ...(p.owner?.is_dangerous === true && { ownerIsDangerous: true }),
  // PetOwnerNested に住所フィールドは無い（detail/list 共有の軽量サマリ）。
  address: undefined as string | undefined,
  phone: p.owner?.phone || p.phone || "",
  petNumber: p.pet_number,
  name: p.name ?? "",
  petNameKana: p.pet_name_kana || undefined,
  species: p.animal_species?.name ?? "",
  animalSpeciesId: p.animal_species_id != null ? String(p.animal_species_id) : undefined,
  breed: p.breed,
  color: p.color,
  // #158 レガシー EMR 準拠: 血液型 / マイクロチップ番号（未記録は undefined → 表示側で "-"）。
  bloodType: p.blood_type ?? undefined,
  microchipNumber: p.microchip_number ?? undefined,
  gender: p.gender ? (PET_GENDER_MAP[p.gender] ?? p.gender) : undefined,
  status: mapPetStatusLabel(p.status),
  birthDate: p.birth_date ? p.birth_date.split("T")[0] : undefined,
  neuteredDate: p.neutered_date ? p.neutered_date.split("T")[0] : undefined,
  weight: p.weight?.toString(),
  food: p.food,
  environment: p.environment,
  acquisitionType: p.acquisition_type
    ? (ACQUISITION_TYPE_MAP[p.acquisition_type] ?? p.acquisition_type)
    : undefined,
  dangerLevel: p.danger_level ? (DANGER_LEVEL_MAP[p.danger_level] ?? p.danger_level) : undefined,
  dangerReason: p.danger_reason,
  // last_visit は birth_date / neutered_date と同じ date 型。兄弟フィールドと同様に
  // 日付部分のみへ正規化し、変換層の非対称を解消する（全消費者は formatDate 経由で無回帰）。
  lastVisit: p.last_visit ? p.last_visit.split("T")[0] : undefined,
  insuranceId: p.insurance_id != null ? String(p.insurance_id) : undefined,
  insuranceName: p.insurance?.name,
  insuranceDetails:
    p.insurance?.coverage_rate != null ? `${p.insurance.coverage_rate}%補償` : undefined,
  remarks: p.remarks,
  // EMR-174: 名前の由来 / 出逢いのストーリー。NULL=未記録は undefined（捏造しない）。
  nameOrigin: p.name_origin ?? undefined,
  meetingStory: p.meeting_story ?? undefined,
  // PR#186 P2-2 Bug#1: 死亡記録日時。null許容 (未死亡 = undefined)。
  // BUG-003: staff GET /pets/{id} の deceased_reason を deceasedReason へ（owner/LIFF は別契約）。
  deceasedAt: p.deceased_at,
  deceasedReason: p.deceased_reason,
});

/**
 * Pet フロントエンド型 — transformBackendPetToFrontend の戻り値から自動導出
 * wire 正本は PetResponse（pet-responses.ts）
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
  dangerReason?: string;
  originalDangerReason?: string;
  status?: "alive" | "deceased";
  insuranceId?: string;
  remarks?: string;
  // EMR-174: 名前の由来 / 出逢いのストーリー（update は original* で差分検知 → tri-state）
  nameOrigin?: string;
  meetingStory?: string;
  originalNameOrigin?: string;
  originalMeetingStory?: string;
};

/**
 * フロントエンドフォームデータからバックエンド CreatePetRequest に変換
 */
export const transformCreatePetRequest = (
  data: PetFormInput & {
    ownerId: string;
    name: string;
    animalSpeciesId: string;
  },
): CreatePetRequest => ({
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
  acquisition_type: data.acquisitionType
    ? (ACQUISITION_TYPE_REVERSE_MAP[data.acquisitionType] ?? data.acquisitionType)
    : undefined,
  danger_level: data.dangerLevel
    ? (DANGER_LEVEL_REVERSE_MAP[data.dangerLevel] ?? data.dangerLevel)
    : undefined,
  ...(data.dangerReason?.trim() ? { danger_reason: data.dangerReason.trim() } : {}),
  // EMR-174: 空・空白のみの叙述値はキーを送らず NULL（未記録）を保持する
  ...(data.nameOrigin?.trim() ? { name_origin: data.nameOrigin.trim() } : {}),
  ...(data.meetingStory?.trim() ? { meeting_story: data.meetingStory.trim() } : {}),
  status: data.status,
  insurance_id: data.insuranceId ? Number(data.insuranceId) : undefined,
  remarks: data.remarks,
});

/**
 * nullable text フィールドの PATCH tri-state 組立（danger_reason / name_origin / meeting_story 共通）。
 * - value 未指定 → キーを送らない（変更なし）
 * - original 同値（trim 比較）→ キーを送らない
 * - それ以外 → trim 値を送る。空は null（既存値クリア）
 */
function transformNullableTextUpdate<K extends string>(
  key: K,
  value: string | undefined,
  original: string | undefined,
  hasOriginal: boolean,
): Partial<Record<K, string | null>> {
  if (value === undefined || (hasOriginal && value.trim() === (original ?? "").trim())) {
    return {};
  }
  return { [key]: value.trim() || null } as Partial<Record<K, string | null>>;
}

/**
 * フロントエンドフォームデータからバックエンド UpdatePetRequest に変換
 *
 * status は意図的に送信しない(BUG-415)。生死ステータスの変更は監査付きの
 * 死亡登録/取消エンドポイント(PetDeceasedRecordButton → /:id/death)に一本化されており、
 * generic PATCH 経由での status 書込は backend 側でも除去済み(buildPetUpdate)。
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
  acquisition_type: data.acquisitionType
    ? (ACQUISITION_TYPE_REVERSE_MAP[data.acquisitionType] ?? data.acquisitionType)
    : undefined,
  danger_level: data.dangerLevel
    ? (DANGER_LEVEL_REVERSE_MAP[data.dangerLevel] ?? data.dangerLevel)
    : undefined,
  ...transformNullableTextUpdate(
    "danger_reason",
    data.dangerReason,
    data.originalDangerReason,
    "originalDangerReason" in data,
  ),
  // EMR-174: tri-state — original* プロパティ有無で「比較可能か」を判定
  ...transformNullableTextUpdate(
    "name_origin",
    data.nameOrigin,
    data.originalNameOrigin,
    "originalNameOrigin" in data,
  ),
  ...transformNullableTextUpdate(
    "meeting_story",
    data.meetingStory,
    data.originalMeetingStory,
    "originalMeetingStory" in data,
  ),
  insurance_id: data.insuranceId ? Number(data.insuranceId) : undefined,
  remarks: data.remarks,
});
