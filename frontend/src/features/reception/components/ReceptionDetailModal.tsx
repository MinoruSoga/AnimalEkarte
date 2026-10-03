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

function isHospitalizationReservationType(reservationType: string): boolean {
  return reservationType.includes("入院") || reservationType.includes("ホテル");
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
  const ownerName = appointment?.ownerName;
  const visitDate = appointment?.visitDate;

  const navigateAndClose = useCallback(
    (path: string, extraState?: Record<string, unknown>) => {
      // 表示中日付のボードへ戻れるよう、from は当日固定ではなく予約日を保持する
      // （?date=<当日> は Reception.tsx の isToday 判定で / と同じボードになる）
      const from = visitDate ? `/?date=${visitDate}` : "/";
      navigate(path, { state: { from, ...extraState } });
      onClose();
    },
    [navigate, onClose, visitDate],
  );

  const handleCreateMedicalRecord = useCallback(
    (tab?: string) => {
      const params = new URLSearchParams();
      if (petId) params.set("petId", petId);
      if (appointmentId) params.set("appointmentId", appointmentId);
      if (visitDate) params.set("visitDate", visitDate);
      if (tab) params.set("tab", tab);
      const query = params.toString();
      const basePath = petId
        ? paths.medicalRecords.new.getHref()
        : paths.medicalRecords.selectPet.getHref();
      const base = query ? `${basePath}?${query}` : basePath;
      navigateAndClose(base, { appointmentId, visitDate });
    },
    [petId, appointmentId, visitDate, navigateAndClose],
  );

  const handleCreateTrimming = useCallback(() => {
    const params = new URLSearchParams();
    if (petId) params.set("petId", petId);
    if (appointmentId) params.set("appointmentId", appointmentId);
    if (visitDate) params.set("visitDate", visitDate);
    const query = params.toString();
    const basePath = petId ? paths.trimming.new.getHref() : paths.trimming.selectPet.getHref();
    const path = query ? `${basePath}?${query}` : basePath;
    navigateAndClose(path, {
      appointmentId,
      visitDate,
    });
  }, [petId, appointmentId, visitDate, navigateAndClose]);

  const handleCreateHospitalization = useCallback(
    () =>
      navigateAndClose(
        petId
          ? `${paths.hospitalization.new.getHref()}?petId=${petId}`
          : paths.hospitalization.new.getHref(),
      ),
    [petId, navigateAndClose],
  );

  const handleCreateAccounting = useCallback(
    () =>
      navigateAndClose(
        petId ? `${paths.accounting.new.getHref()}?petId=${petId}` : paths.accounting.new.getHref(),
        {
          appointmentId,
        },
      ),
    [petId, appointmentId, navigateAndClose],
  );

  const handleOpenOwnerDetail = useCallback(() => {
    if (ownerId) {
      navigateAndClose(paths.owners.detail.getHref(ownerId));
      return;
    }
    // /pets/:id 相当のページは存在しないため、ownerId 未連携のデータは
    // 飼主一覧の name 検索へ逃がす（backend は name/name_kana/phone/email を検索）。
    const params = new URLSearchParams();
    if (ownerName) params.set("search", ownerName);
    const query = params.toString();
    navigateAndClose(query ? `${paths.owners.getHref()}?${query}` : paths.owners.getHref());
  }, [ownerId, ownerName, navigateAndClose]);

  if (!appointment) return null;

  const isTrimming = appointment.reservationCategory === "trimming";
  const isHospitalization = isHospitalizationReservationType(appointment.reservationType);
  const isMedical = !isTrimming && !isHospitalization;
  const canOpenMedicalRecordFromRelatedPages =
    isMedical && canCreateMedicalRecord === true && currentStatus === "診療中";

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent
        className={`${LAYOUT.modal.sm} max-h-[85vh] p-0 gap-0 overflow-hidden ${C.bgWhite} flex flex-col`}
      >
        <ReceptionDialogHeader appointment={appointment} currentStatus={currentStatus} />
        <ReceptionDialogBody
          appointment={appointment}
          isTrimming={isTrimming}
          onCreateMedicalRecord={handleCreateMedicalRecord}
          onCreateTrimming={handleCreateTrimming}
          onCreateAccounting={handleCreateAccounting}
          onCreateHospitalization={handleCreateHospitalization}
          canCreateMedicalRecord={canOpenMedicalRecordFromRelatedPages}
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
