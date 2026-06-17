import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";

interface PetSwitcherProps {
  pets: Pet[];
  selectedPetId: string | undefined;
  onSelect: (petId: string) => void;
}

/**
 * #158 R5: ペット切替タブ。選択でページ遷移せず（state 更新 + URL ?petId= 同期は親が担当）。
 */
export function PetSwitcher({ pets, selectedPetId, onSelect }: PetSwitcherProps) {
  return (
    <div className="flex flex-wrap gap-2" role="tablist" aria-label="ペット切替">
      {pets.map((pet) => {
        const isActive = pet.id === selectedPetId;
        return (
          <button
            key={pet.id}
            type="button"
            role="tab"
            aria-selected={isActive}
            onClick={() => onSelect(pet.id)}
            className={`rounded-full border px-3 py-1 text-sm transition-colors ${
              isActive
                ? `${C.bgBrand} border-transparent text-white`
                : `${C.bgWhite} ${C.borderLight} ${C.text} ${C.hoverBgPage}`
            }`}
          >
            {pet.name}
            {pet.species ? <span className="ml-1 text-xs opacity-70">{pet.species}</span> : null}
          </button>
        );
      })}
    </div>
  );
}
