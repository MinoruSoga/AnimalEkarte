import { PageLayout } from "@/components/shared/PageLayout";
import { PetSelectionSearchForm, PetSelectionResultsTable } from "@/components/shared/PetSelection";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";

export const MedicalRecordPetSelection = () => {
  const { searchParams, setSearchParams, filteredPets, handleSearch, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: "/medical-records/new", backPath: "/medical-records" });

  return (
    <PageLayout title="カルテ作成 - ペット選択" onBack={handleBack} maxWidth="max-w-full">
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} />
    </PageLayout>
  );
};
