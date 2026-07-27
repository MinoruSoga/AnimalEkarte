import { memo } from "react";
import { C, ICON } from "@/lib/design-tokens";
import { Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/components/ui/utils";
import { useAnimalSpecies } from "@/hooks/use-animal-species";

export interface PetSelectionSearchParams {
  search: string;
  ownerId: string;
  ownerName: string;
  ownerNameKana: string;
  phone: string;
  petName: string;
  petNameKana: string;
  species: string;
  address: string;
}

interface PetSelectionSearchFormProps {
  searchParams: PetSelectionSearchParams;
  setSearchParams: (params: PetSelectionSearchParams) => void;
  onSearch: () => void;
  onClear: () => void;
}

const FIELD_DEFS = [
  {
    id: "search",
    label: "全体検索（ペット名・飼主名・よみ・電話）",
    placeholder: "例: もも、山田、090",
  },
  { id: "ownerId", label: "飼主No", placeholder: "例: 30042" },
  { id: "ownerName", label: "飼主名", placeholder: "例: 林 文明" },
  { id: "ownerNameKana", label: "飼主名よみ", placeholder: "例: はやし ふみあき" },
  { id: "phone", label: "電話番号", placeholder: "例: 090-1234-5678" },
  { id: "petName", label: "ペット名", placeholder: "例: Iris" },
  { id: "petNameKana", label: "ペット名よみ", placeholder: "例: いりす" },
  { id: "address", label: "住所", placeholder: "例: 東京都" },
] as const;

const ALL_SPECIES_VALUE = "__all__";

function SpeciesSelectField({
  searchParams,
  setSearchParams,
}: Pick<
  PetSelectionSearchFormProps,
  "searchParams" | "setSearchParams"
>) {
  const { activeSpecies, isLoading, isError } = useAnimalSpecies();
  const statusMessage = isError
    ? "種別を取得できません"
    : isLoading
      ? "種別を読み込み中です"
      : undefined;

  return (
    <div className="space-y-1.5">
      <Label htmlFor="species" className={cn("text-sm", C.text60)}>
        種別
      </Label>
      <Select
        value={searchParams.species || ALL_SPECIES_VALUE}
        onValueChange={(value) =>
          setSearchParams({
            ...searchParams,
            species: value === ALL_SPECIES_VALUE ? "" : value,
          })
        }
        disabled={isLoading || isError}
      >
        <SelectTrigger
          id="species"
          aria-describedby={statusMessage ? "species-status" : undefined}
          className={cn("h-11 bg-white text-sm", C.text, C.borderMedium)}
        >
          <SelectValue placeholder="すべて" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL_SPECIES_VALUE}>すべて</SelectItem>
          {activeSpecies.map((species) => (
            <SelectItem key={species.id} value={String(species.id)}>
              {species.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {statusMessage ? (
        <p
          id="species-status"
          role={isError ? "alert" : "status"}
          className={cn("text-xs", isError ? C.danger : C.text60)}
        >
          {statusMessage}
        </p>
      ) : null}
    </div>
  );
}

export const PetSelectionSearchForm = memo(function PetSelectionSearchForm({ searchParams, setSearchParams, onSearch, onClear }: PetSelectionSearchFormProps) {
  return (
    <div className={cn("mb-4 rounded-lg bg-white p-4 border", C.borderMedium)}>
      <h2 className={cn("mb-2 text-sm font-medium", C.text)}>検索条件</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-4">
        {FIELD_DEFS.map(({ id, label, placeholder }) => (
          <div key={id} className="space-y-1.5">
            <Label htmlFor={id} className={cn("text-sm", C.text60)}>
              {label}
            </Label>
            <Input
              id={id}
              placeholder={placeholder}
              value={searchParams[id as keyof PetSelectionSearchParams]}
              onChange={(e) =>
                setSearchParams({ ...searchParams, [id]: e.target.value })
              }
              className={cn("text-sm h-11 bg-white focus-visible:ring-1", C.text, C.borderMedium)}
            />
          </div>
        ))}
        <SpeciesSelectField
          searchParams={searchParams}
          setSearchParams={setSearchParams}
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={onClear}
          className={cn("h-11 text-sm", C.borderMedium, C.text60, C.hoverBgPage, C.hoverText)}
        >
          クリア
        </Button>
        {/* docs/spec/design-system.md button-primary: brand と同じ primary teal + pill。 */}
        <Button
          size="sm"
          onClick={onSearch}
          className={cn("gap-2 h-11 text-sm rounded-full", C.textOnActionPrimary, C.bgActionPrimary, C.hoverBgActionPrimary, C.hoverTextOnActionPrimary)}
        >
          <Search className={ICON.action} />
          検索
        </Button>
      </div>
    </div>
  );
});
