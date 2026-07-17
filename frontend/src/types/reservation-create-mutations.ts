import type { Owner, Pet } from "@/types";
import type { CreateOwnerRequest } from "@/types/owner";
import type { CreatePetRequest } from "@/types/pet";

/**
 * useReservationActions への依存性注入インターフェース
 * reservations feature が owners / pets feature を直接 import しないための DI（FE-refactor #23）
 */
export interface ReservationCreateMutations {
  /** 予約作成フローでの新規飼主登録用（raw fetch function） */
  createOwnerFn: (data: CreateOwnerRequest) => Promise<Owner>;
  /** 予約作成フローでの新規ペット登録用（raw fetch function） */
  createPetFn: (req: CreatePetRequest) => Promise<Pet>;
}
