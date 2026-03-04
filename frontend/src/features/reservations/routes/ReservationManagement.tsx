// React/Framework
import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router";

// External
import {
  ChevronLeft,
  ChevronRight,
  Calendar as CalendarIcon,
  Plus
} from "lucide-react";
import { format, addWeeks, subWeeks, addMonths, subMonths, addHours } from "date-fns";
import { ja } from "date-fns/locale";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FormHeader } from "@/components/shared/Form";

// Shared
import { ReservationFormModal } from "@/components/shared/ReservationFormModal";

// Relative
import { ReservationDetailModal } from "../components/ReservationDetailModal";
import { MonthView } from "../components/MonthView";
import { WeekView } from "../components/WeekView";
import {
  useGetReservations,
  useCreateReservation,
  useUpdateReservation,
  useDeleteReservation,
  transformToCreateRequest,
} from "../api";

// Types
import type { ReservationAppointment, Pet } from "@/types";
import type { UpdateReservationRequest } from "../api";

export const ReservationManagement = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [currentDate, setCurrentDate] = useState(new Date());
  const [view, setView] = useState<"month" | "week">("week");

  // API hooks
  const { data: appointments = [], isLoading } = useGetReservations();
  const createReservationMutation = useCreateReservation();
  const updateReservationMutation = useUpdateReservation();
  const deleteReservationMutation = useDeleteReservation();

  // Edit/Create Modal State
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingAppointment, setEditingAppointment] = useState<Partial<ReservationAppointment> | null>(null);

  // Detail Modal State
  const [isDetailOpen, setIsDetailOpen] = useState(false);
  const [detailAppointment, setDetailAppointment] = useState<ReservationAppointment | null>(null);

  const navigateToday = () => setCurrentDate(new Date());

  const navigatePrevious = () => {
    if (view === "month") {
      setCurrentDate(subMonths(currentDate, 1));
    } else {
      setCurrentDate(subWeeks(currentDate, 1));
    }
  };

  const navigateNext = () => {
    if (view === "month") {
      setCurrentDate(addMonths(currentDate, 1));
    } else {
      setCurrentDate(addWeeks(currentDate, 1));
    }
  };

  // Check for overlaps against server-fetched appointments
  const checkOverlap = (
    newStart: Date,
    newEnd: Date,
    doctor: string,
    excludeId?: string
  ): boolean => {
    return appointments.some(app => {
      if (excludeId && app.id === excludeId) return false;
      if (app.status === 'cancelled') return false;
      if (app.doctor !== doctor) return false;
      return newStart < app.end && newEnd > app.start;
    });
  };

  // Open Form Modal (Create or Edit)
  const handleOpenForm = (appointment?: Partial<ReservationAppointment>) => {
    if (appointment) {
      setEditingAppointment(appointment);
    } else {
      setEditingAppointment(null);
    }
    setIsFormOpen(true);
    setIsDetailOpen(false);
  };

  useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    const petId = searchParams.get("petId");

    if (petId && !isFormOpen) {
      const now = new Date();
      now.setMinutes(0, 0, 0);
      const start = addHours(now, 1);

      const stub: Partial<ReservationAppointment> = {
        start: start,
        end: addHours(start, 1),
        status: "confirmed",
        visitType: "first",
        type: "treatment",
        doctor: "医師A",
        isDesignated: false,
        petId: petId,
      };
      handleOpenForm(stub);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.search]);

  // Open Detail Modal
  const handleOpenDetail = (appointment: ReservationAppointment) => {
    setDetailAppointment(appointment);
    setIsDetailOpen(true);
  };

  const handleTimeSlotClick = (date: Date) => {
    const newAppointmentStub: Partial<ReservationAppointment> = {
      start: date,
      end: addHours(date, 1),
      status: "confirmed",
      visitType: "first",
      type: "treatment",
      doctor: "医師A",
      isDesignated: false,
    };
    handleOpenForm(newAppointmentStub);
  };

  const handleCloseForm = () => {
    setIsFormOpen(false);
    setEditingAppointment(null);
  };

  const handleCloseDetail = () => {
    setIsDetailOpen(false);
    setDetailAppointment(null);
  };

  const handleSave = (data: Partial<ReservationAppointment>, selectedPets: Pet[]) => {
    if (!data.start || !data.end || selectedPets.length === 0) return;

    const targetDoctor = data.doctor || editingAppointment?.doctor || "医師A";
    const hasOverlap = checkOverlap(
      data.start,
      data.end,
      targetDoctor,
      editingAppointment?.id
    );

    if (hasOverlap) {
      toast.error("指定された時間帯には既に予約が入っています");
      return;
    }

    if (editingAppointment && editingAppointment.id) {
      // Edit mode: Update via API
      const req: UpdateReservationRequest = {
        start_time: data.start.toISOString(),
        end_time: data.end.toISOString(),
        visit_type: data.visitType,
        service_type: data.type,
        is_designated: data.isDesignated,
        status: data.status,
        notes: data.notes,
      };

      updateReservationMutation.mutate(
        { id: editingAppointment.id, req },
        {
          onSuccess: () => {
            toast.success("予約を更新しました");
            handleCloseForm();
          },
          onError: () => {
            toast.error("予約の更新に失敗しました");
          },
        }
      );
    } else {
      // Create new: one per selected pet
      const primaryPet = selectedPets[0];
      const req = transformToCreateRequest(data, primaryPet.id, primaryPet.ownerId);

      createReservationMutation.mutate(req, {
        onSuccess: () => {
          toast.success("予約を作成しました");
          handleCloseForm();

          if (location.state?.from) {
            setTimeout(() => {
              navigate(location.state.from);
            }, 500);
          }
        },
        onError: () => {
          toast.error("予約の作成に失敗しました");
        },
      });
    }
  };

  const handleAppointmentUpdate = (appointment: ReservationAppointment, newStart: Date, newEnd: Date) => {
    const hasOverlap = checkOverlap(newStart, newEnd, appointment.doctor, appointment.id);

    if (hasOverlap) {
      toast.error("移動先に予約が重複しています");
      return;
    }

    const req: UpdateReservationRequest = {
      start_time: newStart.toISOString(),
      end_time: newEnd.toISOString(),
    };

    updateReservationMutation.mutate(
      { id: appointment.id, req },
      {
        onSuccess: () => {
          toast.success("予約時間を変更しました");
        },
        onError: () => {
          toast.error("予約時間の変更に失敗しました");
        },
      }
    );
  };

  // Status Change Handler
  const handleStatusChange = (appointment: ReservationAppointment, status: string) => {
    const req: UpdateReservationRequest = {
      status,
    };

    updateReservationMutation.mutate(
      { id: appointment.id, req },
      {
        onSuccess: (updated) => {
          setDetailAppointment(updated);
          toast.success("ステータスを更新しました");
        },
        onError: () => {
          toast.error("ステータスの更新に失敗しました");
        },
      }
    );
  };

  // Delete handler
  const handleDelete = (appointment: ReservationAppointment) => {
    if (window.confirm("本当にこの予約を削除しますか？")) {
      deleteReservationMutation.mutate(appointment.id, {
        onSuccess: () => {
          handleCloseDetail();
          toast.success("予約を削除しました");
        },
        onError: () => {
          toast.error("予約の削除に失敗しました");
        },
      });
    }
  };

  // Create Record Handler
  const handleCreateRecord = (appointment: ReservationAppointment) => {
    let targetPath = '/medical-records/new';
    if (appointment.type === 'trimming' || appointment.type === 'トリミング') {
      targetPath = '/trimming/new';
    } else if (appointment.type === 'hotel' || appointment.type === '入院' || appointment.type === 'ホテル') {
      targetPath = '/hospitalization/new';
    }

    if (appointment.petId) {
      navigate(`${targetPath}?petId=${appointment.petId}`, { state: { from: "/reservations" } });
    } else {
      if (window.confirm("この予約にはペットIDが紐付いていません。ペット選択画面へ移動しますか？")) {
        let selectPath = '/medical-records/select-pet';
        if (appointment.type === 'trimming' || appointment.type === 'トリミング') {
          selectPath = '/trimming/select-pet';
        } else if (appointment.type === 'hotel' || appointment.type === '入院' || appointment.type === 'ホテル') {
          selectPath = '/hospitalization/select-pet';
        }
        navigate(selectPath, { state: { from: "/reservations" } });
      }
    }
  };

  return (
    <div className="flex-1 flex flex-col min-h-0">
      <FormHeader
        title="予約管理"
        icon={<CalendarIcon className="size-5 text-[#37352F]" />}
        action={
          <Button className="bg-[#37352F] hover:bg-[#37352F]/90 text-white gap-2 h-11 text-sm" onClick={() => handleOpenForm()}>
            <Plus className="size-4" />
            <span className="hidden sm:inline">新規予約</span>
            <span className="sm:hidden">予約</span>
          </Button>
        }
      />

      <div className="flex-1 flex flex-col p-2 min-h-0">
        {/* Toolbar */}
        <div className="flex items-center justify-between mb-4 flex-shrink-0">
          <div className="flex items-center gap-4">
            <div className="flex items-center bg-white rounded-md border border-[rgba(55,53,47,0.16)] p-1 shadow-sm">
              <Button variant="ghost" size="icon" onClick={navigatePrevious}>
                <ChevronLeft className="size-5" />
              </Button>
              <Button variant="ghost" size="sm" className="px-4 font-medium" onClick={navigateToday}>
                今日
              </Button>
              <Button variant="ghost" size="icon" onClick={navigateNext}>
                <ChevronRight className="size-5" />
              </Button>
            </div>
            <h2 className="text-xl font-bold text-[#37352F]">
              {format(currentDate, "yyyy年 M月", { locale: ja })}
            </h2>
          </div>

          <Select value={view} onValueChange={(v: "month" | "week") => setView(v)}>
            <SelectTrigger className="w-[140px] bg-white border-[rgba(55,53,47,0.16)] h-11 text-sm">
              <SelectValue placeholder="表示切替" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="month">月表示</SelectItem>
              <SelectItem value="week">週表示</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Loading state */}
        {isLoading ? (
          <div className="flex-1 flex items-center justify-center text-sm text-muted-foreground">
            予約データを読み込み中...
          </div>
        ) : (
          /* Calendar View */
          view === "month" ? (
            <MonthView
              currentDate={currentDate}
              appointments={appointments}
              onAppointmentClick={handleOpenDetail}
            />
          ) : (
            <div className="flex-1 min-h-0 h-full">
              <WeekView
                currentDate={currentDate}
                appointments={appointments}
                onAppointmentClick={handleOpenDetail}
                onTimeSlotClick={handleTimeSlotClick}
                onAppointmentUpdate={handleAppointmentUpdate}
              />
            </div>
          )
        )}
      </div>

      {/* Create/Edit Form */}
      <ReservationFormModal
        isOpen={isFormOpen}
        onClose={handleCloseForm}
        onSave={handleSave}
        initialData={editingAppointment}
      />

      {/* Detail Modal */}
      <ReservationDetailModal
        isOpen={isDetailOpen}
        onClose={handleCloseDetail}
        appointment={detailAppointment}
        onEdit={handleOpenForm}
        onDelete={handleDelete}
        onCreateRecord={handleCreateRecord}
        onStatusChange={handleStatusChange}
      />
    </div>
  );
};
