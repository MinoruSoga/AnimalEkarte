import { useState, useEffect, useTransition, useCallback, useMemo, useRef, useActionState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { useGetTrimming } from "../api/get-trimming";
import { useCreateTrimming } from "../api/create-trimming";
import { useUpdateTrimming } from "../api/update-trimming";
import { useDeleteTrimming } from "../api/delete-trimming";
import type { CreateTrimmingRequest, UpdateTrimmingRequest, TrimmingFormData } from "@/types/trimming";
import type { Pet } from "@/types";
import { paths } from "@/config/paths";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";

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
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  const { data: existingTrimming, isLoading: isTrimmingLoading } = useGetTrimming(id ?? "");
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const createMutation = useCreateTrimming();
  const updateMutation = useUpdateTrimming();
  const deleteMutation = useDeleteTrimming();

  // useTransition: save/delete の pending 管理 (rerender-transitions)
  const [isDeleteTransitionPending, startDeleteTransition] = useTransition();

  // BUG-027: inline field validation errors
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  // localOverrides・formData を useActionState の前に宣言: callback 内で formData を参照するため
  const [localOverrides, setLocalOverrides] = useState<Partial<TrimmingFormData>>({});

  // --- Draft Persistence (Local Storage) ---
  const DRAFT_KEY = `trimming-draft-${id || "new"}`;

  // Load draft on mount
  useEffect(() => {
    const saved = localStorage.getItem(DRAFT_KEY);
    if (saved) {
      try {
        const draft = JSON.parse(saved);
        // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: one-time draft restore on mount
        setLocalOverrides((prev) => ({ ...prev, ...draft }));
        toast.info("未保存の下書きを復元しました", { duration: 2000 });
      } catch { /* ignore */ }
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
      // 既存画像URLをプレビューとして復元
      if (existingTrimming.styleImage) {
        setStyleImagePreview(existingTrimming.styleImage);
      }
      if (existingTrimming.completedImage) {
        setCompletedImagePreview(existingTrimming.completedImage);
      }
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

  /**
   * React 19 useActionState を使用したフォームアクション
   */
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      try {
        if (isEdit && id) {
          const req: UpdateTrimmingRequest = {
            style_request: formData.styleRequest || undefined,
            body_weight: formData.bw ? Number(formData.bw) : undefined,
            bw_unit: formData.bwUnit || undefined,
            body_temperature: formData.bt ? Number(formData.bt) : undefined,
            used_shampoo: formData.usedShampoo || undefined,
            used_ribbon: formData.usedRibbon || undefined,
            remarks: formData.remarks || undefined,
            option_ids: formData.optionIds.length > 0 ? formData.optionIds.map(Number) : undefined,
          };
          await updateMutation.mutateAsync({ id, req });
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
          if (Object.keys(errors).length > 0) {
            setFieldErrors(errors);
            return { success: false, fieldErrors: errors, timestamp: Date.now() };
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
    startDeleteTransition(() => {
      deleteMutation.mutate(id, {
        onSuccess: () => {
          toast.success("トリミング情報を削除しました");
          onSuccess?.();
        },
        onError: (error) => {
          handleApiError(error, "削除");
        },
      });
    });
  }, [isEdit, id, deleteMutation, startDeleteTransition]);

  const isSaving = isPending;
  const isDeleting = deleteMutation.isPending || isDeleteTransitionPending;
  const mode = isEdit ? ("edit" as const) : ("new" as const);

  const isLoading = isEdit ? isTrimmingLoading : isPetLoading;
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
