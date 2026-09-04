import { C, ICON } from "@/lib/design-tokens";
import { memo, useMemo } from "react";
// External
import { useDroppable } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import Plus from "lucide-react/dist/esm/icons/plus";

// Internal
import { Button } from "@/components/ui/button";
import { getReceptionColumnColor } from "@/lib/status-helpers";

// Relative
import { AppointmentCard } from "./AppointmentCard";

// Types
import type { ColumnData } from "@/types";
import type { ReceptionAppointment } from "../api/types";

interface KanbanColumnProps {
  data: ColumnData;
  onAddClick?: () => void;
  onCardClick: (appointment: ReceptionAppointment) => void;
  onRecordOpen?: (appointment: ReceptionAppointment, columnTitle: string) => void;
}

export const KanbanColumn = memo(function KanbanColumn({
  data,
  onAddClick,
  onCardClick,
  onRecordOpen,
}: KanbanColumnProps) {
  const colors = getReceptionColumnColor(data.title);

  const { setNodeRef, isOver } = useDroppable({
    id: `column-${data.title}`,
    data: { columnTitle: data.title },
  });

  const itemIds = useMemo(() => data.appointments.map((appt) => appt.id), [data.appointments]);

  const appointmentItems = useMemo(
    () =>
      data.appointments.map((appointment) => (
        <AppointmentCard
          key={appointment.id}
          appointment={appointment}
          columnTitle={data.title}
          onCardClick={onCardClick}
          onRecordOpen={onRecordOpen}
        />
      )),
    [data.appointments, data.title, onCardClick, onRecordOpen],
  );

  return (
    <div
      ref={setNodeRef}
      className={`${colors.bg} rounded-md p-2 flex flex-col gap-3 w-full lg:flex-1 lg:min-w-[220px] xl:min-w-[280px] transition-colors ${isOver ? `ring-2 ${C.ringPrimary20} ${C.bgActive}` : ""}`}
      role="region"
      aria-label={`${data.title} — ${data.appointments.length}件`}
    >
      <div className="flex items-center gap-2 px-1 py-1">
        <div className={`${ICON.dot} rounded-full ${colors.dot}`} aria-hidden="true" />
        <h3 className={`text-base font-bold ${colors.text}`}>{data.title}</h3>
        <span
          className={`text-base ${colors.text60} ml-auto`}
          aria-live="polite"
          aria-atomic="true"
        >
          {data.appointments.length}
        </span>
      </div>
      <SortableContext items={itemIds} strategy={verticalListSortingStrategy}>
        <div className="flex flex-col gap-2 flex-1 overflow-y-auto min-h-[100px] lg:min-h-[50px]">
          {appointmentItems}
          {/* Spacer for drop target at end of column */}
          <div className="flex-1 min-h-[2rem]" />
        </div>
      </SortableContext>
      {onAddClick ? (
        <Button
          variant="ghost"
          size="sm"
          className={`w-full justify-start gap-2 h-10 ${colors.text60} ${colors.hoverBgPage} ${colors.hoverText} -mx-1 px-2`}
          onClick={onAddClick}
        >
          <Plus className={ICON.page} />
          <span className="text-sm">新規追加</span>
        </Button>
      ) : null}
    </div>
  );
});
