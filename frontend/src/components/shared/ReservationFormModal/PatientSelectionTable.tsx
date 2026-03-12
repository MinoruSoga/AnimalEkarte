// React/Framework
import { useState, useMemo } from "react";

// External
import { Check, Filter, Search, SearchX, RotateCcw } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { usePetSearch } from "@/hooks/use-pet";

// Types
import type { Pet } from "@/types";

interface PatientSelectionTableProps {
  onSelect: (pet: Pet) => void;
  selectedPets: Pet[];
}

const FIELDS = [
  { key: "ownerId", label: "飼主No", placeholder: "例: 30042" },
  { key: "ownerName", label: "飼主名", placeholder: "例: 林 文明" },
  { key: "ownerNameKana", label: "飼主名(カナ)", placeholder: "例: ハヤシ" },
  { key: "phone", label: "電話番号", placeholder: "例: 090..." },
  { key: "petName", label: "ペット名", placeholder: "例: Iris" },
  { key: "petNameKana", label: "ペット名(カナ)", placeholder: "例: イリス" },
  { key: "species", label: "種別", placeholder: "例: 犬" },
] as const;

const MAX_RESULTS = 20;

export function PatientSelectionTable({ onSelect, selectedPets }: PatientSelectionTableProps) {
  const [searchParams, setSearchParams] = useState({
    ownerId: "",
    ownerName: "",
    ownerNameKana: "",
    phone: "",
    petName: "",
    petNameKana: "",
    species: "",
  });
  const [hasSearched, setHasSearched] = useState(false);

  const { pets: allPets, isLoading } = usePetSearch();

  const hasSearchConditions = useMemo(
    () => Object.values(searchParams).some((value) => value.trim() !== ""),
    [searchParams]
  );

  const filteredPets = useMemo(() => {
    if (!hasSearched) return [];
    const result: typeof allPets = [];
    for (const pet of allPets) {
      if (result.length >= MAX_RESULTS) break;
      if (searchParams.ownerId && !pet.ownerId.includes(searchParams.ownerId)) continue;
      if (searchParams.ownerName && !pet.ownerName.includes(searchParams.ownerName)) continue;
      if (searchParams.phone && (!pet.phone || !pet.phone.includes(searchParams.phone))) continue;
      if (searchParams.petName && !pet.name.includes(searchParams.petName)) continue;
      if (searchParams.species && !pet.species.includes(searchParams.species)) continue;
      result.push(pet);
    }
    return result;
  }, [hasSearched, allPets, searchParams]);

  const handleSearch = () => {
    setHasSearched(true);
  };

  const handleClear = () => {
    setSearchParams({
      ownerId: "",
      ownerName: "",
      ownerNameKana: "",
      phone: "",
      petName: "",
      petNameKana: "",
      species: "",
    });
    setHasSearched(false);
  };

  const selectedPetIds = useMemo(
    () => new Set(selectedPets.map((p) => p.id)),
    [selectedPets]
  );
  const isSelected = (pet: Pet) => selectedPetIds.has(pet.id);

  return (
    <div className="flex flex-col gap-4 h-full">
      {/* Search Criteria */}
      <div className="rounded-lg bg-white p-3 shadow-sm border border-[rgba(55,53,47,0.16)] shrink-0">
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2 mb-3">
          {FIELDS.map((field) => (
            <div key={field.key} className="space-y-0.5">
              <Label htmlFor={field.key} className="text-[12px] text-[#37352F]/60">
                {field.label}
              </Label>
              <Input
                id={field.key}
                placeholder={field.placeholder}
                value={searchParams[field.key as keyof typeof searchParams]}
                onChange={(e) => {
                  const value = e.target.value;
                  setSearchParams((prev) => ({ ...prev, [field.key]: value }));
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    handleSearch();
                  }
                }}
                className="text-[12px] h-9 bg-white border-[rgba(55,53,47,0.12)] text-[#37352F]"
              />
            </div>
          ))}
        </div>
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            onClick={handleSearch}
            className="h-9 text-sm bg-[#37352F] text-white hover:bg-[#37352F]/90"
          >
            <Search className="mr-1.5 size-3" />
            検索
          </Button>
          {hasSearchConditions && (
            <Button
              size="sm"
              variant="outline"
              onClick={handleClear}
              className="h-9 text-sm border-[rgba(55,53,47,0.12)]"
            >
              <RotateCcw className="mr-1.5 size-3" />
              クリア
            </Button>
          )}
        </div>
      </div>

      {/* Results Table */}
      <div className="flex-1 rounded-lg bg-white overflow-hidden shadow-sm border border-[rgba(55,53,47,0.16)] min-h-0">
        <div className="overflow-auto h-full flex flex-col">
          {!hasSearched ? (
            <div className="flex flex-col items-center justify-center flex-1 text-center gap-3">
              <Search className="size-8 text-[#37352F]/20" />
              <div className="text-sm text-[#37352F]/40">検索条件を入力して検索してください</div>
            </div>
          ) : isLoading ? (
            <div className="flex flex-col items-center justify-center flex-1 text-center gap-3">
              <div className="animate-spin">
                <Search className="size-8 text-[#37352F]/20" />
              </div>
              <div className="text-sm text-[#37352F]/40">検索中...</div>
            </div>
          ) : filteredPets.length === 0 ? (
            <div className="flex flex-col items-center justify-center flex-1 text-center gap-3">
              <SearchX className="size-8 text-[#37352F]/20" />
              <div className="text-sm text-[#37352F]/40">該当する患者が見つかりませんでした</div>
            </div>
          ) : (
            <Table>
              <TableHeader className="bg-[#F7F6F3] sticky top-0 z-10">
                <TableRow className="border-b border-[rgba(55,53,47,0.16)] h-9 hover:bg-[#F7F6F3]">
                  <TableHead className="min-w-[80px] text-[12px] text-[#37352F]/40 h-9">飼主No</TableHead>
                  <TableHead className="min-w-[120px] text-[12px] text-[#37352F]/40 h-9">飼主名</TableHead>
                  <TableHead className="min-w-[100px] text-[12px] text-[#37352F]/40 h-9">ペット名</TableHead>
                  <TableHead className="min-w-[60px] text-[12px] text-[#37352F]/40 h-9">種別</TableHead>
                  <TableHead className="min-w-[60px] text-[12px] text-[#37352F]/40 h-9">性別</TableHead>
                  <TableHead className="min-w-[80px] text-[12px] text-[#37352F]/40 h-9">生年月日</TableHead>
                  <TableHead className="min-w-[60px] text-[12px] text-[#37352F]/40 h-9">体重</TableHead>
                  <TableHead className="min-w-[60px] text-[12px] text-[#37352F]/40 h-9">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredPets.map((pet) => (
                  <TableRow
                    key={pet.id}
                    className={`transition-colors hover:bg-[rgba(55,53,47,0.06)] cursor-pointer h-9 ${
                      isSelected(pet) ? "bg-[#F7F6F3]" : ""
                    }`}
                    onClick={() => onSelect(pet)}
                  >
                    <TableCell className="text-sm py-1 font-mono text-[#37352F]">{pet.ownerId}</TableCell>
                    <TableCell className="text-sm py-1 font-medium text-[#37352F]">{pet.ownerName}</TableCell>
                    <TableCell className="text-sm py-1 font-bold text-[#37352F]">{pet.name}</TableCell>
                    <TableCell className="text-sm py-1 text-[#37352F]">{pet.species}</TableCell>
                    <TableCell className="text-sm py-1 text-[#37352F]">{pet.gender || "-"}</TableCell>
                    <TableCell className="text-sm py-1 font-mono text-[#37352F]">{pet.birthDate || "-"}</TableCell>
                    <TableCell className="text-sm py-1 font-mono text-[#37352F]">{pet.weight || "-"}</TableCell>
                    <TableCell className="py-1">
                      <Button
                        size="sm"
                        className={`h-9 gap-1 text-sm px-2 transition-colors ${
                          isSelected(pet)
                            ? "bg-[#37352F] text-white hover:bg-[#37352F]/90"
                            : "bg-white border border-[rgba(55,53,47,0.12)] text-[#37352F] hover:bg-[#FAFAF8]"
                        }`}
                        onClick={(e) => {
                          e.stopPropagation();
                          onSelect(pet);
                        }}
                      >
                        <Check className={`size-3 ${isSelected(pet) ? "" : "opacity-0"}`} />
                        {isSelected(pet) ? "選択中" : "選択"}
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
      </div>
    </div>
  );
}
