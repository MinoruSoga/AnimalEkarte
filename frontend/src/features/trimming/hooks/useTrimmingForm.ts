import { useState, useEffect, useTransition, useCallback } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { usePetInfo } from "@/hooks/use-pet";
import { useGetTrimming } from "../api/get-trimming";
import { useCreateTrimming } from "../api/create-trimming";
import { useUpdateTrimming } from "../api/update-trimming";
import { useDeleteTrimming } from "../api/delete-trimming";
import type { CreateTrimmingRequest, UpdateTrimmingRequest } from "../api/types";
import type { Pet } from "@/types";

export interface TrimmingFormData {
  styleRequest: string;
  memo: string;
  eggs: string;
  parts: {
    nail: boolean;
    analGland: boolean;
    eye: boolean;
    ear: boolean;
    skin: boolean;
    oral: boolean;
  };
  styleImage: File | null;
  bw: string;
  bwUnit: "Kg" | "g";
  bt: string;
  usedShampoo: string;
  usedRibbon: string;
  treatment: string;
  medicine: string;
  charge: string;
  finalCheck: string;
  remarks: string;
  completedImage: File | null;
  courseId: string;
  optionIds: string[];
  staffName: string;
}

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
  treatment: "",
  medicine: "",
  charge: "",
  finalCheck: "",
  remarks: "",
  completedImage: null,
  courseId: "",
  optionIds: [],
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
  const { pet: petFromQuery, isLoading: isPetLoading } = usePetInfo(petId ?? "");
  const createMutation = useCreateTrimming();
  const updateMutation = useUpdateTrimming();
  const deleteMutation = useDeleteTrimming();

  // useTransition: save/delete の pending 管理 (rerender-transitions)
  const [isSaveTransitionPending, startSaveTransition] = useTransition();
  const [isDeleteTransitionPending, startDeleteTransition] = useTransition();

  const [localOverrides, setLocalOverrides] = useState<Partial<TrimmingFormData>>({});

  // 編集モード: サーバーデータを全フィールド復元（初回のみ）
  const [serverDataLoaded, setServerDataLoaded] = useState(false);
  useEffect(() => {
    if (isEdit && existingTrimming && !serverDataLoaded) {
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
        staffName: existingTrimming.staff ?? "",
      });
      setServerDataLoaded(true);
    }
  }, [isEdit, existingTrimming, serverDataLoaded]);

  const formData: TrimmingFormData = { ...defaultFormData, ...localOverrides };

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
        navigate("/trimming/select-pet");
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

  const handleSave = useCallback((onSuccess?: () => void): boolean => {
    const redirectPath: string =
      typeof location.state?.from === "string" ? location.state.from : "/trimming";

    startSaveTransition(() => {
      if (isEdit && id) {
        const req: UpdateTrimmingRequest = {
          style_request: formData.styleRequest || undefined,
          bw: formData.bw || undefined,
          bw_unit: formData.bwUnit || undefined,
          bt: formData.bt || undefined,
          used_shampoo: formData.usedShampoo || undefined,
          used_ribbon: formData.usedRibbon || undefined,
          treatment: formData.treatment || undefined,
          charge: formData.charge || undefined,
          remarks: formData.remarks || undefined,
          option_ids: formData.optionIds.length > 0 ? formData.optionIds.map(Number) : undefined,
        };
        updateMutation.mutate(
          { id, req },
          {
            onSuccess: () => {
              toast.success("トリミング情報を更新しました");
              onSuccess?.();
              navigate(redirectPath);
            },
          }
        );
      } else {
        const pet = selectedPets[0];
        if (!pet) return;
        const req: CreateTrimmingRequest = {
          pet_id: pet.id,
          appointment_date: new Date().toISOString(),
          course_id: formData.courseId ? Number(formData.courseId) : null,
          style_request: formData.styleRequest || undefined,
          remarks: formData.remarks || undefined,
        };
        createMutation.mutate(req, {
          onSuccess: () => {
            toast.success("トリミング情報を登録しました");
            onSuccess?.();
            navigate(redirectPath);
          },
        });
      }
    });
    return true;
  }, [isEdit, id, formData, selectedPets, location.state, updateMutation, createMutation, navigate, startSaveTransition]);

  const isSaving = createMutation.isPending || updateMutation.isPending || isSaveTransitionPending;
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
    handleSave,
    handleDelete,
    isSaving,
    isDeleting,
  };
}
