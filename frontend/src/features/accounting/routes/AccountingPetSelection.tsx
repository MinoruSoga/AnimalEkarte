import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";
import { ResourceAccounting } from "@/types/generated/models";
import { paths } from "@/config/paths";
import { LAYOUT } from "@/lib/design-tokens";

export function AccountingPetSelection() {
  const { searchParams, setSearchParams, filteredPets, handleSearch, handleClear, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: paths.accounting.new.getHref(), backPath: paths.accounting.getHref() });

  return (
    <PageLayout title="会計登録 - ペット選択" onBack={handleBack} resource={ResourceAccounting} maxWidth={LAYOUT.pageContentMaxWidth.full}>
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} onClear={handleClear} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} />
    </PageLayout>
  );
}
