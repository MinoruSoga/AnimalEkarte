import { UserPlus } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { useAnimalSpecies } from "@/features/owners";
import type { NewOwnerFormData } from "@/features/reservations";

interface NewOwnerInlineFormProps {
  value: NewOwnerFormData;
  onChange: (data: NewOwnerFormData) => void;
  errors: Record<string, string>;
}

export function NewOwnerInlineForm({ value, onChange, errors }: NewOwnerInlineFormProps) {
  const { activeSpecies, isLoading: isLoadingSpecies } = useAnimalSpecies();

  return (
    <div className="flex flex-col gap-4 overflow-y-auto flex-1 min-h-0 pb-2">
      <div className="flex items-center gap-2 shrink-0">
        <UserPlus className={`${ICON.action} ${C.text60}`} />
        <Label className={`text-sm font-bold ${C.text}`}>新規飼主・ペット情報</Label>
      </div>

      <div className="flex flex-col gap-1">
        <Label className={`text-xs ${C.text60}`}>
          飼主名 <span aria-hidden="true" style={{ color: "var(--color-danger, #e03e3e)" }}>*</span>
        </Label>
        <Input
          data-testid="new-owner-name"
          value={value.ownerName}
          onChange={(e) => onChange({ ...value, ownerName: e.target.value })}
          placeholder="例: 山田 太郎"
          className="h-9 text-sm"
        />
        <FormFieldError message={errors.ownerName} />
      </div>

      <div className="flex flex-col gap-1">
        <Label className={`text-xs ${C.text60}`}>
          電話番号 <span aria-hidden="true" style={{ color: "var(--color-danger, #e03e3e)" }}>*</span>
        </Label>
        <Input
          data-testid="new-owner-phone"
          value={value.phone}
          onChange={(e) => onChange({ ...value, phone: e.target.value })}
          placeholder="例: 090-1234-5678"
          className="h-9 text-sm"
          type="tel"
        />
        <FormFieldError message={errors.phone} />
      </div>

      <div className="flex flex-col gap-1">
        <Label className={`text-xs ${C.text60}`}>
          ペット名 <span aria-hidden="true" style={{ color: "var(--color-danger, #e03e3e)" }}>*</span>
        </Label>
        <Input
          data-testid="new-owner-pet-name"
          value={value.petName}
          onChange={(e) => onChange({ ...value, petName: e.target.value })}
          placeholder="例: ポチ"
          className="h-9 text-sm"
        />
        <FormFieldError message={errors.petName} />
      </div>

      <div className="flex flex-col gap-1">
        <Label className={`text-xs ${C.text60}`}>
          動物種 <span aria-hidden="true" style={{ color: "var(--color-danger, #e03e3e)" }}>*</span>
        </Label>
        <Select
          value={value.animalSpeciesId ? String(value.animalSpeciesId) : ""}
          onValueChange={(v) => onChange({ ...value, animalSpeciesId: Number(v) })}
          disabled={isLoadingSpecies}
        >
          <SelectTrigger data-testid="new-owner-species" className="h-9 text-sm">
            <SelectValue placeholder={isLoadingSpecies ? "読み込み中..." : "選択してください"} />
          </SelectTrigger>
          <SelectContent>
            {activeSpecies.map((s) => (
              <SelectItem key={s.id} value={String(s.id)}>
                {s.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <FormFieldError message={errors.animalSpeciesId} />
      </div>

      <div className="flex flex-col gap-1">
        <Label className={`text-xs ${C.text60}`}>
          主訴 <span aria-hidden="true" style={{ color: "var(--color-danger, #e03e3e)" }}>*</span>
        </Label>
        <Textarea
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
