import { useState, useEffect, useTransition, useCallback, useMemo, useRef, useActionState } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { useGetTrimming } from "../api/get-trimming";
import { useCreateTrimming } from "../api/create-trimming";
import { useUpdateTrimming } from "../api/update-trimming";
import { useDeleteTrimming } from "../api/delete-trimming";
import type { CreateTrimmingRequest, UpdateTrimmingRequest, TrimmingFormData } from "@/types/trimming";
import type { Pet } from "@/types";
import { paths } from "@/config/paths";

const defaultFormData: TrimmingFormData = {
  styleRequest: "",
  memo: "",
  eggs: "",
  parts: {
    nail: false,
    analGland: false,
    eye: false,
    ear: false,
    skin: false,
    oral: false,
  },
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

export function useTrimmingForm(id?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  const { data: existingTrimming } = useGetTrimming(id ?? "");
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const createMutation = useCreateTrimming();
  const updateMutation = useUpdateTrimming();
  const deleteMutation = useDeleteTrimming();

  // useTransition: save/delete の pending 管理 (rerender-transitions)
  const [isDeleteTransitionPending, startDeleteTransition] = useTransition();

  // BUG-027: inline field validation errors
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  interface FormState {
    success: boolean;
    timestamp: number;
  }

  /**
   * React 19 useActionState を使用したフォームアクション
   */
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      try {
        if (isEdit && id) {
          const req: UpdateTrimmingRequest = {
            style_request: formData.styleRequest || undefined,
            bw: formData.bw ? Number(formData.bw) : undefined,
            bw_unit: formData.bwUnit || undefined,
            bt: formData.bt ? Number(formData.bt) : undefined,
            used_shampoo: formData.usedShampoo || undefined,
            used_ribbon: formData.usedRibbon || undefined,
            remarks: formData.remarks || undefined,
            option_ids: formData.optionIds.length > 0 ? formData.optionIds.map(Number) : undefined,
          };
          await updateMutation.mutateAsync({ id, req });
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
          if (Object.keys(errors).length > 0) {
            setFieldErrors(errors);
            return { success: false, timestamp: Date.now() };
          }
          setFieldErrors({});
          const req: CreateTrimmingRequest = {
            pet_id: Number(pet.id),
            staff_id: Number(formData.staffId),
            course_id: Number(formData.courseId),
            date: new Date().toISOString(),
            style_request: formData.styleRequest || undefined,
            remarks: formData.remarks || undefined,
          };
          await createMutation.mutateAsync(req);
          toast.success("トリミング情報を登録しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        toast.error("保存に失敗しました");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
  );

  const [localOverrides, setLocalOverrides] = useState<Partial<TrimmingFormData>>({});

  // 編集モード: サーバーデータを全フィールド復元（初回のみ）
  // rerender-use-ref-transient-values: フラグを useState → useRef に変更して setState-in-effect を排除
  const serverDataLoadedRef = useRef(false);
  useEffect(() => {
    if (isEdit && existingTrimming && !serverDataLoadedRef.current) {
      serverDataLoadedRef.current = true;
      // eslint-disable-next-line react-hooks/set-state-in-effect -- 非同期サーバーデータでフォームを初期化するパターン（初回1回のみ）
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
    }
  }, [isEdit, existingTrimming]);

  // useMemo: formData の参照を安定化して handleSave 等の deps を最小化 (rerender-dependencies)
  const formData = useMemo<TrimmingFormData>(
    () => ({ ...defaultFormData, ...localOverrides }),
    [localOverrides]
  );

  const setFormData = useCallback((next: Partial<TrimmingFormData>) => {
    setLocalOverrides((prev) => ({ ...prev, ...next }));
  }, []);

  const [styleImagePreview, setStyleImagePreview] = useState<string | null>(null);
  const [completedImagePreview, setCompletedImagePreview] = useState<string | null>(null);

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
    startDeleteTransition(() => {
      deleteMutation.mutate(id, {
        onSuccess: () => {
          toast.success("トリミング情報を削除しました");
          onSuccess?.();
        },
      });
    });
  }, [isEdit, id, deleteMutation, startDeleteTransition]);

  const isSaving = isPending;
  const isDeleting = deleteMutation.isPending || isDeleteTransitionPending;
  const mode = isEdit ? ("edit" as const) : ("new" as const);

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
  };
}
