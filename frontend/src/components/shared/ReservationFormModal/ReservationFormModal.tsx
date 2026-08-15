// React/Framework
import { LAYOUT } from "@/lib/design-tokens";
import { useState, useEffect, useLayoutEffect, useCallback, useMemo, memo } from "react";
import { format as dateFnsFormat } from "date-fns";

// External
// Internal
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { useGetPet } from "@/hooks/use-pet";
import { usePetSelection } from "@/hooks/use-pet-selection";

import { useGetClinicHolidays } from "@/hooks/use-clinic-holidays";
import { useGetOwnerLineTags } from "@/hooks/use-owner-line-tags";
import { useGetReservation } from "@/hooks/use-get-reservation";
import type { NewOwnerFormData } from "@/types/reservation-form";

// Relative
import {
  ReservationDetailsPanel,
  ReservationModalFooter,
  ReservationModalHeader,
  ReservationPatientPanel,
  type MobilePanel,
  type OwnerMode,
} from "./ReservationFormModalPanels";

// Types
import type { Pet, Reservation } from "@/types";

const EMPTY_NEW_OWNER: NewOwnerFormData = {
  ownerName: "",
  phone: "",
  petName: "",
  chiefComplaint: "",
  animalSpeciesId: 0,
};
const RESERVATION_FORM_DESCRIPTION_ID = "reservation-form-description";
const OWNER_PHONE_PATTERN = /^0\d{1,4}-?\d{1,4}-?\d{4}$/;

function isValidOwnerPhone(phone: string): boolean {
  return OWNER_PHONE_PATTERN.test(phone);
}

interface ReservationFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: Partial<Reservation>, selectedPets: Pet[], newOwnerData?: NewOwnerFormData) => void;
  initialData: Partial<Reservation> | null;
  canCreate?: boolean;
  canEdit?: boolean;
}
export const ReservationFormModal = memo(function ReservationFormModal({
  isOpen,
  onClose,
  onSave,
  initialData,
  canCreate = false,
  canEdit = false,
}: ReservationFormModalProps) {
  const [formData, setFormData] = useState<Partial<Reservation>>({});
  const [pendingPetId, setPendingPetId] = useState<string | null>(null);
  const [mobilePanel, setMobilePanel] = useState<MobilePanel>("search");
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const [calendarMonth, setCalendarMonth] = useState<string>(() => dateFnsFormat(new Date(), "yyyy-MM"));
  const [ownerMode, setOwnerMode] = useState<OwnerMode>("existing");
  const [newOwnerData, setNewOwnerData] = useState<NewOwnerFormData>(EMPTY_NEW_OWNER);
  const [newOwnerErrors, setNewOwnerErrors] = useState<Record<string, string>>({});

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

  const { data: ownerLineData } = useGetOwnerLineTags(selectedPets[0]?.ownerId ?? "");
  const lstepStatus = selectedPets.length === 0 || ownerLineData === undefined
    ? undefined
    : ownerLineData.lstep_opt_out
      ? ("opt-out" as const)
      : ownerLineData.is_linked
        ? ("synced" as const)
        : ("not-linked" as const);

  const { data: loadedPet } = useGetPet(pendingPetId ?? "");

  useEffect(() => {
    if (loadedPet && pendingPetId) {
      setSelectedPets([loadedPet]);
      setPendingPetId(null);
    }
  }, [loadedPet, pendingPetId, setSelectedPets]);

  // useLayoutEffect: ペイント前に同期実行し、クリックした日時がフォームに即反映されるようにする (Issue #52)
  useLayoutEffect(() => {
    if (!isOpen) return;
    setValidationErrors({});
    setNewOwnerErrors({});
    setOwnerMode("existing");
    setNewOwnerData(EMPTY_NEW_OWNER);
    setMobilePanel("search");
    setCalendarMonth(dateFnsFormat(new Date(), "yyyy-MM")); // BUG-343: 月またぎ表示リセット
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

  const isEditMode = Boolean(initialData?.id);
  const canSave = isEditMode ? canEdit : canCreate;

  const handleNewOwnerChange = useCallback((data: NewOwnerFormData) => {
    setNewOwnerData(data);
    if (!isValidOwnerPhone(data.phone)) return;

    setNewOwnerErrors((previous) => {
      if (!previous.phone) return previous;
      const next = { ...previous };
      delete next.phone;
      return next;
    });
  }, []);

  // edit mode: subscribe to single-reservation query and sync reservationRoute into formData
  const reservationQueryId = isEditMode ? String(initialData?.id) : "";
  const { data: latestReservation } = useGetReservation(reservationQueryId);
  useEffect(() => {
    if (!latestReservation) return;
    setFormData((prev) => ({
      ...prev,
      reservationRoute: latestReservation.reservationRoute ?? null,
    }));
  }, [latestReservation]);

  const handleSave = useCallback(() => {
    const errors: Record<string, string> = {};
    const noErrors: Record<string, string> = {};

    if (ownerMode === "new") {
      // 新規飼主モードのバリデーション
      const noe: Record<string, string> = {};
      if (!newOwnerData.ownerName.trim()) noe.ownerName = "飼主名を入力してください";
      if (!newOwnerData.phone.trim()) {
        noe.phone = "電話番号を入力してください";
      } else if (!isValidOwnerPhone(newOwnerData.phone)) {
        noe.phone = "電話番号の形式が正しくありません（例：090-1234-5678 または 09012345678）";
      }
      if (!newOwnerData.petName.trim()) noe.petName = "ペット名を入力してください";
      if (!newOwnerData.animalSpeciesId) noe.animalSpeciesId = "動物種を選択してください";
      if (!newOwnerData.chiefComplaint.trim()) noe.chiefComplaint = "主訴を入力してください";

      if (!formData.start) errors.date = "日付を選択してください";
      if (!formData.type) errors.type = "予約区分を選択してください";
      if (formData.start && formData.end && formData.end <= formData.start) {
        errors.time = "終了時刻は開始時刻より後に設定してください";
      }
      if (!isEditMode && formData.start) {
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        const reservationDate = new Date(formData.start);
        reservationDate.setHours(0, 0, 0, 0);
        if (reservationDate < today) errors.date = "本日以降の日付を選択してください";
      }

      if (Object.keys(noe).length > 0 || Object.keys(errors).length > 0) {
        setNewOwnerErrors(noe);
        setValidationErrors(errors);
        return;
      }
      setNewOwnerErrors(noErrors);
      setValidationErrors(noErrors);
      onSave(formData, [], newOwnerData);
      return;
    }

    // 既存飼主モードのバリデーション
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
      return;
    }

    setValidationErrors(noErrors);
    onSave(formData, selectedPets);
  }, [formData, selectedPets, onSave, isEditMode, ownerMode, newOwnerData]);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent
        aria-describedby={RESERVATION_FORM_DESCRIPTION_ID}
        className={`${LAYOUT.modal.full} flex flex-col p-0 gap-0 bg-white overflow-hidden rounded-xl`}
      >
        <ReservationModalHeader
          isEditMode={isEditMode}
          mobilePanel={mobilePanel}
          onMobilePanelChange={setMobilePanel}
          descriptionId={RESERVATION_FORM_DESCRIPTION_ID}
        />

        <div className="flex flex-col lg:flex-row flex-1 overflow-hidden">
          <ReservationPatientPanel
            isEditMode={isEditMode}
            ownerMode={ownerMode}
            mobilePanel={mobilePanel}
            selectedPets={selectedPets}
            newOwnerData={newOwnerData}
            newOwnerErrors={newOwnerErrors}
            onOwnerModeChange={setOwnerMode}
            onPetSelect={togglePetSelection}
            onNewOwnerChange={handleNewOwnerChange}
          />

          <ReservationDetailsPanel
            mobilePanel={mobilePanel}
            selectedPets={selectedPets}
            lstepStatus={lstepStatus}
            formData={formData}
            validationErrors={validationErrors}
            holidayDates={holidayDates}
            isEditMode={isEditMode}
            reservationId={reservationQueryId || undefined}
            canEdit={canEdit}
            onSelectedPetsChange={setSelectedPets}
            onFormChange={(data) => {
              setFormData(data);
              setValidationErrors((prev) => {
                const next = { ...prev };
                if (data.start) delete next.date;
                if (data.type) delete next.type;
                return next;
              });
            }}
            onClearError={(field) =>
              setValidationErrors((prev) => {
                const next = { ...prev };
                delete next[field];
                return next;
              })
            }
            onMonthChange={handleCalendarMonthChange}
          />
        </div>

        <ReservationModalFooter
          ownerMode={ownerMode}
          selectedPetsCount={selectedPets.length}
          isEditMode={isEditMode}
          canSave={canSave}
          onClose={onClose}
          onSave={handleSave}
        />
      </DialogContent>
    </Dialog>
  );
});
