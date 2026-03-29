/**
 * OwnersListPage - app層での feature 合成
 * owners feature と pets feature を app層でのみ合成し、
 * feature 間 import を排除する（依存逆転）。
 */

import { OwnersList } from "@/features/owners";
import { updatePet } from "@/features/pets/api/update-pet";

export function OwnersListPage() {
  return <OwnersList onUpdatePet={updatePet} />;
}
