// React/Framework
import { useState, useMemo } from "react";
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
import { PageLayout } from "@/components/shared/PageLayout";
import { PrimaryButton } from "@/components/shared/Form";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { MasterLink } from "@/components/shared/MasterLink";
import { MasterSelectModal, MasterSelectTrigger } from "@/components/shared/MasterSelectModal";
import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import type { SortOrder } from "@/types";
import { useMasterItems } from "@/hooks/use-master-items";
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";

// Relative
import { useTrimmingForm, type TrimmingFormData } from "../hooks/useTrimmingForm";
import { useGetTrimmingsByPetId } from "../api";

// ✅ React 19: function宣言を使用 (CLAUDE.md準拠)
export function TrimmingForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");

  const { data: courses } = useMasterItems("trimmingCourse");
  const { data: options } = useMasterItems("trimmingOption");
  const { data: staffItems } = useMasterItems("staff");

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

  // Unsaved changes
  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // Modal states
  const [courseModalOpen, setCourseModalOpen] = useState(false);
  const [staffModalOpen, setStaffModalOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  // History filter
  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [historySortOrder, setHistorySortOrder] = useState<SortOrder>("desc");
  const [historyDateRange, setHistoryDateRange] = useState({ from: "", to: "" });

  // History data
  const { data: petTrimmings = [] } = useGetTrimmingsByPetId(selectedPet?.id ?? "");

  // ✅ useMemo でメモ化 — filter + sort を毎レンダリングで再実行しない (rerender-memo)
  const sortedHistory = useMemo(() => {
    const filtered = petTrimmings.filter((t) => {
      if (historySearchTerm && !t.styleRequest.toLowerCase().includes(historySearchTerm.toLowerCase())) {
        return false;
      }
      const recordDate = t.date.slice(0, 10);
      if (historyDateRange.from && recordDate < historyDateRange.from) return false;
      if (historyDateRange.to && recordDate > historyDateRange.to) return false;
      return true;
    });
    return filtered.toSorted((a, b) => {
      const dateA = new Date(a.date).getTime();
      const dateB = new Date(b.date).getTime();
      return historySortOrder === "desc" ? dateB - dateA : dateA - dateB;
    });
  }, [petTrimmings, historySearchTerm, historyDateRange, historySortOrder]);

  // ✅ useTrimmingForm 内の useEffect と重複するため削除 (重複ガード除去)

  if (!selectedPet && mode === "new" && petId) return null;
  if (!selectedPet && mode === "new") return null;

  const handleBack = () => {
    if (location.state?.from) {
      navigate(location.state.from);
    } else {
      navigate("/trimming");
    }
  };

  const handleFormChange = (updates: Partial<TrimmingFormData>) => {
    markDirty();
    setFormData(updates);
  };

  // ✅ markClean は保存成功後のコールバックで呼ぶ（失敗時にdirtyフラグが消えないよう）
  const handleSaveClick = () => {
    handleSave(markClean);
  };

  const handleDeleteClick = () => {
    handleDelete(() => {
      markClean();
      navigate("/trimming");
    });
  };

  const selectedCourse = courses.find((c) => c.id === formData.courseId);

  return (
    <PageLayout
      title={mode === "new" ? "トリミング登録" : "トリミング編集"}
      onBack={handleBack}
      icon={<Scissors className="h-4 w-4 text-[#37352F]" />}
      maxWidth="max-w-[1400px]"
      headerAction={
        <div className="flex gap-2">
          {mode === "edit" && (
            <Button
              onClick={() => setDeleteConfirmOpen(true)}
              variant="ghost"
              className="h-10 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-[6px] text-sm px-4"
              disabled={isDeleting}
            >
              <Trash2 className="mr-1.5 size-4" />
              削除
            </Button>
          )}
          <PrimaryButton
            onClick={handleSaveClick}
            disabled={isSaving}
            className="h-10"
          >
            {isSaving ? "保存中..." : "保存"}
          </PrimaryButton>
        </div>
      }
    >
      <NavigationBlocker when={isDirty} />

      {selectedPet && (
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
            onStaffClick={() => setStaffModalOpen(true)}
          />

          {/* Main Content */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Left Column: Course & Options & Images */}
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
                  onClick={() => setCourseModalOpen(true)}
                  variant="block"
                />
              </div>

              <div>
                <Label className="text-sm text-[#37352F]/60 mb-2 block">スタイルの希望</Label>
                <Textarea
                  value={formData.styleRequest}
                  onChange={(e) => handleFormChange({ styleRequest: e.target.value })}
                  placeholder="スタイルの希望を入力..."
                  className="min-h-[80px] text-sm"
                />
              </div>

              <div>
                <Label className="text-sm text-[#37352F]/60 mb-2 block">メモ</Label>
                <Textarea
                  value={formData.memo}
                  onChange={(e) => handleFormChange({ memo: e.target.value })}
                  placeholder="メモを入力..."
                  className="min-h-[60px] text-sm"
                />
              </div>

              <div>
                <div className="flex items-center gap-2 mb-2">
                  <Label className="text-sm text-[#37352F]/60">オプション</Label>
                  <MasterLink category="trimming_option" label="マスタ管理" />
                </div>
                {options.length > 0 && (
                  <div className="space-y-2">
                    {options.map((option) => (
                    <div key={option.id} className="flex items-center gap-2">
                      <Checkbox
                        id={`option-${option.id}`}
                        checked={formData.optionIds.includes(option.id)}
                        onCheckedChange={(checked) => {
                          if (checked) {
                            handleFormChange({
                              optionIds: [...formData.optionIds, option.id],
                            });
                          } else {
                            handleFormChange({
                              optionIds: formData.optionIds.filter((id) => id !== option.id),
                            });
                          }
                        }}
                      />
                      <label htmlFor={`option-${option.id}`} className="text-sm text-[#37352F] cursor-pointer">
                        {option.name}
                      </label>
                      {option.price != null && (
                        <span className="text-xs text-[#37352F]/60 ml-auto">
                          ¥{option.price.toLocaleString()}
                        </span>
                      )}
                    </div>
                  ))}
                  </div>
                )}
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
                      onClick={removeStyleImage}
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
                    <input
                      type="file"
                      accept="image/*"
                      onChange={handleStyleImageChange}
                      className="hidden"
                    />
                  </label>
                )}
              </div>
            </div>

            {/* Middle Column: Body Data */}
            <div className="bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] p-3 space-y-4">
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-2 block">体重 (BW)</Label>
                <div className="flex gap-2">
                  <Input
                    type="number"
                    value={formData.bw}
                    onChange={(e) => handleFormChange({ bw: e.target.value })}
                    placeholder="体重"
                    className="flex-1 text-sm"
                  />
                  <Select
                    value={formData.bwUnit}
                    onValueChange={(val) =>
                      handleFormChange({ bwUnit: val as "Kg" | "g" })
                    }
                  >
                    <SelectTrigger className="w-[80px]">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Kg">Kg</SelectItem>
                      <SelectItem value="g">g</SelectItem>
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
                  onChange={(e) => handleFormChange({ bt: e.target.value })}
                  placeholder="体温"
                  className="text-sm"
                />
              </div>

              <div>
                <Label className="text-sm text-[#37352F]/60 mb-2 block">使用シャンプー</Label>
                <Input
                  value={formData.usedShampoo}
                  onChange={(e) => handleFormChange({ usedShampoo: e.target.value })}
                  placeholder="シャンプー名"
                  className="text-sm"
                />
              </div>

              <div>
                <Label className="text-sm text-[#37352F]/60 mb-2 block">使用リボン</Label>
                <Input
                  value={formData.usedRibbon}
                  onChange={(e) => handleFormChange({ usedRibbon: e.target.value })}
                  placeholder="リボン"
                  className="text-sm"
                />
              </div>

              <div>
                <Label className="text-sm text-[#37352F]/60 mb-2 block">トリートメント</Label>
                <Input
                  value={formData.treatment}
                  onChange={(e) => handleFormChange({ treatment: e.target.value })}
                  placeholder="トリートメント"
                  className="text-sm"
                />
              </div>

              <div>
                <Label className="text-sm text-[#37352F]/60 mb-2 block">備考</Label>
                <Textarea
                  value={formData.remarks}
                  onChange={(e) => handleFormChange({ remarks: e.target.value })}
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
                      onClick={removeCompletedImage}
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
                    <input
                      type="file"
                      accept="image/*"
                      onChange={handleCompletedImageChange}
                      className="hidden"
                    />
                  </label>
                )}
              </div>
            </div>

            {/* Right Column: History */}
            <div className="bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] p-3 space-y-4">
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-2 block">施術履歴</Label>
                <HistoryFilterPanel
                  searchTerm={historySearchTerm}
                  onSearchTermChange={setHistorySearchTerm}
                  sortOrder={historySortOrder}
                  onSortOrderChange={setHistorySortOrder}
                  onClear={() => {
                    setHistorySearchTerm("");
                    setHistorySortOrder("desc");
                    setHistoryDateRange({ from: "", to: "" });
                  }}
                  showDateRange={true}
                  filterStartDate={historyDateRange.from}
                  onFilterStartDateChange={(val) =>
                    setHistoryDateRange({ ...historyDateRange, from: val })
                  }
                  filterEndDate={historyDateRange.to}
                  onFilterEndDateChange={(val) =>
                    setHistoryDateRange({ ...historyDateRange, to: val })
                  }
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
                      onClick={() => {
                        handleFormChange({
                          styleRequest: hist.styleRequest,
                          staffName: hist.staff,
                        });
                      }}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <div className="text-xs text-[#37352F]/60 mb-1">
                            {hist.date}
                          </div>
                          <div className="text-sm text-[#37352F] font-medium truncate">
                            {hist.styleRequest}
                          </div>
                          <div className="text-xs text-[#37352F]/60 mt-1">
                            {hist.staff}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Course Modal */}
      <MasterSelectModal
        open={courseModalOpen}
        onOpenChange={setCourseModalOpen}
        title="コース選択"
        description="施術するコースを選択してください"
        items={courses}
        selectedValue={formData.courseId}
        matchBy="id"
        onSelect={(item) => {
          handleFormChange({ courseId: item.id });
        }}
        masterCategory="trimming_course"
      />

      {/* Staff Modal */}
      <MasterSelectModal
        open={staffModalOpen}
        onOpenChange={setStaffModalOpen}
        title="担当スタッフ選択"
        description="担当するスタッフを選択してください"
        items={staffItems.filter((s) => s.status === "active")}
        selectedValue={formData.staffName}
        matchBy="name"
        onSelect={(item) => {
          handleFormChange({ staffName: item.name });
        }}
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
    </PageLayout>
  );
};
