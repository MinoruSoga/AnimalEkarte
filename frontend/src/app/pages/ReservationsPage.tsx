/**
 * ReservationsPage - app層での feature 合成
 * reservations feature が owners feature / pets feature を直接 import しないよう、
 * 新規オーナー・ペット作成 mutation を app 層でのみ合成し DI 注入する（FE-refactor #23）。
 */
import { createOwner } from "@/features/owners";
import { createPet } from "@/features/pets";
import { ReservationManagement } from "@/features/reservations";
import type { ReservationCreateMutations } from "@/types/reservation-create-mutations";

const createMutations: ReservationCreateMutations = {
  createOwnerFn: createOwner,
  createPetFn: createPet,
};

export function ReservationsPage() {
  return <ReservationManagement createMutations={createMutations} />;
}
