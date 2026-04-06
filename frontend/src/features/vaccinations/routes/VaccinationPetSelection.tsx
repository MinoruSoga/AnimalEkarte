import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";
import { ResourceVaccinations } from "@/types/generated/models";

export function VaccinationPetSelection() {
  const { searchParams, setSearchParams, filteredPets, handleSearch, handleClear, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: "/vaccinations/new", backPath: "/vaccinations" });

  return (
    <PageLayout title="ワクチン接種 - ペット選択" onBack={handleBack} resource={ResourceVaccinations} maxWidth="max-w-full">
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} onClear={handleClear} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} />
    </PageLayout>
  );
}
