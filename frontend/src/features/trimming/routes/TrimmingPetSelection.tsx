import { PageLayout } from "@/components/shared/PageLayout";
import { PetSelectionSearchForm, PetSelectionResultsTable } from "@/components/shared/PetSelection";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";

export const TrimmingPetSelection = () => {
  const { searchParams, setSearchParams, filteredPets, handleSearch, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: "/trimming/new", backPath: "/trimming" });

  return (
    <PageLayout title="トリミング登録 - ペット選択" onBack={handleBack} maxWidth="max-w-full">
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} />
    </PageLayout>
  );
};
