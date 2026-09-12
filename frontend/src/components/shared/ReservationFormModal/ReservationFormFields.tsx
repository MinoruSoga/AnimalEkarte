import { memo, useMemo, useCallback, useState, useEffect } from "react";
import { isBefore, startOfDay, format } from "date-fns";
import { useGetMasterItems } from "@/hooks/use-master-items";
import {
  getCurrentClinicId,
  isLineReservationSettingsUnsetError,
  useGetReservationTypesGrouped,
  useGetOnDutyStaffs,
  useGetReservationStaffs,
  useGetReservationAvailableTimes,
} from "@/hooks/use-reservation-types";
import { useGetUnavailableTimes } from "@/hooks/use-reservation-type-unavailable-times";
import { DISPLAY_TIME_FORMAT } from "@/lib/format/date";
import { toJSTWallDate } from "@/lib/jst-date";
import type { SearchableSelectOption } from "@/components/ui/searchable-select";
import type { ReservationTypePickerGroup } from "@/components/shared/ReservationFormModal/ReservationTypePickerDialog";
import type { Reservation } from "@/types";
import {
  getApplicableUnavailableTimes,
  isStartTimeUnavailable,
  slotTimeToSelectValue,
  TIME_OPTIONS,
} from "./reservation-time-utils";
import {
  filterStaffCandidatesByCapability,
  resolveStaffSelectionEligibility,
} from "./filter-staff-candidates";
import { ReservationDateTimeFields } from "./ReservationDateTimeFields";
import { ReservationTypeAndStaffFields } from "./ReservationTypeAndStaffFields";
import { ReservationNotesField } from "./ReservationNotesField";

export interface StaffSelectionState {
  isConfirmedOrphan: boolean;
  reasonMessage: string | null;
}

interface ReservationFormFieldsProps {
  formData: Partial<Reservation>;
  onChange: (data: Partial<Reservation>) => void;
  validationErrors?: Record<string, string>;
  onClearError?: (field: string) => void;
  /** 定休日の日付文字列セット (YYYY-MM-DD 形式) — BUG-343 */
  holidayDates?: Set<string>;
  /** カレンダーの月が変わったときに呼ばれるコールバック (YYYY-MM 形式) — BUG-343 */
  onMonthChange?: (yearMonth: string) => void;
  /** Notify modal submit guard when type/date filters confirm an options-orphan doctor. */
  onStaffSelectionStateChange?: (state: StaffSelectionState) => void;
}

export const ReservationFormFields = memo(function ReservationFormFields({
  formData,
  onChange,
  validationErrors,
  onClearError: _onClearError,
  holidayDates,
  onMonthChange,
  onStaffSelectionStateChange,
}: ReservationFormFieldsProps) {
  // BUG-344: 選択日に出勤しているスタッフのみに絞り込む
  const selectedDateStr = formData.start ? format(formData.start, "yyyy-MM-dd") : null;
  const selectedReservationTypeId = formData.type ? String(formData.type) : null;

  // BUG-341/BUG-015: グループ情報付き取得。編集中の無効区分 ID のみ表示用に残す
  const { data: groupedReservationTypes = [] } =
    useGetReservationTypesGrouped(selectedReservationTypeId);

  const handleMonthChange = useCallback(
    (month: Date) => {
      onMonthChange?.(format(month, "yyyy-MM"));
    },
    [onMonthChange],
  );

  const isCalendarDateDisabled = useCallback(
    (date: Date): boolean => {
      if (isBefore(date, startOfDay(toJSTWallDate(new Date())))) return true;
      if (holidayDates) return holidayDates.has(format(date, "yyyy-MM-dd"));
      return false;
    },
    [holidayDates],
  );

  const { data: staffItems, isLoading: isStaffMasterLoading } = useGetMasterItems("staff");
  // useMemo で参照を安定化（staffOptions の deps が毎レンダー新参照を受け取るのを防ぐ）
  const activeStaff = useMemo(() => staffItems.filter((s) => s.status === "active"), [staffItems]);

  const {
    data: onDutyStaffs,
    isError: isOnDutyError,
    isFetching: isOnDutyFetching,
  } = useGetOnDutyStaffs(selectedDateStr);
  const {
    data: reservationStaffs,
    isError: isReservationStaffError,
    isFetching: isReservationStaffFetching,
  } = useGetReservationStaffs();
  const {
    data: availableTimeSlots,
    isError: isAvailableTimesError,
    error: availableTimesError,
  } = useGetReservationAvailableTimes(
    selectedReservationTypeId,
    selectedDateStr,
    formData.doctor || null,
  );
  const isSettingsUnset = isLineReservationSettingsUnsetError(availableTimesError);
  const currentClinicId = getCurrentClinicId();
  const { data: unavailableTimes = [] } = useGetUnavailableTimes(
    currentClinicId,
    selectedReservationTypeId ?? "",
  );
  const applicableUnavailableTimes = useMemo(
    () => getApplicableUnavailableTimes(unavailableTimes, formData.start),
    [unavailableTimes, formData.start],
  );
  const availableTimeSlotMap = useMemo(() => {
    if (availableTimeSlots === undefined) return undefined;
    return new Map(
      availableTimeSlots.map((slot) => [
        slotTimeToSelectValue(slot.start_time),
        slotTimeToSelectValue(slot.end_time),
      ]),
    );
  }, [availableTimeSlots]);
  const startTimeOptions = useMemo(() => {
    let options: string[];
    const hasTypeAndDate =
      selectedReservationTypeId !== null && selectedDateStr !== null;
    if (isSettingsUnset) {
      // Guided manual entry only when LINE settings are unset.
      options = TIME_OPTIONS.filter(
        (time) => !isStartTimeUnavailable(time, applicableUnavailableTimes),
      );
    } else if (availableTimeSlotMap !== undefined && hasTypeAndDate) {
      // Success (including holiday/full → []) uses computed slots only — never invent hours.
      options = [...availableTimeSlotMap.keys()];
    } else if (isAvailableTimesError && hasTypeAndDate) {
      // Transport/internal errors stay errors — do not fall back to full-day TIME_OPTIONS.
      options = [];
    } else if (hasTypeAndDate) {
      // Loading with type+date selected: wait for API; do not invent slots.
      options = [];
    } else {
      options = TIME_OPTIONS.filter(
        (time) => !isStartTimeUnavailable(time, applicableUnavailableTimes),
      );
    }
    // BUG-015: keep the current edit start even when the slot map is empty/missing.
    if (formData.start) {
      const currentStart = format(formData.start, DISPLAY_TIME_FORMAT);
      if (!options.includes(currentStart)) {
        options = [...options, currentStart];
      }
    }
    return options;
  }, [
    availableTimeSlotMap,
    selectedReservationTypeId,
    selectedDateStr,
    applicableUnavailableTimes,
    formData.start,
    isSettingsUnset,
    isAvailableTimesError,
  ]);
  const settingsUnsetGuidance = isSettingsUnset
    ? "LINE予約の空き枠設定が未登録のため、時刻を手動で入力してください"
    : null;
  const availableTimesErrorMessage =
    isAvailableTimesError && !isSettingsUnset ? "空き枠の取得に失敗しました" : null;
  const reservationStaffMap = useMemo(() => {
    if (reservationStaffs === undefined) return undefined;
    return new Map(reservationStaffs.map((staff) => [String(staff.id), staff]));
  }, [reservationStaffs]);
  const staffOptions = useMemo(() => {
    let options = activeStaff;
    if (selectedDateStr !== null && onDutyStaffs !== undefined) {
      // 出勤スタッフの ID セットで絞り込む
      const onDutyIdSet = new Set(onDutyStaffs.map((s) => String(s.id)));
      options = options.filter((s) => onDutyIdSet.has(String(s.id)));
    }
    // TASK-021 Stage B: affirmative capabilities; missing/pending metadata fail-closed
    options = filterStaffCandidatesByCapability(
      options,
      selectedReservationTypeId,
      reservationStaffMap,
    );
    return options;
  }, [selectedDateStr, onDutyStaffs, activeStaff, selectedReservationTypeId, reservationStaffMap]);

  const [typePickerOpen, setTypePickerOpen] = useState(false);

  // 予約区分ピッカー(サブダイアログ)用: 色・所要時間付きでグループ化(参照安定のため memo 化)
  // BUG-015: 無効区分は選択中のものだけ残り、名前に（無効）を付与する
  const reservationTypePickerGroups = useMemo<ReservationTypePickerGroup[]>(
    () =>
      groupedReservationTypes.map((group) => ({
        label: group.label,
        items: group.types.map((t) => ({
          id: String(t.id),
          name: t.is_active ? t.name : `${t.name}（無効）`,
          color: t.color,
          durationMinutes: t.duration_minutes,
        })),
      })),
    [groupedReservationTypes],
  );
  // トリガー表示用: 選択中の予約区分(色ドット + 名前)
  const selectedReservationType = useMemo(() => {
    if (selectedReservationTypeId === null) return null;
    for (const group of groupedReservationTypes) {
      const found = group.types.find((t) => String(t.id) === selectedReservationTypeId);
      if (found) {
        return {
          color: found.color,
          name: found.name,
          isActive: found.is_active,
        };
      }
    }
    return null;
  }, [groupedReservationTypes, selectedReservationTypeId]);
  const staffSelectOptions = useMemo<SearchableSelectOption[]>(
    () => staffOptions.map((s) => ({ value: String(s.id), label: s.name })),
    [staffOptions],
  );
  const staffEmptyMessage =
    selectedReservationTypeId !== null
      ? "この条件で対応可能なスタッフがいません"
      : selectedDateStr !== null
        ? "この日に出勤しているスタッフがいません"
        : "スタッフが登録されていません";

  const staffNameById = useMemo(() => {
    const names = new Map<string, string>();
    for (const staff of staffItems) {
      names.set(String(staff.id), staff.name);
    }
    for (const staff of reservationStaffs ?? []) {
      names.set(String(staff.id), staff.name);
    }
    for (const staff of onDutyStaffs ?? []) {
      names.set(String(staff.id), staff.name);
    }
    return names;
  }, [staffItems, reservationStaffs, onDutyStaffs]);

  const eligibleOptionIds = useMemo(
    () => new Set(staffSelectOptions.map((option) => option.value)),
    [staffSelectOptions],
  );

  const onDutyReady = selectedDateStr === null || onDutyStaffs !== undefined;
  const capabilityReady =
    selectedReservationTypeId === null || reservationStaffs !== undefined;
  const hasQueryError =
    (selectedDateStr !== null && isOnDutyError) ||
    (selectedReservationTypeId !== null && isReservationStaffError);
  // Pending fetch (undefined data / still fetching) must not confirm orphan eligibility.
  const candidatesSettled =
    !isStaffMasterLoading &&
    onDutyReady &&
    capabilityReady &&
    !(selectedDateStr !== null && isOnDutyFetching && onDutyStaffs === undefined) &&
    !(
      selectedReservationTypeId !== null &&
      isReservationStaffFetching &&
      reservationStaffs === undefined
    ) &&
    !hasQueryError;

  const staffEligibility = useMemo(
    () =>
      resolveStaffSelectionEligibility({
        doctorId: formData.doctor ? String(formData.doctor) : "",
        eligibleOptionIds,
        nameById: staffNameById,
        candidatesSettled,
        hasQueryError,
      }),
    [formData.doctor, eligibleOptionIds, staffNameById, candidatesSettled, hasQueryError],
  );

  useEffect(() => {
    onStaffSelectionStateChange?.({
      isConfirmedOrphan: staffEligibility.isConfirmedOrphan,
      reasonMessage: staffEligibility.reasonMessage,
    });
  }, [
    onStaffSelectionStateChange,
    staffEligibility.isConfirmedOrphan,
    staffEligibility.reasonMessage,
  ]);

  return (
    <div className="space-y-4">
      {/* Date + Time Group */}
      <ReservationDateTimeFields
        formData={formData}
        onChange={onChange}
        validationErrors={validationErrors}
        isCalendarDateDisabled={isCalendarDateDisabled}
        handleMonthChange={handleMonthChange}
        startTimeOptions={startTimeOptions}
        availableTimeSlotMap={availableTimeSlotMap}
        settingsUnsetGuidance={settingsUnsetGuidance}
        availableTimesErrorMessage={availableTimesErrorMessage}
      />

      <ReservationTypeAndStaffFields
        formData={formData}
        onChange={onChange}
        validationErrors={validationErrors}
        typePickerOpen={typePickerOpen}
        setTypePickerOpen={setTypePickerOpen}
        reservationTypePickerGroups={reservationTypePickerGroups}
        selectedReservationType={selectedReservationType}
        staffSelectOptions={staffSelectOptions}
        staffEmptyMessage={staffEmptyMessage}
        staffFallbackLabel={staffEligibility.displayLabel}
        staffOrphanReason={staffEligibility.reasonMessage}
      />

      <ReservationNotesField formData={formData} onChange={onChange} />
    </div>
  );
});
