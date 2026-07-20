import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";

import { formatPetAge } from "../lib/pet-age";

interface SelectedPetContextProps {
  pet: Pet;
}

export function SelectedPetContext({ pet }: SelectedPetContextProps) {
  const age = pet.birthDate ? formatPetAge(pet.birthDate) : null;
  const identity = [pet.species, pet.breed].filter(Boolean).join("・");
  const details = [age, pet.gender, pet.weight ? `${pet.weight} kg` : ""].filter(Boolean);

  return (
    <div className="flex min-w-0 items-center gap-3">
      <div
        aria-hidden="true"
        className={`flex size-10 shrink-0 items-center justify-center rounded-lg ${C.bgBrandLight} ${C.textBrandDark}`}
      >
        <span className="text-lg font-semibold">{pet.name?.slice(0, 1) || "-"}</span>
      </div>
      <div className="min-w-0">
        <p className={`text-2xs font-medium tracking-wide ${C.text50}`}>選択中のペット</p>
        <div className="flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5">
          <h2 className={`break-words text-lg font-semibold ${C.text}`}>{pet.name || "-"}</h2>
          {identity ? <span className={`text-sm ${C.text70}`}>{identity}</span> : null}
          {details.length > 0 ? (
            <span className={`text-xs ${C.text50}`}>{details.join(" ・ ")}</span>
          ) : null}
        </div>
      </div>
    </div>
  );
}
