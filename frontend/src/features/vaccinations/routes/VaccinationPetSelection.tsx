import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { useCallback } from "react";
import { useLocation, useNavigate } from "react-router";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";
import type { Pet } from "@/types";
import { vaccinationCreateHref } from "./vaccinations-list-model";
import { ResourceVaccinations } from "@/types/generated/models";
import { paths } from "@/config/paths";
import { LAYOUT } from "@/lib/design-tokens";

export function VaccinationPetSelection() {
  const navigate = useNavigate();
  const location = useLocation();
  const { searchParams, setSearchParams, petPage, error, isLoading, handleClear, handleBack } =
    usePetSelectionPage({ selectPath: paths.vaccinations.new.getHref(), backPath: paths.vaccinations.getHref() });

  const handleSelect = useCallback(
    (pet: Pet) => {
      if (pet.status !== "生存") return;
      navigate(vaccinationCreateHref(pet.id), { state: location.state });
    },
    [location.state, navigate],
  );

  return (
    <PageLayout title="予防接種登録 - ペット選択" onBack={handleBack} resource={ResourceVaccinations} maxWidth={LAYOUT.pageContentMaxWidth.full}>
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onClear={handleClear} />
      <PetSelectionResultsTable pets={petPage} onSelect={handleSelect} isError={Boolean(error)} isLoading={isLoading} />
    </PageLayout>
  );
}
