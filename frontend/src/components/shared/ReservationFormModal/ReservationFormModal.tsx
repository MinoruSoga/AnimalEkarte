// React/Framework
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useState, useEffect, useCallback, useMemo, memo } from "react";
import { format as dateFnsFormat } from "date-fns";

// External
import { Calendar, CalendarCheck, PawPrint, X, Search } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { useGetPet } from "@/hooks/use-pet";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { FormFieldError } from "@/components/shared/FormFieldError";

// Features
import { useGetClinicHolidays } from "@/features/shifts";

// Relative
import { PatientSelectionTable } from "./PatientSelectionTable";
import { ReservationFormFields } from "./ReservationFormFields";

// Types
import type { Pet, Appointment } from "@/types";

interface ReservationFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: Partial<Appointment>, selectedPets: Pet[]) => void;
  initialData: Partial<Appointment> | null;
  canCreate?: boolean;
  canEdit?: boolean;
}


const StepIndicator = memo(function StepIndicator({ step, label, active }: { step: number; label: string; active: boolean }) {
  return (
    <div className={`flex items-center gap-1.5 text-xs ${active ? C.accent : C.text30}`}>
      <span
        className={`w-5 h-5 rounded-full flex items-center justify-center text-[11px] font-bold transition-colors ${
          active ? `${C.bgAccent} ${C.textWhite}` : `${C.bgPrimary10} ${C.text30}`
        }`}
      >
        {step}
      </span>
      {label}
    </div>
  );
});

const SelectedPetChip = memo(function SelectedPetChip({ pet, onRemove }: { pet: Pet; onRemove: () => void }) {
  return (
    <div className={`flex items-center gap-2 bg-white p-2 rounded-lg border ${C.borderMediumLight} shadow-sm`}>
      <PawPrint className={`${ICON.action} ${C.text60} flex-shrink-0`} />
      <span className={`text-sm font-bold ${C.text}`}>{pet.name}</span>
      <Badge variant="outline" className={`text-[11px] font-normal ${C.text60} ${C.bgPage} ${C.borderMediumLight} h-5`}>
        {pet.species}
      </Badge>
      <span className={`text-[11px] ${C.text60} ml-auto`}>
        No. {pet.ownerId} {pet.ownerName}
      </span>
      <button
        onClick={onRemove}
        className={`ml-1 p-1 ${C.hoverBgDanger5} rounded transition-colors`}
      >
        <X className={`${ICON.action} ${C.danger} ${C.hoverTextDanger}`} />
      </button>
    </div>
  );
});

export const ReservationFormModal = memo(function ReservationFormModal({
  isOpen,
  onClose,
  onSave,
  initialData,
  canCreate = false,
  canEdit = false,
}: ReservationFormModalProps) {
  const [formData, setFormData] = useState<Partial<Appointment>>({});
  const [pendingPetId, setPendingPetId] = useState<string | null>(null);
  const [mobilePanel, setMobilePanel] = useState<"search" | "form">("search");
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const [calendarMonth, setCalendarMonth] = useState<string>(() => dateFnsFormat(new Date(), "yyyy-MM"));

  // BUG-343: 定休日を取得して Calendar で disabled にする
  const { data: clinicHolidays = [] } = useGetClinicHolidays(calendarMonth);
  const holidayDates = useMemo(
    () => new Set(clinicHolidays.map((h) => h.date)),
    [clinicHolidays]
  );
  const handleCalendarMonthChange = useCallback((yearMonth: string) => {
    setCalendarMonth(yearMonth);
  }, []);

  const {
    selectedPets,
    setSelectedPets,
    togglePetSelection,
  } = usePetSelection([], "multiple-same-owner");

  const { data: loadedPet } = useGetPet(pendingPetId ?? "");

  useEffect(() => {
    if (loadedPet && pendingPetId) {
      /* eslint-disable react-hooks/set-state-in-effect -- 非同期ペットデータのロード完了後に選択状態を設定するパターン */
      setSelectedPets([loadedPet]);
      setPendingPetId(null);
      /* eslint-enable react-hooks/set-state-in-effect */
    }
  }, [loadedPet, pendingPetId, setSelectedPets]);

  useEffect(() => {
    if (!isOpen) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- モーダル open 時にフォームをリセット。key prop パターンの代替
    setValidationErrors({});
    setMobilePanel("search");
    if (initialData) {
      setFormData({ ...initialData });
      if (initialData.petId) {
        setPendingPetId(initialData.petId);
      } else {
        setSelectedPets([]);
        setPendingPetId(null);
      }
    } else {
      const defaultStart = new Date();
      defaultStart.setHours(10, 0, 0, 0);
      const defaultEnd = new Date(defaultStart);
      defaultEnd.setHours(11, 0, 0, 0);
      setFormData({
        start: defaultStart,
        end: defaultEnd,
        visitType: "first",
        doctor: "",
        isDesignated: false,
        status: "confirmed",
      });
      setSelectedPets([]);
      setPendingPetId(null);
    }
  }, [isOpen, initialData, setSelectedPets]);

  const isEditMode = initialData && initialData.id;
  const canSave = isEditMode ? canEdit : canCreate;

  const handleSave = useCallback(() => {
    const errors: Record<string, string> = {};
    if (selectedPets.length === 0) {
      errors.patient = "患者を選択してください";
    }
    if (!formData.start) {
      errors.date = "日付を選択してください";
    }
    if (!formData.type) {
      errors.type = "予約区分を選択してください";
    }
    // BUG-034: end_time > start_time バリデーション
    if (formData.start && formData.end && formData.end <= formData.start) {
      errors.time = "終了時刻は開始時刻より後に設定してください";
    }
    // BUG-098: 新規予約は過去日付不可
    if (!isEditMode && formData.start) {
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      const reservationDate = new Date(formData.start);
      reservationDate.setHours(0, 0, 0, 0);
      if (reservationDate < today) {
        errors.date = "本日以降の日付を選択してください";
      }
    }

    if (Object.keys(errors).length > 0) {
      setValidationErrors(errors);
      toast.error("入力内容を確認してください", {
        description: Object.values(errors).join("、"),
      });
      return;
    }

    setValidationErrors({});
    onSave(formData, selectedPets);
  }, [formData, selectedPets, onSave, isEditMode]);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className={`${LAYOUT.modal.full} flex flex-col p-0 gap-0 bg-white overflow-hidden rounded-xl`}>
        <DialogHeader className="p-4 border-b shrink-0 h-auto flex flex-col gap-3 space-y-0">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {isEditMode ? (
                <CalendarCheck className={`${ICON.page} ${C.textNotice}`} />
              ) : (
                <Calendar className={`${ICON.page} ${C.accent}`} />
              )}
              <DialogTitle className={`text-sm font-bold ${C.text}`}>
                {isEditMode ? "予約編集" : "新規予約作成"}
              </DialogTitle>
            </div>
            <DialogDescription className="sr-only">
              左側のリストからペットを選択し、右側のフォームで予約情報を入力してください
            </DialogDescription>
          </div>
          <div className="flex items-center gap-6">
            <StepIndicator step={1} label="患者選択" active={mobilePanel === "search"} />
            <StepIndicator step={2} label="予約情報" active={mobilePanel === "form"} />
          </div>
          {/* Mobile Tab Bar */}
          <div className="flex gap-2 lg:hidden border-t pt-3 -mx-4 px-4">
            <Button
              size="sm"
              variant={mobilePanel === "search" ? "default" : "outline"}
              onClick={() => setMobilePanel("search")}
              className="flex-1 h-9 text-sm"
            >
              患者選択
            </Button>
            <Button
              size="sm"
              variant={mobilePanel === "form" ? "default" : "outline"}
              onClick={() => setMobilePanel("form")}
              className="flex-1 h-9 text-sm"
            >
              予約情報
            </Button>
          </div>
        </DialogHeader>

        <div className="flex flex-col lg:flex-row flex-1 overflow-hidden">
          {/* Left Panel: Patient Selection Table */}
          <div
            className={cn(
              `w-full lg:w-7/12 border-b lg:border-b-0 lg:border-r ${C.bgSubtle} p-4 flex flex-col overflow-hidden min-h-[300px] lg:min-h-auto flex-1`,
              mobilePanel !== "search" && "hidden lg:flex"
            )}
          >
            <div className="mb-3 flex items-center gap-2 shrink-0">
              <Search className={`${ICON.action} ${C.text60}`} />
              <Label className={`text-sm font-bold ${C.text}`}>患者検索</Label>
            </div>
            <div className="flex-1 overflow-hidden flex flex-col min-h-0">
              <PatientSelectionTable
                onSelect={togglePetSelection}
                selectedPets={selectedPets}
              />
            </div>
          </div>

          {/* Right Panel: Reservation Form */}
          <div
            className={cn(
              "w-full lg:w-5/12 bg-white flex flex-col overflow-hidden min-h-[300px] lg:min-h-auto flex-1",
              mobilePanel !== "form" && "hidden lg:flex"
            )}
          >
            <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
              {/* Selected Patient Summary (Top of Form) */}
              <div className={`rounded-lg border p-3 transition-colors ${selectedPets.length > 0 ? `${C.bgAccentLight50} ${C.borderAccentLight}` : `${C.bgPage} ${C.borderMediumLight}`}`}>
                <Label className={`text-[12px] ${C.text40} font-bold tracking-widest uppercase block mb-3`}>
                  予約対象（選択中）
                </Label>

                {selectedPets.length > 0 ? (
                  <div className="flex flex-col gap-2">
                    {selectedPets.map(pet => (
                      <SelectedPetChip
                        key={pet.id}
                        pet={pet}
                        onRemove={() => {
                          setSelectedPets(prev => prev.filter(p => p.id !== pet.id));
                        }}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center h-20 text-center">
                    <PawPrint className={`${ICON.lg} ${C.text15} mb-2`} />
                    <div className={`text-[12px] ${C.text40}`}>
                      左側から患者を選択してください
                    </div>
                  </div>
                )}
                {selectedPets.length === 0 ? (
                  <FormFieldError message={validationErrors.patient} />
                ) : null}
              </div>

              {/* Form Fields */}
              <div className="space-y-4">
                <Label className={`text-sm font-bold ${C.text}`}>予約詳細</Label>
                <ReservationFormFields
                  formData={formData}
                  onChange={(data) => {
                    setFormData(data);
                    setValidationErrors((prev) => {
                      const next = { ...prev };
                      if (data.start) delete next.date;
                      if (data.type) delete next.type;
                      return next;
                    });
                  }}
                  validationErrors={validationErrors}
                  onClearError={(field) =>
                    setValidationErrors((prev) => {
                      const next = { ...prev };
                      delete next[field];
                      return next;
                    })
                  }
                  holidayDates={holidayDates}
                  onMonthChange={handleCalendarMonthChange}
                />
              </div>
            </div>
          </div>
        </div>

        <DialogFooter className="p-4 border-t bg-white shrink-0 h-14 flex items-center justify-between gap-2">
          <div className="flex items-center gap-1.5">
            <PawPrint className={`${ICON.action} ${C.text60}`} />
            <span className={`text-sm ${C.text60}`}>
              {selectedPets.length}頭 選択中
            </span>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onClose} className="h-10 text-sm">
              キャンセル
            </Button>
            {canSave ? (
              <Button
                onClick={handleSave}
                className={`${C.bgAccent} ${C.textWhite} ${C.bgAccentHover} h-10 text-sm min-w-[100px]`}
              >
                {isEditMode ? "更新する" : "予約を確定"}
              </Button>
            ) : null}
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
});
