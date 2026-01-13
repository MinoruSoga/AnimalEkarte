import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { Button } from "../../../components/ui/button";
import { 
  ChevronLeft, 
  ChevronRight, 
  Calendar as CalendarIcon, 
  Plus
} from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";
import { format, addWeeks, subWeeks, addMonths, subMonths, addHours } from "date-fns";
import { ja } from "date-fns/locale";
import { toast } from "sonner@2.0.3";
import { ReservationAppointment, Pet } from "../../../types";
import { generateMockAppointments, MOCK_PETS } from "../../../lib/constants";
import { ReservationFormModal } from "../components/ReservationFormModal";
import { ReservationDetailModal } from "../components/ReservationDetailModal";
import { MonthView } from "../components/MonthView";
import { WeekView } from "../components/WeekView";

import FormHeader from "../../../components/shared/FormHeader";

export const ReservationManagement = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [currentDate, setCurrentDate] = useState(new Date());
  const [view, setView] = useState<"month" | "week">("week");
  const [appointments, setAppointments] = useState<ReservationAppointment[]>(generateMockAppointments());
  
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

  // Check for overlaps
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
      
      // Check time overlap: (StartA < EndB) and (EndA > StartB)
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
    // Ensure detail is closed if open
    setIsDetailOpen(false); 
  };

  useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    const petId = searchParams.get("petId");
    
    if (petId && !isFormOpen) {
       const pet = MOCK_PETS.find(p => p.id === petId);
       const now = new Date();
       now.setMinutes(0, 0, 0);
       const start = addHours(now, 1);
       
       const stub: Partial<ReservationAppointment> = {
         start: start,
         end: addHours(start, 1),
         status: "confirmed",
         visitType: pet?.lastVisit ? "revisit" : "first",
         type: "treatment",
         doctor: "医師A",
         isDesignated: false,
         petId: petId
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
    // Create a new appointment stub starting at the clicked time
    const newAppointmentStub: Partial<ReservationAppointment> = {
        start: date,
        end: addHours(date, 1), // Default 1 hour duration
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

    // Validate overlap
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

    const newAppointmentsList: ReservationAppointment[] = [];

    if (editingAppointment && editingAppointment.id) {
        // Edit mode: Update existing
        const primaryPet = selectedPets[0];
        setAppointments(appointments.map(app => 
            app.id === editingAppointment.id ? { 
                ...app, 
                ...data,
                ownerName: primaryPet.ownerName,
                petName: primaryPet.name,
                petId: primaryPet.id
            } as ReservationAppointment : app
        ));
        toast.success("予約を更新しました");
    } else {
        // Create new mode: One per pet
        selectedPets.forEach(pet => {
            const newAppointment: ReservationAppointment = {
                id: Math.random().toString(36).substr(2, 9),
                ...data as ReservationAppointment,
                ownerName: pet.ownerName,
                petName: pet.name,
                petId: pet.id
            };
            newAppointmentsList.push(newAppointment);
        });
        
        setAppointments([...appointments, ...newAppointmentsList]);
        toast.success("予約を作成しました");
    }
    handleCloseForm();

    if (location.state?.from) {
        setTimeout(() => {
            navigate(location.state.from);
        }, 500);
    }
  };

  const handleAppointmentUpdate = (appointment: ReservationAppointment, newStart: Date, newEnd: Date) => {
    // Check overlap
    const hasOverlap = checkOverlap(newStart, newEnd, appointment.doctor, appointment.id);
    
    if (hasOverlap) {
        toast.error("移動先に予約が重複しています");
        return;
    }

    setAppointments(appointments.map(app => 
        app.id === appointment.id 
            ? { ...app, start: newStart, end: newEnd } 
            : app
    ));
    toast.success("予約時間を変更しました");
  };

  // Status Change Handler
  const handleStatusChange = (appointment: ReservationAppointment, status: string) => {
      const updatedAppointment = { ...appointment, status: status as ReservationAppointment['status'] };
      
      setAppointments(appointments.map(app => 
          app.id === appointment.id ? updatedAppointment : app
      ));
      
      setDetailAppointment(updatedAppointment);
      toast.success("ステータスを更新しました");
  };

  // Delete handler (mock)
  const handleDelete = (appointment: ReservationAppointment) => {
      if (window.confirm("本当にこの予約を削除しますか？")) {
          setAppointments(appointments.filter(a => a.id !== appointment.id));
          handleCloseDetail();
          toast.success("予約を削除し���した");
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
          // Fallback if no pet ID (e.g. manually entered name without ID)
          if (window.confirm("この予約にはペットIDが紐付いていません。ペット選択画面へ移動しますか？")) {
               // Determine selection path based on type
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
    <div className="flex-1 flex flex-col h-full bg-[#F7F6F3] min-w-0 w-full">
      <FormHeader 
         title="予約管理" 
         icon={<CalendarIcon className="size-5 text-[#37352F]" />}
         action={
            <div className="flex items-center gap-2">
                 <Button className="bg-[#37352F] hover:bg-[#37352F]/90 text-white gap-2 h-10 text-sm" onClick={() => handleOpenForm()}>
                    <Plus className="size-4" />
                    <span className="hidden sm:inline">新規予約</span>
                    <span className="sm:hidden">予約</span>
                </Button>
            </div>
         }
      />

      <div className="flex-1 flex flex-col p-4 overflow-hidden w-full min-w-0">
        {/* Toolbar */}
        <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-4">
                <div className="flex items-center bg-white rounded-md border border-[rgba(55,53,47,0.16)] p-1 shadow-sm">
                    <Button variant="ghost" size="icon" className="h-10 w-10" onClick={navigatePrevious}>
                        <ChevronLeft className="size-5" />
                    </Button>
                    <Button variant="ghost" size="sm" className="h-10 px-4 text-sm font-medium" onClick={navigateToday}>
                        今日
                    </Button>
                    <Button variant="ghost" size="icon" className="h-10 w-10" onClick={navigateNext}>
                        <ChevronRight className="size-5" />
                    </Button>
                </div>
                <h2 className="text-xl font-bold text-[#37352F] flex items-center gap-2">
                    {format(currentDate, "yyyy年 M月", { locale: ja })}
                </h2>
            </div>

            <div className="flex items-center gap-2">
                <Select value={view} onValueChange={(v: any) => setView(v)}>
                    <SelectTrigger className="w-[140px] bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm">
                        <SelectValue placeholder="表示切替" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="month">月表示</SelectItem>
                        <SelectItem value="week">週表示</SelectItem>
                    </SelectContent>
                </Select>
            </div>
        </div>

        {/* Calendar View */}
        <div className="flex-1 overflow-hidden flex flex-col">
            {view === "month" ? (
                <MonthView 
                    currentDate={currentDate} 
                    appointments={appointments} 
                    onAppointmentClick={handleOpenDetail} 
                />
            ) : (
                <WeekView 
                    currentDate={currentDate} 
                    appointments={appointments} 
                    onAppointmentClick={handleOpenDetail} 
                    onTimeSlotClick={handleTimeSlotClick}
                    onAppointmentUpdate={handleAppointmentUpdate}
                />
            )}
        </div>
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
}
