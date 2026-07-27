import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";
import { ResourceVaccinations } from "@/types/generated/models";
import { paths } from "@/config/paths";
import { LAYOUT } from "@/lib/design-tokens";

export function VaccinationPetSelection() {
  const { searchParams, setSearchParams, filteredPets, error, isLoading, handleSearch, handleClear, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: paths.vaccinations.new.getHref(), backPath: paths.vaccinations.getHref() });

  return (
    <PageLayout title="ワクチン接種 - ペット選択" onBack={handleBack} resource={ResourceVaccinations} maxWidth={LAYOUT.pageContentMaxWidth.full}>
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} onClear={handleClear} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} isError={Boolean(error)} isLoading={isLoading} />
    </PageLayout>
  );
}
