import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";
import { ResourceTrimming } from "@/types/generated/models";
import { paths } from "@/config/paths";
import { LAYOUT } from "@/lib/design-tokens";

export function TrimmingPetSelection() {
  const { searchParams, setSearchParams, filteredPets, error, isLoading, handleSearch, handleClear, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: paths.trimming.new.getHref(), backPath: paths.trimming.getHref() });

  return (
    <PageLayout title="トリミング登録 - ペット選択" onBack={handleBack} resource={ResourceTrimming} maxWidth={LAYOUT.pageContentMaxWidth.full}>
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onSearch={handleSearch} onClear={handleClear} />
      <PetSelectionResultsTable pets={filteredPets} onSelect={handleSelect} isError={Boolean(error)} isLoading={isLoading} />
    </PageLayout>
  );
}
