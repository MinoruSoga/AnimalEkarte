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
import type { TrimmingFormData } from "@/types/trimming";
import type { Pet } from "@/types";
import { paths } from "@/config/paths";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import {
  buildCreateTrimmingRequest,
  buildUpdateTrimmingRequest,
  findDefaultTrimmingReservationTypeId,
  formatJSTDate,
  normalizeVisitDate,
} from "./trimming-form-utils";
import { useTrimmingFormValidation } from "./use-trimming-form-validation";

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
  initialStatus: "in_consultation",
  nextScheduleType: "4weeks",
  nextDate: "",
};

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
  // #233: 既存予約（受付経由 or 同日再利用可能なトリミング予約）に紐付く新規作成かどうか。
  // true の場合はカルテ画面からの登録時ステータス選択を無効化し、予約側のステータスを維持する。
  const hasExistingAppointment = Number.isFinite(appointmentIdFromState) || Number.isFinite(reusableAppointmentId);

  // BUG-027: inline field validation errors
  const { fieldErrors, validate } = useTrimmingFormValidation();

  // localOverrides・formData を useActionState の前に宣言: callback 内で formData を参照するため
  const [localOverrides, setLocalOverrides] = useState<Partial<TrimmingFormData>>({});

  const [styleImagePreview, setStyleImagePreview] = useState<string | null>(null);
  const [completedImagePreview, setCompletedImagePreview] = useState<string | null>(null);

  // 編集モード: サーバーデータを全フィールド復元（初回のみ）
  // rerender-use-ref-transient-values: フラグを useState → useRef に変更して setState-in-effect を排除
  // ⚠️ レンダー中比較 (inline-comparison) に書き換えてはならない:
  // localOverrides は下の useActionState action closure (formData 経由) から参照される。
  // render-phase setState はマウント時の再レンダーパスで action closure を更新しないため、
  // マウント後に再レンダーなしでサブミットすると action がデフォルト formData を見る
  // (use-trimming-form.test.ts「受付から既存トリミング記録を開いた保存」で実証済み)。
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
        // eslint-disable-next-line react-hooks/set-state-in-effect -- 上記と同じ理由（初回のみ・effect同期が必須）
        setStyleImagePreview(existingTrimming.styleImage);
      }
      if (existingTrimming.completedImage) {
        setCompletedImagePreview(existingTrimming.completedImage);
      }
    }
  }, [isEdit, existingTrimming]);

  // appointment 由来の初期反映（初回のみ）。上記と同じ理由で effect 同期が必須。
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
      // eslint-disable-next-line react-hooks/set-state-in-effect -- 上記と同じ理由（初回のみ・effect同期が必須）
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
          toast.success("トリミング情報を更新しました");
        } else if ((existingAppointmentHasDetail && existingAppointmentId) || reusableTrimming?.hasDetail) {
          const req = buildUpdateTrimmingRequest(formData);
          await updateMutation.mutateAsync({ id: existingAppointmentId || reusableTrimming?.id || "", req });
          toast.success("トリミング情報を更新しました");
        } else {
          const pet = selectedPets[0];
          if (!pet) return { success: false, timestamp: Date.now() };
          // BUG-027: バリデーション: staff と course は必須（インラインエラー + toast）
          const validation = validate(formData, defaultTrimmingReservationTypeId);
          if (!validation.valid) {
            return { success: false, fieldErrors: validation.errors, timestamp: Date.now() };
          }
          const resolvedReservationTypeId = validation.reservationTypeId;
          // 日時: フォームから選択していない場合は指定日（未指定なら当日）10:00〜11:30
          const fallbackDate = visitDateFromState ?? formatJSTDate(new Date());
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
            // #233: カルテ画面から直接新規作成する場合のみ、登録時ステータスをユーザーが選択できる
            // （デフォルトは in_consultation）。既存予約に紐付く経路（上記分岐）はここに来ない。
            req.status = formData.initialStatus;
            req.reservation_route = "record_shortcut";
          }
          await createMutation.mutateAsync(req);
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
    hasExistingAppointment,
  };
}
