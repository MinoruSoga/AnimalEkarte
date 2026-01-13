import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../../../components/ui/button";
import { DashboardDetailModal } from "../components/DashboardDetailModal";
import { KanbanColumn } from "../components/KanbanColumn";
import { Appointment, ReservationAppointment, Pet } from "../../../types";
import FormHeader from "../../../components/shared/FormHeader";
import { Checkbox } from "../../../components/ui/checkbox";
import { Label } from "../../../components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../../../components/ui/select";
import { Filter } from "lucide-react";
import { toast } from "sonner@2.0.3";
import { useDashboardKanban } from "../hooks/useDashboardKanban";
import { format, addHours } from "date-fns";
import { ja } from "date-fns/locale";
import { ReservationFormModal } from "../../reservations/components/ReservationFormModal";

export const Dashboard = () => {
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

  const [isFilterOpen, setIsFilterOpen] = useState(false);

  const today = format(new Date(), "yyyy年M月d日 (E)", { locale: ja });

  const doctors = [
    { id: "all", name: "全て" },
    { id: "医師A", name: "医師A" },
    { id: "医師B", name: "医師B" },
    { id: "医師C", name: "医師C" },
    { id: "医師指名なし", name: "医師指名なし" },
  ];

  const handleAddClick = (columnTitle: string) => {
    if (columnTitle === "受付予約") {
        navigate("/reservations");
    } else {
        toast.info("新規登録", {
            description: "新規予約・受付画面へ移動します。",
            duration: 2000
        });
        navigate("/reservations");
    }
  };

  const handleCardClick = (appointment: Appointment) => {
    setSelectedAppointment(appointment);
    setModalOpen(true);
  };

  const handleConfirm = () => {
    if (!selectedAppointment) return;
    advanceStatus(selectedAppointment);
    setModalOpen(false);
    setSelectedAppointment(null);
  };

  const handleEdit = (appointment: Appointment) => {
    // Convert Appointment to ReservationAppointment stub
    const now = new Date();
    const [hours, minutes] = appointment.time.split(':').map(Number);
    const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), hours, minutes);
    
    const stub: Partial<ReservationAppointment> = {
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
    
    setEditingAppointment(stub);
    setIsEditModalOpen(true);
    setModalOpen(false);
  };

  const handleEditSave = (data: Partial<ReservationAppointment>, selectedPets: Pet[]) => {
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
  };

  const handleCancel = (appointment: Appointment) => {
      if (window.confirm("本当にこの予約を取り消しますか？")) {
          cancelAppointment(appointment.id);
          toast.success("予約を取り消しました");
          setModalOpen(false);
      }
  };

  return (
    <div className="flex-1 flex flex-col h-full bg-[#F7F6F3]">
        <FormHeader 
            title="当日の受付" 
            description={`${today} - 受付状況をリアルタイムで確認`}
            action={
                <div className="flex gap-2">
                    <Button 
                        variant={isFilterOpen ? "secondary" : "outline"} 
                        size="sm" 
                        className="gap-2 bg-white h-10 text-sm"
                        onClick={() => setIsFilterOpen(!isFilterOpen)}
                    >
                        <Filter className="size-4" />
                        フィルター
                    </Button>
                    <Button 
                        size="sm" 
                        className="bg-[#37352F] hover:bg-[#37352F]/90 text-white h-10 text-sm"
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
                        <h4 className="font-bold text-sm text-[#37352F]">診察区分</h4>
                        <div className="flex gap-4">
                            {["初診", "再診"].map(type => (
                                <div key={type} className="flex items-center space-x-2">
                                    <Checkbox 
                                        id={`visit-${type}`} 
                                        checked={filters.selectedVisitTypes.includes(type)}
                                        onCheckedChange={() => filters.toggleVisitType(type)}
                                        className="size-4"
                                    />
                                    <Label htmlFor={`visit-${type}`} className="text-sm font-normal cursor-pointer text-[#37352F]">{type}</Label>
                                </div>
                            ))}
                        </div>
                    </div>
                    
                    {/* Doctor/Designation Selection */}
                    <div className="space-y-2">
                        <h4 className="font-bold text-sm text-[#37352F]">指名</h4>
                        <Select value={filters.selectedDoctor} onValueChange={filters.setSelectedDoctor}>
                            <SelectTrigger className="w-[200px] h-10 text-sm bg-white border-input">
                                <SelectValue placeholder="指名を選択" />
                            </SelectTrigger>
                            <SelectContent>
                                {doctors.map((doctor) => (
                                    <SelectItem key={doctor.id} value={doctor.id}>
                                        {doctor.name}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>

                    {/* Trimming */}
                    <div className="space-y-2">
                        <h4 className="font-bold text-sm text-[#37352F]">種類</h4>
                        <div className="flex items-center space-x-2 pt-0.5">
                            <Checkbox  
                                id="trimming-only" 
                                checked={filters.isTrimmingOnly}
                                onCheckedChange={(c) => filters.setIsTrimmingOnly(!!c)}
                                className="size-4"
                            />
                            <Label htmlFor="trimming-only" className="text-sm font-normal cursor-pointer text-[#37352F]">トリミングのみ表示</Label>
                        </div>
                    </div>
                </div>
            </div>
        )}

        <div className="flex-1 overflow-hidden p-[20px] pt-4">
            <div className="flex gap-2 h-full w-full overflow-x-auto pb-2 bg-transparent">
                {filteredColumns.map((column, index) => (
                <KanbanColumn 
                    key={index} 
                    data={column} 
                    moveCard={moveCard} 
                    onAddClick={["診療中", "会計待ち", "会計済"].includes(column.title) ? undefined : () => handleAddClick(column.title)} 
                    onCardClick={handleCardClick} 
                />
                ))}
            </div>
        </div>
        
        <DashboardDetailModal 
            isOpen={modalOpen} 
            onClose={() => setModalOpen(false)}
            onConfirm={handleConfirm}
            onEdit={handleEdit}
            onCancel={handleCancel}
            appointment={selectedAppointment}
            currentStatus={selectedAppointment ? filteredColumns.find(c => c.appointments.some(a => a.id === selectedAppointment.id))?.title : undefined}
        />
        
        <ReservationFormModal 
            isOpen={isEditModalOpen}
            onClose={() => {
                setIsEditModalOpen(false);
                setEditingAppointment(null);
            }}
            onSave={handleEditSave}
            initialData={editingAppointment}
        />
    </div>
  );
}
