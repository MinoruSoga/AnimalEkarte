import type { Pet } from "@/lib/transforms/pet";

/**
 * ペット作成リクエスト（バックエンドAPI）
 * features/pets/api/types.ts から移管した共有型
 */
export interface CreatePetRequest {
  owner_id: number;
  animal_species_id: number;
  pet_number?: string;
  name: string;
  pet_name_kana?: string;
  gender?: string;
  birth_date?: string;
  breed?: string;
  color?: string;
  weight?: number;
  food?: string;
  environment?: string;
  neutered_date?: string;
  acquisition_type?: string;
  danger_level?: string;
  status?: "alive" | "deceased";
  insurance_id?: number;
  remarks?: string;
}

/**
 * ペット更新リクエスト（バックエンドAPI）
 */
export interface UpdatePetRequest {
  owner_id?: number;
  animal_species_id?: number;
  pet_number?: string;
  name?: string;
  pet_name_kana?: string;
  gender?: string;
  birth_date?: string;
  breed?: string;
  color?: string;
  weight?: number;
  food?: string;
  environment?: string;
  neutered_date?: string;
  acquisition_type?: string;
  danger_level?: string;
  status?: "alive" | "deceased";
  insurance_id?: number;
  remarks?: string;
}

/**
 * useOwnerForm への依存性注入インターフェース
 * owners feature が pets feature を直接 import しないための DI
 */
export interface PetMutationCallbacks {
  onSuccess: () => void;
  onError: () => void;
}

export interface PetCreateCallbacks {
  onSuccess: (data: Pet) => void;
  onError: () => void;
}

export interface PetMutations {
  /** 新規飼主登録時の pending ペット一括作成用（raw fetch function） */
  createPetFn: (req: CreatePetRequest) => Promise<Pet>;
  /** 既存飼主へのペット即時追加用（React Query mutation） */
  createPetMutate: (req: CreatePetRequest, callbacks: PetCreateCallbacks) => void;
  /** ペット更新用（React Query mutation） */
  updatePetMutate: (
    args: { id: string; req: UpdatePetRequest },
    callbacks: PetMutationCallbacks
  ) => void;
  /** ペット削除用（React Query mutation） */
  deletePetMutate: (id: string, callbacks: PetMutationCallbacks) => void;
}
