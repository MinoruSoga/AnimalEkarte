import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { 
  Calendar, 
  Clock, 
  Stethoscope,
  Pencil,
  Trash2,
  FilePlus2,
  FileText,
  Activity,
  User,
  Scissors,
  Building2
} from "lucide-react";
import { Button } from "../../../components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "../../../components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";
import { ReservationAppointment } from "../../../types";
import { getReservationTypeName } from "../../../lib/status-helpers";

interface ReservationDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  onEdit: (appointment: ReservationAppointment) => void;
  onDelete?: (appointment: ReservationAppointment) => void;
  onCreateRecord?: (appointment: ReservationAppointment) => void;
  onStatusChange?: (appointment: ReservationAppointment, status: string) => void;
  appointment: ReservationAppointment | null;
}

const STATUS_OPTIONS = [
  { value: 'confirmed', label: '予約確定', color: 'bg-green-500' },
  { value: 'checked_in', label: '受付済', color: 'bg-blue-500' },
  { value: 'in_consultation', label: '診療中', color: 'bg-purple-500' },
  { value: 'accounting', label: '会計待ち', color: 'bg-orange-500' },
  { value: 'completed', label: '完了', color: 'bg-gray-500' },
  { value: 'cancelled', label: 'キャンセル', color: 'bg-red-500' },
];

export const ReservationDetailModal = ({
  isOpen,
  onClose,
  onEdit,
  onDelete,
  onCreateRecord,
  onStatusChange,
  appointment,
}: ReservationDetailModalProps) => {
  if (!appointment) return null;

  const getActionConfig = (type: string) => {
    if (type === 'trimming' || type === 'トリミング') {
        return { label: 'トリミング記録作成', icon: <Scissors className="size-4 mr-1.5" /> };
    }
    if (type === 'hotel' || type === '入院' || type === 'ホテル') {
        return { label: '入院・ホテル登録', icon: <Building2 className="size-4 mr-1.5" /> };
    }
    // Default
    return { label: 'カルテ作成', icon: <FilePlus2 className="size-4 mr-1.5" /> };
  };

  const actionConfig = getActionConfig(appointment.type);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[400px] p-0 gap-0 overflow-hidden bg-white">
        <DialogHeader className="p-4 pb-2 pr-12">
          <DialogTitle className="flex items-center justify-between w-full">
             <div className="flex items-center gap-2 text-base font-bold text-[#37352F]">
                <span className={`w-2.5 h-2.5 rounded-full ${appointment.visitType === 'first' ? 'bg-red-500' : 'bg-blue-500'}`} />
                {getReservationTypeName(appointment.type)}
             </div>
             
             {/* Status Selector */}
             {onStatusChange && (
               <Select 
                 value={appointment.status} 
                 onValueChange={(val) => onStatusChange(appointment, val)}
               >
                 <SelectTrigger className="h-10 w-[160px] text-sm border border-[rgba(55,53,47,0.16)] bg-white shadow-sm">
                   <div className="flex items-center gap-1.5 truncate">
                      <span className={`w-1.5 h-1.5 rounded-full ${STATUS_OPTIONS.find(o => o.value === appointment.status)?.color || 'bg-gray-400'}`} />
                      <SelectValue placeholder="ステータス" />
                   </div>
                 </SelectTrigger>
                 <SelectContent>
                   {STATUS_OPTIONS.map((opt) => (
                     <SelectItem key={opt.value} value={opt.value} className="text-sm">
                        <div className="flex items-center gap-2">
                          <span className={`w-2 h-2 rounded-full ${opt.color}`} />
                          {opt.label}
                        </div>
                     </SelectItem>
                   ))}
                 </SelectContent>
               </Select>
             )}
          </DialogTitle>
          <DialogDescription className="sr-only">
             予約の詳細情報
          </DialogDescription>
        </DialogHeader>

        <div className="p-4 pt-2 space-y-4">
            {/* Time & Date */}
            <div className="flex items-start gap-3 p-3 bg-[#F7F6F3] rounded-md border border-[rgba(55,53,47,0.09)]">
                <Calendar className="size-4 text-[#37352F]/60 mt-0.5" />
                <div>
                    <div className="font-bold text-sm text-[#37352F]">
                        {format(appointment.start, "yyyy年 M月 d日 (E)", { locale: ja })}
                    </div>
                    <div className="flex items-center gap-2 text-sm text-[#37352F]/80">
                        <Clock className="size-3" />
                        {format(appointment.start, "H:mm")} - {format(appointment.end, "H:mm")}
                    </div>
                </div>
            </div>

            {/* Pet & Owner Info */}
            <div className="space-y-2">
                <h3 className="text-sm font-bold text-[#37352F]/60 uppercase tracking-wider flex items-center gap-1">
                    <User className="size-3" />
                    患者情報
                </h3>
                <div className="flex items-center justify-between border-b border-[rgba(55,53,47,0.09)] pb-1.5">
                    <span className="text-sm text-[#37352F]/80">ペット名</span>
                    <span className="font-bold text-sm text-[#37352F]">{appointment.petName}</span>
                </div>
                <div className="flex items-center justify-between border-b border-[rgba(55,53,47,0.09)] pb-1.5">
                    <span className="text-sm text-[#37352F]/80">飼い主名</span>
                    <span className="font-medium text-sm text-[#37352F]">{appointment.ownerName}</span>
                </div>
                {appointment.petId && (
                     <div className="flex items-center justify-between border-b border-[rgba(55,53,47,0.09)] pb-1.5">
                        <span className="text-sm text-[#37352F]/80">カルテNo.</span>
                        <span className="font-mono text-sm text-[#37352F]">{appointment.petId}</span>
                    </div>
                )}
            </div>

            {/* Medical Info */}
            <div className="space-y-2">
                <h3 className="text-sm font-bold text-[#37352F]/60 uppercase tracking-wider flex items-center gap-1">
                    <Activity className="size-3" />
                    診療詳細
                </h3>
                <div className="grid grid-cols-2 gap-3">
                    <div className="space-y-0.5">
                        <span className="text-sm text-[#37352F]/60">担当医</span>
                        <div className="flex items-center gap-1.5 font-medium text-sm text-[#37352F]">
                            <Stethoscope className="size-3 text-[#37352F]/40" />
                            {appointment.doctor}
                            {appointment.isDesignated && <span className="text-sm bg-orange-100 text-orange-700 px-1 rounded">指名</span>}
                        </div>
                    </div>
                     <div className="space-y-0.5">
                        <span className="text-sm text-[#37352F]/60">来院区分</span>
                        <div className="font-medium text-sm text-[#37352F]">
                            {appointment.visitType === 'first' ? '初診' : '再診'}
                        </div>
                    </div>
                </div>
                
                {appointment.notes && (
                    <div className="mt-2 p-2 bg-yellow-50/50 rounded border border-yellow-100 text-sm text-[#37352F]">
                        <div className="flex items-center gap-1 text-yellow-700 mb-1 font-medium">
                            <FileText className="size-3" />
                            メモ
                        </div>
                        <p className="whitespace-pre-wrap">{appointment.notes}</p>
                    </div>
                )}
            </div>
        </div>

        <DialogFooter className="p-3 bg-[#F7F6F3] flex flex-col sm:flex-row gap-2 sm:justify-between items-center border-t border-[rgba(55,53,47,0.09)]">
            <div className="flex gap-2 w-full sm:w-auto">
                {onDelete && (
                    <Button variant="ghost" size="sm" className="h-10 w-10 p-0 text-red-500 hover:text-red-600 hover:bg-red-50" onClick={() => onDelete(appointment)}>
                        <Trash2 className="size-5" />
                    </Button>
                )}
            </div>
            
            <div className="flex gap-2 w-full sm:w-auto justify-end">
                <Button variant="outline" size="sm" onClick={() => onEdit(appointment)} className="h-10 text-sm border-[rgba(55,53,47,0.16)] bg-white text-[#37352F]">
                    <Pencil className="size-4 mr-1.5" />
                    編集
                 </Button>
                 {onCreateRecord && (
                     <Button 
                        size="sm"
                        className="bg-[#37352F] text-white hover:bg-[#37352F]/90 h-10 text-sm shadow-sm border-transparent" 
                        onClick={() => onCreateRecord(appointment)}
                     >
                        {actionConfig.icon}
                        {actionConfig.label}
                     </Button>
                 )}
            </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
