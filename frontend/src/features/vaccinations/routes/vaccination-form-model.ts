import { normalizeKana } from "@/lib/normalize-kana";
import type { SortOrder } from "@/types";
import type { VaccinationRecord } from "../api/transforms";

export const VACCINATION_PRIORITY_FIELDS = ["date", "vaccineId"] as const;
export const VACCINATION_FIELD_ID_MAP: Record<string, string> = {
  date: "vaccination-date",
  vaccineId: "vaccine-select",
};

export type VaccinationFormGate =
  | { kind: "new-no-pet" }
  | { kind: "edit-loading" }
  | { kind: "edit-not-found" }
  | { kind: "edit-error"; retryRead?: () => void };

export function resolveVaccinationFormGate(input: {
  hasSelectedPet: boolean;
  isEdit: boolean;
  isReadLoading: boolean;
  isReadNotFound: boolean;
  isReadError: boolean;
  retryRead?: () => void;
}): VaccinationFormGate | null {
  if (!input.hasSelectedPet && !input.isEdit) return { kind: "new-no-pet" };
  if (input.isEdit && input.isReadLoading) return { kind: "edit-loading" };
  if (input.isEdit && input.isReadNotFound) return { kind: "edit-not-found" };
  if (input.isEdit && input.isReadError) return { kind: "edit-error", retryRead: input.retryRead };
  return null;
}

export function filterVaccinationHistory(input: {
  records: readonly VaccinationRecord[];
  historyPetId: string | undefined;
  excludeId: string | undefined;
  historySearchTerm: string;
  filterStartDate: string;
  filterEndDate: string;
  sortOrder: SortOrder;
}): VaccinationRecord[] {
  if (!input.historyPetId) return [];

  let result = input.records.filter((record) => record.id !== input.excludeId);

  const term = normalizeKana(input.historySearchTerm).toLowerCase();
  if (term) {
    result = result.filter((record) =>
      normalizeKana(record.vaccineName).toLowerCase().includes(term),
    );
  }

  if (input.filterStartDate) {
    result = result.filter((record) => record.date >= input.filterStartDate);
  }
  if (input.filterEndDate) {
    result = result.filter((record) => record.date <= input.filterEndDate);
  }

  return [...result].sort((left, right) =>
    input.sortOrder === "asc"
      ? left.date.localeCompare(right.date)
      : right.date.localeCompare(left.date),
  );
}
