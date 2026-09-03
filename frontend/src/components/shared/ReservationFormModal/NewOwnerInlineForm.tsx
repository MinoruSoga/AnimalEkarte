import { UserPlus } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { useAnimalSpecies } from "@/hooks/use-animal-species";
import type { NewOwnerFormData } from "@/types/reservation-form";

interface NewOwnerInlineFormProps {
  value: NewOwnerFormData;
  onChange: (data: NewOwnerFormData) => void;
  errors: Record<string, string>;
}

export function NewOwnerInlineForm({ value, onChange, errors }: NewOwnerInlineFormProps) {
  const {
    activeSpecies,
    isLoading: isLoadingSpecies,
    isError: isSpeciesError,
  } = useAnimalSpecies();
  const isSpeciesUnavailable = isSpeciesError || isLoadingSpecies || activeSpecies.length === 0;

  return (
    <div className="flex flex-col gap-4 overflow-y-auto flex-1 min-h-0 pb-2">
      <div className="flex items-center gap-2 shrink-0">
        <UserPlus className={`${ICON.action} ${C.text60}`} />
        <Label className={`text-sm font-bold ${C.text}`}>新規飼主・ペット情報</Label>
      </div>

      <div className="flex flex-col gap-1">
        <Label htmlFor="new-owner-name" className={`text-xs ${C.text60}`}>
          飼主名{" "}
          <span aria-hidden="true" className={C.textRequired}>
            *
          </span>
        </Label>
        <Input
          id="new-owner-name"
          data-testid="new-owner-name"
          value={value.ownerName}
          onChange={(e) => onChange({ ...value, ownerName: e.target.value })}
          placeholder="例: 山田 太郎"
          className="h-9 text-sm"
        />
        <FormFieldError message={errors.ownerName} />
      </div>

      <div className="flex flex-col gap-1">
        <Label htmlFor="new-owner-phone" className={`text-xs ${C.text60}`}>
          電話番号{" "}
          <span aria-hidden="true" className={C.textRequired}>
            *
          </span>
        </Label>
        <Input
          id="new-owner-phone"
          data-testid="new-owner-phone"
          value={value.phone}
          onChange={(e) => onChange({ ...value, phone: e.target.value })}
          placeholder="例: 090-1234-5678"
          className="h-9 text-sm"
          type="tel"
          aria-invalid={errors.phone ? true : undefined}
          aria-describedby={errors.phone ? "new-owner-phone-error" : undefined}
        />
        <FormFieldError id="new-owner-phone-error" message={errors.phone} />
      </div>

      <div className="flex flex-col gap-1">
        <Label htmlFor="new-owner-pet-name" className={`text-xs ${C.text60}`}>
          ペット名{" "}
          <span aria-hidden="true" className={C.textRequired}>
            *
          </span>
        </Label>
        <Input
          id="new-owner-pet-name"
          data-testid="new-owner-pet-name"
          value={value.petName}
          onChange={(e) => onChange({ ...value, petName: e.target.value })}
          placeholder="例: ポチ"
          className="h-9 text-sm"
        />
        <FormFieldError message={errors.petName} />
      </div>

      <div className="flex flex-col gap-1">
        <Label htmlFor="new-owner-species" className={`text-xs ${C.text60}`}>
          動物種{" "}
          <span aria-hidden="true" className={C.textRequired}>
            *
          </span>
        </Label>
        <SearchableSelect
          id="new-owner-species"
          value={value.animalSpeciesId ? String(value.animalSpeciesId) : ""}
          onValueChange={(v) => onChange({ ...value, animalSpeciesId: Number(v) })}
          options={activeSpecies.map((s) => ({ value: String(s.id), label: s.name }))}
          disabled={isSpeciesUnavailable}
          placeholder={isLoadingSpecies ? "読み込み中..." : "選択してください"}
          searchPlaceholder="動物種を検索..."
          triggerTestId="new-owner-species"
        />
        {isSpeciesError ? (
          <p role="alert" aria-atomic="true" className={`text-xs ${C.danger}`}>
            動物種の取得に失敗しました。
          </p>
        ) : isLoadingSpecies ? (
          <p role="status" aria-live="polite" aria-atomic="true" className={`text-xs ${C.text50}`}>
            動物種を読み込み中です。
          </p>
        ) : activeSpecies.length === 0 ? (
          <p role="status" aria-live="polite" aria-atomic="true" className={`text-xs ${C.text50}`}>
            動物種マスタが登録されていません。
          </p>
        ) : null}
        <FormFieldError message={errors.animalSpeciesId} />
      </div>

      <div className="flex flex-col gap-1">
        <Label htmlFor="new-owner-chief-complaint" className={`text-xs ${C.text60}`}>
          主訴{" "}
          <span aria-hidden="true" className={C.textRequired}>
            *
          </span>
        </Label>
        <Textarea
          id="new-owner-chief-complaint"
          data-testid="new-owner-chief-complaint"
          value={value.chiefComplaint}
          onChange={(e) => onChange({ ...value, chiefComplaint: e.target.value })}
          placeholder="例: 食欲不振、元気がない"
          className="text-sm resize-none"
          rows={3}
        />
        <FormFieldError message={errors.chiefComplaint} />
      </div>
    </div>
  );
}
