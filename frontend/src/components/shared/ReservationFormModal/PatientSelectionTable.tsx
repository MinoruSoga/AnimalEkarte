import { memo } from "react";

import { Pagination } from "@/components/shared/Pagination";
import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";

import { PatientSearchFilters } from "./patient-selection-filters";
import { PatientSelectionResults } from "./patient-selection-results";
import { usePatientSelectionTable } from "./use-patient-selection-table";

interface PatientSelectionTableProps {
  onSelect: (pet: Pet) => void;
  selectedPets: Pet[];
  /** 直接記録の患者変更では死亡 sentinel も表示し、選択不可を明示する。 */
  includeDeceased?: boolean;
}

/**
 * 検索条件は backend の述語と1対1で対応させる（BUG-452）。
 * `search` は pets.name / pets.name_kana / owners.name / owners.name_kana /
 * owners.phone を横断する ILIKE 検索（backend/internal/pet/repository.go:123-139）、
 * `species` は animal_species_id、`ownerId` は owner_id。
 *
 * 旧実装は「飼主名 / 飼主名よみ / 電話番号 / ペット名 / ペット名よみ」の5欄を持ち、
 * 取得済みの先頭20件に対してクライアント側で絞り込んでいた。5欄は backend の
 * `search` 1個でカバーされる。backend に述語の無い条件（住所など）を足してはならない
 * — FE 側で補うと総件数と描画行が食い違い、利用者は「その条件で全件を検索した」と
 * 誤解する。
 */
export const PatientSelectionTable = memo(function PatientSelectionTable({
  onSelect,
  selectedPets,
  includeDeceased = false,
}: PatientSelectionTableProps) {
  const table = usePatientSelectionTable({ selectedPets, includeDeceased });

  return (
    <div className="flex flex-col gap-4 h-full">
      <PatientSearchFilters
        searchParams={table.searchParams}
        onTextFieldChange={table.handleTextFieldChange}
        onSpeciesChange={table.handleSpeciesChange}
        onClear={table.handleClear}
      />

      <div className="flex items-center justify-between gap-2 shrink-0">
        <span
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-xs ${C.text60}`}
        >
          {table.selectionText}
        </span>
        <span
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={
            table.isRangeShownByPagination ? "sr-only" : `text-xs ${C.text60}`
          }
        >
          {table.statusText}
        </span>
      </div>

      <PatientSelectionResults
        hasSearchConditions={table.hasSearchConditions}
        isSearchPending={table.isSearchPending}
        error={table.error}
        isBusy={table.isBusy}
        pets={table.pets}
        selectedPetIds={table.selectedPetIds}
        onSelect={onSelect}
      />

      {!table.error && table.totalPages > 1 ? (
        <fieldset aria-busy={table.isBusy} className="border-0 p-0 m-0 shrink-0">
          <Pagination
            currentPage={table.responsePage}
            totalPages={table.totalPages}
            totalCount={table.totalCount}
            startIndex={table.startIndex}
            endIndex={table.endIndex}
            onPageChange={table.handlePageChange}
            onPrev={() => table.handlePageChange(table.responsePage - 1)}
            onNext={() => table.handlePageChange(table.responsePage + 1)}
          />
        </fieldset>
      ) : null}
    </div>
  );
});
