import { Suspense, type ComponentProps, type ReactNode } from "react";
import { DndContext, type DragEndEvent } from "@dnd-kit/core";
import type { CollisionDetection } from "@dnd-kit/core";
import Filter from "lucide-react/dist/esm/icons/filter";
import { C } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { FormHeader } from "@/components/shared/Form/FormHeader";
import { PermissionBadges } from "@/components/shared/PermissionBadges/PermissionBadges";
import { ResourceReception } from "@/types/generated/models";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { ReceptionFilterPanel } from "../components/ReceptionFilterPanel";
import { ReceptionTelemetryStrip } from "../components/ReceptionTelemetryStrip";
import type { ReceptionAppointment } from "../api/types";
import type { ReceptionTelemetryResult } from "../hooks/use-reception-telemetry";
import { ReceptionDetailModal, ReservationFormModal } from "./ReceptionLazyModals";
import type { Reservation } from "@/types";

interface ReceptionPageBodyProps {
  todayLabel: string;
  isFilterOpen: boolean;
  onToggleFilter: () => void;
  canCreateReservation: boolean;
  onNewReception: () => void;
  telemetry: ReceptionTelemetryResult;
  selectedVisitTypes: string[];
  selectedDoctor: string;
  doctors: { id: string; name: string }[];
  isTrimmingOnly: boolean;
  onToggleVisitType: (type: string) => void;
  onSelectedDoctorChange: (doctor: string) => void;
  onTrimmingOnlyChange: (value: boolean) => void;
  sensors: ComponentProps<typeof DndContext>["sensors"];
  collisionDetection: CollisionDetection;
  onDragEnd: (event: DragEndEvent) => void;
  isUpdatingStatus: boolean;
  columnElements: ReactNode;
  modalOpen: boolean;
  onCloseModal: () => void;
  onConfirm: (() => void) | undefined;
  onEdit: ((appointment: ReceptionAppointment) => void) | undefined;
  onCancel: ((appointment: ReceptionAppointment) => void) | undefined;
  selectedAppointment: ReceptionAppointment | null;
  currentStatus: string | undefined;
  canCreateMedicalRecord: boolean;
  canCreateAccounting: boolean;
  canCreateHospitalization: boolean;
  isEditModalOpen: boolean;
  onCloseEditModal: () => void;
  onEditSave: ComponentProps<typeof ReservationFormModal>["onSave"];
  editingAppointment: Partial<Reservation> | null;
  canEditReservation: boolean;
  cancelConfirmOpen: boolean;
  onCloseCancelConfirm: () => void;
  onConfirmCancel: () => void;
  cancelDescription: string;
}

export function ReceptionPageBody({
  todayLabel,
  isFilterOpen,
  onToggleFilter,
  canCreateReservation,
  onNewReception,
  telemetry,
  selectedVisitTypes,
  selectedDoctor,
  doctors,
  isTrimmingOnly,
  onToggleVisitType,
  onSelectedDoctorChange,
  onTrimmingOnlyChange,
  sensors,
  collisionDetection,
  onDragEnd,
  isUpdatingStatus,
  columnElements,
  modalOpen,
  onCloseModal,
  onConfirm,
  onEdit,
  onCancel,
  selectedAppointment,
  currentStatus,
  canCreateMedicalRecord,
  canCreateAccounting,
  canCreateHospitalization,
  isEditModalOpen,
  onCloseEditModal,
  onEditSave,
  editingAppointment,
  canEditReservation,
  cancelConfirmOpen,
  onCloseCancelConfirm,
  onConfirmCancel,
  cancelDescription,
}: ReceptionPageBodyProps) {
  return (
    <div className={`flex-1 flex flex-col h-full ${C.bgPage}`}>
      <FormHeader
        title="当日の受付"
        description={`${todayLabel} - 受付状況をリアルタイムで確認`}
        action={
          <div className="flex flex-wrap items-center gap-2">
            <PermissionBadges resource={ResourceReception} />
            <Button
              variant={isFilterOpen ? "secondary" : "outline"}
              className={`gap-2 ${C.bgWhite} h-11 text-base ${C.text} ${C.borderMedium}`}
              onClick={onToggleFilter}
            >
              <Filter className="size-[17.5px]" />
              フィルター
            </Button>
            {canCreateReservation === true ? (
              <Button
                className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} rounded-full px-4 shadow-none border-transparent h-11 text-base`}
                onClick={onNewReception}
              >
                新規予約登録
              </Button>
            ) : null}
          </div>
        }
      />

      <ReceptionTelemetryStrip totalCount={telemetry.totalCount} waitStats={telemetry} />

      {isFilterOpen ? (
        <ReceptionFilterPanel
          selectedVisitTypes={selectedVisitTypes}
          selectedDoctor={selectedDoctor}
          doctors={doctors}
          isTrimmingOnly={isTrimmingOnly}
          onToggleVisitType={onToggleVisitType}
          onSelectedDoctorChange={onSelectedDoctorChange}
          onTrimmingOnlyChange={onTrimmingOnlyChange}
        />
      ) : null}

      <div className="flex-1 overflow-hidden p-6 pt-4">
        {/* 確定はドロップ時のみ。onDragOver でのライブ移動は measureRects の再計測ループ
            と通過時の誤ステータス API 発火を招くため撤去（commit-on-drop） */}
        <DndContext sensors={sensors} collisionDetection={collisionDetection} onDragEnd={onDragEnd}>
          {/* タブレット: 2-3列グリッド、デスクトップ: 5列flex */}
          <div
            className={`grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:flex gap-4 h-full w-full overflow-y-auto lg:overflow-x-auto lg:overflow-y-hidden pb-2 bg-transparent transition-opacity${isUpdatingStatus ? " opacity-70 pointer-events-none" : ""}`}
          >
            {columnElements}
          </div>
        </DndContext>
      </div>

      {modalOpen ? (
        <Suspense fallback={null}>
          <ReceptionDetailModal
            isOpen={modalOpen}
            onClose={onCloseModal}
            onConfirm={onConfirm}
            onEdit={onEdit}
            onCancel={onCancel}
            appointment={selectedAppointment}
            currentStatus={currentStatus}
            canCreateMedicalRecord={canCreateMedicalRecord}
            canCreateAccounting={canCreateAccounting}
            canCreateHospitalization={canCreateHospitalization}
          />
        </Suspense>
      ) : null}

      {isEditModalOpen ? (
        <Suspense fallback={null}>
          <ReservationFormModal
            isOpen={isEditModalOpen}
            onClose={onCloseEditModal}
            onSave={onEditSave}
            initialData={editingAppointment}
            canCreate={canCreateReservation}
            canEdit={canEditReservation}
          />
        </Suspense>
      ) : null}

      <ConfirmDialog
        open={cancelConfirmOpen}
        onClose={onCloseCancelConfirm}
        onConfirm={onConfirmCancel}
        title="予約を取り消しますか？"
        description={cancelDescription}
        confirmLabel="取り消す"
        cancelLabel="キャンセル"
        variant="destructive"
      />
    </div>
  );
}
