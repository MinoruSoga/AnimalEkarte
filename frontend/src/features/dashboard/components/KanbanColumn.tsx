import { useDrop } from "react-dnd";
import { Plus } from "lucide-react";
import { Button } from "../../../components/ui/button";
import { AppointmentCard } from "./AppointmentCard";
import { Appointment, ColumnData } from "../../../types";
import { getDashboardColumnColor } from "../../../lib/status-helpers";

interface KanbanColumnProps {
  data: ColumnData;
  moveCard: (dragIndex: number, hoverIndex: number, sourceColumn: string, targetColumn: string) => boolean;
  onAddClick?: () => void;
  onCardClick: (appointment: Appointment) => void;
}

export const KanbanColumn = ({ 
  data, 
  moveCard,
  onAddClick,
  onCardClick
}: KanbanColumnProps) => {
  const colors = getDashboardColumnColor(data.title);

  const [{ isOver }, drop] = useDrop(() => ({
    accept: 'appointment',
    drop: (item: any, monitor) => {
      if (monitor.didDrop()) {
        return;
      }
      
      const sourceColumn = item.columnTitle;
      const targetColumn = data.title;

      if (sourceColumn !== targetColumn) {
        if (moveCard(item.index, data.appointments.length, sourceColumn, targetColumn)) {
          item.columnTitle = targetColumn;
          item.index = data.appointments.length;
        }
      }
    },
    // Main column drop only handles empty state or fallback
    hover: (item: any, monitor) => {
      if (!monitor.isOver({ shallow: true })) return;
      if (data.appointments.length > 0) return;

      const sourceColumn = item.columnTitle;
      const targetColumn = data.title;

      if (sourceColumn !== targetColumn) {
        if (moveCard(item.index, data.appointments.length, sourceColumn, targetColumn)) {
          item.columnTitle = targetColumn;
          item.index = data.appointments.length;
        }
      }
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
    }),
  }));

  // Dedicated drop zone for the bottom of the list
  // This handles "dropping at the end" without the flickering issues of using the main column wrapper
  const [{ isOverBottom }, dropBottom] = useDrop(() => ({
    accept: 'appointment',
    hover: (item: any, monitor) => {
      if (!monitor.isOver({ shallow: true })) return;

      const sourceColumn = item.columnTitle;
      const targetColumn = data.title;
      const targetIndex = data.appointments.length;

      // Optimization: if we are already targeting the end of this column
      // For same column: if item is at last position (length - 1)
      if (sourceColumn === targetColumn && item.index === targetIndex - 1) return;
      
      // For cross column: if item is already logically at the end (index == length)
      // (This happens if we just moved it here and haven't re-rendered yet)
      if (sourceColumn !== targetColumn && item.columnTitle === targetColumn && item.index === targetIndex) return;

      if (moveCard(item.index, targetIndex, sourceColumn, targetColumn)) {
        item.columnTitle = targetColumn;
        // If same column, the item ends up at length - 1 (because it was removed from list)
        // If different column, it ends up at length (appended to existing items)
        // Note: data.appointments.length is the length BEFORE the move for cross-column (sort of, until re-render)
        // For same column, it includes the item.
        item.index = sourceColumn === targetColumn ? targetIndex - 1 : targetIndex;
      }
    },
    collect: (monitor) => ({
      isOverBottom: monitor.isOver(),
    }),
  }));

  return (
    <div 
      ref={drop}
      className={`${colors.bg} rounded-md p-2 flex flex-col gap-3 flex-1 min-w-[280px] transition-colors shadow-sm ${isOver ? 'ring-2 ring-primary/20 bg-accent/50' : ''}`}
    >
      <div className="flex items-center gap-2 px-1">
        <div className={`size-2 rounded-full ${colors.dot}`} />
        <h3 className={`text-sm font-bold ${colors.text}`}>{data.title}</h3>
        <span className="text-sm text-muted-foreground ml-auto">{data.appointments.length}</span>
      </div>
      <div className="flex flex-col gap-2 flex-1 overflow-y-auto min-h-[50px]">
        {data.appointments.map((appointment, index) => (
          <AppointmentCard 
            key={appointment.id} 
            appointment={appointment} 
            columnTitle={data.title}
            index={index}
            moveCard={moveCard}
            onCardClick={onCardClick}
          />
        ))}
        {/* Drop zone for end of list - fills remaining space */}
        <div 
          ref={dropBottom}
          className={`flex-1 min-h-[2rem] transition-colors rounded-sm ${isOverBottom ? 'bg-primary/5' : 'bg-transparent'}`} 
        />
      </div>
      {onAddClick && (
        <Button 
          variant="ghost" 
          size="sm"
          className="w-full justify-start gap-2 h-10 text-muted-foreground hover:bg-accent hover:text-accent-foreground -mx-1 px-2"
          onClick={onAddClick}
        >
          <Plus className="size-4" />
          <span className="text-sm">新規追加</span>
        </Button>
      )}
    </div>
  );
};
