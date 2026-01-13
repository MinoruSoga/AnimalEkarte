import { Plus, LayoutGrid, List } from "lucide-react";
import { Tabs, TabsList, TabsTrigger } from "../../../components/ui/tabs";
import { PageLayout } from "../../../components/shared/PageLayout";
import { SearchFilterBar } from "../../../components/shared/SearchFilterBar";
import { PrimaryButton } from "../../../components/shared/PrimaryButton";
import { HospitalizationBoard } from "../components/HospitalizationBoard";
import { HospitalizationListView } from "../components/HospitalizationListView";
import { ToggleGroup, ToggleGroupItem } from "../../../components/ui/toggle-group";
import { useHospitalizationList } from "../hooks/useHospitalizationList";
import { HOSPITALIZATION_FILTER_STATUS } from "../constants";

export const HospitalizationList = () => {
  const { 
    searchTerm, setSearchTerm,
    statusFilter, setStatusFilter,
    viewMode, setViewMode,
    filteredHospitalizations,
    cages,
    movePet,
    handleNavigateToForm
  } = useHospitalizationList();

  return (
    <PageLayout
      title="入院・ホテル管理"
      headerAction={
        <PrimaryButton onClick={() => handleNavigateToForm()}>
          <Plus className="mr-1.5 size-4" />
          新規入院登録
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Status Tabs */}
        <Tabs value={statusFilter} onValueChange={(v) => setStatusFilter(v as any)} className="w-full">
            <TabsList className="grid w-[400px] grid-cols-4">
                <TabsTrigger value={HOSPITALIZATION_FILTER_STATUS.ACTIVE}>入院中</TabsTrigger>
                <TabsTrigger value={HOSPITALIZATION_FILTER_STATUS.RESERVED}>予約</TabsTrigger>
                <TabsTrigger value={HOSPITALIZATION_FILTER_STATUS.DISCHARGED}>退院済</TabsTrigger>
                <TabsTrigger value={HOSPITALIZATION_FILTER_STATUS.ALL}>すべて</TabsTrigger>
            </TabsList>
        </Tabs>

        {/* Search & View Toggle */}
        <div className="flex items-center gap-4">
            <div className="flex-1">
                <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="飼主名、ペット名、入院No..."
                count={filteredHospitalizations.length}
                />
            </div>
            <div className="bg-white rounded-md border border-[rgba(55,53,47,0.16)] p-1 h-10 flex items-center">
                <ToggleGroup type="single" value={viewMode} onValueChange={(v) => v && setViewMode(v as any)}>
                    <ToggleGroupItem value="board" size="sm" aria-label="Board View">
                        <LayoutGrid className="h-4 w-4" />
                    </ToggleGroupItem>
                    <ToggleGroupItem value="list" size="sm" aria-label="List View">
                        <List className="h-4 w-4" />
                    </ToggleGroupItem>
                </ToggleGroup>
            </div>
        </div>

        {/* Content */}
        {viewMode === "board" ? (
            <HospitalizationBoard 
                cages={cages}
                hospitalizations={filteredHospitalizations}
                onNavigateToForm={handleNavigateToForm}
                onMovePet={movePet}
            />
        ) : (
            <HospitalizationListView 
                hospitalizations={filteredHospitalizations}
                onNavigate={handleNavigateToForm}
            />
        )}
      </div>
    </PageLayout>
  );
}
