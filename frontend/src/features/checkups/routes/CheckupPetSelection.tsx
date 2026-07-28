import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PetSelectionSearchForm } from "@/components/shared/PetSelection/PetSelectionSearchForm";
import { PetSelectionResultsTable } from "@/components/shared/PetSelection/PetSelectionResultsTable";
import { usePetSelectionPage } from "@/hooks/use-pet-selection-page";
import { ResourceCheckups } from "@/types/generated/models";
import { paths } from "@/config/paths";
import { LAYOUT } from "@/lib/design-tokens";

export function CheckupPetSelection() {
  const { searchParams, setSearchParams, petPage, error, isLoading, handleClear, handleSelect, handleBack } =
    usePetSelectionPage({ selectPath: paths.checkups.new.getHref(), backPath: paths.checkups.getHref() });

  return (
    <PageLayout title="定期健診登録 - ペット選択" onBack={handleBack} resource={ResourceCheckups} maxWidth={LAYOUT.pageContentMaxWidth.full}>
      <PetSelectionSearchForm searchParams={searchParams} setSearchParams={setSearchParams} onClear={handleClear} />
      <PetSelectionResultsTable pets={petPage} onSelect={handleSelect} isError={Boolean(error)} isLoading={isLoading} />
    </PageLayout>
  );
}
