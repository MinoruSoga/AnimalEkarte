import { memo, useMemo } from "react";
import { C, ICON } from "@/lib/design-tokens";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Search, Check, X } from "lucide-react";
import { Pet } from "@/types";
import { cn } from "@/components/ui/utils";

interface PetSelectionProps {
  searchQuery: string;
  onSearchChange: (value: string) => void;
  filteredPets: Pet[];
  selectedPets: Pet[];
  onTogglePet: (pet: Pet) => void;
  className?: string;
  listClassName?: string;
}

export const PetSelection = memo(function PetSelection({
  searchQuery,
  onSearchChange,
  filteredPets,
  selectedPets,
  onTogglePet,
  className,
  listClassName,
}: PetSelectionProps) {
  // js-set-map-lookups: O(1) ルックアップのため Set を事前構築（レンダーループ内の O(n) some() を排除）
  const selectedPetIdSet = useMemo(
    () => new Set(selectedPets.map((p) => p.id)),
    [selectedPets]
  );

  return (
    <div className={cn("flex flex-col gap-2 border rounded-lg p-2 bg-white", className)}>
      <div>
        <Label className={cn("mb-1 block text-sm font-bold", C.text)}>ペット選択</Label>
        <div className="relative">
          <Search className={`absolute left-2.5 top-1/2 -translate-y-1/2 ${ICON.action} text-muted-foreground`} />
          <Input
            placeholder="飼主名、ペット名で検索..."
            className={cn("pl-9 bg-white h-11 text-sm", C.borderMedium)}
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
          />
        </div>
      </div>

      <div className={cn("flex-1 overflow-y-auto bg-white rounded-md border min-h-[300px]", listClassName)}>
        {filteredPets.length === 0 ? (
            <div className="p-4 text-center text-sm text-muted-foreground">
                該当するペットが見つかりません
            </div>
        ) : (
            filteredPets.map((pet) => {
            const isSelected = selectedPetIdSet.has(pet.id);
            return (
                <div
                key={pet.id}
                className={cn(
                  "p-2 border-b flex items-start justify-between cursor-pointer transition-colors",
                  C.borderLight,
                  isSelected ? C.bgAccent8 : C.hoverBgPage
                )}
                onClick={() => onTogglePet(pet)}
                >
                <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-0.5">
                        <span className={cn("font-medium text-sm", C.text)}>{pet.ownerName}</span>
                        <span className="text-sm text-muted-foreground font-mono">
                            ID:{pet.ownerId}
                        </span>
                    </div>
                    <div className="flex items-center gap-2">
                        <span className={cn("text-sm font-bold", C.text)}>{pet.name}</span>
                        <span className={cn("text-sm text-muted-foreground border rounded px-1.5 bg-white", C.borderMedium)}>
                            {pet.species}
                        </span>
                    </div>
                </div>
                {isSelected ? <Check className={`${ICON.xs} ${C.accent} flex-shrink-0 ml-2`} /> : null}
                </div>
            );
            })
        )}
      </div>

      {selectedPets.length > 0 ? (
        <div className={`${C.bgAccent8} p-2 rounded-md border ${C.borderAccentLight}`}>
          <div className={`text-sm font-bold ${C.textAccentDark} mb-1`}>選択中 ({selectedPets.length})</div>
          <div className="flex flex-wrap gap-1.5">
            {selectedPets.map((p) => (
              <div
                key={p.id}
                className={`text-sm bg-white px-2 py-1 rounded border ${C.borderAccentLight} ${C.textAccentDark} flex items-center gap-1 shadow-sm`}
              >
                <span className="font-bold">{p.name}</span>
                <span className={C.accent}>|</span>
                <span>{p.ownerName}</span>
                <button
                  className={`${C.hoverBgAccent8} rounded-full p-0.5 ml-1`}
                  onClick={(e) => {
                    e.stopPropagation();
                    onTogglePet(p);
                  }}
                >
                  <X className={ICON.xs} />
                </button>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
});
