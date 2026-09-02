import { useState, useCallback, useMemo, useActionState } from "react";
import { useLocation, useSearchParams } from "react-router";
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
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import {
  buildCreateTrimmingRequest,
  buildUpdateTrimmingRequest,
  DEFAULT_TRIMMING_FORM_DATA,
  findDefaultTrimmingReservationTypeId,
  defaultRecordShortcutTimes,
  formatJSTDate,
  normalizeVisitDate,
  parseTrimmingAppointmentId,
} from "./trimming-form-utils";
import { useTrimmingFormValidation } from "./use-trimming-form-validation";
import {
  createTrimmingDeleteHandler,
  useTrimmingFormHydration,
  useTrimmingFormImages,
  useTrimmingFormPetSync,
} from "./use-trimming-form-helpers";

export function useTrimmingForm(id?: string) {
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;
  const { appointmentIdFromState, existingAppointmentId } = parseTrimmingAppointmentId(
    location.state?.appointmentId,
    searchParams.get("appointmentId"),
  );
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
  const hasExistingAppointment = Number.isFinite(appointmentIdFromState) || Number.isFinite(reusableAppointmentId);
  const { fieldErrors, validate } = useTrimmingFormValidation();
  const [localOverrides, setLocalOverrides] = useState<Partial<TrimmingFormData>>({});
  const images = useTrimmingFormImages(setLocalOverrides);

  useTrimmingFormHydration({
    isEdit,
    existingTrimming,
    existingAppointmentTrimming,
    setLocalOverrides,
    setStyleImagePreview: images.setStyleImagePreview,
    setCompletedImagePreview: images.setCompletedImagePreview,
  });

  const formData = useMemo<TrimmingFormData>(
    () => ({ ...DEFAULT_TRIMMING_FORM_DATA, ...localOverrides }),
    [localOverrides]
  );
  const setFormData = useCallback((next: Partial<TrimmingFormData>) => {
    setLocalOverrides((prev) => ({ ...prev, ...next }));
  }, []);

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      try {
        if (isEdit && id) {
          await updateMutation.mutateAsync({ id, req: buildUpdateTrimmingRequest(formData) });
          toast.success("トリミング情報を更新しました");
        } else if ((existingAppointmentHasDetail && existingAppointmentId) || reusableTrimming?.hasDetail) {
          await updateMutation.mutateAsync({
            id: existingAppointmentId || reusableTrimming?.id || "",
            req: buildUpdateTrimmingRequest(formData),
          });
          toast.success("トリミング情報を更新しました");
        } else {
          const pet = selectedPets[0];
          if (!pet) return { success: false, timestamp: Date.now() };
          const validation = validate(formData, defaultTrimmingReservationTypeId);
          if (!validation.valid) {
            return { success: false, fieldErrors: validation.errors, timestamp: Date.now() };
          }
          const fallbackDate = visitDateFromState ?? formatJSTDate(new Date());
          const fallbackTimes = defaultRecordShortcutTimes(fallbackDate);
          const startDate = formData.startTime || (hasExistingAppointment ? undefined : fallbackTimes.start);
          const endDate = formData.endTime || (hasExistingAppointment ? undefined : fallbackTimes.end);
          const req = buildCreateTrimmingRequest(
            formData,
            Number(pet.id),
            validation.reservationTypeId,
            startDate,
            endDate,
            Number.isFinite(appointmentIdFromState) ? appointmentIdFromState : reusableAppointmentId,
          );
          if (!hasExistingAppointment) {
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

  useTrimmingFormPetSync({
    isEdit,
    petId,
    petFromQuery,
    isPetLoading,
    existingTrimming,
    setSelectedPets,
  });

  const handleDelete = useCallback((onSuccess?: () => void) => {
    createTrimmingDeleteHandler({
      isEdit,
      id,
      deleteTrimming: (trimmingId, opts) => deleteMutation.mutate(trimmingId, opts),
    })(onSuccess);
  }, [isEdit, id, deleteMutation]);

  return {
    mode: isEdit ? ("edit" as const) : ("new" as const),
    formData,
    setFormData,
    styleImagePreview: images.styleImagePreview,
    completedImagePreview: images.completedImagePreview,
    petSelection,
    handleStyleImageChange: images.handleStyleImageChange,
    handleCompletedImageChange: images.handleCompletedImageChange,
    removeStyleImage: images.removeStyleImage,
    removeCompletedImage: images.removeCompletedImage,
    formAction,
    formState,
    handleDelete,
    isSaving: isPending,
    isDeleting: deleteMutation.isPending,
    fieldErrors,
    isLoading: isEdit ? isTrimmingLoading : isPetLoading || isAppointmentLoading || isSameDayTrimmingsLoading,
    notFound: isEdit && !isTrimmingLoading && !existingTrimming && !!id,
    hasExistingAppointment,
  };
}
