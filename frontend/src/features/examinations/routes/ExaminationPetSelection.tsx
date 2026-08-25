import { useCallback } from "react";
import { useLocation, useNavigate } from "react-router";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";
import { ResourceExaminations } from "@/types/generated/models";
import { paths } from "@/config/paths";
import { LAYOUT } from "@/lib/design-tokens";
import type { Pet } from "@/types";
import { examinationCreateHref } from "./examinations-list-model";

export function ExaminationPetSelection() {
  const navigate = useNavigate();
  const location = useLocation();
  const { searchParams, setSearchParams, petPage, error, isLoading, handleClear, handleBack } =
    usePetSelectionPage({
      selectPath: paths.examinations.new.getHref(),
      backPath: paths.examinations.getHref(),
    });

  const handleSelect = useCallback(
    (pet: Pet) => {
      if (pet.status !== "生存") return;
      navigate(examinationCreateHref(pet.id), { state: location.state });
    },
    [location.state, navigate],
  );

  return (
    <PageLayout title="検査登録 - ペット選択" onBack={handleBack} resource={ResourceExaminations} maxWidth={LAYOUT.pageContentMaxWidth.full}>
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onClear={handleClear} />
      <PetSelectionResultsTable pets={petPage} onSelect={handleSelect} isError={Boolean(error)} isLoading={isLoading} />
    </PageLayout>
  );
}
