import { PageLayout } from "@/components/shared/PageLayout";
import { PetSelectionSearchForm, PetSelectionResultsTable } from "@/components/shared/PetSelection";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";

export const AccountingPetSelection = () => {
  const { searchParams, setSearchParams, filteredPets, handleSearch, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: "/accounting/new", backPath: "/accounting" });

  return (
    <PageLayout title="会計登録 - ペット選択" onBack={handleBack} maxWidth="max-w-full">
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} />
    </PageLayout>
  );
};
