// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { useState, useMemo, useCallback, memo } from "react";

// External
import { Check, Search, SearchX, RotateCcw } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { EmptyState } from "@/components/shared/DataStates";
import { useSearchPets } from "@/hooks/use-pet";
import { normalizedIncludes } from "@/lib/normalize-kana";

// Types
import type { Pet } from "@/types";

interface PatientSelectionTableProps {
  onSelect: (pet: Pet) => void;
  selectedPets: Pet[];
}

const FIELDS = [
  { key: "ownerId", label: "飼主No", placeholder: "例: 30042" },
  { key: "ownerName", label: "飼主名", placeholder: "例: 林 文明" },
  { key: "ownerNameKana", label: "飼主名よみ", placeholder: "例: はやし" },
  { key: "phone", label: "電話番号", placeholder: "例: 090..." },
  { key: "petName", label: "ペット名", placeholder: "例: Iris" },
  { key: "petNameKana", label: "ペット名よみ", placeholder: "例: いりす" },
  { key: "species", label: "種別", placeholder: "例: 犬" },
] as const;

const MAX_RESULTS = 20;

export const PatientSelectionTable = memo(function PatientSelectionTable({ onSelect, selectedPets }: PatientSelectionTableProps) {
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

  const { pets: allPets, isLoading } = useSearchPets();

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
      if (searchParams.ownerName && !normalizedIncludes(pet.ownerName, searchParams.ownerName)) continue;
      if (searchParams.ownerNameKana && !normalizedIncludes(pet.ownerNameKana ?? "", searchParams.ownerNameKana)) continue;
      if (searchParams.phone && (!pet.phone || !pet.phone.includes(searchParams.phone))) continue;
      if (searchParams.petName && !normalizedIncludes(pet.name, searchParams.petName)) continue;
      if (searchParams.petNameKana && !normalizedIncludes(pet.petNameKana ?? "", searchParams.petNameKana)) continue;
      if (searchParams.species && !normalizedIncludes(pet.species, searchParams.species)) continue;
      result.push(pet);
    }
    return result;
  }, [hasSearched, allPets, searchParams]);

  const handleSearch = useCallback(() => {
    setHasSearched(true);
  }, []);

  const handleClear = useCallback(() => {
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
  }, []);

  const selectedPetIds = useMemo(
    () => new Set(selectedPets.map((p) => p.id)),
    [selectedPets]
  );
  const isSelected = (pet: Pet) => selectedPetIds.has(pet.id);

  return (
    <div className="flex flex-col gap-4 h-full">
      {/* Search Criteria */}
      <div className={`rounded-lg bg-white p-3 border ${C.borderMedium} shrink-0`}>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2 mb-3">
          {FIELDS.map((field) => (
            <div key={field.key} className="space-y-0.5">
              <Label htmlFor={field.key} className={`text-xs ${C.text60}`}>
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
                className={`text-xs h-9 bg-white ${C.borderMediumLight} ${C.text}`}
              />
            </div>
          ))}
        </div>
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            onClick={handleSearch}
            className={`h-9 text-sm ${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand}`}
          >
            <Search className={`mr-1.5 ${ICON.xs}`} />
            検索
          </Button>
          {hasSearchConditions ? (
            <Button
              size="sm"
              variant="outline"
              onClick={handleClear}
              className={`h-9 text-sm ${C.borderMediumLight}`}
            >
              <RotateCcw className={`mr-1.5 ${ICON.xs}`} />
              クリア
            </Button>
          ) : null}
        </div>
      </div>

      {/* Results Table */}
      <div className={`flex-1 rounded-lg bg-white overflow-hidden border ${C.borderMedium} min-h-0`}>
        <div className="overflow-auto h-full flex flex-col">
          {!hasSearched ? (
            <div className="flex flex-col items-center justify-center flex-1 text-center gap-3">
              <Search className={`${ICON.xl} ${C.text20}`} />
              <div className={`text-sm ${C.text40}`}>検索条件を入力して検索してください</div>
            </div>
          ) : isLoading ? (
            <div className="flex flex-col items-center justify-center flex-1 text-center gap-3">
              <div className="animate-spin">
                <Search className={`${ICON.xl} ${C.text20}`} />
              </div>
              <div className={`text-sm ${C.text40}`}>検索中...</div>
            </div>
          ) : filteredPets.length === 0 ? (
            <div className="flex-1 flex items-center justify-center">
              <EmptyState icon={<SearchX className={ICON.xl} />} message="該当する患者が見つかりませんでした" />
            </div>
          ) : (
            <Table>
              <TableHeader className={`${C.bgPage} sticky top-0 z-10`}>
                <TableRow className={`border-b ${C.borderMedium} h-9 ${C.hoverBgPage}`}>
                  <TableHead className={`min-w-[80px] ${C.text40} h-9`}>飼主No</TableHead>
                  <TableHead className={`min-w-[120px] ${C.text40} h-9`}>飼主名</TableHead>
                  <TableHead className={`min-w-[100px] ${C.text40} h-9`}>ペット名</TableHead>
                  <TableHead className={`min-w-[60px] ${C.text40} h-9`}>種別</TableHead>
                  <TableHead className={`min-w-[60px] ${C.text40} h-9`}>性別</TableHead>
                  <TableHead className={`min-w-[80px] ${C.text40} h-9`}>生年月日</TableHead>
                  <TableHead className={`min-w-[60px] ${C.text40} h-9`}>体重</TableHead>
                  <TableHead className={`min-w-[60px] ${C.text40} h-9`}>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredPets.map((pet) => {
                  const isDeceased = pet.status === "死亡";
                  return (
                    <TableRow
                      key={pet.id}
                      className={`transition-colors ${C.hoverBgLight} h-12 ${
                        isSelected(pet) ? C.bgPage : ""
                      } ${isDeceased ? "opacity-50 grayscale-[0.5]" : ""}`}
                    >
                      <TableCell className={`text-sm font-mono ${C.text}`}>{pet.ownerId}</TableCell>
                      <TableCell className={`text-sm font-medium ${C.text}`}>{pet.ownerName}</TableCell>
                      <TableCell className={`text-sm ${C.text}`}>
                        <span className="font-bold">{pet.name}</span>
                      </TableCell>
                      <TableCell className={`text-sm ${C.text}`}>{pet.species}</TableCell>
                      <TableCell className={`text-sm ${C.text}`}>{pet.gender || "-"}</TableCell>
                      <TableCell className={`text-sm font-mono ${C.text}`}>{pet.birthDate || "-"}</TableCell>
                      <TableCell className={`text-sm font-mono ${C.text}`}>{pet.weight || "-"}</TableCell>
                      <TableCell>
                        <Button
                          size="sm"
                          disabled={isDeceased}
                          aria-label={
                            isDeceased
                              ? `死亡・選択不可: ${pet.name} (ID ${pet.id})`
                              : isSelected(pet)
                                ? `選択中: ${pet.name} (ID ${pet.id})`
                                : `選択: ${pet.name} (ID ${pet.id})`
                          }
                          className={`h-11 min-w-11 gap-1 text-sm px-2 transition-colors ${
                            isDeceased
                              ? `${C.bgPage} ${C.textStatusGray} border-transparent cursor-not-allowed`
                              : isSelected(pet)
                              ? `${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand}`
                              : `bg-white border ${C.borderMediumLight} ${C.text} ${C.hoverBgSubtle}`
                          }`}
                          onClick={() => {
                            if (!isDeceased) onSelect(pet);
                          }}
                        >
                          <Check className={`${ICON.xs} ${isSelected(pet) ? "" : "opacity-0"}`} />
                          {isDeceased ? "死亡" : isSelected(pet) ? "選択中" : "選択"}
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </div>
      </div>
    </div>
  );
});
