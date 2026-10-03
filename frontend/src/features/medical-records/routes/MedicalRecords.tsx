// React/Framework
import { useState, useCallback, useMemo, useEffect, useLayoutEffect, useRef } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Auth
import { useClinicScope } from "@/hooks/use-clinic-scope";

// Hooks
import { useModalState } from "@/hooks/use-modal-state";
import { useGetStaffs } from "@/hooks/use-staffs";
import {
  useGetAllMedicinesMaster,
  useGetAllProcedures,
  useGetAllConsultations,
} from "@/hooks/use-treatment-master";
import { useGetAllInventoryItems } from "@/hooks/use-inventory-items";

// External
import { paths } from "@/config/paths";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useStaffValidation } from "@/hooks/use-staff-validation";

// Relative
import { useMedicalRecordsList } from "../hooks/use-medical-records";
import { useMedicalRecordsUrlState } from "../hooks/use-medical-records-url-state";
import { useDeleteMedicalRecord } from "../api/delete-medical-record";
import { usePermission } from "@/hooks/use-permission";
import { useAnimalSpecies } from "@/hooks/use-animal-species";
import { useMedicalRecordsColumns } from "../hooks/use-medical-records-columns";
import { MedicalRecordsListPanels } from "./MedicalRecordsListPanels";

// Types
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { ResourceAccounting, ResourceMedicalRecords } from "@/types/generated/models";
import type { MedicalRecord } from "../api/transforms";
import {
  CLINIC_TOGGLE_RESET_PARAMS,
  PAGE_SIZE,
  buildMedicalRecordsFilterProperties,
} from "./medical-records-list-model";

function confirmMedicalRecordDelete(input: {
  item: { id: string; petIsDeceased: boolean } | null;
  recordsById: Map<string, MedicalRecord>;
  canDelete: boolean;
  deleteRecord: (id: string) => void;
}): void {
  const currentRecord = input.item ? input.recordsById.get(input.item.id) : undefined;
  if (
    input.canDelete === true &&
    input.item?.petIsDeceased === false &&
    currentRecord?.petIsDeceased === false
  ) {
    input.deleteRecord(input.item.id);
  }
}

function useMedicalRecordsDeleteFlow(canDelete: boolean, records: MedicalRecord[]) {
  const deleteModal = useModalState<{
    id: string;
    label: string;
    petIsDeceased: boolean;
  }>();
  const { mutate: deleteRecord } = useDeleteMedicalRecord();
  const canDeleteRef = useRef(canDelete);
  const recordsByIdRef = useRef(new Map(records.map((record) => [record.id, record])));

  useLayoutEffect(() => {
    canDeleteRef.current = canDelete;
  }, [canDelete]);
  useLayoutEffect(() => {
    recordsByIdRef.current = new Map(records.map((record) => [record.id, record]));
  }, [records]);

  const onDeleteConfirm = () => {
    confirmMedicalRecordDelete({
      item: deleteModal.item,
      recordsById: recordsByIdRef.current,
      canDelete: canDeleteRef.current,
      deleteRecord,
    });
    deleteModal.close();
  };

  return { deleteModal, onDeleteConfirm };
}

// EMR-247: 検索語は PropertyFilter 側の確定操作（Enter / 検索ボタン / クリア）で
// 確定した値のみが searchTerm に入る。入力途中のクエリ発火はしない
// （1検索=一覧+件数の重いDBクエリ2発が連打されていた 2026-10-01 障害の増幅要因を防ぐ）。

export function MedicalRecords() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("pet_id") || undefined;
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMedicalRecords);
  const { canView: canViewAccounting } = usePermission(ResourceAccounting);
  const {
    assignedClinics,
    selectedClinicIds,
    isMultiClinic,
    clinicNameById,
    currentClinicId,
    handleToggleClinic,
  } = useClinicScope({ resetParamsOnToggle: CLINIC_TOGGLE_RESET_PARAMS });
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  const { data: staffs } = useGetStaffs();
  const { data: medicines } = useGetAllMedicinesMaster();
  const { data: procedures } = useGetAllProcedures();
  const { data: consultations } = useGetAllConsultations();
  const { data: inventories } = useGetAllInventoryItems();
  const {
    activeSpecies,
    isLoading: isSpeciesLoading,
    isError: isSpeciesError,
  } = useAnimalSpecies();

  const filterProperties = useMemo(
    () =>
      buildMedicalRecordsFilterProperties({
        staffs,
        activeSpecies,
        isSpeciesError,
        isSpeciesLoading,
        medicines,
        procedures,
        consultations,
        inventories,
      }),
    [
      staffs,
      activeSpecies,
      isSpeciesError,
      isSpeciesLoading,
      medicines,
      procedures,
      consultations,
      inventories,
    ],
  );

  const resetKey = `${searchTerm}|${JSON.stringify(activeFilters)}|${petId ?? ""}`;
  const {
    currentPage,
    sortKey,
    sortOrder,
    handleSortToggle,
    directionForSort,
    handlePageChange,
    resetPage,
  } = useMedicalRecordsUrlState(resetKey);

  // EMR-245: 検索・フィルタ変更は URL の page 除去と同じイベントで行い、
  // 確定描画から page=1 に戻す（?page=N が残ると検出 pass 捨て後に復活する）。
  const handleSearchChange = useCallback(
    (value: string) => {
      setSearchTerm(value);
      resetPage();
    },
    [resetPage],
  );
  const handleFilterChange = useCallback(
    (next: ActiveFilter[]) => {
      setActiveFilters(next);
      resetPage();
    },
    [resetPage],
  );

  const clinicIdsForApi =
    selectedClinicIds.length === 0 ||
    (selectedClinicIds.length === 1 && selectedClinicIds[0] === currentClinicId)
      ? undefined
      : selectedClinicIds;
  const { records, total, isLoading, isError, isPlaceholderData } = useMedicalRecordsList({
    searchTerm,
    activeFilters,
    clinicIds: clinicIdsForApi,
    petId,
    page: currentPage,
    limit: PAGE_SIZE,
    sort: sortKey,
    order: sortKey ? sortOrder : undefined,
  });

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  // BUG-B1: server-side フィルタで件数が減り currentPage が範囲外になった場合、最終ページへ補正
  useEffect(() => {
    if (total > 0 && currentPage > totalPages) {
      handlePageChange(totalPages);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [totalPages, total]);

  const { isValidStaff } = useStaffValidation();
  const { deleteModal, onDeleteConfirm } = useMedicalRecordsDeleteFlow(canDelete, records);

  const startIndex = total === 0 ? 0 : (currentPage - 1) * PAGE_SIZE + 1;
  const endIndex = Math.min(currentPage * PAGE_SIZE, total);
  // EMR-247: searchTerm は確定済みの値のみ保持するため「確定待ち」状態は存在しない。
  const isFiltering = isPlaceholderData;

  const handleNavigateToForm = useCallback(
    (recordId?: string) => {
      navigate(
        recordId
          ? paths.medicalRecords.detail.getHref(recordId)
          : paths.medicalRecords.selectPet.getHref(),
        { state: { from: paths.medicalRecords.getHref() } },
      );
    },
    [navigate],
  );

  const showClinicColumn = isMultiClinic;
  const COLUMNS = useMedicalRecordsColumns({
    showClinicColumn,
    directionForSort,
    onSortToggle: handleSortToggle,
  });

  if (isLoading) return <LoadingFallback />;
  if (isError) return <ErrorFallback />;

  return (
    <MedicalRecordsListPanels
      canCreate={canCreate}
      onCreate={() => handleNavigateToForm()}
      assignedClinics={assignedClinics}
      selectedClinicIds={selectedClinicIds}
      onToggleClinic={handleToggleClinic}
      isSpeciesError={isSpeciesError}
      isSpeciesLoading={isSpeciesLoading}
      hasSpecies={activeSpecies.length > 0}
      filterProperties={filterProperties}
      activeFilters={activeFilters}
      onFilterChange={handleFilterChange}
      searchTerm={searchTerm}
      onSearchChange={handleSearchChange}
      total={total}
      isFiltering={isFiltering}
      columns={COLUMNS}
      records={records}
      showClinicColumn={showClinicColumn}
      currentClinicId={currentClinicId ?? undefined}
      clinicNameById={clinicNameById}
      canViewAccounting={canViewAccounting}
      canEdit={canEdit}
      canDelete={canDelete}
      isValidStaff={isValidStaff}
      onAccountingOpen={(accountingId) => navigate(paths.accounting.detail.getHref(accountingId))}
      onEdit={(recordId) => handleNavigateToForm(recordId)}
      onDeleteRequest={(item) => deleteModal.open(item)}
      totalPages={totalPages}
      currentPage={currentPage}
      startIndex={startIndex}
      endIndex={endIndex}
      onPageChange={handlePageChange}
      deleteOpen={deleteModal.isOpen}
      deleteLabel={deleteModal.item?.label ?? ""}
      onDeleteClose={deleteModal.close}
      onDeleteConfirm={onDeleteConfirm}
    />
  );
}
