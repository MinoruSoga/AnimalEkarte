// External
import { useDroppable } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { Plus } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { getDashboardColumnColor } from "@/utils/status-helpers";

// Relative
import { AppointmentCard } from "./AppointmentCard";

// Types
import type { Appointment, ColumnData } from "@/types";

interface KanbanColumnProps {
  data: ColumnData;
  onAddClick?: () => void;
  onCardClick: (appointment: Appointment) => void;
}

export const KanbanColumn = ({
  data,
  onAddClick,
  onCardClick
}: KanbanColumnProps) => {
  const colors = getDashboardColumnColor(data.title);

  const { setNodeRef, isOver } = useDroppable({
    id: `column-${data.title}`,
    data: { columnTitle: data.title },
  });

  const itemIds = data.appointments.map(a => a.id);

  return (
    <div
      ref={setNodeRef}
      className={`${colors.bg} rounded-md p-2 flex flex-col gap-3 w-full lg:flex-1 lg:min-w-[200px] xl:min-w-[240px] transition-colors shadow-sm ${isOver ? 'ring-2 ring-primary/20 bg-accent/50' : ''}`}
    >
      <div className="flex items-center gap-2 px-1 min-h-[44px]">
        <div className={`size-2.5 rounded-full ${colors.dot}`} />
        <h3 className={`text-sm font-bold ${colors.text}`}>{data.title}</h3>
        <span className="text-sm text-muted-foreground ml-auto">{data.appointments.length}</span>
      </div>
      <SortableContext items={itemIds} strategy={verticalListSortingStrategy}>
        <div className="flex flex-col gap-2 flex-1 overflow-y-auto min-h-[100px] lg:min-h-[50px]">
          {data.appointments.map((appointment) => (
            <AppointmentCard
              key={appointment.id}
              appointment={appointment}
              columnTitle={data.title}
              onCardClick={onCardClick}
            />
          ))}
          {/* Spacer for drop target at end of column */}
          <div className="flex-1 min-h-[2rem]" />
        </div>
      </SortableContext>
      {onAddClick && (
        <Button
          variant="ghost"
          size="sm"
          className="w-full justify-start gap-2 h-11 min-h-[44px] text-muted-foreground hover:bg-accent hover:text-accent-foreground px-2"
          onClick={onAddClick}
        >
          <Plus className="size-5" />
          <span className="text-sm">新規追加</span>
        </Button>
      )}
    </div>
  );
};
