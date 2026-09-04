// React/Framework
import { useMemo } from "react";

// Internal
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";
import {
  useGetAllMedicinesMaster,
  useGetAllProcedures,
  useGetAllHospitalizationPlansMaster,
} from "@/hooks/use-treatment-master";

// Types
import type { CarePlanItemType } from "../../api/care-plan-items";

interface CarePlanRefSelectProps {
  /** ケアプラン項目の type。medicine/treatment/item 以外は何も描画しない。 */
  type: CarePlanItemType;
  /** 選択中の参照先マスタ ID(未選択時は null)。 */
  value: string | null;
  onChange: (value: string | null) => void;
}

/**
 * ケアプラン項目の type に応じて必須マスタ参照(投薬=薬剤/処置・検査=処置/持ち物=入院プラン)を
 * 選択させる SearchableSelect。DDL の chk_care_plan_item_ref 制約に対応する UI 側の入力欄。
 *
 * 既存パターン(VaccinationForm.tsx の SearchableSelect + マスタ取得 hook)を再利用し、
 * 新しい選択 UI は発明しない。
 */
export function CarePlanRefSelect({ type, value, onChange }: CarePlanRefSelectProps) {
  const { data: medicines, isLoading: isMedicinesLoading } = useGetAllMedicinesMaster();
  const { data: procedures, isLoading: isProceduresLoading } = useGetAllProcedures();
  const { data: plans, isLoading: isPlansLoading } = useGetAllHospitalizationPlansMaster();

  const medicineOptions = useMemo<SearchableSelectOption[]>(
    () => (medicines ?? []).map((m) => ({ value: m.id, label: m.name })),
    [medicines],
  );
  const procedureOptions = useMemo<SearchableSelectOption[]>(
    () => (procedures ?? []).map((p) => ({ value: p.id, label: p.name })),
    [procedures],
  );
  const planOptions = useMemo<SearchableSelectOption[]>(
    () => (plans ?? []).map((p) => ({ value: p.id, label: p.name })),
    [plans],
  );

  const handleChange = (next: string) => onChange(next || null);

  if (type === "medicine") {
    return (
      <SearchableSelect
        value={value ?? ""}
        onValueChange={handleChange}
        options={medicineOptions}
        disabled={isMedicinesLoading}
        placeholder={isMedicinesLoading ? "読み込み中..." : "薬剤を選択"}
        searchPlaceholder="薬剤を検索..."
      />
    );
  }

  if (type === "treatment") {
    return (
      <SearchableSelect
        value={value ?? ""}
        onValueChange={handleChange}
        options={procedureOptions}
        disabled={isProceduresLoading}
        placeholder={isProceduresLoading ? "読み込み中..." : "処置・検査を選択"}
        searchPlaceholder="処置・検査を検索..."
      />
    );
  }

  if (type === "item") {
    return (
      <SearchableSelect
        value={value ?? ""}
        onValueChange={handleChange}
        options={planOptions}
        disabled={isPlansLoading}
        placeholder={isPlansLoading ? "読み込み中..." : "入院プランを選択"}
        searchPlaceholder="入院プランを検索..."
      />
    );
  }

  return null;
}
