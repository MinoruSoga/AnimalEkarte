import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";

/** ペットタブと対応するレポート本文（tabpanel）を関連付ける id 群。OwnerReport と共有する。 */
export const OWNER_REPORT_TABPANEL_ID = "owner-report-tabpanel";
export const ownerReportPetTabId = (petId: string) => `owner-report-pet-tab-${petId}`;

interface PetSwitcherProps {
  pets: Pet[];
  selectedPetId: string | undefined;
  onSelect: (petId: string) => void;
}

/**
 * #158 R5: ペット切替タブ。選択でページ遷移せず（state 更新 + URL ?petId= 同期は親が担当）。
 * tablist/tab パターンを完成させるため、各タブは aria-controls で本文 tabpanel を指す。
 */
export function PetSwitcher({ pets, selectedPetId, onSelect }: PetSwitcherProps) {
  return (
    <div className="flex flex-wrap gap-1.5" role="tablist" aria-label="ペット切替">
      {pets.map((pet) => {
        const isActive = pet.id === selectedPetId;
        return (
          <button
            key={pet.id}
            type="button"
            role="tab"
            id={ownerReportPetTabId(pet.id)}
            aria-selected={isActive}
            aria-controls={OWNER_REPORT_TABPANEL_ID}
            onClick={() => onSelect(pet.id)}
            className={`rounded-full border px-2.5 py-0.5 text-sm whitespace-nowrap transition-colors ${
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
