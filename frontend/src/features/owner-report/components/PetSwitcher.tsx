import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";

interface PetSwitcherProps {
  pets: Pet[];
  selectedPetId: string | undefined;
  onSelect: (petId: string) => void;
}

/** 同居ペットをタブ化せず、固定ヘッダー内の単一選択欄で切り替える。 */
export function PetSwitcher({ pets, selectedPetId, onSelect }: PetSwitcherProps) {
  const hasSelectedPet = pets.some((pet) => pet.id === selectedPetId);
  const value = hasSelectedPet ? selectedPetId : pets[0]?.id;

  return (
    <div className="flex min-w-0 items-center gap-2">
      <label
        htmlFor="owner-report-pet-switcher"
        className={`shrink-0 text-2xs font-semibold ${C.text60}`}
      >
        ペット切替
      </label>
      <select
        id="owner-report-pet-switcher"
        name="petId"
        className={`min-h-11 min-w-0 max-w-56 rounded-xxs border px-2 text-sm ${C.borderLight} ${C.bgWhite} ${C.text} focus-visible:outline-none focus-visible:ring-2 ${C.focusRingBrand}`}
        value={value ?? ""}
        onChange={(event) => onSelect(event.target.value)}
      >
        {pets.map((pet) => (
          <option key={pet.id} value={pet.id}>
            {pet.name || "-"}
            {pet.species ? `（${pet.species}）` : ""}
          </option>
        ))}
      </select>
    </div>
  );
}
