import { PageLayout } from "@/components/shared/PageLayout";
import { PetSelectionSearchForm, PetSelectionResultsTable } from "@/components/shared/PetSelection";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";

export const VaccinationPetSelection = () => {
  const { searchParams, setSearchParams, filteredPets, handleSearch, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: "/vaccinations/new", backPath: "/vaccinations" });

  return (
    <PageLayout title="ワクチン接種 - ペット選択" onBack={handleBack} maxWidth="max-w-full">
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} />
    </PageLayout>
  );
};
