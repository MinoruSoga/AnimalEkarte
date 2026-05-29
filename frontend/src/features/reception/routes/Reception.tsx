// React/Framework
import { useState, useCallback, useMemo, Suspense } from "react";
import { useNavigate } from "react-router";

// External
import { DndContext, PointerSensor, useSensor, useSensors, pointerWithin } from "@dnd-kit/core";
import Filter from "lucide-react/dist/esm/icons/filter";
import { toast } from "sonner";
import { format } from "date-fns";
import { ja } from "date-fns/locale";

// Internal
import { paths } from "@/config/paths";
import { C, STYLE } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { FormHeader } from "@/components/shared/Form/FormHeader";
import { PermissionBadges } from "@/components/shared/PermissionBadges/PermissionBadges";
import { ResourceReception, ResourceReservations, ResourceMedicalRecords, ResourceAccounting, ResourceHospitalization } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// Shared
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";

import { KanbanColumn } from "../components/KanbanColumn";
import { ReceptionFilterPanel } from "../components/ReceptionFilterPanel";
import type { ReceptionAppointment } from "../api/types";
import { useReceptionKanban } from "../hooks/use-reception-kanban";
import { ReceptionDetailModal, ReservationFormModal } from "./ReceptionLazyModals";
import { NO_ADD_BUTTON_COLUMNS } from "./ReceptionModel";
import { useReceptionDragHandlers } from "./useReceptionDragHandlers";
import { useReceptionModalHandlers } from "./useReceptionModalHandlers";

export function Reception() {
    const navigate = useNavigate();
    const { canCreate: canCreateReservation, canEdit: canEditReservation, canDelete: canDeleteReservation } = usePermission(ResourceReservations);
    const { canCreate: canCreateMedicalRecord } = usePermission(ResourceMedicalRecords);
    const { canCreate: canCreateAccounting } = usePermission(ResourceAccounting);
    const { canCreate: canCreateHospitalization } = usePermission(ResourceHospitalization);
    const {
        columns,
        filteredColumns,
        isUpdatingStatus,
        isLoading,
        isError,
        staffs,
        moveCard,
        advanceStatus,
        cancelAppointment,
        updateAppointment,
        filters
    } = useReceptionKanban();

    // スタッフAPIから医師フィルター選択肢を動的生成
    const doctors = useMemo(() => [
      { id: "all", name: "全て" },
      ...staffs.flatMap((s) => s.isActive ? [{ id: s.name, name: s.name }] : []),
      { id: "医師指名なし", name: "医師指名なし" },
    ], [staffs]);

    const {
        cancelConfirmOpen,
        cancelTarget,
        closeCancelConfirm,
        closeEditModal,
        editingAppointment,
        executeCancel,
        handleAdvanceStatus,
        handleCancelAppointment,
        handleCardClick,
        handleEditAppointment,
        handleEditSave,
        isEditModalOpen,
        modalOpen,
        selectedAppointment,
        setModalOpen,
    } = useReceptionModalHandlers({
        advanceStatus,
        cancelAppointment,
        updateAppointment,
    });

    const [isFilterOpen, setIsFilterOpen] = useState(false);

    const sensors = useSensors(
        useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
    );

    const { handleDragEnd, handleDragOver } = useReceptionDragHandlers(columns, moveCard);

    const todayLabel = format(new Date(), "yyyy年M月d日 (E)", { locale: ja });

    const handleRecordOpen = useCallback((appointment: ReceptionAppointment, columnTitle: string) => {
        if (columnTitle === "受付済" && canEditReservation) {
            advanceStatus(appointment);
        }
    }, [advanceStatus, canEditReservation]);

    const handleAddClick = useCallback((columnTitle: string) => {
        if (columnTitle === "受付予約") {
            navigate(paths.reservations.getHref());
        } else {
            toast.info("新規登録", {
                description: "新規予約・受付画面へ移動します。",
                duration: 2000
            });
            navigate(`${paths.reservations.getHref()}?reception=1`, { state: { from: paths.home.getHref() } });
        }
    }, [navigate]);

    const addClickHandlers = useMemo(() => {
        const handlers = new Map<string, (() => void) | undefined>();
        // BUG-132: create 権限がない場合は「新規追加」ボタンを非表示
        if (!canCreateReservation) return handlers;
        for (const column of filteredColumns) {
            handlers.set(
                column.title,
                NO_ADD_BUTTON_COLUMNS.has(column.title) ? undefined : () => handleAddClick(column.title)
            );
        }
        return handlers;
    }, [filteredColumns, handleAddClick, canCreateReservation]);

    const columnElements = useMemo(() =>
        filteredColumns.map((column) => (
            <KanbanColumn
                key={column.title}
                data={column}
                onAddClick={addClickHandlers.get(column.title)}
                onCardClick={handleCardClick}
                onRecordOpen={handleRecordOpen}
            />
        )),
        [filteredColumns, addClickHandlers, handleCardClick, handleRecordOpen]
    );

    // js-set-map-lookups: レンダーパスの O(n²) find+some を O(1) Map ルックアップへ変換
    const appointmentColumnTitleMap = useMemo(() => {
        const map = new Map<string, string>();
        for (const col of filteredColumns) {
            for (const apt of col.appointments) {
                map.set(apt.id, col.title);
            }
        }
        return map;
    }, [filteredColumns]);

    if (isLoading) {
        return <LoadingFallback />;
    }

    if (isError) {
        return <ErrorFallback message="受付データの取得に失敗しました" />;
    }

    return (
        <div className={`flex-1 flex flex-col h-full ${C.bgPage}`}>
            <FormHeader
                title="当日の受付"
                description={`${todayLabel} - 受付状況をリアルタイムで確認`}
                action={
                    <div className="flex items-center gap-2">
                        <PermissionBadges resource={ResourceReception} />
                        <Button
                            variant={isFilterOpen ? "secondary" : "outline"}
                            className={`gap-2 ${C.bgWhite} h-11 text-base tracking-[var(--tracking-notion)] ${C.text} ${C.borderMedium}`}
                            onClick={() => setIsFilterOpen(prev => !prev)}
                        >
                            <Filter className="size-[17.5px]" />
                            フィルター
                        </Button>
                        {canCreateReservation ? (
                            <Button
                                className={`${STYLE.confirmPrimary} h-11 text-base tracking-[var(--tracking-notion)]`}
                                onClick={() => navigate(`${paths.reservations.getHref()}?reception=1`, { state: { from: paths.home.getHref() } })}
                            >
                                新規予約登録
                            </Button>
                        ) : null}
                    </div>
                }
            />

            {isFilterOpen ? (
                <ReceptionFilterPanel
                    selectedVisitTypes={filters.selectedVisitTypes}
                    selectedDoctor={filters.selectedDoctor}
                    doctors={doctors}
                    isTrimmingOnly={filters.isTrimmingOnly}
                    onToggleVisitType={filters.toggleVisitType}
                    onSelectedDoctorChange={filters.setSelectedDoctor}
                    onTrimmingOnlyChange={filters.setIsTrimmingOnly}
                />
            ) : null}

            <div className="flex-1 overflow-hidden p-5 pt-4">
                <DndContext sensors={sensors} collisionDetection={pointerWithin} onDragOver={handleDragOver} onDragEnd={handleDragEnd}>
                    {/* タブレット: 2-3列グリッド、デスクトップ: 5列flex */}
                    <div className={`grid grid-cols-2 md:grid-cols-3 lg:flex gap-4 h-full w-full overflow-y-auto lg:overflow-x-auto lg:overflow-y-hidden pb-2 bg-transparent transition-opacity${isUpdatingStatus ? " opacity-70 pointer-events-none" : ""}`}>
                        {columnElements}
                    </div>
                </DndContext>
            </div>

            {modalOpen ? (
              <Suspense fallback={null}>
                <ReceptionDetailModal
                    isOpen={modalOpen}
                    onClose={() => setModalOpen(false)}
                    onConfirm={canEditReservation ? handleAdvanceStatus : undefined}
                    onEdit={canEditReservation ? handleEditAppointment : undefined}
                    onCancel={canDeleteReservation ? handleCancelAppointment : undefined}
                    appointment={selectedAppointment}
                    currentStatus={selectedAppointment ? appointmentColumnTitleMap.get(selectedAppointment.id) : undefined}
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
                    onClose={closeEditModal}
                    onSave={handleEditSave}
                    initialData={editingAppointment}
                    canCreate={canCreateReservation}
                    canEdit={canEditReservation}
                />
              </Suspense>
            ) : null}

            <ConfirmDialog
                open={cancelConfirmOpen}
                onClose={closeCancelConfirm}
                onConfirm={executeCancel}
                title="予約を取り消しますか？"
                description={cancelTarget ? `${cancelTarget.petName}（${cancelTarget.ownerName}）の予約を取り消します。` : ""}
                confirmLabel="取り消す"
                cancelLabel="キャンセル"
                variant="destructive"
            />
        </div>
    );
}
