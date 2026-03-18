// React/Framework
import { useDeferredValue, lazy, memo, Suspense, useCallback, useMemo, useState } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// External
import { Scissors, Upload, X, Trash2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { MasterLink } from "@/components/shared/MasterLink";
import { MasterSelectTrigger } from "@/components/shared/MasterSelectModal";
import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import type { SortOrder } from "@/types";
import { useMasterItems } from "@/hooks/use-master-items";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { paths } from "@/config/paths";

// Relative (direct file import, no barrel — bundle-barrel-imports)
import { useTrimmingForm } from "../hooks/use-trimming-form";
import type { TrimmingFormData } from "@/types/trimming";
import { useGetTrimmingsByPetId } from "../api/get-trimming";

// bundle-dynamic-imports: 重いモーダルは lazy() + Suspense で遅延ロード
const MasterSelectModal = lazy(() =>
  import("@/components/shared/MasterSelectModal").then((m) => ({ default: m.MasterSelectModal }))
);
const ConfirmDialog = lazy(() =>
  import("@/components/shared/ConfirmDialog/ConfirmDialog").then((m) => ({ default: m.ConfirmDialog }))
);

// ─── 静的JSXはモジュール定数に巻き上げ (rendering-hoist-jsx) ────────────────
const BW_UNIT_ITEMS = (
  <>
    <SelectItem value="Kg">Kg</SelectItem>
    <SelectItem value="g">g</SelectItem>
  </>
);

// ─── memo化されたフォームセクション ─────────────────────────────────────────

interface MasterItem {
  id: string;
  name: string;
  price?: number;
  status?: string;
}

interface LeftColumnProps {
  formData: TrimmingFormData;
  courses: MasterItem[];
  options: MasterItem[];
  styleImagePreview: string | null;
  onFormChange: (updates: Partial<TrimmingFormData>) => void;
  onCourseModalOpen: () => void;
  onStyleImageChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onRemoveStyleImage: () => void;
}

// rerender-memo: フォームセクションを memo() で独立させる
const LeftColumn = memo(function LeftColumn({
  formData,
  courses,
  options,
  styleImagePreview,
  onFormChange,
  onCourseModalOpen,
  onStyleImageChange,
  onRemoveStyleImage,
}: LeftColumnProps) {
  const selectedCourse = courses.find((c) => c.id === formData.courseId);

  return (
    <div className="bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] p-3 space-y-4">
      <div>
        <div className="flex items-center gap-2 mb-2">
          <Label className="text-sm text-[#37352F]/60">コース</Label>
          <MasterLink category="trimming_course" label="マスタ管理" />
        </div>
        <MasterSelectTrigger
          selectedItem={selectedCourse ? { name: selectedCourse.name, price: selectedCourse.price } : undefined}
          placeholder="コースを選択"
          icon={<Scissors className="size-4" />}
          onClick={onCourseModalOpen}
          variant="block"
        />
      </div>

      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">スタイルの希望</Label>
        <Textarea
          value={formData.styleRequest}
          onChange={(e) => onFormChange({ styleRequest: e.target.value })}
          placeholder="スタイルの希望を入力..."
          className="min-h-[80px] text-sm"
        />
      </div>

      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">メモ</Label>
        <Textarea
          value={formData.memo}
          onChange={(e) => onFormChange({ memo: e.target.value })}
          placeholder="メモを入力..."
          className="min-h-[60px] text-sm"
        />
      </div>

      <div>
        <div className="flex items-center gap-2 mb-2">
          <Label className="text-sm text-[#37352F]/60">オプション</Label>
          <MasterLink category="trimming_option" label="マスタ管理" />
        </div>
        {options.length > 0 ? (
          <div className="space-y-2">
            {options.map((option) => (
              <div key={option.id} className="flex items-center gap-2">
                <Checkbox
                  id={`option-${option.id}`}
                  checked={formData.optionIds.includes(option.id)}
                  onCheckedChange={(checked) => {
                    if (checked) {
                      onFormChange({ optionIds: [...formData.optionIds, option.id] });
                    } else {
                      onFormChange({ optionIds: formData.optionIds.filter((id) => id !== option.id) });
                    }
                  }}
                />
                <label htmlFor={`option-${option.id}`} className="text-sm text-[#37352F] cursor-pointer">
                  {option.name}
                </label>
                {option.price != null ? (
                  <span className="text-xs text-[#37352F]/60 ml-auto">
                    ¥{option.price.toLocaleString()}
                  </span>
                ) : null}
              </div>
            ))}
          </div>
        ) : null}
      </div>

      {/* Style Image */}
      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">希望スタイル画像</Label>
        {styleImagePreview ? (
          <div className="relative">
            <img
              src={styleImagePreview}
              alt="Style preview"
              className="w-full h-32 object-cover rounded-md border border-[#37352F]/20"
            />
            <button
              onClick={onRemoveStyleImage}
              className="absolute top-1 right-1 p-1 bg-white rounded-full shadow-sm hover:bg-gray-100"
            >
              <X className="size-4 text-[#37352F]" />
            </button>
          </div>
        ) : (
          <label className="flex items-center justify-center w-full h-32 border-2 border-dashed border-[rgba(55,53,47,0.16)] rounded-md cursor-pointer hover:bg-[#F7F6F3]">
            <div className="flex flex-col items-center">
              <Upload className="size-6 text-[#37352F]/40 mb-1" />
              <span className="text-sm text-[#37352F]/60">画像をアップロード</span>
            </div>
            <input type="file" accept="image/*" onChange={onStyleImageChange} className="hidden" />
          </label>
        )}
      </div>
    </div>
  );
});

interface MiddleColumnProps {
  formData: TrimmingFormData;
  completedImagePreview: string | null;
  onFormChange: (updates: Partial<TrimmingFormData>) => void;
  onCompletedImageChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onRemoveCompletedImage: () => void;
}

const MiddleColumn = memo(function MiddleColumn({
  formData,
  completedImagePreview,
  onFormChange,
  onCompletedImageChange,
  onRemoveCompletedImage,
}: MiddleColumnProps) {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] p-3 space-y-4">
      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">体重 (BW)</Label>
        <div className="flex gap-2">
          <Input
            type="number"
            value={formData.bw}
            onChange={(e) => onFormChange({ bw: e.target.value })}
            placeholder="体重"
            className="flex-1 text-sm"
          />
          <Select
            value={formData.bwUnit}
            onValueChange={(val) => onFormChange({ bwUnit: val as "Kg" | "g" })}
          >
            <SelectTrigger className="w-[80px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {BW_UNIT_ITEMS}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">体温 (BT)</Label>
        <Input
          type="number"
          step="0.1"
          value={formData.bt}
          onChange={(e) => onFormChange({ bt: e.target.value })}
          placeholder="体温"
          className="text-sm"
        />
      </div>

      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">使用シャンプー</Label>
        <Input
          value={formData.usedShampoo}
          onChange={(e) => onFormChange({ usedShampoo: e.target.value })}
          placeholder="シャンプー名"
          className="text-sm"
        />
      </div>

      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">使用リボン</Label>
        <Input
          value={formData.usedRibbon}
          onChange={(e) => onFormChange({ usedRibbon: e.target.value })}
          placeholder="リボン"
          className="text-sm"
        />
      </div>

      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">備考</Label>
        <Textarea
          value={formData.remarks}
          onChange={(e) => onFormChange({ remarks: e.target.value })}
          placeholder="備考を入力..."
          className="min-h-[60px] text-sm"
        />
      </div>

      {/* Completed Image */}
      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">完成画像</Label>
        {completedImagePreview ? (
          <div className="relative">
            <img
              src={completedImagePreview}
              alt="Completed preview"
              className="w-full h-32 object-cover rounded-md border border-[#37352F]/20"
            />
            <button
              onClick={onRemoveCompletedImage}
              className="absolute top-1 right-1 p-1 bg-white rounded-full shadow-sm hover:bg-gray-100"
            >
              <X className="size-4 text-[#37352F]" />
            </button>
          </div>
        ) : (
          <label className="flex items-center justify-center w-full h-32 border-2 border-dashed border-[rgba(55,53,47,0.16)] rounded-md cursor-pointer hover:bg-[#F7F6F3]">
            <div className="flex flex-col items-center">
              <Upload className="size-6 text-[#37352F]/40 mb-1" />
              <span className="text-sm text-[#37352F]/60">画像をアップロード</span>
            </div>
            <input type="file" accept="image/*" onChange={onCompletedImageChange} className="hidden" />
          </label>
        )}
      </div>
    </div>
  );
});

interface HistoryItem {
  id: string;
  date: string;
  styleRequest: string;
  staff: string;
}

interface RightColumnProps {
  sortedHistory: HistoryItem[];
  historySearchTerm: string;
  historySortOrder: SortOrder;
  historyDateRange: { from: string; to: string };
  onSearchTermChange: (val: string) => void;
  onSortOrderChange: (val: SortOrder) => void;
  onClear: () => void;
  onFilterStartDateChange: (val: string) => void;
  onFilterEndDateChange: (val: string) => void;
  onHistoryClick: (updates: Partial<TrimmingFormData>) => void;
}

const RightColumn = memo(function RightColumn({
  sortedHistory,
  historySearchTerm,
  historySortOrder,
  historyDateRange,
  onSearchTermChange,
  onSortOrderChange,
  onClear,
  onFilterStartDateChange,
  onFilterEndDateChange,
  onHistoryClick,
}: RightColumnProps) {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] p-3 space-y-4">
      <div>
        <Label className="text-sm text-[#37352F]/60 mb-2 block">施術履歴</Label>
        <HistoryFilterPanel
          searchTerm={historySearchTerm}
          onSearchTermChange={onSearchTermChange}
          sortOrder={historySortOrder}
          onSortOrderChange={onSortOrderChange}
          onClear={onClear}
          showDateRange={true}
          filterStartDate={historyDateRange.from}
          onFilterStartDateChange={onFilterStartDateChange}
          filterEndDate={historyDateRange.to}
          onFilterEndDateChange={onFilterEndDateChange}
          searchPlaceholder="スタイル希望で検索..."
        />
      </div>

      {/* History Cards */}
      <div className="space-y-2 max-h-[600px] overflow-y-auto">
        {sortedHistory.length === 0 ? (
          <div className="text-center py-8 text-sm text-[#37352F]/40">
            施術履歴がありません
          </div>
        ) : (
          sortedHistory.map((hist) => (
            <div
              key={hist.id}
              className="p-3 border border-[rgba(55,53,47,0.16)] rounded-lg bg-white hover:bg-[#F7F6F3] transition-colors cursor-pointer"
              onClick={() => onHistoryClick({ styleRequest: hist.styleRequest, staffName: hist.staff })}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <div className="text-xs text-[#37352F]/60 mb-1">{hist.date}</div>
                  <div className="text-sm text-[#37352F] font-medium truncate">{hist.styleRequest}</div>
                  <div className="text-xs text-[#37352F]/60 mt-1">{hist.staff}</div>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
});

// ─── メインコンポーネント ────────────────────────────────────────────────────

export function TrimmingForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");

  const { data: courses = [] } = useMasterItems("trimmingCourse");
  const { data: options = [] } = useMasterItems("trimmingOption");
  const { data: staffItems = [] } = useMasterItems("staff");

  const {
    mode,
    formData,
    setFormData,
    styleImagePreview,
    completedImagePreview,
    petSelection,
    handleStyleImageChange,
    handleCompletedImageChange,
    removeStyleImage,
    removeCompletedImage,
    handleSave,
    handleDelete,
    isSaving,
    isDeleting,
  } = useTrimmingForm(id);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  const [courseModalOpen, setCourseModalOpen] = useState(false);
  const [staffModalOpen, setStaffModalOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [historySortOrder, setHistorySortOrder] = useState<SortOrder>("desc");
  // rerender-dependencies: object state を 2 つの primitive state に分割
  const [historyDateRangeFrom, setHistoryDateRangeFrom] = useState("");
  const [historyDateRangeTo, setHistoryDateRangeTo] = useState("");
  const deferredHistorySearch = useDeferredValue(historySearchTerm);

  const { data: petTrimmings = [] } = useGetTrimmingsByPetId(selectedPet?.id ?? "");

  const sortedHistory = useMemo(() => {
    const filtered = petTrimmings.filter((t) => {
      if (deferredHistorySearch && !t.styleRequest.toLowerCase().includes(deferredHistorySearch.toLowerCase())) {
        return false;
      }
      const recordDate = t.date.slice(0, 10);
      if (historyDateRangeFrom && recordDate < historyDateRangeFrom) return false;
      if (historyDateRangeTo && recordDate > historyDateRangeTo) return false;
      return true;
    });
    return filtered.toSorted((a, b) => {
      const dateA = new Date(a.date).getTime();
      const dateB = new Date(b.date).getTime();
      return historySortOrder === "desc" ? dateB - dateA : dateA - dateB;
    });
  }, [petTrimmings, deferredHistorySearch, historyDateRangeFrom, historyDateRangeTo, historySortOrder]);

  // rerender-functional-setstate: useCallback でハンドラを安定化
  // ※ React hooks はコンポーネントのトップレベルで呼ぶ（ガード節の前に定義）
  const handleFormChange = useCallback((updates: Partial<TrimmingFormData>) => {
    markDirty();
    setFormData(updates);
  }, [markDirty, setFormData]);

  const handleSaveClick = useCallback(() => {
    handleSave(markClean);
  }, [handleSave, markClean]);

  const handleDeleteClick = useCallback(() => {
    handleDelete(() => {
      markClean();
      navigate(paths.trimming.getHref());
    });
  }, [handleDelete, markClean, navigate]);

  // rerender-dependencies: location.state (object) から primitive を抽出して deps を安定化
  const fromPath = location.state?.from as string | undefined;
  const handleBack = useCallback(() => {
    navigate(fromPath ?? "/trimming");
  }, [fromPath, navigate]);

  const handleOpenCourseModal = useCallback(() => setCourseModalOpen(true), []);
  const handleOpenStaffModal = useCallback(() => setStaffModalOpen(true), []);

  const handleHistoryClick = useCallback((updates: Partial<TrimmingFormData>) => {
    handleFormChange(updates);
  }, [handleFormChange]);

  const handleHistoryClear = useCallback(() => {
    setHistorySearchTerm("");
    setHistorySortOrder("desc");
    setHistoryDateRangeFrom("");
    setHistoryDateRangeTo("");
  }, []);

  const handleHistoryStartDateChange = useCallback((val: string) => {
    setHistoryDateRangeFrom(val);
  }, []);

  const handleHistoryEndDateChange = useCallback((val: string) => {
    setHistoryDateRangeTo(val);
  }, []);

  const activeStaffItems = useMemo(
    () => staffItems.filter((s) => s.status === "active"),
    [staffItems]
  );

  // rerender-dependencies: primitive から object を再構築して RightColumn に渡す (安定参照)
  const historyDateRange = useMemo(
    () => ({ from: historyDateRangeFrom, to: historyDateRangeTo }),
    [historyDateRangeFrom, historyDateRangeTo]
  );

  if (!selectedPet && mode === "new" && petId) return null;
  if (!selectedPet && mode === "new") return null;

  return (
    <PageLayout
      title={mode === "new" ? "トリミング登録" : "トリミング編集"}
      onBack={handleBack}
      icon={<Scissors className="h-4 w-4 text-[#37352F]" />}
      maxWidth="max-w-[1400px]"
      headerAction={
        <div className="flex gap-2">
          {/* rendering-conditional-render: && → ? ... : null */}
          {mode === "edit" ? (
            <Button
              onClick={() => setDeleteConfirmOpen(true)}
              variant="ghost"
              className="h-10 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-[6px] text-sm px-4"
              disabled={isDeleting}
            >
              <Trash2 className="mr-1.5 size-4" />
              削除
            </Button>
          ) : null}
          <PrimaryButton onClick={handleSaveClick} disabled={isSaving} className="h-10">
            {isSaving ? "保存中..." : "保存"}
          </PrimaryButton>
        </div>
      }
    >
      {/* NavigationBlocker: isSaving 中はブロック無効化 */}
      <NavigationBlocker when={isDirty && !isSaving} />

      {/* rendering-conditional-render: && → ? ... : null */}
      {selectedPet ? (
        <div className="space-y-6">
          {/* Patient Info Card */}
          <PatientInfoCard
            ownerName={selectedPet.ownerName}
            petName={selectedPet.name}
            petNumber={selectedPet.petNumber || ""}
            weight={selectedPet.weight || ""}
            staffName={formData.staffName}
            serviceType="トリミング"
            nextVisitDate="-"
            nextVisitContent="-"
            onStaffClick={handleOpenStaffModal}
          />

          {/* Main Content - 3 column layout */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* rerender-memo: 各カラムを独立したmemo化コンポーネントに分離 */}
            <LeftColumn
              formData={formData}
              courses={courses}
              options={options}
              styleImagePreview={styleImagePreview}
              onFormChange={handleFormChange}
              onCourseModalOpen={handleOpenCourseModal}
              onStyleImageChange={handleStyleImageChange}
              onRemoveStyleImage={removeStyleImage}
            />
            <MiddleColumn
              formData={formData}
              completedImagePreview={completedImagePreview}
              onFormChange={handleFormChange}
              onCompletedImageChange={handleCompletedImageChange}
              onRemoveCompletedImage={removeCompletedImage}
            />
            <RightColumn
              sortedHistory={sortedHistory}
              historySearchTerm={historySearchTerm}
              historySortOrder={historySortOrder}
              historyDateRange={historyDateRange}
              onSearchTermChange={setHistorySearchTerm}
              onSortOrderChange={setHistorySortOrder}
              onClear={handleHistoryClear}
              onFilterStartDateChange={handleHistoryStartDateChange}
              onFilterEndDateChange={handleHistoryEndDateChange}
              onHistoryClick={handleHistoryClick}
            />
          </div>
        </div>
      ) : null}

      <Suspense fallback={null}>
        {/* Course Modal */}
        <MasterSelectModal
          open={courseModalOpen}
          onOpenChange={setCourseModalOpen}
          title="コース選択"
          description="施術するコースを選択してください"
          items={courses}
          selectedValue={formData.courseId}
          matchBy="id"
          onSelect={(item) => handleFormChange({ courseId: item.id })}
          masterCategory="trimming_course"
        />

        {/* Staff Modal */}
        <MasterSelectModal
          open={staffModalOpen}
          onOpenChange={setStaffModalOpen}
          title="担当スタッフ選択"
          description="担当するスタッフを選択してください"
          items={activeStaffItems}
          selectedValue={formData.staffName}
          matchBy="name"
          onSelect={(item) => handleFormChange({ staffName: item.name, staffId: item.id })}
          masterCategory="staff"
        />

        {/* Delete Confirmation Dialog */}
        <ConfirmDialog
          open={deleteConfirmOpen}
          onClose={() => setDeleteConfirmOpen(false)}
          title="削除確認"
          description="このトリミング情報を削除してもよろしいですか？"
          confirmLabel="削除"
          variant="destructive"
          onConfirm={handleDeleteClick}
        />
      </Suspense>
    </PageLayout>
  );
}
