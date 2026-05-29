import { useState, useEffect, useCallback, useMemo, useRef, useActionState } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useGetTrimming } from "../api/get-trimming";
import { useGetTrimmings } from "../api/get-trimmings";
import { useCreateTrimming } from "../api/create-trimming";
import { useUpdateTrimming } from "../api/update-trimming";
import { useDeleteTrimming } from "../api/delete-trimming";
import type { CreateTrimmingRequest, UpdateTrimmingRequest, TrimmingFormData } from "@/types/trimming";
import type { Pet } from "@/types";
import { paths } from "@/config/paths";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";

const defaultFormData: TrimmingFormData = {
  reservationTypeId: "",
  startTime: "",
  endTime: "",
  styleRequest: "",
  styleImage: null,
  bw: "",
  bwUnit: "Kg",
  bt: "",
  usedShampoo: "",
  usedRibbon: "",
  remarks: "",
  completedImage: null,
  courseId: "",
  optionIds: [],
  staffId: "",
  staffName: "",
};
const JST_OFFSET_MS = 9 * 60 * 60 * 1000;

interface TrimmingReservationType {
  id: number;
  category: string;
  is_internal: boolean;
  sort_order: number;
}

interface TrimmingReservationTypeGroup {
  types: TrimmingReservationType[];
}

function optionalNumber(value: string): number | undefined {
  if (value === "") return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function optionalDateTime(value: string): string | undefined {
  return value === "" ? undefined : value;
}

function padDatePart(value: number): string {
  return String(value).padStart(2, "0");
}

function formatJSTDate(date: Date): string {
  const jstDate = new Date(date.getTime() + JST_OFFSET_MS);
  return `${jstDate.getUTCFullYear()}-${padDatePart(jstDate.getUTCMonth() + 1)}-${padDatePart(jstDate.getUTCDate())}`;
}

function normalizeVisitDate(value: unknown): string | undefined {
  if (typeof value !== "string") return undefined;
  return /^\d{4}-\d{2}-\d{2}$/.test(value) ? value : undefined;
}

function findDefaultTrimmingReservationTypeId(
  groups: readonly TrimmingReservationTypeGroup[] | undefined,
): number | undefined {
  return groups
    ?.flatMap((group) => group.types)
    .filter((type) => type.category === "trimming" && !type.is_internal)
    .sort((a, b) => a.sort_order - b.sort_order)[0]?.id;
}

function buildUpdateTrimmingRequest(formData: TrimmingFormData): UpdateTrimmingRequest {
  return {
    start_time: optionalDateTime(formData.startTime),
    end_time: optionalDateTime(formData.endTime),
    staff_id: optionalNumber(formData.staffId),
    course_id: optionalNumber(formData.courseId),
    style_request: formData.styleRequest,
    bw: optionalNumber(formData.bw),
    bw_unit: formData.bwUnit,
    bt: optionalNumber(formData.bt),
    used_shampoo: formData.usedShampoo,
    used_ribbon: formData.usedRibbon,
    remarks: formData.remarks,
    option_ids: (formData.optionIds ?? []).map(Number),
  };
}

function buildCreateTrimmingRequest(
  formData: TrimmingFormData,
  petID: number,
  reservationTypeID: number,
  startTime: string | undefined,
  endTime: string | undefined,
  appointmentID?: number,
): CreateTrimmingRequest {
  return {
    appointment_id: appointmentID,
    reservation_type_id: reservationTypeID,
    start_time: startTime,
    end_time: endTime,
    pet_id: petID,
    staff_id: optionalNumber(formData.staffId),
    course_id: optionalNumber(formData.courseId),
    style_request: formData.styleRequest,
    bw: optionalNumber(formData.bw),
    bw_unit: formData.bwUnit,
    bt: optionalNumber(formData.bt),
    used_shampoo: formData.usedShampoo,
    used_ribbon: formData.usedRibbon,
    remarks: formData.remarks,
    option_ids: (formData.optionIds ?? []).map(Number),
  };
}

export function useTrimmingForm(id?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;
  const appointmentIdFromState = typeof location.state?.appointmentId === "string"
    ? Number(location.state.appointmentId)
    : typeof location.state?.appointmentId === "number"
      ? location.state.appointmentId
      : Number(searchParams.get("appointmentId") ?? NaN);
  const existingAppointmentId = Number.isFinite(appointmentIdFromState)
    ? String(appointmentIdFromState)
    : "";
  const visitDateFromState = normalizeVisitDate(location.state?.visitDate)
    ?? normalizeVisitDate(searchParams.get("visitDate"));

  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  const { data: existingTrimming, isLoading: isTrimmingLoading } = useGetTrimming(id ?? "");
  const { data: existingAppointmentTrimming, isLoading: isAppointmentLoading } = useGetTrimming(
    !isEdit ? existingAppointmentId : "",
  );
  const { data: reservationTypeGroups } = useGetReservationTypesGrouped();
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const existingLookupDate = visitDateFromState ?? formatJSTDate(new Date());
  const lookupPetId = petId ?? selectedPets[0]?.id ?? "";
  const { data: sameDayTrimmings = [], isLoading: isSameDayTrimmingsLoading } = useGetTrimmings({
    startDate: existingLookupDate,
    endDate: existingLookupDate,
    petId: lookupPetId,
    enabled: !isEdit && existingAppointmentId === "" && lookupPetId !== "",
  });
  const createMutation = useCreateTrimming();
  const updateMutation = useUpdateTrimming();
  const deleteMutation = useDeleteTrimming();
  const existingAppointmentHasDetail = existingAppointmentTrimming?.hasDetail ?? false;
  const defaultTrimmingReservationTypeId = findDefaultTrimmingReservationTypeId(reservationTypeGroups);
  const reusableTrimming = sameDayTrimmings.find((trimming) =>
    trimming.status !== "完了" && trimming.status !== "キャンセル"
  );
  const reusableAppointmentId = reusableTrimming?.id ? Number(reusableTrimming.id) : undefined;

  // BUG-027: inline field validation errors
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  // localOverrides・formData を useActionState の前に宣言: callback 内で formData を参照するため
  const [localOverrides, setLocalOverrides] = useState<Partial<TrimmingFormData>>({});

  // --- Draft Persistence (Local Storage) ---
  const draftScope = id
    ? id
    : existingAppointmentId
      ? `appointment-${existingAppointmentId}`
      : petId
        ? `pet-${petId}`
        : "new";
  const DRAFT_KEY = `trimming-draft-${draftScope}`;

  // Load draft on mount
  useEffect(() => {
    const saved = localStorage.getItem(DRAFT_KEY);
    if (saved) {
      try {
        const draft = JSON.parse(saved);
        // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: one-time draft restore on mount
        setLocalOverrides((prev) => ({ ...prev, ...draft }));
        toast.info("未保存の下書きを復元しました", { duration: 2000 });
      } catch {
        // localStorage の下書きが破損している場合は静かにスキップ（復元失敗は非致命的）
        localStorage.removeItem(DRAFT_KEY);
      }
    }
  }, [DRAFT_KEY]);

  // Save draft on changes
  useEffect(() => {
    const draft: Partial<TrimmingFormData> = { ...localOverrides };
    delete draft.styleImage;
    delete draft.completedImage;
    if (Object.keys(draft).length > 0) {
      localStorage.setItem(DRAFT_KEY, JSON.stringify(draft));
    }
  }, [DRAFT_KEY, localOverrides]);

  const [styleImagePreview, setStyleImagePreview] = useState<string | null>(null);
  const [completedImagePreview, setCompletedImagePreview] = useState<string | null>(null);

  // 編集モード: サーバーデータを全フィールド復元（初回のみ）
  // rerender-use-ref-transient-values: フラグを useState → useRef に変更して setState-in-effect を排除
  const serverDataLoadedRef = useRef(false);
  useEffect(() => {
    if (isEdit && existingTrimming && !serverDataLoadedRef.current) {
      serverDataLoadedRef.current = true;
      setLocalOverrides({
        styleRequest: existingTrimming.styleRequest,
        courseId: existingTrimming.courseId ?? "",
        optionIds: existingTrimming.optionIds ?? [],
        bw: existingTrimming.bw ?? "",
        bwUnit: existingTrimming.bwUnit ?? "Kg",
        bt: existingTrimming.bt ?? "",
        usedShampoo: existingTrimming.usedShampoo ?? "",
        usedRibbon: existingTrimming.usedRibbon ?? "",
        remarks: existingTrimming.remarks ?? "",
        staffId: existingTrimming.staffId ?? "",
        staffName: existingTrimming.staff ?? "",
      });
      // 既存画像URLをプレビューとして復元
      if (existingTrimming.styleImage) {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setStyleImagePreview(existingTrimming.styleImage);
      }
      if (existingTrimming.completedImage) {
        setCompletedImagePreview(existingTrimming.completedImage);
      }
    }
  }, [isEdit, existingTrimming]);

  const appointmentDataLoadedRef = useRef(false);
  useEffect(() => {
    if (isEdit || !existingAppointmentTrimming || appointmentDataLoadedRef.current) return;
    appointmentDataLoadedRef.current = true;
    setLocalOverrides((prev) => ({
      ...prev,
      reservationTypeId: existingAppointmentTrimming.reservationTypeId || prev.reservationTypeId || "",
      styleRequest: existingAppointmentTrimming.styleRequest || prev.styleRequest || "",
      courseId: existingAppointmentTrimming.courseId || prev.courseId || "",
      optionIds: (existingAppointmentTrimming.optionIds?.length ?? 0) > 0
        ? existingAppointmentTrimming.optionIds
        : (prev.optionIds ?? []),
      bw: existingAppointmentTrimming.bw || prev.bw || "",
      bwUnit: existingAppointmentTrimming.bwUnit || prev.bwUnit || "Kg",
      bt: existingAppointmentTrimming.bt || prev.bt || "",
      usedShampoo: existingAppointmentTrimming.usedShampoo || prev.usedShampoo || "",
      usedRibbon: existingAppointmentTrimming.usedRibbon || prev.usedRibbon || "",
      remarks: existingAppointmentTrimming.remarks || prev.remarks || "",
      staffId: existingAppointmentTrimming.staffId || prev.staffId || "",
      staffName: existingAppointmentTrimming.staff || prev.staffName || "",
    }));
    if (existingAppointmentTrimming.styleImage) {
      setStyleImagePreview(existingAppointmentTrimming.styleImage);
    }
    if (existingAppointmentTrimming.completedImage) {
      setCompletedImagePreview(existingAppointmentTrimming.completedImage);
    }
  }, [isEdit, existingAppointmentTrimming]);

  // useMemo: formData の参照を安定化して handleSave 等の deps を最小化 (rerender-dependencies)
  const formData = useMemo<TrimmingFormData>(
    () => ({ ...defaultFormData, ...localOverrides }),
    [localOverrides]
  );

  const setFormData = useCallback((next: Partial<TrimmingFormData>) => {
    setLocalOverrides((prev) => ({ ...prev, ...next }));
  }, []);

  /**
   * React 19 useActionState を使用したフォームアクション
   */
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      try {
        if (isEdit && id) {
          const req = buildUpdateTrimmingRequest(formData);
          await updateMutation.mutateAsync({ id, req });
          localStorage.removeItem(DRAFT_KEY);
          toast.success("トリミング情報を更新しました");
        } else if ((existingAppointmentHasDetail && existingAppointmentId) || reusableTrimming?.hasDetail) {
          const req = buildUpdateTrimmingRequest(formData);
          await updateMutation.mutateAsync({ id: existingAppointmentId || reusableTrimming?.id || "", req });
          localStorage.removeItem(DRAFT_KEY);
          toast.success("トリミング情報を更新しました");
        } else {
          const pet = selectedPets[0];
          if (!pet) return { success: false, timestamp: Date.now() };
          // BUG-027: バリデーション: staff と course は必須（インラインエラー + toast）
          const errors: Record<string, string> = {};
          if (!formData.staffId) {
            errors.staffId = "担当者を選択してください";
          }
          if (!formData.courseId) {
            errors.courseId = "コースを選択してください";
          }
          const reservationTypeId = formData.reservationTypeId
            ? Number(formData.reservationTypeId)
            : defaultTrimmingReservationTypeId;
          if (!reservationTypeId || !Number.isFinite(reservationTypeId)) {
            errors.reservationTypeId = "トリミング予約区分が設定されていません";
          }
          if (Object.keys(errors).length > 0) {
            setFieldErrors(errors);
            return { success: false, fieldErrors: errors, timestamp: Date.now() };
          }
          setFieldErrors({});
          const resolvedReservationTypeId = Number(reservationTypeId);
          // 日時: フォームから選択していない場合は指定日（未指定なら当日）10:00〜11:30
          const fallbackDate = visitDateFromState ?? formatJSTDate(new Date());
          const hasExistingAppointment = Number.isFinite(appointmentIdFromState) || Number.isFinite(reusableAppointmentId);
          const startDate = formData.startTime || (hasExistingAppointment ? undefined : `${fallbackDate}T10:00:00+09:00`);
          const endDate = formData.endTime || (hasExistingAppointment ? undefined : `${fallbackDate}T11:30:00+09:00`);
          const req = buildCreateTrimmingRequest(
            formData,
            Number(pet.id),
            resolvedReservationTypeId,
            startDate,
            endDate,
            Number.isFinite(appointmentIdFromState) ? appointmentIdFromState : reusableAppointmentId,
          );
          if (!hasExistingAppointment) {
            req.status = "in_consultation";
            req.reservation_route = "record_shortcut";
          }
          await createMutation.mutateAsync(req);
          localStorage.removeItem(DRAFT_KEY);
          toast.success("トリミング情報を登録しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  // Edit mode: restore pet from fetched trimming data
  useEffect(() => {
    if (isEdit && existingTrimming?.petId) {
      setSelectedPets([
        {
          id: existingTrimming.petId,
          ownerId: existingTrimming.ownerId ?? "",
          name: existingTrimming.petName,
          petNumber: existingTrimming.petNumber,
          ownerName: existingTrimming.ownerName,
          species: existingTrimming.species,
          weight: existingTrimming.weight,
        } as Pet,
      ]);
    }
  }, [isEdit, existingTrimming, setSelectedPets]);

  // New mode: populate pet selection from petId query param
  useEffect(() => {
    if (!isEdit) {
      if (petFromQuery) {
        setSelectedPets([petFromQuery]);
      } else if (!petId && !isPetLoading) {
        navigate(paths.trimming.selectPet.getHref());
      }
    }
  }, [isEdit, petId, petFromQuery, isPetLoading, setSelectedPets, navigate]);

  const handleStyleImageChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setLocalOverrides((prev) => ({ ...prev, styleImage: file }));
      const reader = new FileReader();
      reader.onloadend = () => setStyleImagePreview(reader.result as string);
      reader.readAsDataURL(file);
    }
  }, []);

  const handleCompletedImageChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setLocalOverrides((prev) => ({ ...prev, completedImage: file }));
      const reader = new FileReader();
      reader.onloadend = () => setCompletedImagePreview(reader.result as string);
      reader.readAsDataURL(file);
    }
  }, []);

  const removeStyleImage = useCallback(() => {
    setLocalOverrides((prev) => ({ ...prev, styleImage: null }));
    setStyleImagePreview(null);
  }, []);

  const removeCompletedImage = useCallback(() => {
    setLocalOverrides((prev) => ({ ...prev, completedImage: null }));
    setCompletedImagePreview(null);
  }, []);

  const handleDelete = useCallback((onSuccess?: () => void) => {
    if (!isEdit || !id) return;
    deleteMutation.mutate(id, {
      onSuccess: () => {
        toast.success("トリミング情報を削除しました");
        onSuccess?.();
      },
      onError: (error) => {
        handleApiError(error, "削除");
      },
    });
  }, [isEdit, id, deleteMutation]);

  const isSaving = isPending;
  const isDeleting = deleteMutation.isPending;
  const mode = isEdit ? ("edit" as const) : ("new" as const);

  const isLoading = isEdit ? isTrimmingLoading : isPetLoading || isAppointmentLoading || isSameDayTrimmingsLoading;
  const notFound = isEdit && !isTrimmingLoading && !existingTrimming && !!id;

  return {
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
    formAction,
    formState,
    handleDelete,
    isSaving,
    isDeleting,
    fieldErrors,
    isLoading,
    notFound,
  };
}
