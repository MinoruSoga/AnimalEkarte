import { useState, useEffect, useCallback, useRef } from "react";
import { useNavigate, useLocation } from "react-router";
import { addHours } from "date-fns";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";

import { getReservationStatusLabel } from "@/utils/status-helpers";
import type {
  ReservationAppointment,
  ReservationFormData,
  ReservationStatus,
  Pet,
  NavigationState,
} from "../types";
import { useGetReservations } from "../api/get-reservations";
import { useCreateReservation } from "../api/create-reservation";
import { useUpdateReservation } from "../api/update-reservation";
import { useDeleteReservation } from "../api/delete-reservation";
import { transformToCreateRequest } from "../api/transforms";

const EMPTY_APPOINTMENTS: ReservationAppointment[] = [];

/** Maps reservation type → record creation path */
const RECORD_PATH: Record<string, string> = {
  "トリミング": "/trimming/new",
  "入院": "/hospitalization/new",
  "ホテル": "/hospitalization/new",
};

/** Maps reservation type → pet-select path */
const SELECT_PATH: Record<string, string> = {
  "トリミング": "/trimming/select-pet",
  "入院": "/hospitalization/select-pet",
  "ホテル": "/hospitalization/select-pet",
};

export function useReservationManagement() {
  const navigate = useNavigate();
  const location = useLocation();
  const locationFrom = (location.state as NavigationState | null)?.from ?? null;

  // API mutations
  const { data: appointments = EMPTY_APPOINTMENTS, isLoading } = useGetReservations();
  const createMutation = useCreateReservation();
  const updateMutation = useUpdateReservation();
  const deleteMutation = useDeleteReservation();

  // Edit/Create Modal State
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingAppointment, setEditingAppointment] = useState<ReservationFormData | null>(null);

  // Detail Modal State
  const [isDetailOpen, setIsDetailOpen] = useState(false);
  const [detailAppointment, setDetailAppointment] = useState<ReservationAppointment | null>(null);

  // Confirm Dialog State
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<ReservationAppointment | null>(null);
  const [petSelectConfirmOpen, setPetSelectConfirmOpen] = useState(false);
  const [petSelectPath, setPetSelectPath] = useState("");

  // Check for overlaps
  const checkOverlap = useCallback(
    (newStart: Date, newEnd: Date, doctor: string, excludeId?: string): boolean =>
      appointments.some((app) => {
        if (excludeId && app.id === excludeId) return false;
        if (app.status === "cancelled") return false;
        if (app.doctor !== doctor) return false;
        return newStart < app.end && newEnd > app.start;
      }),
    [appointments]
  );

  // Open Form Modal (Create or Edit)
  const handleOpenForm = useCallback((appointment?: ReservationFormData) => {
    setEditingAppointment(appointment ?? null);
    setIsFormOpen(true);
    setIsDetailOpen(false);
  }, []);

  const isFormOpenRef = useRef(false);
  useEffect(() => {
    isFormOpenRef.current = isFormOpen;
  }, [isFormOpen]);

  const editingAppointmentRef = useRef<ReservationFormData | null>(null);
  useEffect(() => {
    editingAppointmentRef.current = editingAppointment;
  }, [editingAppointment]);

  // Auto-open form when petId is in URL query
  useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    const petId = searchParams.get("petId");

    if (petId && !isFormOpenRef.current) {
      const now = new Date();
      now.setMinutes(0, 0, 0);
      const start = addHours(now, 1);

      const stub: ReservationFormData = {
        start,
        end: addHours(start, 1),
        status: "confirmed",
        visitType: "first",
        doctor: "",
        isDesignated: false,
        petId,
      };
      // eslint-disable-next-line react-hooks/set-state-in-effect -- URLパラメータに応じてフォームを自動オープン。間接的なsetStateだが意図的
      handleOpenForm(stub);
    }
  }, [location.search, handleOpenForm]);

  // Open Detail Modal
  const handleOpenDetail = useCallback((appointment: ReservationAppointment) => {
    setDetailAppointment(appointment);
    setIsDetailOpen(true);
  }, []);

  const handleTimeSlotClick = useCallback(
    (date: Date) => {
      const stub: ReservationFormData = {
        start: date,
        end: addHours(date, 1),
        status: "confirmed",
        visitType: "first",
        doctor: "",
        isDesignated: false,
      };
      handleOpenForm(stub);
    },
    [handleOpenForm]
  );

  const handleCloseForm = useCallback(() => {
    setIsFormOpen(false);
    setEditingAppointment(null);
  }, []);

  const handleCloseDetail = useCallback(() => {
    setIsDetailOpen(false);
    setDetailAppointment(null);
  }, []);

  const handleSave = useCallback(
    async (data: ReservationFormData, selectedPets: Pet[]) => {
      if (!data.start || !data.end || selectedPets.length === 0) return;

      const currentEditing = editingAppointmentRef.current;
      const targetDoctor = data.doctor || currentEditing?.doctor || "";
      const hasOverlap = checkOverlap(data.start, data.end, targetDoctor, currentEditing?.id);

      if (hasOverlap) {
        toast.error("指定された時間帯には既に予約が入っています", {
          description: "時間帯または担当医を変更してください。",
        });
        return;
      }

      if (currentEditing?.id) {
        // Edit mode
        const updatePayload = {
          id: currentEditing.id,
          req: {
            start_time: data.start.toISOString(),
            end_time: data.end.toISOString(),
            visit_type: data.visitType || "first",
            service_type_id: Number(data.type || 0),
            doctor_id: targetDoctor,
            is_designated: data.isDesignated ?? false,
            status: data.status || ("confirmed" as const),
            notes: data.notes,
          },
        };
        updateMutation.mutate(updatePayload, {
          onSuccess: () => {
            toast.success("予約を更新しました", { description: `担当医: ${targetDoctor}` });
            handleCloseForm();
            if (locationFrom) {
              navigate(locationFrom);
            }
          },
        });
      } else {
        // Create mode
        try {
          await Promise.all(
            selectedPets.map((pet) => {
              const createPayload = transformToCreateRequest(data, pet.id, pet.ownerId);
              return createMutation.mutateAsync(createPayload);
            })
          );
          toast.success("予約を作成しました", {
            description: `担当医: ${targetDoctor} / ${selectedPets.length}件`,
          });
          handleCloseForm();
          if (locationFrom) {
            navigate(locationFrom);
          }
        } catch (error) {
          handleApiError(error, "作成");
        }
      }
    },
    [checkOverlap, handleCloseForm, locationFrom, navigate, updateMutation, createMutation]
    // editingAppointment をオブジェクト参照ではなくrefで参照するため依存から除外
  );

  const handleAppointmentUpdate = useCallback(
    (appointment: ReservationAppointment, newStart: Date, newEnd: Date) => {
      const hasOverlap = checkOverlap(newStart, newEnd, appointment.doctor, appointment.id);

      if (hasOverlap) {
        toast.error("移動先に予約が重複しています", {
          description: "別の時間帯にドラッグしてください。",
        });
        return;
      }

      const updatePayload = {
        id: appointment.id,
        req: {
          start_time: newStart.toISOString(),
          end_time: newEnd.toISOString(),
          visit_type: appointment.visitType,
          service_type: appointment.type,
          doctor_id: appointment.doctor,
          is_designated: appointment.isDesignated,
          status: appointment.status,
          notes: appointment.notes,
        },
      };

      updateMutation.mutate(updatePayload, {
        onSuccess: () => {
          toast.success("予約時間を変更しました", {
            description: `${appointment.petName} / ${appointment.doctor}`,
          });
        },
      });
    },
    [checkOverlap, updateMutation]
  );

  // Status Change Handler
  const handleStatusChange = useCallback(
    (appointment: ReservationAppointment, status: ReservationStatus) => {
      const updatePayload = {
        id: appointment.id,
        req: {
          start_time: appointment.start.toISOString(),
          end_time: appointment.end.toISOString(),
          visit_type: appointment.visitType,
          service_type: appointment.type,
          doctor_id: appointment.doctor,
          is_designated: appointment.isDesignated,
          status,
          notes: appointment.notes,
        },
      };

      updateMutation.mutate(updatePayload, {
        onSuccess: () => {
          setDetailAppointment((prev) => (prev ? { ...prev, status } : null));
          const statusLabel = getReservationStatusLabel(status);
          toast.success("ステータスを更新しました", {
            description: `${appointment.petName} → ${statusLabel}`,
          });
        },
      });
    },
    [updateMutation]
  );

  // Delete handler
  const handleDelete = useCallback((appointment: ReservationAppointment) => {
    setDeleteTarget(appointment);
    setDeleteConfirmOpen(true);
  }, []);

  const executeDelete = useCallback(() => {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id, {
      onSuccess: () => {
        setDeleteConfirmOpen(false);
        setDeleteTarget(null);
        handleCloseDetail();
        toast.success("予約を削除しました", {
          description: `${deleteTarget.petName} (${deleteTarget.ownerName}様)`,
        });
      },
      onError: (error: unknown) => {
        handleApiError(error, "削除");
      },
    });
  }, [deleteTarget, deleteMutation, handleCloseDetail]);

  // Create Record Handler
  const handleCreateRecord = useCallback(
    (appointment: ReservationAppointment) => {
      const targetPath = RECORD_PATH[appointment.type] || "/medical-records/new";

      if (appointment.petId) {
        navigate(`${targetPath}?petId=${appointment.petId}`, { state: { from: "/reservations" } });
      } else {
        const selectPath = SELECT_PATH[appointment.type] || "/medical-records/select-pet";
        setPetSelectPath(selectPath);
        setPetSelectConfirmOpen(true);
      }
    },
    [navigate]
  );

  const handlePetSelectConfirm = useCallback(() => {
    setPetSelectConfirmOpen(false);
    navigate(petSelectPath, { state: { from: "/reservations" } });
  }, [navigate, petSelectPath]);

  const handleDeleteConfirmClose = useCallback(() => {
    setDeleteConfirmOpen(false);
    setDeleteTarget(null);
  }, []);

  return {
    // Data
    appointments,
    isLoading,

    // Form modal
    isFormOpen,
    editingAppointment,
    handleOpenForm,
    handleCloseForm,
    handleSave,

    // Detail modal
    isDetailOpen,
    detailAppointment,
    handleOpenDetail,
    handleCloseDetail,
    handleStatusChange,
    handleDelete,
    handleCreateRecord,

    // Calendar interactions
    handleTimeSlotClick,
    handleAppointmentUpdate,

    // Delete confirm
    deleteConfirmOpen,
    deleteTarget,
    executeDelete,
    handleDeleteConfirmClose,

    // Pet select confirm
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handlePetSelectConfirm,
  };
}
