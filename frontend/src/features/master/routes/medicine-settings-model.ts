import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import type { CreateMedicineRequest, UpdateMedicineRequest } from "@/types/medicine";
import type { Medicine } from "@/types";

import type { MedicineFormData } from "../components/medicine-side-panel-model";
import { normalizeKana } from "@/lib/normalize-kana";

export interface MedicineGroups {
  groupedMedicines: Map<string, { header: Medicine; items: Medicine[] }>;
  ungroupedMedicines: Medicine[];
  totalCount: number;
}

export type MedicineDragResolution =
  | { type: "none" }
  | { type: "same-category" }
  | {
      type: "move-category";
      activeItemId: string;
      overParentId: string | null;
      updateRequest: UpdateMedicineRequest;
    };

export function isCategoryMedicine(medicine: Medicine | null): boolean {
  return medicine !== null && !medicine.parentId && medicine.price === 0;
}

export function applyMedicineCategoryOverrides(
  medicines: Medicine[],
  overrideCategories: Map<string, string | undefined>
) {
  if (overrideCategories.size === 0) return medicines;
  return medicines.map((medicine) =>
    overrideCategories.has(medicine.id)
      ? { ...medicine, parentId: overrideCategories.get(medicine.id) }
      : medicine
  );
}

export function getCategoryMedicines(medicines: Medicine[]) {
  return medicines.filter((medicine) => !medicine.parentId && medicine.price === 0);
}

export function groupFilteredMedicines({
  orderedMedicines,
  activeFilters,
  searchTerm,
  medicinesById,
}: {
  orderedMedicines: Medicine[];
  activeFilters: ActiveFilter[];
  searchTerm: string;
  medicinesById: Map<string, Medicine>;
}): MedicineGroups {
  let items = orderedMedicines;
  for (const filter of activeFilters) {
    if (filter.key === "status" && typeof filter.value === "string") {
      const want = filter.value === "active";
      items = items.filter((medicine) =>
        filter.condition === "is" ? medicine.isActive === want : medicine.isActive !== want
      );
    }
  }

  const normalizedTerm = normalizeKana(searchTerm).toLowerCase();
  const filtered = items.filter(
    (medicine) => !searchTerm || normalizeKana(medicine.name).toLowerCase().includes(normalizedTerm)
  );

  const groupedMedicines = new Map<string, { header: Medicine; items: Medicine[] }>();
  const ungroupedMedicines: Medicine[] = [];

  for (const medicine of filtered) {
    if (medicine.parentId) {
      const existing = groupedMedicines.get(medicine.parentId);
      if (existing) {
        existing.items.push(medicine);
      } else {
        const parent = medicinesById.get(medicine.parentId);
        if (parent) {
          groupedMedicines.set(medicine.parentId, { header: parent, items: [medicine] });
        } else {
          ungroupedMedicines.push(medicine);
        }
      }
    } else if (medicine.price === 0) {
      if (!groupedMedicines.has(medicine.id)) {
        groupedMedicines.set(medicine.id, { header: medicine, items: [] });
      }
    } else {
      ungroupedMedicines.push(medicine);
    }
  }

  return { groupedMedicines, ungroupedMedicines, totalCount: filtered.length };
}

export function resolveMedicineDrag({
  activeItemId,
  overItemId,
  orderedMedicinesById,
}: {
  activeItemId: string;
  overItemId: string;
  orderedMedicinesById: ReadonlyMap<string, Medicine>;
}): MedicineDragResolution {
  const activeMedicine = orderedMedicinesById.get(activeItemId);
  const overMedicine = orderedMedicinesById.get(overItemId);
  if (!activeMedicine || !overMedicine) return { type: "none" };

  const activeParentId = activeMedicine.parentId ?? null;
  const overParentId = overMedicine.parentId ?? null;

  if (activeParentId === overParentId) return { type: "same-category" };

  return {
    type: "move-category",
    activeItemId,
    overParentId,
    updateRequest: overParentId
      ? { parent_id: Number(overParentId) }
      : { clear_parent_id: true },
  };
}

export function buildMedicineCreateRequest(
  data: MedicineFormData,
  isCategory: boolean
): CreateMedicineRequest {
  const effectivePrice = isCategory ? 0 : data.price;
  return {
    name: data.name,
    dosage_form: data.dosageForm || undefined,
    medicine_unit: data.medicineUnit || undefined,
    price: effectivePrice,
    description: data.description,
    is_active: data.isActive,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
    is_non_insurance: data.isNonInsurance,
    ...(data.parentId ? { parent_id: Number(data.parentId) } : {}),
  };
}

export function buildMedicineUpdateRequest({
  data,
  isCategory,
  selectedMedicine,
}: {
  data: MedicineFormData;
  isCategory: boolean;
  selectedMedicine: Medicine | null;
}): UpdateMedicineRequest {
  const effectivePrice = isCategory ? 0 : data.price;
  const request: UpdateMedicineRequest = {
    name: data.name,
    dosage_form: data.dosageForm || undefined,
    medicine_unit: data.medicineUnit || undefined,
    price: effectivePrice,
    description: data.description,
    is_active: data.isActive,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
    is_non_insurance: data.isNonInsurance,
  };

  if (data.parentId) {
    request.parent_id = Number(data.parentId);
  } else if (selectedMedicine?.parentId) {
    request.clear_parent_id = true;
  }

  return request;
}
