import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";

export function HospitalizationPetSelection() {
  const { searchParams, setSearchParams, filteredPets, handleSearch, handleClear, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: "/hospitalization/new", backPath: "/hospitalization" });

  return (
    <PageLayout title="入院・ホテル登録 - ペット選択" onBack={handleBack} maxWidth="max-w-full">
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} onClear={handleClear} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} />
    </PageLayout>
  );
}
