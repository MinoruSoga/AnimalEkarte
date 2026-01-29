import { MOCK_PETS } from "@/config/mock-data";
import type { Pet } from "@/types";

/** 全角・半角スペースを除去して文字列を正規化 */
function normalize(s: string): string {
  return s.replace(/[\s\u3000]/g, "");
}

/** MOCK_PETSからペット名+飼主名の部分一致でPetを検索 */
export function findPetByRecord(
  petName: string,
  ownerName: string
): Pet | undefined {
  return MOCK_PETS.find((p) => {
    const pName = normalize(p.name);
    const rName = normalize(petName);
    const pOwner = normalize(p.ownerName);
    const rOwner = normalize(ownerName);
    return (
      (pName.includes(rName) || rName.includes(pName)) &&
      (pOwner.includes(rOwner) || rOwner.includes(pOwner))
    );
  });
}
