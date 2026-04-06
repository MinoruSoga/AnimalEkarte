import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";
import { ResourceExaminations } from "@/types/generated/models";

export function ExaminationPetSelection() {
  const { searchParams, setSearchParams, filteredPets, handleSearch, handleClear, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: "/examinations/new", backPath: "/examinations" });

  return (
    <PageLayout title="検査登録 - ペット選択" onBack={handleBack} resource={ResourceExaminations} maxWidth="max-w-full">
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} onClear={handleClear} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} />
    </PageLayout>
  );
}
