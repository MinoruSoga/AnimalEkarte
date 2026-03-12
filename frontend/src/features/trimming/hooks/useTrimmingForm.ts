import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/features/pets/api/get-pet";
import {
  useGetTrimming,
  useCreateTrimming,
  useUpdateTrimming,
  useDeleteTrimming,
} from "../api";
import type { CreateTrimmingRequest, UpdateTrimmingRequest } from "../api";

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

  // New fields for Master Selection
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

  // API hooks
  const { data: existingTrimming } = useGetTrimming(id ?? "");
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(
    petId ?? ""
  );
  const createMutation = useCreateTrimming();
  const updateMutation = useUpdateTrimming();
  const deleteMutation = useDeleteTrimming();

  // Local overrides applied on top of server data (tracks user edits in edit mode)
  const [localOverrides, setLocalOverrides] = useState<
    Partial<TrimmingFormData>
  >({});

  // Merge: server data as base + user edits on top
  const formData: TrimmingFormData =
    isEdit && existingTrimming
      ? {
          ...defaultFormData,
          styleRequest: existingTrimming.styleRequest,
          ...localOverrides,
        }
      : { ...defaultFormData, ...localOverrides };

  const setFormData = (next: Partial<TrimmingFormData>) => {
    setLocalOverrides((prev) => ({ ...prev, ...next }));
  };

  const [styleImagePreview, setStyleImagePreview] = useState<string | null>(
    null
  );
  const [completedImagePreview, setCompletedImagePreview] = useState<
    string | null
  >(null);

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

  const handleStyleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setLocalOverrides((prev) => ({ ...prev, styleImage: file }));
      const reader = new FileReader();
      reader.onloadend = () => {
        setStyleImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleCompletedImageChange = (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = e.target.files?.[0];
    if (file) {
      setLocalOverrides((prev) => ({ ...prev, completedImage: file }));
      const reader = new FileReader();
      reader.onloadend = () => {
        setCompletedImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const removeStyleImage = () => {
    setLocalOverrides((prev) => ({ ...prev, styleImage: null }));
    setStyleImagePreview(null);
  };

  const removeCompletedImage = () => {
    setLocalOverrides((prev) => ({ ...prev, completedImage: null }));
    setCompletedImagePreview(null);
  };

  const handleDelete = (onSuccess?: () => void) => {
    if (isEdit && id) {
      deleteMutation.mutate(id, {
        onSuccess: () => {
          toast.success("トリミング情報を削除しました");
          onSuccess?.();
        },
      });
    }
  };

  // ✅ onSuccess コールバックを受け取る（markClean を保存完了後に呼ぶため）
  const handleSave = (onSuccess?: () => void): boolean => {
    const redirectPath: string =
      typeof location.state?.from === "string"
        ? location.state.from
        : "/trimming";

    if (isEdit && id) {
      const req: UpdateTrimmingRequest = {
        style_request: formData.styleRequest,
        notes: formData.remarks,
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
      if (!pet) return false;
      const req: CreateTrimmingRequest = {
        pet_id: pet.id,
        owner_id: pet.ownerId,
        appointment_date: new Date().toISOString(),
        course: formData.courseId,
        options: JSON.stringify(formData.optionIds),
        style_request: formData.styleRequest,
        notes: formData.remarks,
      };
      createMutation.mutate(req, {
        onSuccess: () => {
          toast.success("トリミング情報を登録しました");
          onSuccess?.();
          navigate(redirectPath);
        },
      });
    }
    return true;
  };

  const isSaving = createMutation.isPending || updateMutation.isPending;
  const isDeleting = deleteMutation.isPending;
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
