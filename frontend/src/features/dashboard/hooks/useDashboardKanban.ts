import { useState, useMemo, useRef } from "react";
import { toast } from "sonner@2.0.3";
import { Appointment, ColumnData } from "../../../types";
import { INITIAL_DASHBOARD_COLUMNS } from "../../../lib/constants";

export const useDashboardKanban = () => {
  const [columns, setColumns] = useState<ColumnData[]>(INITIAL_DASHBOARD_COLUMNS);
  
  // Filter States
  const [selectedVisitTypes, setSelectedVisitTypes] = useState<string[]>(["初診", "再診"]);
  const [selectedDoctor, setSelectedDoctor] = useState<string>("all");
  const [isTrimmingOnly, setIsTrimmingOnly] = useState(false);
  
  const lastAlertRef = useRef(0);

  // Filter Logic
  const filteredColumns = useMemo(() => {
    return columns.map(col => ({
        ...col,
        appointments: col.appointments.filter(app => {
            // 1. Visit Type Filter
            if (!selectedVisitTypes.includes(app.visitType)) return false;

            // 2. Doctor/Designation Filter
            if (selectedDoctor !== "all") {
                if (selectedDoctor === "医師指名なし") {
                    // Show only non-designated appointments
                    if (app.isDesignated) return false;
                } else {
                    // Show specific designated doctor
                    if (app.doctor !== selectedDoctor || !app.isDesignated) return false;
                }
            }

            // 3. Trimming Filter
            if (isTrimmingOnly) {
                if (!app.serviceType.includes("トリミング")) return false;
            }

            return true;
        })
    }));
  }, [columns, selectedVisitTypes, selectedDoctor, isTrimmingOnly]);

  const toggleVisitType = (type: string) => {
      setSelectedVisitTypes(prev => 
          prev.includes(type) ? prev.filter(t => t !== type) : [...prev, type]
      );
  };

  const moveCard = (dragIndex: number, hoverIndex: number, sourceColumn: string, targetColumn: string) => {
    // Validation: Check if moving from "受付済" to "診療中"
    if (sourceColumn === "受付済" && targetColumn === "診療中") {
        const now = Date.now();
        if (now - lastAlertRef.current > 3000) {
            toast.error("カルテ作成が必要です", {
                description: "このステータスに変更するには、詳細画面からカルテを作成してください。",
                duration: 4000,
            });
            lastAlertRef.current = now;
        }
        return false;
    }

    // Determine the actual card being dragged based on the filtered view
    const sourceColFiltered = filteredColumns.find(col => col.title === sourceColumn);
    const targetColFiltered = filteredColumns.find(col => col.title === targetColumn);
    
    if (!sourceColFiltered || !targetColFiltered) return false;
    
    // The card being dragged
    const draggedCard = sourceColFiltered.appointments[dragIndex];
    if (!draggedCard) return false;

    setColumns((prevColumns) => {
      const newColumns = prevColumns.map((col) => ({
        ...col,
        appointments: [...col.appointments],
      }));

      const sourceCol = newColumns.find((col) => col.title === sourceColumn);
      const targetCol = newColumns.find((col) => col.title === targetColumn);

      if (!sourceCol || !targetCol) return prevColumns;

      // Find index of dragged card in the REAL source column
      const realDragIndex = sourceCol.appointments.findIndex(app => app.id === draggedCard.id);
      if (realDragIndex === -1) return prevColumns;

      // Remove from source
      sourceCol.appointments.splice(realDragIndex, 1);

      // Find insertion index in the REAL target column
      let realHoverIndex = targetCol.appointments.length; // Default to end

      // If inserting into a specific position (not end)
      if (hoverIndex < targetColFiltered.appointments.length) {
          const referenceCard = targetColFiltered.appointments[hoverIndex];
          const refIndex = targetCol.appointments.findIndex(app => app.id === referenceCard.id);
          if (refIndex !== -1) {
              realHoverIndex = refIndex;
          }
      }

      // Insert at calculated position
      targetCol.appointments.splice(realHoverIndex, 0, draggedCard);

      return newColumns;
    });
    return true;
  };

  const advanceStatus = (appointment: Appointment) => {
    const currentColumnTitle = filteredColumns.find(c => c.appointments.some(a => a.id === appointment.id))?.title;
    if (!currentColumnTitle) return;

    let nextColumnTitle = "";
    switch (currentColumnTitle) {
      case "受付予約": nextColumnTitle = "受付済"; break;
      case "受付済": nextColumnTitle = "診療中"; break;
      case "診療中": nextColumnTitle = "会計待ち"; break;
      case "会計待ち": nextColumnTitle = "会計済"; break;
      case "会計済": 
        // Remove from list
        setColumns(prev => {
            const newColumns = prev.map(col => ({ ...col, appointments: [...col.appointments] }));
            const sourceCol = newColumns.find(c => c.title === currentColumnTitle);
            
            if (sourceCol) {
                const cardIndex = sourceCol.appointments.findIndex(a => a.id === appointment.id);
                if (cardIndex > -1) {
                    sourceCol.appointments.splice(cardIndex, 1);
                    toast.success("手続きを完了し、リストから削除しました");
                }
            }
            return newColumns;
        });
        return;
      default: return;
    }

    // Move the card
    setColumns(prev => {
        const newColumns = prev.map(col => ({ ...col, appointments: [...col.appointments] }));
        const sourceCol = newColumns.find(c => c.title === currentColumnTitle);
        const targetCol = newColumns.find(c => c.title === nextColumnTitle);

        if (sourceCol && targetCol) {
            const cardIndex = sourceCol.appointments.findIndex(a => a.id === appointment.id);
            if (cardIndex > -1) {
                const [card] = sourceCol.appointments.splice(cardIndex, 1);
                targetCol.appointments.push(card);
                
                toast.success(`ステータスを「${nextColumnTitle}」に変更しました`, {
                    description: `${appointment.petName}ちゃんのステータスを更新しました。`
                });
            }
        }
        return newColumns;
    });
  };

  const cancelAppointment = (appointmentId: string) => {
    setColumns(prev => {
        const newColumns = prev.map(col => ({ ...col, appointments: [...col.appointments] }));
        for (const col of newColumns) {
            const index = col.appointments.findIndex(a => a.id === appointmentId);
            if (index > -1) {
                col.appointments.splice(index, 1);
                return newColumns;
            }
        }
        return prev;
    });
  };

  const updateAppointment = (updatedAppointment: Appointment) => {
    setColumns(prev => {
        const newColumns = prev.map(col => ({ ...col, appointments: [...col.appointments] }));
        for (const col of newColumns) {
            const index = col.appointments.findIndex(a => a.id === updatedAppointment.id);
            if (index > -1) {
                // Keep existing appointment data but overwrite with updates
                col.appointments[index] = {
                    ...col.appointments[index],
                    ...updatedAppointment
                };
                return newColumns;
            }
        }
        return prev;
    });
    toast.success("予約情報を更新しました");
  };

  return {
    columns,
    filteredColumns,
    moveCard,
    advanceStatus,
    cancelAppointment,
    updateAppointment,
    filters: {
        selectedVisitTypes,
        selectedDoctor,
        isTrimmingOnly,
        setSelectedDoctor,
        setIsTrimmingOnly,
        toggleVisitType
    }
  };
};
