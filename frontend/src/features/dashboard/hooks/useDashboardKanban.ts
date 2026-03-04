import { useState, useMemo, useRef, useEffect } from "react";
import { toast } from "sonner";
import type { Appointment, ColumnData } from "@/types";
import { useDashboardData, useUpdateAppointmentStatus, todayISO, COLUMN_TITLE_TO_STATUS } from "../api";
import type { DashboardColumn, DashboardAppointment } from "../api";

/** DashboardColumn → ColumnData（@/types）変換 */
function toColumnData(col: DashboardColumn): ColumnData {
  return {
    title: col.title,
    appointments: col.appointments.map(toAppointment),
  };
}

/** DashboardAppointment → Appointment（@/types）変換 */
function toAppointment(a: DashboardAppointment): Appointment {
  return {
    id: a.id,
    time: a.time,
    ownerName: a.ownerName,
    petType: a.petType,
    petName: a.petName,
    visitType: a.visitType,
    serviceType: a.serviceType,
    nextAppointment: a.nextAppointment,
    isDesignated: a.isDesignated,
    doctor: a.doctor,
    petId: a.petId,
    ownerId: a.ownerId,
  };
}

export const useDashboardKanban = () => {
  const today = todayISO();
  const { data: apiColumns, isLoading } = useDashboardData(today);
  const updateStatusMutation = useUpdateAppointmentStatus();

  // API データを ColumnData[] に変換。ロード中またはデータがない場合は空カラムを用意
  const apiColumnData: ColumnData[] = useMemo(() => {
    if (!apiColumns) {
      return [
        { title: "受付予約", appointments: [] },
        { title: "受付済", appointments: [] },
        { title: "診療中", appointments: [] },
        { title: "会計待ち", appointments: [] },
        { title: "会計済", appointments: [] },
      ];
    }
    return apiColumns.map(toColumnData);
  }, [apiColumns]);

  // ローカル状態: API から取得したデータを元にドラッグ操作のために保持
  const [columns, setColumns] = useState<ColumnData[]>(apiColumnData);

  // API データが更新されたらローカル状態を同期
  useEffect(() => {
    setColumns(apiColumnData);
  }, [apiColumnData]);

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
            if (app.isDesignated) return false;
          } else {
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
    // 受付済 → 診療中 への直接ドラッグは禁止（カルテ作成が必要）
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

    const sourceColFiltered = filteredColumns.find(col => col.title === sourceColumn);
    const targetColFiltered = filteredColumns.find(col => col.title === targetColumn);

    if (!sourceColFiltered || !targetColFiltered) return false;

    const draggedCard = sourceColFiltered.appointments[dragIndex];
    if (!draggedCard) return false;

    // カラムが変わった場合は API でステータス更新
    if (sourceColumn !== targetColumn) {
      const newStatus = COLUMN_TITLE_TO_STATUS[targetColumn];
      if (newStatus) {
        updateStatusMutation.mutate(
          { id: draggedCard.id, status: newStatus },
          {
            onError: () => {
              toast.error("ステータスの更新に失敗しました");
              // 楽観的更新を元に戻すためにローカル状態を再同期
              setColumns(apiColumnData);
            },
          }
        );
      }
    }

    // 楽観的 UI 更新
    setColumns((prevColumns) => {
      const newColumns = prevColumns.map((col) => ({
        ...col,
        appointments: [...col.appointments],
      }));

      const sourceCol = newColumns.find((col) => col.title === sourceColumn);
      const targetCol = newColumns.find((col) => col.title === targetColumn);

      if (!sourceCol || !targetCol) return prevColumns;

      const realDragIndex = sourceCol.appointments.findIndex(app => app.id === draggedCard.id);
      if (realDragIndex === -1) return prevColumns;

      sourceCol.appointments.splice(realDragIndex, 1);

      let realHoverIndex = targetCol.appointments.length;

      if (hoverIndex < targetColFiltered.appointments.length) {
        const referenceCard = targetColFiltered.appointments[hoverIndex];
        const refIndex = targetCol.appointments.findIndex(app => app.id === referenceCard.id);
        if (refIndex !== -1) {
          realHoverIndex = refIndex;
        }
      }

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
        // API でステータスを completed に更新してリストから削除
        updateStatusMutation.mutate(
          { id: appointment.id, status: "completed" },
          {
            onSuccess: () => {
              toast.success("手続きを完了し、リストから削除しました");
            },
            onError: () => {
              toast.error("ステータスの更新に失敗しました");
            },
          }
        );
        // 楽観的: ローカルからも除外
        setColumns(prev => {
          const newColumns = prev.map(col => ({ ...col, appointments: [...col.appointments] }));
          const sourceCol = newColumns.find(c => c.title === currentColumnTitle);
          if (sourceCol) {
            const cardIndex = sourceCol.appointments.findIndex(a => a.id === appointment.id);
            if (cardIndex > -1) {
              sourceCol.appointments.splice(cardIndex, 1);
            }
          }
          return newColumns;
        });
        return;
      default: return;
    }

    // API でステータス更新
    const newStatus = COLUMN_TITLE_TO_STATUS[nextColumnTitle];
    if (newStatus) {
      updateStatusMutation.mutate(
        { id: appointment.id, status: newStatus },
        {
          onError: () => {
            toast.error("ステータスの更新に失敗しました");
            setColumns(apiColumnData);
          },
        }
      );
    }

    // 楽観的 UI 更新
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
    // API でキャンセルステータスに更新
    updateStatusMutation.mutate(
      { id: appointmentId, status: "canceled" },
      {
        onError: () => {
          toast.error("予約の取り消しに失敗しました");
          setColumns(apiColumnData);
        },
      }
    );

    // 楽観的: ローカルからも除外
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
    isLoading,
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
