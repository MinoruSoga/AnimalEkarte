import * as React from "react";
import { Check, ChevronsUpDown, Search, X } from "lucide-react";
import { cn } from "../../../components/ui/utils";
import { Button } from "../../../components/ui/button";
import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "../../../components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "../../../components/ui/popover";
import { Badge } from "../../../components/ui/badge";
import { Pet } from "../../../types";

interface PatientSearchProps {
  filteredPets: Pet[];
  selectedPets: Pet[];
  onSelectPet: (pet: Pet) => void;
  searchQuery: string;
  onSearchChange: (value: string) => void;
  className?: string;
}

export function PatientSearch({
  filteredPets,
  selectedPets,
  onSelectPet,
  searchQuery,
  onSearchChange,
  className,
}: PatientSearchProps) {
  const [open, setOpen] = React.useState(false);

  return (
    <div className={cn("space-y-2", className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-full justify-between h-10 text-sm border-[rgba(55,53,47,0.16)] bg-white text-[#37352F]"
          >
            <div className="flex items-center gap-2 text-muted-foreground">
              <Search className="h-4 w-4" />
              {searchQuery ? searchQuery : "飼い主名、ペット名で検索..."}
            </div>
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[400px] p-0" align="start">
          <Command shouldFilter={false}>
            <CommandInput
              placeholder="検索..."
              value={searchQuery}
              onValueChange={onSearchChange}
              className="h-10 text-sm"
            />
            <CommandList>
              {filteredPets.length === 0 && (
                <div className="py-6 text-center text-sm text-muted-foreground">
                  {searchQuery ? "該当するペットが見つかりません" : "検索キーワードを入力してください"}
                </div>
              )}
              {filteredPets.length > 0 && (
                <CommandGroup heading="検索結果">
                  {filteredPets.map((pet) => {
                    const isSelected = selectedPets.some((p) => p.id === pet.id);
                    return (
                      <CommandItem
                        key={pet.id}
                        value={`${pet.name} ${pet.ownerName}`}
                        onSelect={() => {
                          onSelectPet(pet);
                          // Keep open for multiple selection or close?
                          // Usually keep open if multiple, but user might want to move on.
                          // Let's keep it open to allow searching another one easily, or just rely on the toggle.
                        }}
                        className="text-sm cursor-pointer"
                      >
                        <Check
                          className={cn(
                            "mr-2 h-4 w-4",
                            isSelected ? "opacity-100" : "opacity-0"
                          )}
                        />
                        <div className="flex flex-col">
                          <div className="flex items-center gap-2">
                            <span className="font-bold">{pet.name}</span>
                            <span className="text-sm text-muted-foreground bg-muted px-1 rounded">
                              {pet.species}
                            </span>
                          </div>
                          <span className="text-sm text-muted-foreground">
                            {pet.ownerName} (ID: {pet.ownerId})
                          </span>
                        </div>
                      </CommandItem>
                    );
                  })}
                </CommandGroup>
              )}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {/* Selected Patients Chips */}
      {selectedPets.length > 0 && (
        <div className="flex flex-wrap gap-2 p-2 bg-blue-50/30 rounded-md border border-blue-100/50">
          {selectedPets.map((pet) => (
            <Badge
              key={pet.id}
              variant="secondary"
              className="h-10 text-sm gap-1 bg-white border-blue-200 text-blue-800 hover:bg-blue-50 pl-2 pr-1"
            >
              <span className="font-bold">{pet.name}</span>
              <span className="text-blue-300">|</span>
              <span className="font-normal">{pet.ownerName}</span>
              <Button
                variant="ghost"
                size="icon"
                className="h-5 w-5 ml-1 hover:bg-blue-100 rounded-full text-blue-800"
                onClick={(e) => {
                  e.stopPropagation();
                  onSelectPet(pet);
                }}
              >
                <X className="h-3 w-3" />
              </Button>
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}
