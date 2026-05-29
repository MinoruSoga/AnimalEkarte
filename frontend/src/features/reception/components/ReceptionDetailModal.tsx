// React/Framework
import { memo, useCallback } from "react";
import { useNavigate } from "react-router";

// Internal
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { C, LAYOUT } from "@/lib/design-tokens";

// Internal
import { paths } from "@/config/paths";

// Types
import type { ReceptionAppointment as Appointment } from "../api/types";
import {
  ReceptionDialogBody,
  ReceptionDialogFooter,
  ReceptionDialogHeader,
} from "./ReceptionDetailModalParts";

interface ReceptionDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  appointment: Appointment | null;
  onConfirm?: () => void;
  onEdit?: (appointment: Appointment) => void;
  onCancel?: (appointment: Appointment) => void;
  currentStatus?: string;
  canCreateMedicalRecord?: boolean;
  canCreateAccounting?: boolean;
  canCreateHospitalization?: boolean;
}

export const ReceptionDetailModal = memo(function ReceptionDetailModal({
  isOpen,
  onClose,
  appointment,
  onConfirm,
  onEdit,
  onCancel,
  currentStatus,
  canCreateMedicalRecord,
  canCreateAccounting,
  canCreateHospitalization,
}: ReceptionDetailModalProps) {
  const navigate = useNavigate();

  // Extract primitives before hooks so useCallback deps stay stable
  const petId = appointment?.petId;
  const appointmentId = appointment?.id;
  const ownerId = appointment?.ownerId;

  const navigateAndClose = useCallback((path: string, extraState?: Record<string, unknown>) => {
    navigate(path, { state: { from: "/", ...extraState } });
    onClose();
  }, [navigate, onClose]);

  const handleCreateMedicalRecord = useCallback((tab?: string) => {
    const params = new URLSearchParams();
    if (petId) params.set("petId", petId);
    if (appointmentId) params.set("appointmentId", appointmentId);
    if (tab) params.set("tab", tab);
    const query = params.toString();
    const basePath = petId
      ? paths.medicalRecords.new.getHref()
      : paths.medicalRecords.selectPet.getHref();
    const base = query ? `${basePath}?${query}` : basePath;
    navigateAndClose(base, { appointmentId });
  }, [petId, appointmentId, navigateAndClose]);

  const handleCreateTrimming = useCallback(() => {
    const params = new URLSearchParams();
    if (petId) params.set("petId", petId);
    if (appointmentId) params.set("appointmentId", appointmentId);
    const query = params.toString();
    const basePath = petId
      ? paths.trimming.new.getHref()
      : paths.trimming.selectPet.getHref();
    const path = query ? `${basePath}?${query}` : basePath;
    navigateAndClose(path, {
      appointmentId,
    });
  }, [petId, appointmentId, navigateAndClose]);

  const handleCreateHospitalization = useCallback(() =>
    navigateAndClose(petId ? `${paths.hospitalization.new.getHref()}?petId=${petId}` : paths.hospitalization.new.getHref()),
  [petId, navigateAndClose]);

  const handleCreateAccounting = useCallback(() =>
    navigateAndClose(petId ? `${paths.accounting.new.getHref()}?petId=${petId}` : paths.accounting.new.getHref(), {
      appointmentId,
    }),
  [petId, appointmentId, navigateAndClose]);

  const handleOpenOwnerDetail = useCallback(() => {
    if (ownerId) {
      navigateAndClose(`/owners/${ownerId}`);
    } else if (petId) {
      navigateAndClose(`/pets/${petId}`);
    }
  }, [ownerId, petId, navigateAndClose]);

  if (!appointment) return null;

  const isTrimming = appointment.reservationCategory === "trimming";
  const isHospitalization = appointment.reservationType.includes("入院");
  const isMedical = appointment.reservationCategory === "general" || (!isTrimming && !isHospitalization);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent
        className={`${LAYOUT.modal.sm} p-0 gap-0 overflow-hidden ${C.bgWhite}`}
      >
        <ReceptionDialogHeader appointment={appointment} currentStatus={currentStatus} />
        <ReceptionDialogBody
          appointment={appointment}
          isTrimming={isTrimming}
          onCreateMedicalRecord={handleCreateMedicalRecord}
          onCreateTrimming={handleCreateTrimming}
          onCreateAccounting={handleCreateAccounting}
          onCreateHospitalization={handleCreateHospitalization}
          canCreateMedicalRecord={isTrimming ? false : canCreateMedicalRecord}
          canCreateAccounting={canCreateAccounting}
          canCreateHospitalization={canCreateHospitalization}
        />
        <ReceptionDialogFooter
          currentStatus={currentStatus}
          appointment={appointment}
          isTrimming={isTrimming}
          isHospitalization={isHospitalization}
          isMedical={isMedical}
          onConfirm={onConfirm}
          onEdit={onEdit}
          onCancel={onCancel}
          onOpenOwnerDetail={handleOpenOwnerDetail}
          onCreateMedicalRecord={handleCreateMedicalRecord}
          onCreateTrimming={handleCreateTrimming}
          onCreateAccounting={handleCreateAccounting}
          onCreateHospitalization={handleCreateHospitalization}
        />
      </DialogContent>
    </Dialog>
  );
});
