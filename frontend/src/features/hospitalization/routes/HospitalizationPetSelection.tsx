// Internal
import { PageLayout } from "../../../components/shared/PageLayout";

// Relative
import { useHospitalizationPetSelection } from "../hooks/useHospitalizationPetSelection";
import { PetSelectionSearchForm } from "../components/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "../components/PetSelectionResultsTable";

export const HospitalizationPetSelection = () => {
  const {
    searchParams,
    setSearchParams,
    filteredPets,
    handleSearch,
    handleSelect,
    handleBack
  } = useHospitalizationPetSelection();

  return (
    <PageLayout
      title="入院・ホテル登録 - ペット選択"
      onBack={handleBack}
      maxWidth="max-w-full"
    >
        <PetSelectionSearchForm 
            searchParams={searchParams}
            setSearchParams={setSearchParams}
            onSearch={handleSearch}
        />
        <PetSelectionResultsTable 
            pets={filteredPets}
            onSelect={handleSelect}
        />
    </PageLayout>
  );
};
