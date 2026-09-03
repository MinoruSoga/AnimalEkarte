import { RotateCcw } from "lucide-react";

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
import { C, ICON } from "@/lib/design-tokens";
import { useAnimalSpecies } from "@/hooks/use-animal-species";

import {
  ALL_SPECIES_VALUE,
  TEXT_FIELDS,
  type PatientSearchParams,
} from "./patient-selection-table-model";

function SpeciesSelectField({
  value,
  onChange,
}: {
  value: string;
  onChange: (species: string) => void;
}) {
  const { activeSpecies, isLoading, isError } = useAnimalSpecies();
  const isUnavailable = isLoading || isError;
  const hasStatusMessage = isError || isLoading || activeSpecies.length === 0;

  return (
    <div className="space-y-0.5">
      <Label htmlFor="species" className={`text-xs ${C.text60}`}>
        種別
      </Label>
      <Select
        value={value || ALL_SPECIES_VALUE}
        onValueChange={(next) => onChange(next === ALL_SPECIES_VALUE ? "" : next)}
        disabled={isUnavailable}
      >
        <SelectTrigger
          id="species"
          aria-describedby={hasStatusMessage ? "species-status" : undefined}
          className={`text-xs h-11 bg-white ${C.borderMediumLight} ${C.text}`}
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
      {isError ? (
        <p id="species-status" role="alert" aria-atomic="true" className={`text-xs ${C.danger}`}>
          動物種の取得に失敗しました。
        </p>
      ) : isLoading ? (
        <p
          id="species-status"
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-xs ${C.text50}`}
        >
          動物種を読み込み中です。
        </p>
      ) : activeSpecies.length === 0 ? (
        <p
          id="species-status"
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-xs ${C.text50}`}
        >
          動物種マスタが登録されていません。
        </p>
      ) : null}
    </div>
  );
}

interface PatientSearchFiltersProps {
  searchParams: PatientSearchParams;
  onTextFieldChange: (key: "search" | "ownerId", value: string) => void;
  onSpeciesChange: (species: string) => void;
  onClear: () => void;
}

export function PatientSearchFilters({
  searchParams,
  onTextFieldChange,
  onSpeciesChange,
  onClear,
}: PatientSearchFiltersProps) {
  return (
    <div className={`rounded-lg bg-white p-3 border ${C.borderMedium} shrink-0`}>
      <p className={`mb-2 text-xs ${C.text60}`}>入力すると自動で検索します</p>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2 mb-3">
        {TEXT_FIELDS.map((field) => (
          <div key={field.key} className="space-y-0.5">
            <Label htmlFor={field.key} className={`text-xs ${C.text60}`}>
              {field.label}
            </Label>
            <Input
              id={field.key}
              placeholder={field.placeholder}
              value={searchParams[field.key]}
              onChange={(e) => onTextFieldChange(field.key, e.target.value)}
              className={`text-xs h-11 bg-white ${C.borderMediumLight} ${C.text}`}
            />
          </div>
        ))}
        <SpeciesSelectField value={searchParams.species} onChange={onSpeciesChange} />
      </div>
      <Button
        size="sm"
        variant="outline"
        onClick={onClear}
        className={`h-11 min-w-11 text-sm ${C.borderMediumLight}`}
      >
        <RotateCcw className={`mr-1.5 ${ICON.xs}`} />
        クリア
      </Button>
    </div>
  );
}
