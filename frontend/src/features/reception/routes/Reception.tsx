// React/Framework
import { useState, useCallback, useMemo, Suspense } from "react";
import { useNavigate } from "react-router";

// External
import { DndContext, PointerSensor, useSensor, useSensors, pointerWithin } from "@dnd-kit/core";
import Filter from "lucide-react/dist/esm/icons/filter";
import { format } from "date-fns";
import { ja } from "date-fns/locale";

// Internal
import { paths } from "@/config/paths";
import { C } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
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
import { ReceptionTelemetryStrip } from "../components/ReceptionTelemetryStrip";
import type { ReceptionAppointment } from "../api/types";
import { useReceptionKanban } from "../hooks/use-reception-kanban";
import { useReceptionTelemetry } from "../hooks/use-reception-telemetry";
import { ReceptionDetailModal, ReservationFormModal } from "./ReceptionLazyModals";
import { NO_ADD_BUTTON_COLUMNS } from "./reception-model";
import { useReceptionDragHandlers } from "../hooks/use-reception-drag-handlers";
import { useReceptionModalHandlers } from "../hooks/use-reception-modal-handlers";

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
    } = useReceptionKanban({
        canEditReservation,
        canDeleteReservation,
    });

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
        canEditReservation,
        canDeleteReservation,
    });

    const [isFilterOpen, setIsFilterOpen] = useState(false);

    const sensors = useSensors(
        useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
    );

    const { handleDragEnd } = useReceptionDragHandlers(columns, moveCard);

    // 受付ヘッダー テレメトリ（change-ui.md）: 集計は必ず columns（フィルタ非適用）から算出する。
    // filteredColumns を渡すと「本日受付」件数がフィルタ操作で変動してしまう。
    const telemetry = useReceptionTelemetry(columns);

    const todayLabel = format(toJSTWallDate(new Date()), "yyyy年M月d日 (E)", { locale: ja });

    const handleRecordOpen = useCallback((appointment: ReceptionAppointment, columnTitle: string) => {
        if (columnTitle === "受付済" && canEditReservation === true) {
            advanceStatus(appointment);
        }
    }, [advanceStatus, canEditReservation]);

    // 当日受付ページから新規予約作成モーダルを自動オープンする遷移ヘルパー。
    const goToNewReservation = useCallback((query: string) => {
        navigate(`${paths.reservations.getHref()}?${query}`, { state: { from: paths.home.getHref() } });
    }, [navigate]);

    // 受付予約ボード → 通常の新規予約（confirmed → 受付予約カラム）。
    // 受付済ボード → 受付 walk-in（checked_in → 受付済カラム、route=reception）。
    const handleAddClick = useCallback((columnTitle: string) => {
        goToNewReservation(columnTitle === "受付予約" ? "newReservation=1" : "reception=1");
    }, [goToNewReservation]);

    const addClickHandlers = useMemo(() => {
        const handlers = new Map<string, (() => void) | undefined>();
        // BUG-132: create 権限がない場合は「新規追加」ボタンを非表示
        if (canCreateReservation !== true) return handlers;
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
                    <div className="flex flex-wrap items-center gap-2">
                        <PermissionBadges resource={ResourceReception} />
                        <Button
                            variant={isFilterOpen ? "secondary" : "outline"}
                            className={`gap-2 ${C.bgWhite} h-11 text-base ${C.text} ${C.borderMedium}`}
                            onClick={() => setIsFilterOpen(prev => !prev)}
                        >
                            <Filter className="size-[17.5px]" />
                            フィルター
                        </Button>
                        {canCreateReservation === true ? (
                            <Button
                                className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} rounded-full px-4 shadow-none border-transparent h-11 text-base`}
                                onClick={() => goToNewReservation("reception=1")}
                            >
                                新規予約登録
                            </Button>
                        ) : null}
                    </div>
                }
            />

            <ReceptionTelemetryStrip
                totalCount={telemetry.totalCount}
                waitStats={telemetry}
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

            <div className="flex-1 overflow-hidden p-6 pt-4">
                {/* 確定はドロップ時のみ。onDragOver でのライブ移動は measureRects の再計測ループ
                    と通過時の誤ステータス API 発火を招くため撤去（commit-on-drop） */}
                <DndContext sensors={sensors} collisionDetection={pointerWithin} onDragEnd={handleDragEnd}>
                    {/* タブレット: 2-3列グリッド、デスクトップ: 5列flex */}
                    <div className={`grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:flex gap-4 h-full w-full overflow-y-auto lg:overflow-x-auto lg:overflow-y-hidden pb-2 bg-transparent transition-opacity${isUpdatingStatus ? " opacity-70 pointer-events-none" : ""}`}>
                        {columnElements}
                    </div>
                </DndContext>
            </div>

            {modalOpen ? (
              <Suspense fallback={null}>
                <ReceptionDetailModal
                    isOpen={modalOpen}
                    onClose={() => setModalOpen(false)}
                    onConfirm={canEditReservation === true ? handleAdvanceStatus : undefined}
                    onEdit={canEditReservation === true ? handleEditAppointment : undefined}
                    onCancel={canDeleteReservation === true ? handleCancelAppointment : undefined}
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
