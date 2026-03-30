import { useState, useEffect, useCallback, useMemo, useActionState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { addWeeks, addYears, format } from "date-fns";
import { toast } from "sonner";
import { paths } from "@/config/paths";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { useGetVaccination } from "../api/get-vaccination";
import { useCreateVaccination } from "../api/create-vaccination";
import { useUpdateVaccination } from "../api/update-vaccination";
import { useDeleteVaccination } from "../api/delete-vaccination";
import type { CreateVaccinationRequest, UpdateVaccinationRequest } from "../api/types";

interface VaccinationFormState {
  vaccineId: string;
  date: string;
  supplemental: string;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  nextScheduleType: string;
  nextDate: string;
  remarks: string;
}

const DEFAULT_NEXT_SCHEDULE_TYPE = "1year" as const;

// BUG-026: vaccine interval mapping (vaccine_id → schedule type)
const VACCINE_SCHEDULE_MAP: Record<string, string> = {
  "1": "1year", // 混合ワクチン（年1回）
  "2": "1year", // 狂犬病ワクチン（年1回）
};

const DEFAULT_FORM: VaccinationFormState = {
  vaccineId: "",
  date: "",
  supplemental: "",
  lot1: "",
  lot2: "",
  lot3: "",
  lot4: "",
  nextScheduleType: DEFAULT_NEXT_SCHEDULE_TYPE,
  nextDate: "",
  remarks: "",
};

// BUG-026: calculate next date based on vaccination date and schedule type
function calculateNextDate(vaccinationDate: string, scheduleType: string): string {
  if (!vaccinationDate || scheduleType === "other") return "";
  const date = new Date(vaccinationDate);
  if (isNaN(date.getTime())) return "";
  switch (scheduleType) {
    case "3weeks":
      return format(addWeeks(date, 3), "yyyy-MM-dd");
    case "4weeks":
      return format(addWeeks(date, 4), "yyyy-MM-dd");
    case "1year":
      return format(addYears(date, 1), "yyyy-MM-dd");
    default:
      return "";
  }
}

export function useVaccinationForm(id?: string) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  // Pet Selection
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  // API hooks
  const { data: existingVaccination } = useGetVaccination(id ?? "");
  // 編集時: レコードに紐づくペットIDを解決するため、existingVaccination.petId から取得
  const editPetId = isEdit ? (existingVaccination?.petId ?? "") : "";
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const { data: petFromEdit } = useGetPet(editPetId);
  const createMutation = useCreateVaccination();
  const updateMutation = useUpdateVaccination();
  const deleteMutation = useDeleteVaccination();

  // BUG-024/074: validation errors
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  // Local overrides: tracks user edits on top of server data
  const [localOverrides, setLocalOverrides] = useState<Partial<VaccinationFormState>>({});

  // Merge: server data as base + user edits on top
  const formData: VaccinationFormState = isEdit && existingVaccination
    ? {
        vaccineId: existingVaccination.vaccineId,
        date: existingVaccination.date ? existingVaccination.date.slice(0, 10) : "",
        supplemental: "",
        lot1: "",
        lot2: "",
        lot3: "",
        lot4: "",
        nextScheduleType: "4weeks",
        nextDate: existingVaccination.nextDate ? existingVaccination.nextDate.slice(0, 10) : "",
        remarks: "",
        ...localOverrides,
      }
    : { ...DEFAULT_FORM, ...localOverrides };

  const setField = useCallback(<K extends keyof VaccinationFormState>(key: K, value: VaccinationFormState[K]) => {
    setLocalOverrides((prev) => ({ ...prev, [key]: value }));
  }, []);

  interface FormState {
    success: boolean;
    timestamp: number;
  }

  /**
   * React 19 useActionState を使用したフォームアクション
   */
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      // BUG-024/074: バリデーション
      const errors: Record<string, string> = {};
      if (!isEdit) {
        if (!formData.vaccineId || formData.vaccineId === "0") {
          errors.vaccineId = "ワクチン種別を選択してください";
        }
        if (!formData.date) {
          errors.date = "接種日を入力してください";
        }
      } else {
        if (!formData.date) {
          errors.date = "接種日を入力してください";
        }
      }
      if (Object.keys(errors).length > 0) {
        setFieldErrors(errors);
        return { success: false, timestamp: Date.now() };
      }
      setFieldErrors({});

      try {
        if (isEdit && id) {
          const toRFC3339 = (d: string) => d ? `${d}T00:00:00Z` : undefined;
          const req: UpdateVaccinationRequest = {
            date: toRFC3339(formData.date),
            next_date: formData.nextDate ? `${formData.nextDate}T00:00:00Z` : null,
            lot1: formData.lot1 || undefined,
            remarks: formData.remarks || undefined,
          };
          await updateMutation.mutateAsync({ id, req });
          toast.success("予防接種情報を更新しました");
        } else {
          const pet = selectedPets[0];
          if (!pet) return { success: false, timestamp: Date.now() };
          const req: CreateVaccinationRequest = {
            medical_record_id: null,
            pet_id: Number(pet.id),
            vaccine_id: Number(formData.vaccineId),
            date: formData.date ? `${formData.date}T00:00:00Z` : new Date().toISOString(),
            next_date: formData.nextDate ? `${formData.nextDate}T00:00:00Z` : null,
            lot1: formData.lot1 || undefined,
            remarks: formData.remarks || undefined,
          };
          await createMutation.mutateAsync(req);
          toast.success("予防接種を登録しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch {
        // BUG-024/074: API error toast
        toast.error("保存に失敗しました");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
  );

  // History Filter State
  const [filterStartDate, setFilterStartDate] = useState("");
  const [filterEndDate, setFilterEndDate] = useState("");
  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");

  // New mode: populate pet selection from petId query param
  useEffect(() => {
    if (!isEdit) {
      if (petFromQuery) {
        setSelectedPets([petFromQuery]);
      } else if (!petId && !isPetLoading) {
        navigate(paths.vaccinations.selectPet.getHref());
      }
    }
  }, [isEdit, petId, petFromQuery, isPetLoading, setSelectedPets, navigate]);

  // Edit mode: populate pet selection from the vaccination record's pet_id
  useEffect(() => {
    if (isEdit && petFromEdit) {
      setSelectedPets([petFromEdit]);
    }
  }, [isEdit, petFromEdit, setSelectedPets]);

  const handleClearHistoryFilter = () => {
    setHistorySearchTerm("");
  };

  const isSaving = isPending;

  // BUG-026: when vaccine type changes, auto-update nextScheduleType and recalculate nextDate
  const setVaccineId = useCallback((v: string) => {
    setLocalOverrides((prev) => {
      const scheduleType = VACCINE_SCHEDULE_MAP[v] ?? DEFAULT_NEXT_SCHEDULE_TYPE;
      const currentDate = prev.date ?? "";
      const calculated = calculateNextDate(currentDate, scheduleType);
      return {
        ...prev,
        vaccineId: v,
        nextScheduleType: scheduleType,
        ...(calculated ? { nextDate: calculated } : {}),
      };
    });
  }, []);

  // BUG-026: auto-calculate nextDate when date changes
  const setDate = useCallback((v: string) => {
    setLocalOverrides((prev) => {
      const scheduleType = prev.nextScheduleType ?? DEFAULT_NEXT_SCHEDULE_TYPE;
      const calculated = calculateNextDate(v, scheduleType);
      return { ...prev, date: v, ...(calculated ? { nextDate: calculated } : {}) };
    });
  }, []);

  const setSupplemental = useCallback((v: string) => setField("supplemental", v), [setField]);
  const setLot1 = useCallback((v: string) => setField("lot1", v), [setField]);
  const setLot2 = useCallback((v: string) => setField("lot2", v), [setField]);
  const setLot3 = useCallback((v: string) => setField("lot3", v), [setField]);
  const setLot4 = useCallback((v: string) => setField("lot4", v), [setField]);

  // BUG-026: auto-calculate nextDate when schedule type changes
  const setNextScheduleType = useCallback((v: string) => {
    setLocalOverrides((prev) => {
      const currentDate = prev.date ?? "";
      const calculated = calculateNextDate(currentDate, v);
      return { ...prev, nextScheduleType: v, ...(calculated ? { nextDate: calculated } : {}) };
    });
  }, []);

  const setNextDate = useCallback((v: string) => setField("nextDate", v), [setField]);
  const setRemarks = useCallback((v: string) => setField("remarks", v), [setField]);

  // BUG-025: delete handler
  const handleDelete = useCallback((onSuccess?: () => void) => {
    if (!isEdit || !id) return;
    deleteMutation.mutate(id, {
      onSuccess: () => {
        toast.success("予防接種情報を削除しました");
        onSuccess?.();
      },
      onError: () => {
        toast.error("削除に失敗しました");
      },
    });
  }, [isEdit, id, deleteMutation]);

  const isDeleting = deleteMutation.isPending;

  const form = useMemo(() => ({
    vaccineId: formData.vaccineId,
    setVaccineId,
    date: formData.date,
    setDate,
    supplemental: formData.supplemental,
    setSupplemental,
    lot1: formData.lot1,
    setLot1,
    lot2: formData.lot2,
    setLot2,
    lot3: formData.lot3,
    setLot3,
    lot4: formData.lot4,
    setLot4,
    nextScheduleType: formData.nextScheduleType,
    setNextScheduleType,
    nextDate: formData.nextDate,
    setNextDate,
    remarks: formData.remarks,
    setRemarks,
  }), [
    formData.vaccineId, setVaccineId,
    formData.date, setDate,
    formData.supplemental, setSupplemental,
    formData.lot1, setLot1,
    formData.lot2, setLot2,
    formData.lot3, setLot3,
    formData.lot4, setLot4,
    formData.nextScheduleType, setNextScheduleType,
    formData.nextDate, setNextDate,
    formData.remarks, setRemarks,
  ]);

  return {
    isEdit,
    petSelection,
    form,
    historyFilter: {
      filterStartDate, setFilterStartDate,
      filterEndDate, setFilterEndDate,
      historySearchTerm, setHistorySearchTerm,
      sortOrder, setSortOrder,
      handleClearHistoryFilter,
    },
    formAction,
    formState,
    isSaving,
    fieldErrors,
    handleDelete,
    isDeleting,
  };
}
