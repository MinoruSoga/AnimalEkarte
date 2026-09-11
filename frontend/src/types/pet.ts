import type { Pet as BackendPet } from "@/types/generated/models";
import type { Pet } from "@/lib/transforms/pet";

/**
 * リクエストで送信可能なペットフィールド（明示許可リスト）。
 * transformCreatePetRequest / transformUpdatePetRequest が送るキーに固定する。
 * BackendPet へサーバー専用列が増えても入力型へ自動混入しない（TASK-444）。
 * version / deceased_at / deceased_reason は死亡専用 API 側の責務であり含めない。
 */
type PetWritable = Pick<
  BackendPet,
  | "owner_id"
  | "animal_species_id"
  | "name"
  | "pet_number"
  | "name_kana"
  | "breed"
  | "color"
  | "blood_type"
  | "microchip_number"
  | "gender"
  | "birth_date"
  | "weight"
  | "food"
  | "environment"
  | "neutered_date"
  | "acquisition_type"
  | "danger_level"
  | "danger_reason"
  | "status"
  | "insurance_id"
  | "remarks"
>;

/**
 * ペット作成リクエスト（バックエンドAPI）
 * owner_id / animal_species_id / name のみ必須、残りはoptional
 */
export type CreatePetRequest = Pick<PetWritable, "owner_id" | "animal_species_id" | "name"> &
  Partial<Omit<PetWritable, "owner_id" | "animal_species_id" | "name">>;

/**
 * ペット更新リクエスト（バックエンドAPI）
 * PATCH: 全フィールドoptional
 *
 * status は意図的に除外する(BUG-415)。generic update から status を書けなくする
 * backend 側の除去(buildPetUpdate)を型レベルでも多層防御する。status 変更は
 * 監査付きの死亡登録/取消エンドポイント(/:id/death)に一本化されている。
 * CreatePetRequest は PetWritable を直接参照するため、この除外の影響を受けない。
 */
export type UpdatePetRequest = Omit<Partial<PetWritable>, "status" | "danger_reason"> & {
  /**
   * tri-state: key不在=変更なし / null=クリア / 値=更新。
   * backend側は nullableStringRequestField (pet_request.go) で null と absent を区別する。
   * 生成基底の danger_reason?: string のままでは null クリアが型落ちするため、ここで上書きする。
   */
  danger_reason?: string | null;
};

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
