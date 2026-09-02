import { useState, useEffect, useCallback, useRef } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import type { TrimmingFormData } from "@/types/trimming";
import type { Pet } from "@/types";

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
    optionIds: (existing.optionIds?.length ?? 0) > 0
      ? existing.optionIds ?? []
      : (prev.optionIds ?? []),
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
    next: Partial<TrimmingFormData> | ((prev: Partial<TrimmingFormData>) => Partial<TrimmingFormData>),
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
  }, [isEdit, existingAppointmentTrimming, setLocalOverrides, setStyleImagePreview, setCompletedImagePreview]);
}

export function useTrimmingFormImages(setLocalOverrides: (
  updater: (prev: Partial<TrimmingFormData>) => Partial<TrimmingFormData>,
) => void) {
  const [styleImagePreview, setStyleImagePreview] = useState<string | null>(null);
  const [completedImagePreview, setCompletedImagePreview] = useState<string | null>(null);

  const handleStyleImageChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setLocalOverrides((prev) => ({ ...prev, styleImage: file }));
      const reader = new FileReader();
      reader.onloadend = () => setStyleImagePreview(reader.result as string);
      reader.readAsDataURL(file);
    }
  }, [setLocalOverrides]);

  const handleCompletedImageChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setLocalOverrides((prev) => ({ ...prev, completedImage: file }));
      const reader = new FileReader();
      reader.onloadend = () => setCompletedImagePreview(reader.result as string);
      reader.readAsDataURL(file);
    }
  }, [setLocalOverrides]);

  const removeStyleImage = useCallback(() => {
    setLocalOverrides((prev) => ({ ...prev, styleImage: null }));
    setStyleImagePreview(null);
  }, [setLocalOverrides]);

  const removeCompletedImage = useCallback(() => {
    setLocalOverrides((prev) => ({ ...prev, completedImage: null }));
    setCompletedImagePreview(null);
  }, [setLocalOverrides]);

  return {
    styleImagePreview,
    setStyleImagePreview,
    completedImagePreview,
    setCompletedImagePreview,
    handleStyleImageChange,
    handleCompletedImageChange,
    removeStyleImage,
    removeCompletedImage,
  };
}

export function useTrimmingFormPetSync(input: {
  isEdit: boolean;
  petId: string | null;
  petFromQuery: Pet | undefined;
  isPetLoading: boolean;
  existingTrimming: {
    petId?: string;
    ownerId?: string;
    petName?: string;
    petNumber?: string;
    ownerName?: string;
    species?: string;
    weight?: string;
  } | undefined;
  setSelectedPets: (pets: Pet[]) => void;
}) {
  const {
    isEdit,
    petId,
    petFromQuery,
    isPetLoading,
    existingTrimming,
    setSelectedPets,
  } = input;
  const navigate = useNavigate();

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
  deleteTrimming: (
    id: string,
    opts: { onSuccess: () => void; onError: (error: unknown) => void },
  ) => void;
}): (onSuccess?: () => void) => void {
  return (onSuccess?: () => void) => {
    if (!input.isEdit || !input.id) return;
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
