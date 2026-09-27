import { useCallback } from "react";
import { useLocation, useNavigate } from "react-router";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";
import { ResourceMedicalRecords } from "@/types/generated/models";
import { paths } from "@/config/paths";
import { LAYOUT } from "@/lib/design-tokens";
import type { Pet } from "@/types";

export function MedicalRecordPetSelection() {
  const navigate = useNavigate();
  const location = useLocation();
  const {
    searchParams,
    setSearchParams,
    petPage,
    error,
    isLoading,
    handleClear,
    handleSelect: selectForNewRecord,
    handleBack,
  } = usePetSelectionPage({
    selectPath: paths.medicalRecords.new.getHref(),
    backPath: paths.medicalRecords.getHref(),
  });

  const handleSelect = useCallback(
    (pet: Pet) => {
      // EMR-177: 死亡ペットは新規カルテを作成できないため、新規作成画面ではなく
      // そのペットのカルテ一覧へ遷移する（既存カルテの閲覧と確定済カルテへの追記が可能）。
      if (pet.status === "死亡") {
        const params = new URLSearchParams({ pet_id: pet.id });
        navigate(`${paths.medicalRecords.getHref()}?${params.toString()}`, {
          state: location.state,
        });
        return;
      }
      selectForNewRecord(pet);
    },
    [selectForNewRecord, navigate, location.state],
  );

  return (
    <PageLayout
      title="カルテ登録 - ペット選択"
      onBack={handleBack}
      resource={ResourceMedicalRecords}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <PetSelectionSearchForm
        searchParams={searchParams}
        setSearchParams={setSearchParams}
        onClear={handleClear}
      />
      <PetSelectionResultsTable
        pets={petPage}
        onSelect={handleSelect}
        isError={Boolean(error)}
        isLoading={isLoading}
      />
    </PageLayout>
  );
}
