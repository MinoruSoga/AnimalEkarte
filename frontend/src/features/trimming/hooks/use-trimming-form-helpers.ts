import { useState, useEffect, useCallback, useRef } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import type { TrimmingFormData } from "@/types/trimming";
import type { Pet } from "@/types";
import { uploadTrimmingImage } from "../api/upload-trimming-image";

type TrimmingImageField = "styleImage" | "completedImage";

type TrimmingImageOverrides = (
  updater: (prev: Partial<TrimmingFormData>) => Partial<TrimmingFormData>,
) => void;

function selectedImageFile(event: React.ChangeEvent<HTMLInputElement>): File | null {
  const file = event.target.files?.[0] ?? null;
  event.target.value = "";
  return file;
}

function readImageDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onloadend = () => {
      if (typeof reader.result === "string" && reader.result !== "") {
        resolve(reader.result);
        return;
      }
      reject(reader.error ?? new Error("image preview unavailable"));
    };
    reader.readAsDataURL(file);
  });
}

async function applyTrimmingImageSelection(input: {
  file: File;
  isCurrent: () => boolean;
  onStored: (key: string, preview: string) => void;
}): Promise<void> {
  try {
    const uploaded = await uploadTrimmingImage(input.file);
    const preview =
      uploaded.signed_url !== "" ? uploaded.signed_url : await readImageDataUrl(input.file);
    if (!input.isCurrent()) return;
    input.onStored(uploaded.key, preview);
  } catch (error) {
    if (!input.isCurrent()) return;
    handleApiError(error, "画像アップロード");
  }
}

function useTrimmingImageField(
  field: TrimmingImageField,
  setPreview: (value: string | null) => void,
  setLocalOverrides: TrimmingImageOverrides,
) {
  const requestIdRef = useRef(0);

  const handleChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = selectedImageFile(event);
      if (!file) return undefined;
      const requestId = requestIdRef.current + 1;
      requestIdRef.current = requestId;
      return applyTrimmingImageSelection({
        file,
        isCurrent: () => requestIdRef.current === requestId,
        onStored: (key, preview) => {
          setLocalOverrides((prev) => ({ ...prev, [field]: key }));
          setPreview(preview);
        },
      });
    },
    [field, setLocalOverrides, setPreview],
  );

  const remove = useCallback(() => {
    requestIdRef.current += 1;
    setLocalOverrides((prev) => ({ ...prev, [field]: "" }));
    setPreview(null);
  }, [field, setLocalOverrides, setPreview]);

  return { handleChange, remove };
}

type TrimmingHydrateSource = {
  styleRequest?: string;
  courseId?: string | null;
  optionIds?: string[] | null;
  bw?: string | null;
  bwUnit?: string | null;
  bt?: string | null;
  usedShampoo?: string | null;
  usedRibbon?: string | null;
  remarks?: string | null;
  staffId?: string | null;
  staff?: string | null;
  styleImage?: string | null;
  completedImage?: string | null;
  reservationTypeId?: string;
};

function toBwUnit(value: string | null | undefined): TrimmingFormData["bwUnit"] {
  return value === "g" ? "g" : "Kg";
}

function existingTrimmingOverrides(existing: TrimmingHydrateSource): Partial<TrimmingFormData> {
  return {
    styleRequest: existing.styleRequest,
    courseId: existing.courseId ?? "",
    optionIds: existing.optionIds ?? [],
    bw: existing.bw ?? "",
    bwUnit: toBwUnit(existing.bwUnit),
    bt: existing.bt ?? "",
    usedShampoo: existing.usedShampoo ?? "",
    usedRibbon: existing.usedRibbon ?? "",
    remarks: existing.remarks ?? "",
    staffId: existing.staffId ?? "",
    staffName: existing.staff ?? "",
  };
}

function appointmentTrimmingOverrides(
  prev: Partial<TrimmingFormData>,
  existing: TrimmingHydrateSource,
): Partial<TrimmingFormData> {
  return {
    ...prev,
    reservationTypeId: existing.reservationTypeId || prev.reservationTypeId || "",
    styleRequest: existing.styleRequest || prev.styleRequest || "",
    courseId: existing.courseId || prev.courseId || "",
    optionIds:
      (existing.optionIds?.length ?? 0) > 0 ? (existing.optionIds ?? []) : (prev.optionIds ?? []),
    bw: existing.bw || prev.bw || "",
    bwUnit: toBwUnit(existing.bwUnit || prev.bwUnit),
    bt: existing.bt || prev.bt || "",
    usedShampoo: existing.usedShampoo || prev.usedShampoo || "",
    usedRibbon: existing.usedRibbon || prev.usedRibbon || "",
    remarks: existing.remarks || prev.remarks || "",
    staffId: existing.staffId || prev.staffId || "",
    staffName: existing.staff || prev.staffName || "",
  };
}

export function useTrimmingFormHydration(input: {
  isEdit: boolean;
  existingTrimming: TrimmingHydrateSource | undefined;
  existingAppointmentTrimming: TrimmingHydrateSource | undefined;
  setLocalOverrides: (
    next:
      Partial<TrimmingFormData> | ((prev: Partial<TrimmingFormData>) => Partial<TrimmingFormData>),
  ) => void;
  setStyleImagePreview: (value: string | null) => void;
  setCompletedImagePreview: (value: string | null) => void;
}) {
  const {
    isEdit,
    existingTrimming,
    existingAppointmentTrimming,
    setLocalOverrides,
    setStyleImagePreview,
    setCompletedImagePreview,
  } = input;
  const serverDataLoadedRef = useRef(false);
  useEffect(() => {
    if (isEdit && existingTrimming && !serverDataLoadedRef.current) {
      serverDataLoadedRef.current = true;
      setLocalOverrides(existingTrimmingOverrides(existingTrimming));
      if (existingTrimming.styleImage) {
        setStyleImagePreview(existingTrimming.styleImage);
      }
      if (existingTrimming.completedImage) {
        setCompletedImagePreview(existingTrimming.completedImage);
      }
    }
  }, [isEdit, existingTrimming, setLocalOverrides, setStyleImagePreview, setCompletedImagePreview]);

  const appointmentDataLoadedRef = useRef(false);
  useEffect(() => {
    if (isEdit || !existingAppointmentTrimming || appointmentDataLoadedRef.current) return;
    appointmentDataLoadedRef.current = true;
    setLocalOverrides((prev) => appointmentTrimmingOverrides(prev, existingAppointmentTrimming));
    if (existingAppointmentTrimming.styleImage) {
      setStyleImagePreview(existingAppointmentTrimming.styleImage);
    }
    if (existingAppointmentTrimming.completedImage) {
      setCompletedImagePreview(existingAppointmentTrimming.completedImage);
    }
  }, [
    isEdit,
    existingAppointmentTrimming,
    setLocalOverrides,
    setStyleImagePreview,
    setCompletedImagePreview,
  ]);
}

export function useTrimmingFormImages(setLocalOverrides: TrimmingImageOverrides) {
  const [styleImagePreview, setStyleImagePreview] = useState<string | null>(null);
  const [completedImagePreview, setCompletedImagePreview] = useState<string | null>(null);
  const styleImage = useTrimmingImageField("styleImage", setStyleImagePreview, setLocalOverrides);
  const completedImage = useTrimmingImageField(
    "completedImage",
    setCompletedImagePreview,
    setLocalOverrides,
  );

  return {
    styleImagePreview,
    setStyleImagePreview,
    completedImagePreview,
    setCompletedImagePreview,
    handleStyleImageChange: styleImage.handleChange,
    handleCompletedImageChange: completedImage.handleChange,
    removeStyleImage: styleImage.remove,
    removeCompletedImage: completedImage.remove,
  };
}

export function useTrimmingFormPetSync(input: {
  isEdit: boolean;
  petId: string | null;
  petFromQuery: Pet | undefined;
  petFromEdit?: Pet;
  isPetLoading: boolean;
  setSelectedPets: (pets: Pet[]) => void;
}) {
  const { isEdit, petId, petFromQuery, petFromEdit, isPetLoading, setSelectedPets } = input;
  const navigate = useNavigate();

  useEffect(() => {
    if (isEdit && petFromEdit) {
      setSelectedPets([petFromEdit]);
    }
  }, [isEdit, petFromEdit, setSelectedPets]);

  useEffect(() => {
    if (!isEdit) {
      if (petFromQuery) {
        setSelectedPets([petFromQuery]);
      } else if (!petId && !isPetLoading) {
        navigate(paths.trimming.selectPet.getHref());
      }
    }
  }, [isEdit, petId, petFromQuery, isPetLoading, setSelectedPets, navigate]);
}

export function createTrimmingDeleteHandler(input: {
  isEdit: boolean;
  id: string | undefined;
  isMutationAllowed: (action: "canCreate" | "canEdit" | "canDelete") => boolean;
  isEditPetReady: () => boolean;
  isPetDeceased: () => boolean;
  deleteTrimming: (
    id: string,
    opts: { onSuccess: () => void; onError: (error: unknown) => void },
  ) => void;
}): (onSuccess?: () => void) => void {
  return (onSuccess?: () => void) => {
    if (!input.isEdit || !input.id) return;
    if (!input.isMutationAllowed("canDelete")) {
      toast.error("この操作を行う権限がありません");
      return;
    }
    if (!input.isEditPetReady()) {
      toast.error("ペット情報の読み込みが完了してから削除してください");
      return;
    }
    if (input.isPetDeceased()) {
      toast.error("死亡したペットのトリミング記録は削除できません");
      return;
    }
    input.deleteTrimming(input.id, {
      onSuccess: () => {
        toast.success("トリミング情報を削除しました");
        onSuccess?.();
      },
      onError: (error) => {
        handleApiError(error, "削除");
      },
    });
  };
}
