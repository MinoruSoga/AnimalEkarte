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
  const isDeceased = pet.status === "死亡";

  return (
    <div
      className="flex min-w-0 items-center gap-2"
      aria-live="polite"
    >
      <div
        aria-hidden="true"
        className={`flex size-9 shrink-0 items-center justify-center rounded-sm ${C.bgBrand} ${C.textOnBrand}`}
      >
        <span className="text-xl font-semibold">{pet.name?.slice(0, 1) || "-"}</span>
      </div>
      <div className="min-w-0">
        <p className={`text-2xs font-semibold ${C.textBrandDark}`}>
          選択中のペット
        </p>
        <div className="flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5">
          <h2
            className={`break-words text-xl font-semibold ${C.text}`}
          >
            {pet.name || "-"}
          </h2>
          {isDeceased ? (
            <span
              className={`shrink-0 rounded-xxs border px-1.5 py-0.5 text-xs font-semibold ${C.borderDanger} ${C.bgDanger10} ${C.danger}`}
            >
              死亡
            </span>
          ) : null}
          {identity ? <span className={`text-sm ${C.text70}`}>{identity}</span> : null}
          {details.length > 0 ? (
            <span className={`text-xs ${C.text50}`}>{details.join(" ・ ")}</span>
          ) : null}
        </div>
      </div>
    </div>
  );
}
