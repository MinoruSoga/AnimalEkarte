import type {
  AcquisitionType,
  DangerLevel,
  PetGender,
  PetNeuteredStatus,
  PetStatus,
} from "@/types/generated/models";
import type { Pet } from "@/lib/transforms/pet";

// 作成・更新リクエストは保存 model（generated Pet）から Pick せず、HTTP の binding struct が
// 受け取る列のうち transformCreatePetRequest / transformUpdatePetRequest が送るキーだけを
// 明示列挙する。サーバー専用列（clinic_id / version / deceased_at / deceased_reason 等）や
// 関係配列は入力型へ自動混入しない（TASK-444 / EMR-527）。死亡関連は死亡専用 API の責務。
// 体重は wire 上の number で、保存側の精度型を入力へ持ち込まない。

/**
 * ペット作成リクエスト（バックエンドAPI）
 * owner_id / animal_species_id / name のみ必須、残りはoptional
 * @see backend/internal/pet/pet_request.go createPetRequest
 */
export interface CreatePetRequest {
  owner_id: number;
  animal_species_id: number;
  name: string;
  name_kana?: string;
  gender?: PetGender;
  status?: PetStatus;
  birth_date?: string;
  breed?: string;
  color?: string;
  blood_type?: string;
  microchip_number?: string;
  weight?: number;
  neutered_date?: string;
  neutered_status?: PetNeuteredStatus;
  acquisition_type?: AcquisitionType;
  danger_level?: DangerLevel;
  /** 空・空白のみはキーを送らない（NULL のまま）。 */
  danger_reason?: string;
  food?: string;
  environment?: string;
  insurance_id?: number;
  remarks?: string;
  /** EMR-174: 名前の由来（任意記録。空は送らず NULL=未記録） */
  name_origin?: string;
  /** EMR-174: 出逢いのストーリー（任意記録。空は送らず NULL=未記録） */
  meeting_story?: string;
}

/**
 * ペット更新リクエスト（バックエンドAPI）
 * PATCH: 全フィールドoptional
 *
 * status は意図的に持たない(BUG-415)。generic update から status を書けなくする
 * backend 側の除去(updatePetRequest に Status が無い)を型レベルでも多層防御する。
 * status 変更は監査付きの死亡登録/取消エンドポイント(/:id/death)に一本化されている。
 * @see backend/internal/pet/pet_request.go updatePetRequest
 */
export interface UpdatePetRequest {
  owner_id?: number;
  animal_species_id?: number;
  /** 自動採番後も手動変更可（作成リクエストには無い） */
  pet_number?: string;
  name?: string;
  name_kana?: string;
  gender?: PetGender;
  birth_date?: string;
  breed?: string;
  color?: string;
  blood_type?: string;
  microchip_number?: string;
  weight?: number;
  neutered_date?: string;
  neutered_status?: PetNeuteredStatus;
  acquisition_type?: AcquisitionType;
  danger_level?: DangerLevel;
  /**
   * tri-state: key不在=変更なし / null=クリア / 値=更新。
   * backend側は nullableStringRequestField (pet_request.go) で null と absent を区別する。
   */
  danger_reason?: string | null;
  food?: string;
  environment?: string;
  insurance_id?: number;
  remarks?: string;
  /** EMR-174: tri-state（danger_reason と同じ null クリア方式） */
  name_origin?: string | null;
  /** EMR-174: tri-state（danger_reason と同じ null クリア方式） */
  meeting_story?: string | null;
}

/**
 * useOwnerForm への依存性注入インターフェース
 * owners feature が pets feature を直接 import しないための DI
 */
interface PetMutationCallbacks {
  onSuccess: () => void;
  onError: (error?: unknown) => void;
}

interface PetCreateCallbacks {
  onSuccess: (data: Pet) => void;
  onError: (error?: unknown) => void;
}

export interface PetMutations {
  /** 新規飼主登録時の pending ペット一括作成用（raw fetch function） */
  createPetFn: (req: CreatePetRequest) => Promise<Pet>;
  /** 既存飼主へのペット即時追加用（React Query mutation） */
  createPetMutate: (req: CreatePetRequest, callbacks: PetCreateCallbacks) => void;
  /** ペット更新用（React Query mutation） */
  updatePetMutate: (
    args: { id: string; req: UpdatePetRequest },
    callbacks: PetMutationCallbacks,
  ) => void;
  /** ペット削除用（React Query mutation） */
  deletePetMutate: (id: string, callbacks: PetMutationCallbacks) => void;
  /** ペット死亡記録解除用（死亡→生存の遷移時のみ・React Query mutation） */
  revokePetDeathMutate: (petId: string) => void;
}
