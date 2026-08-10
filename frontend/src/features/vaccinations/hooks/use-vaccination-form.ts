import { useState, useEffect, useLayoutEffect, useCallback, useMemo, useActionState, useRef } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { addWeeks, addYears, format } from "date-fns";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { jstDateStartISOString, todayJSTISO } from "@/lib/jst-date";
import { paths } from "@/config/paths";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { useGetAllVaccinesMaster } from "@/hooks/use-treatment-master";
import {
  isNonDisclosureReadStatus,
  resolveEntityReadResult,
  type EntityReadResult,
} from "@/lib/entity-read-result";
import { useGetVaccination } from "../api/get-vaccination";
import { useCreateVaccination } from "../api/create-vaccination";
import { useUpdateVaccination } from "../api/update-vaccination";
import { useDeleteVaccination } from "../api/delete-vaccination";
import type { CreateVaccinationRequest, UpdateVaccinationRequest } from "../api/types";
import type { VaccinationRecord } from "@/types";

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

interface VaccinationMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

const DENIED_MUTATION_PERMISSIONS: Readonly<VaccinationMutationPermissions> = {
  canCreate: false,
  canEdit: false,
  canDelete: false,
};

const DEFAULT_NEXT_SCHEDULE_TYPE = "1year" as const;

// react-reviewer指摘: インラインの `?? []` は毎レンダー新規参照になり vaccineOptions の
// useMemo を不必要に再計算させる（取得未完了/エラー時に顕著）。安定参照にする。
// NonNullable: react-query の data は T[] | undefined。undefined を残したままだと
// 分割代入デフォルト値(= vaccinesMaster = EMPTY_VACCINES_MASTER)の型も undefined を
// 含んだままになり、既定値を与えた意味が消えて possibly-undefined エラーになる。
const EMPTY_VACCINES_MASTER: NonNullable<ReturnType<typeof useGetAllVaccinesMaster>["data"]> = [];

// BUG-401/BUG-026: vaccine interval (vaccines master 実データ) → schedule type。
// 旧実装はハードコードの vaccine_id "1"/"2" をキーにしていたため、実マスタ ID（例: "14"）に
// 切り替えると必ずフォールスルーして誤った次回予定を計算していた（サイレント mis-scheduling）。
// マスタの vaccine.interval（"1年"/"1ヶ月"）から導出することで、実 ID に依存せず正しく動作する。
function scheduleTypeForInterval(interval: string | undefined): string {
  switch (interval) {
    case "1年":
      return "1year";
    case "1ヶ月":
      return "4weeks";
    default:
      return DEFAULT_NEXT_SCHEDULE_TYPE;
  }
}

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
export function calculateNextDate(vaccinationDate: string, scheduleType: string): string {
  if (!vaccinationDate || scheduleType === "other") return "";
  const date = new Date(vaccinationDate + "T00:00:00");
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

export function useVaccinationForm(
  id?: string,
  permissions: Readonly<VaccinationMutationPermissions> = DENIED_MUTATION_PERMISSIONS,
) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  // Pet Selection
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  // BUG-401: ワクチンマスタを実際にクエリする（姉妹フォーム MedicalRecordVaccination.tsx と同じ
  // useGetAllVaccinesMaster パターン）。ハードコードの2択は別クリニックのマスタ行（id=1/2 は
  // カテゴリ用の "ワクチン犬"/"ワクチン猫" スタブ）と衝突し、保存した vaccine_id が選択ラベルと
  // 一致しないデータ破損を起こしていた。species フィルタは姉妹フォームも持たないため本修正では
  // 追加しない（BUG-408 に残置）。
  const { data: vaccinesMaster = EMPTY_VACCINES_MASTER } = useGetAllVaccinesMaster();
  const vaccineOptions = useMemo(
    () => vaccinesMaster.flatMap((v) => (v.isActive ? [{ value: v.id, label: v.name }] : [])),
    [vaccinesMaster],
  );

  // API hooks — BUG-016: classify read failures; never fold into blank edit model
  const {
    data: vaccinationData,
    isLoading: isVaccinationLoading,
    isError: isVaccinationError,
    error: vaccinationError,
    refetch: refetchVaccination,
  } = useGetVaccination(id ?? "");
  const entityRead: EntityReadResult<VaccinationRecord> = resolveEntityReadResult({
    id: isEdit ? id : undefined,
    data: vaccinationData,
    isLoading: isVaccinationLoading,
    isError: isVaccinationError,
    error: vaccinationError,
    refetch: refetchVaccination,
  });
  const existingVaccination =
    entityRead.status === "found" ? entityRead.data : undefined;
  const entityReadRef = useRef(entityRead);
  useLayoutEffect(() => {
    entityReadRef.current = entityRead;
  }, [entityRead]);
  // 編集時: レコードに紐づくペットIDを解決するため、existingVaccination.petId から取得
  const editPetId = isEdit ? (existingVaccination?.petId ?? "") : "";
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const { data: petFromEdit } = useGetPet(editPetId);
  const selectedPetRef = useRef(selectedPets[0]);
  const queryPetRef = useRef(petFromQuery);
  const editPetRef = useRef(petFromEdit);
  useLayoutEffect(() => {
    selectedPetRef.current = selectedPets[0];
    queryPetRef.current = petFromQuery;
  }, [petFromQuery, selectedPets]);
  useLayoutEffect(() => {
    editPetRef.current = petFromEdit;
  }, [petFromEdit]);
  const createMutation = useCreateVaccination();
  const updateMutation = useUpdateVaccination();
  const deleteMutation = useDeleteVaccination();
  const { canCreate, canEdit, canDelete } = permissions;
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete };
  }, [canCreate, canDelete, canEdit]);
  const isMutationAllowed = useCallback(
    (action: keyof VaccinationMutationPermissions) => permissionsRef.current[action] === true,
    [],
  );

  // BUG-024/074: validation errors
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  // Local overrides: tracks user edits on top of server data
  const [localOverrides, setLocalOverrides] = useState<Partial<VaccinationFormState>>({});

  // Merge: server data as base + user edits on top
  // BUG-004: 新規オープン時は接種日を JST 当日で初期表示（編集は既存 date を保持）。
  // todayJSTISO() は merge 時に評価し、モジュール定数に焼き込まない（日跨ぎ対策）。
  // localOverrides.date があれば上書き可能（手動変更 DoD）。
  const formData: VaccinationFormState = isEdit && existingVaccination
    ? {
        vaccineId: existingVaccination.vaccineId,
        date: existingVaccination.date ? existingVaccination.date.slice(0, 10) : "",
        supplemental: existingVaccination.supplemental ?? "",
        lot1: existingVaccination.lot1 ?? "",
        lot2: existingVaccination.lot2 ?? "",
        lot3: existingVaccination.lot3 ?? "",
        lot4: existingVaccination.lot4 ?? "",
        nextScheduleType: existingVaccination.nextScheduleType ?? DEFAULT_NEXT_SCHEDULE_TYPE,
        nextDate: existingVaccination.nextDate ? existingVaccination.nextDate.slice(0, 10) : "",
        remarks: existingVaccination.remarks ?? "",
        ...localOverrides,
      }
    : { ...DEFAULT_FORM, date: todayJSTISO(), ...localOverrides };

  // localOverrides は部分更新のみ。編集時に未タッチの date/type は server base 側にあるため、
  // setDate / setNextScheduleType / setNextDate は formData 合成結果を ref 経由で読む（BUG-005/026）。
  const formDataRef = useRef(formData);
  useLayoutEffect(() => {
    formDataRef.current = formData;
  }, [formData]);

  const setField = useCallback(<K extends keyof VaccinationFormState>(key: K, value: VaccinationFormState[K]) => {
    setLocalOverrides((prev) => ({ ...prev, [key]: value }));
  }, []);

  interface FormState {
    success: boolean;
    timestamp: number;
    fieldErrors?: Record<string, string>;
  }

  /**
   * React 19 useActionState を使用したフォームアクション
   */
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      // BUG-024/074: バリデーション
      const errors: Record<string, string> = {};
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      if (!isEdit) {
        if (!formData.vaccineId || formData.vaccineId === "0") {
          errors.vaccineId = "ワクチン種別を選択してください";
        }
        if (!formData.date) {
          errors.date = "接種日を入力してください";
        } else if (new Date(formData.date + "T00:00:00") > today) {
          // BUG-024: 実施日は今日以前であること
          errors.date = "接種日は今日以前の日付を入力してください";
        }
      } else {
        if (!formData.date) {
          errors.date = "接種日を入力してください";
        } else if (new Date(formData.date + "T00:00:00") > today) {
          // BUG-024: 実施日は今日以前であること
          errors.date = "接種日は今日以前の日付を入力してください";
        }
      }
      // BUG-096: 新規登録時、次回予定日は本日以降
      if (!isEdit && formData.nextDate) {
        if (new Date(formData.nextDate + "T00:00:00") < today) {
          errors.nextDate = "次回予定日は本日以降の日付を入力してください";
        }
      }
      // BUG-024: 次回予定日は実施日より後であること（新規・編集共通）
      if (formData.date && formData.nextDate) {
        const dateVal = new Date(formData.date + "T00:00:00");
        const nextDateVal = new Date(formData.nextDate + "T00:00:00");
        if (!isNaN(dateVal.getTime()) && !isNaN(nextDateVal.getTime()) && nextDateVal <= dateVal) {
          errors.nextDate = "次回予定日は接種日より後の日付を入力してください";
        }
      }
      if (Object.keys(errors).length > 0) {
        setFieldErrors(errors);
        return { success: false, timestamp: Date.now() };
      }
      setFieldErrors({});

      try {
        if (isEdit && id) {
          // BUG-016: edit route without a found entity must not create/update
          if (entityReadRef.current.status !== "found") {
            return { success: false, timestamp: Date.now() };
          }
          const toRFC3339 = (d: string) => d ? jstDateStartISOString(d) : undefined;
          const req: UpdateVaccinationRequest = {
            date: toRFC3339(formData.date),
            next_date: formData.nextDate ? jstDateStartISOString(formData.nextDate) : null,
            lot1: formData.lot1 || undefined,
            lot2: formData.lot2 || undefined,
            lot3: formData.lot3 || undefined,
            lot4: formData.lot4 || undefined,
            remarks: formData.remarks || undefined,
            supplemental: formData.supplemental || undefined,
            next_schedule_type: formData.nextScheduleType || undefined,
          };
          if (
            !isMutationAllowed("canEdit") ||
            editPetRef.current?.status === "死亡"
          ) {
            return { success: false, timestamp: Date.now() };
          }
          await updateMutation.mutateAsync({ id, req });
          toast.success("予防接種情報を更新しました");
        } else {
          const pet = petId ? queryPetRef.current : selectedPetRef.current;
          if (!pet) return { success: false, timestamp: Date.now() };
          const req: CreateVaccinationRequest = {
            medical_record_id: null,
            pet_id: Number(pet.id),
            vaccine_id: Number(formData.vaccineId),
            date: jstDateStartISOString(formData.date || todayJSTISO()),
            next_date: formData.nextDate ? jstDateStartISOString(formData.nextDate) : null,
            lot1: formData.lot1 || undefined,
            lot2: formData.lot2 || undefined,
            lot3: formData.lot3 || undefined,
            lot4: formData.lot4 || undefined,
            remarks: formData.remarks || undefined,
            supplemental: formData.supplemental || undefined,
            next_schedule_type: formData.nextScheduleType || undefined,
          };
          if (
            !isMutationAllowed("canCreate") ||
            pet.status === "死亡"
          ) {
            return { success: false, timestamp: Date.now() };
          }
          await createMutation.mutateAsync(req);
          toast.success("予防接種を登録しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
  );

  // History Filter State
  const [filterStartDate, setFilterStartDate] = useState("");
  const [filterEndDate, setFilterEndDate] = useState("");
  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

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

  // BUG-026/BUG-401: when vaccine type changes, auto-update nextScheduleType and recalculate
  // nextDate — derived from the selected master vaccine's own `interval` field (see
  // scheduleTypeForInterval), not a hardcoded id table.
  const setVaccineId = useCallback((v: string) => {
    setLocalOverrides((prev) => {
      const selected = vaccinesMaster.find((vac) => vac.id === v);
      const scheduleType = scheduleTypeForInterval(selected?.interval);
      const currentDate = prev.date ?? formDataRef.current.date;
      const calculated = calculateNextDate(currentDate, scheduleType);
      return {
        ...prev,
        vaccineId: v,
        nextScheduleType: scheduleType,
        ...(calculated ? { nextDate: calculated } : {}),
      };
    });
  }, [vaccinesMaster]);

  // BUG-026: auto-calculate nextDate when date changes（other のときは手入力 nextDate を維持）
  const setDate = useCallback((v: string) => {
    setLocalOverrides((prev) => {
      const scheduleType =
        prev.nextScheduleType ?? formDataRef.current.nextScheduleType ?? DEFAULT_NEXT_SCHEDULE_TYPE;
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
      const currentDate = prev.date ?? formDataRef.current.date;
      const calculated = calculateNextDate(currentDate, v);
      return { ...prev, nextScheduleType: v, ...(calculated ? { nextDate: calculated } : {}) };
    });
  }, []);

  // BUG-005: 次回予定日の手動上書き時、標準間隔の計算結果と一致しなければ type を other へ切替。
  // type=1year のまま next_date だけずらす矛盾永続化を防ぐ。一致する場合は標準 type を維持。
  const setNextDate = useCallback((v: string) => {
    setLocalOverrides((prev) => {
      const base = formDataRef.current;
      const vaccinationDate = prev.date ?? base.date;
      const currentType =
        prev.nextScheduleType ?? base.nextScheduleType ?? DEFAULT_NEXT_SCHEDULE_TYPE;
      if (currentType !== "other" && vaccinationDate && v) {
        const calculated = calculateNextDate(vaccinationDate, currentType);
        if (calculated && calculated === v) {
          return { ...prev, nextDate: v };
        }
      }
      return { ...prev, nextDate: v, nextScheduleType: "other" };
    });
  }, []);
  const setRemarks = useCallback((v: string) => setField("remarks", v), [setField]);

  // BUG-025: delete handler
  const { mutate: deleteVaccinationFn } = deleteMutation;
  const handleDelete = useCallback((onSuccess?: () => void) => {
    if (!isEdit || !id) return;
    if (
      !isMutationAllowed("canDelete") ||
      editPetRef.current?.status === "死亡"
    ) {
      return;
    }
    deleteVaccinationFn(id, {
      onSuccess: () => {
        toast.success("予防接種情報を削除しました");
        onSuccess?.();
      },
      onError: (error) => {
        handleApiError(error, "削除");
      },
    });
  }, [isEdit, id, isMutationAllowed, deleteVaccinationFn]);

  const isDeleting = deleteMutation.isPending;

  const form = useMemo(() => ({
    doctorName: existingVaccination?.doctor ?? "",
    vaccineId: formData.vaccineId,
    setVaccineId,
    vaccineOptions,
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
    existingVaccination?.doctor,
    formData.vaccineId, setVaccineId, vaccineOptions,
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
    entityRead,
    isReadLoading: entityRead.status === "loading",
    isReadNotFound: isNonDisclosureReadStatus(entityRead.status),
    isReadError: entityRead.status === "error",
    retryRead: entityRead.status === "error" ? entityRead.retry : undefined,
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
