import { memo } from "react";
import { C } from "@/lib/design-tokens";
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

/**
 * 検索条件は backend の述語と1対1で対応させる。
 * `search` は pets.name / pets.name_kana / owners.name / owners.name_kana /
 * owners.phone を横断する ILIKE 検索、`species` は animal_species_id、
 * `ownerId` は owner_id。
 *
 * backend に述語の無い条件（住所など）をここへ足してはならない。FE 側で
 * 補うと総件数と描画行が食い違い、利用者は「その条件で全件を検索した」と
 * 誤解する（BUG-451）。住所検索は述語が存在しないため機能ごと廃止した。
 */
export interface PetSelectionSearchParams {
  search: string;
  ownerId: string;
  species: string;
}

interface PetSelectionSearchFormProps {
  searchParams: PetSelectionSearchParams;
  setSearchParams: (params: PetSelectionSearchParams) => void;
  onClear: () => void;
}

const FIELD_DEFS = [
  {
    id: "search",
    label: "検索（ペット名・飼主名・よみ・電話）",
    placeholder: "例: もも、山田、090",
  },
  { id: "ownerId", label: "飼主No", placeholder: "例: 30042" },
] as const;

const ALL_SPECIES_VALUE = "__all__";

function SpeciesSelectField({
  searchParams,
  setSearchParams,
}: Pick<PetSelectionSearchFormProps, "searchParams" | "setSearchParams">) {
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

export const PetSelectionSearchForm = memo(function PetSelectionSearchForm({
  searchParams,
  setSearchParams,
  onClear,
}: PetSelectionSearchFormProps) {
  return (
    <div className={cn("mb-4 rounded-lg bg-white p-4 border", C.borderMedium)}>
      <h2 className={cn("mb-2 text-sm font-medium", C.text)}>検索条件</h2>
      {/* 入力停止後に自動検索するため、押しても何も起きない検索ボタンは置かない。 */}
      <p className={cn("mb-3 text-xs", C.text60)}>入力すると自動で検索します</p>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4 mb-4">
        {FIELD_DEFS.map(({ id, label, placeholder }) => (
          <div key={id} className="space-y-1.5">
            <Label htmlFor={id} className={cn("text-sm", C.text60)}>
              {label}
            </Label>
            <Input
              id={id}
              placeholder={placeholder}
              value={searchParams[id]}
              onChange={(e) => setSearchParams({ ...searchParams, [id]: e.target.value })}
              className={cn("text-sm h-11 bg-white focus-visible:ring-1", C.text, C.borderMedium)}
            />
          </div>
        ))}
        <SpeciesSelectField searchParams={searchParams} setSearchParams={setSearchParams} />
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
      </div>
    </div>
  );
});
