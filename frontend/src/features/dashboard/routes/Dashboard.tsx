// React/Framework
import { useState, useCallback, useMemo, lazy, Suspense } from "react";
import { useNavigate } from "react-router";

// External
import { DndContext, PointerSensor, useSensor, useSensors, DragOverEvent, DragEndEvent, closestCorners } from "@dnd-kit/core";
import Filter from "lucide-react/dist/esm/icons/filter";
import { toast } from "sonner";
import { format, addHours } from "date-fns";
import { ja } from "date-fns/locale";

// Internal
import { C, STYLE } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FormHeader } from "@/components/shared/Form";

// Shared
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";

// Lazy-loaded modals — only loaded when first opened (bundle splitting)
const DashboardDetailModal = lazy(() =>
  import("../components/DashboardDetailModal").then(m => ({ default: m.DashboardDetailModal }))
);
const ReservationFormModal = lazy(() =>
  import("@/components/shared/ReservationFormModal").then(m => ({ default: m.ReservationFormModal }))
);
import { KanbanColumn } from "../components/KanbanColumn";
import { useDashboardKanban } from "../hooks/useDashboardKanban";

// Types
import type { Appointment, ReservationAppointment, Pet } from "@/types";

// Columns that don't show the "add" button — Set for O(1) lookup
const NO_ADD_BUTTON_COLUMNS = new Set(["診療中", "会計待ち", "会計済"]);

// Module-level constant — stable across renders, not recreated on each call
const DOCTORS = [
  { id: "all", name: "全て" },
  { id: "医師A", name: "医師A" },
  { id: "医師B", name: "医師B" },
  { id: "医師C", name: "医師C" },
  { id: "医師指名なし", name: "医師指名なし" },
] as const;

export function Dashboard() {
    const navigate = useNavigate();
    const {
        filteredColumns,
        moveCard,
        advanceStatus,
        cancelAppointment,
        updateAppointment,
        filters
    } = useDashboardKanban();

    const [modalOpen, setModalOpen] = useState(false);
    const [selectedAppointment, setSelectedAppointment] = useState<Appointment | null>(null);

    // Edit Modal State
    const [isEditModalOpen, setIsEditModalOpen] = useState(false);
    const [editingAppointment, setEditingAppointment] = useState<Partial<ReservationAppointment> | null>(null);

    // Cancel Confirm Dialog State
    const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false);
    const [cancelTarget, setCancelTarget] = useState<Appointment | null>(null);

    const [isFilterOpen, setIsFilterOpen] = useState(false);

    const sensors = useSensors(
        useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
    );

    const findColumnByCardId = useCallback((cardId: string) => {
        return filteredColumns.find(col => col.appointments.some(a => a.id === cardId));
    }, [filteredColumns]);

    const handleDragOver = useCallback((event: DragOverEvent) => {
        const { active, over } = event;
        if (!over) return;

        const activeId = active.id as string;
        const overId = over.id as string;

        // Determine source and target columns
        const sourceColumn = findColumnByCardId(activeId);
        if (!sourceColumn) return;
        const sourceTitle = sourceColumn.title;

        let targetTitle: string;
        let hoverIndex: number;

        if (overId.startsWith("column-")) {
            // Dropped on column itself (empty area)
            targetTitle = overId.replace("column-", "");
            const targetCol = filteredColumns.find(c => c.title === targetTitle);
            hoverIndex = targetCol ? targetCol.appointments.length : 0;
        } else {
            // Dropped on another card
            const targetColumn = findColumnByCardId(overId);
            if (!targetColumn) return;
            targetTitle = targetColumn.title;
            hoverIndex = targetColumn.appointments.findIndex(a => a.id === overId);
        }

        if (sourceTitle === targetTitle && activeId === overId) return;

        const dragIndex = sourceColumn.appointments.findIndex(a => a.id === activeId);
        if (dragIndex === -1) return;

        // Same column, same position
        if (sourceTitle === targetTitle && dragIndex === hoverIndex) return;

        moveCard(dragIndex, hoverIndex, sourceTitle, targetTitle);
    }, [filteredColumns, findColumnByCardId, moveCard]);

    const handleDragEnd = useCallback((event: DragEndEvent) => {
        const { active, over } = event;
        if (!over) return;

        const activeId = active.id as string;
        const overId = over.id as string;

        if (activeId === overId) return;

        const sourceColumn = findColumnByCardId(activeId);
        if (!sourceColumn) return;

        let targetTitle: string;
        let hoverIndex: number;

        if (overId.startsWith("column-")) {
            targetTitle = overId.replace("column-", "");
            const targetCol = filteredColumns.find(c => c.title === targetTitle);
            hoverIndex = targetCol ? targetCol.appointments.length : 0;
        } else {
            const targetColumn = findColumnByCardId(overId);
            if (!targetColumn) return;
            targetTitle = targetColumn.title;
            hoverIndex = targetColumn.appointments.findIndex(a => a.id === overId);
        }

        const dragIndex = sourceColumn.appointments.findIndex(a => a.id === activeId);
        if (dragIndex === -1) return;

        if (sourceColumn.title === targetTitle && dragIndex === hoverIndex) return;

        moveCard(dragIndex, hoverIndex, sourceColumn.title, targetTitle);
    }, [filteredColumns, findColumnByCardId, moveCard]);

    const todayLabel = format(new Date(), "yyyy年M月d日 (E)", { locale: ja });

    const handleAddClick = useCallback((columnTitle: string) => {
        if (columnTitle === "受付予約") {
            navigate("/reservations");
        } else {
            toast.info("新規登録", {
                description: "新規予約・受付画面へ移動します。",
                duration: 2000
            });
            navigate("/reservations");
        }
    }, [navigate]);

    const addClickHandlers = useMemo(() => {
        const handlers = new Map<string, (() => void) | undefined>();
        for (const column of filteredColumns) {
            handlers.set(
                column.title,
                NO_ADD_BUTTON_COLUMNS.has(column.title) ? undefined : () => handleAddClick(column.title)
            );
        }
        return handlers;
    }, [filteredColumns, handleAddClick]);

    const handleCardClick = useCallback((appointment: Appointment) => {
        setSelectedAppointment(appointment);
        setModalOpen(true);
    }, []);

    const handleAdvanceStatus = useCallback(() => {
        if (!selectedAppointment) return;
        advanceStatus(selectedAppointment);
        setModalOpen(false);
        setSelectedAppointment(null);
    }, [selectedAppointment, advanceStatus]);

    const handleEditAppointment = useCallback((appointment: Appointment) => {
        // Appointment を ReservationAppointment のフォームデータに変換
        const now = new Date();
        const [hours, minutes] = appointment.time.split(':').map(Number);
        const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), hours, minutes);

        const reservationFormData: Partial<ReservationAppointment> = {
            id: appointment.id,
            start: start,
            end: addHours(start, 1), // Default 1 hour duration
            status: "confirmed",
            visitType: appointment.visitType === "初診" ? "first" : "revisit",
            type: appointment.serviceType,
            doctor: appointment.doctor || "医師A",
            isDesignated: appointment.isDesignated || false,
            petId: appointment.petId,
            ownerName: appointment.ownerName,
            petName: appointment.petName
        };

        setEditingAppointment(reservationFormData);
        setIsEditModalOpen(true);
        setModalOpen(false);
    }, []);

    const handleEditSave = useCallback((data: Partial<ReservationAppointment>, selectedPets: Pet[]) => {
        if (!editingAppointment?.id || !data.start) return;

        const updatedTime = format(data.start, "HH:mm");

        const updatedAppointment: Appointment = {
            id: editingAppointment.id,
            time: updatedTime,
            ownerName: selectedPets[0]?.ownerName || data.ownerName || "",
            petName: selectedPets[0]?.name || data.petName || "",
            petType: selectedPets[0]?.species || "犬",
            visitType: data.visitType === "first" ? "初診" : "再診",
            serviceType: data.type || "診療",
            doctor: data.doctor,
            isDesignated: data.isDesignated,
            petId: selectedPets[0]?.id || data.petId,
            // Other fields are preserved by updateAppointment merging
        } as Appointment;

        updateAppointment(updatedAppointment);
        setIsEditModalOpen(false);
        setEditingAppointment(null);
    }, [editingAppointment, updateAppointment]);

    const handleCancelAppointment = useCallback((appointment: Appointment) => {
        setCancelTarget(appointment);
        setCancelConfirmOpen(true);
    }, []);

    const executeCancel = useCallback(() => {
        if (!cancelTarget) return;
        cancelAppointment(cancelTarget.id);
        toast.success("予約を取り消しました");
        setModalOpen(false);
        setCancelConfirmOpen(false);
        setCancelTarget(null);
    }, [cancelTarget, cancelAppointment]);

    return (
        <div className={`flex-1 flex flex-col h-full ${C.bgPage}`}>
            <FormHeader
                title="当日の受付"
                description={`${todayLabel} - 受付状況をリアルタイムで確認`}
                action={
                    <div className="flex gap-2">
                        <Button
                            variant={isFilterOpen ? "secondary" : "outline"}
                            className={`gap-2 bg-white h-11 text-base tracking-[var(--tracking-notion)] ${C.text} ${C.borderMedium}`}
                            onClick={() => setIsFilterOpen(prev => !prev)}
                        >
                            <Filter className="size-[17.5px]" />
                            フィルター
                        </Button>
                        <Button
                            className={`${STYLE.confirmPrimary} h-11 text-base tracking-[var(--tracking-notion)]`}
                            onClick={() => navigate("/reservations")}
                        >
                            新規予約
                        </Button>
                    </div>
                }
            />

            {isFilterOpen && (
                <div className="bg-white border-b border-border px-6 py-4 animate-in slide-in-from-top-1 fade-in duration-200">
                    <div className="flex flex-wrap gap-8">
                        {/* Visit Type */}
                        <div className="space-y-2">
                            <h4 className={`font-bold text-sm ${C.text}`}>診察区分</h4>
                            <div className="flex gap-4">
                                {["初診", "再診"].map(type => (
                                    <div key={type} className="flex items-center space-x-2">
                                        <Checkbox
                                            id={`visit-${type}`}
                                            checked={filters.selectedVisitTypes.includes(type)}
                                            onCheckedChange={() => filters.toggleVisitType(type)}
                                            className="size-4"
                                        />
                                        <Label htmlFor={`visit-${type}`} className={`text-sm font-normal cursor-pointer ${C.text}`}>{type}</Label>
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* Doctor/Designation Selection */}
                        <div className="space-y-2">
                            <h4 className={`font-bold text-sm ${C.text}`}>指名</h4>
                            <Select value={filters.selectedDoctor} onValueChange={filters.setSelectedDoctor}>
                                <SelectTrigger className="w-[200px] h-10 text-sm bg-white border-input">
                                    <SelectValue placeholder="指名を選択" />
                                </SelectTrigger>
                                <SelectContent>
                                    {DOCTORS.map((doctor) => (
                                        <SelectItem key={doctor.id} value={doctor.id}>
                                            {doctor.name}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        {/* Trimming */}
                        <div className="space-y-2">
                            <h4 className={`font-bold text-sm ${C.text}`}>種類</h4>
                            <div className="flex items-center space-x-2 pt-0.5">
                                <Checkbox
                                    id="trimming-only"
                                    checked={filters.isTrimmingOnly}
                                    onCheckedChange={(c) => filters.setIsTrimmingOnly(!!c)}
                                    className="size-4"
                                />
                                <Label htmlFor="trimming-only" className={`text-sm font-normal cursor-pointer ${C.text}`}>トリミングのみ表示</Label>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            <div className="flex-1 overflow-hidden p-[20px] pt-4">
                <DndContext sensors={sensors} collisionDetection={closestCorners} onDragOver={handleDragOver} onDragEnd={handleDragEnd}>
                    {/* タブレット: 2-3列グリッド、デスクトップ: 5列flex */}
                    <div className="grid grid-cols-2 md:grid-cols-3 lg:flex gap-2 h-full w-full overflow-y-auto lg:overflow-x-auto lg:overflow-y-hidden pb-2 bg-transparent">
                        {filteredColumns.map((column) => (
                            <KanbanColumn
                                key={column.title}
                                data={column}
                                onAddClick={addClickHandlers.get(column.title)}
                                onCardClick={handleCardClick}
                            />
                        ))}
                    </div>
                </DndContext>
            </div>

            {modalOpen && (
              <Suspense fallback={null}>
                <DashboardDetailModal
                    isOpen={modalOpen}
                    onClose={() => setModalOpen(false)}
                    onConfirm={handleAdvanceStatus}
                    onEdit={handleEditAppointment}
                    onCancel={handleCancelAppointment}
                    appointment={selectedAppointment}
                    currentStatus={selectedAppointment ? filteredColumns.find(c => c.appointments.some(a => a.id === selectedAppointment.id))?.title : undefined}
                />
              </Suspense>
            )}

            {isEditModalOpen && (
              <Suspense fallback={null}>
                <ReservationFormModal
                    isOpen={isEditModalOpen}
                    onClose={() => {
                        setIsEditModalOpen(false);
                        setEditingAppointment(null);
                    }}
                    onSave={handleEditSave}
                    initialData={editingAppointment}
                />
              </Suspense>
            )}

            <ConfirmDialog
                open={cancelConfirmOpen}
                onClose={() => {
                    setCancelConfirmOpen(false);
                    setCancelTarget(null);
                }}
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
