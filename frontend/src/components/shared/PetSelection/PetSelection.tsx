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

export const PetSelection = ({
  searchQuery,
  onSearchChange,
  filteredPets,
  selectedPets,
  onTogglePet,
  className,
  listClassName,
}: PetSelectionProps) => {
  return (
    <div className={cn("flex flex-col gap-2 border rounded-lg p-2 bg-white", className)}>
      <div>
        <Label className="mb-1 block text-sm font-bold text-[#37352F]">ペット選択</Label>
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="飼主名、ペット名で検索..."
            className="pl-9 bg-white h-10 text-sm border-[rgba(55,53,47,0.16)]"
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
            const isSelected = selectedPets.some((p) => p.id === pet.id);
            return (
                <div
                key={pet.id}
                className={`p-2 border-b flex items-start justify-between cursor-pointer transition-colors border-[rgba(55,53,47,0.09)] ${
                    isSelected ? "bg-blue-50/50" : "hover:bg-[#F7F6F3]"
                }`}
                onClick={() => onTogglePet(pet)}
                >
                <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-0.5">
                        <span className="font-medium text-sm text-[#37352F]">{pet.ownerName}</span>
                        <span className="text-sm text-muted-foreground font-mono">
                            ID:{pet.ownerId}
                        </span>
                    </div>
                    <div className="flex items-center gap-2">
                        <span className="text-sm text-[#37352F] font-bold">{pet.name}</span>
                        <span className="text-sm text-muted-foreground border rounded px-1.5 bg-white border-[rgba(55,53,47,0.16)]">
                            {pet.species}
                        </span>
                    </div>
                </div>
                {isSelected && <Check className="size-3.5 text-blue-600 flex-shrink-0 ml-2" />}
                </div>
            );
            })
        )}
      </div>

      {selectedPets.length > 0 && (
        <div className="bg-blue-50/50 p-2 rounded-md border border-blue-100/50">
          <div className="text-sm font-bold text-blue-700 mb-1">選択中 ({selectedPets.length})</div>
          <div className="flex flex-wrap gap-1.5">
            {selectedPets.map((p) => (
              <div
                key={p.id}
                className="text-sm bg-white px-2 py-1 rounded border border-blue-200 text-blue-800 flex items-center gap-1 shadow-sm"
              >
                <span className="font-bold">{p.name}</span>
                <span className="text-blue-400">|</span>
                <span>{p.ownerName}</span>
                <button
                  className="hover:bg-blue-100 rounded-full p-0.5 ml-1"
                  onClick={(e) => {
                    e.stopPropagation();
                    onTogglePet(p);
                  }}
                >
                  <X className="size-3" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
