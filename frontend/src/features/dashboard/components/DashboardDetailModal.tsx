import { useNavigate } from "react-router-dom";
import { MOCK_PETS } from "../../../lib/constants";
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogFooter,
  DialogDescription
} from "../../../components/ui/dialog";
import { Button } from "../../../components/ui/button";
import { Badge } from "../../../components/ui/badge";
import { 
  Clock, 
  User, 
  Dog, 
  Stethoscope, 
  AlertCircle,
  FileText,
  CreditCard,
  Scissors,
  Bed
} from "lucide-react";
import { Appointment } from "../../../types";

interface DashboardDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  appointment: Appointment | null;
  onConfirm?: () => void; // Optional action (e.g. status change)
  onEdit?: (appointment: Appointment) => void;
  onCancel?: (appointment: Appointment) => void;
  currentStatus?: string;
}

export const DashboardDetailModal = ({
  isOpen,
  onClose,
  appointment,
  onConfirm,
  onEdit,
  onCancel,
  currentStatus,
}: DashboardDetailModalProps) => {
  const navigate = useNavigate();

  if (!appointment) return null;

  const handleCreateMedicalRecord = () => {
      if (appointment.petId) {
          navigate(`/medical-records/new?petId=${appointment.petId}`, { 
              state: { 
                  from: "/",
                  appointmentId: appointment.id 
              } 
          });
      } else {
          navigate("/medical-records/select-pet", { state: { from: "/" } });
      }
      onClose();
  };

  const isTrimming = appointment.serviceType.includes("トリミング");
  const isHospitalization = appointment.serviceType.includes("入院");
  const isMedical = appointment.serviceType.includes("診療") || (!isTrimming && !isHospitalization);

  const handleCreateTrimming = () => {
      if (appointment.petId) {
          navigate(`/trimming/new?petId=${appointment.petId}`, { state: { from: "/" } });
      } else {
          navigate("/trimming/new", { state: { from: "/" } });
      }
      onClose();
  };

  const handleCreateHospitalization = () => {
      if (appointment.petId) {
          navigate(`/hospitalization/new?petId=${appointment.petId}`, { state: { from: "/" } });
      } else {
          navigate("/hospitalization/new", { state: { from: "/" } });
      }
      onClose();
  };

  const handleCreateAccounting = () => {
    // Navigate to Accounting New page with mock data/params
    if (appointment.petId) {
        navigate(`/accounting/new?petId=${appointment.petId}`, { 
            state: { 
                from: "/",
                appointmentId: appointment.id 
            } 
        });
    } else {
        navigate("/accounting/new", { state: { from: "/" } });
    }
    onClose();
  };

  const handleOpenOwnerDetail = () => {
      if (appointment.petId) {
          const pet = MOCK_PETS.find(p => p.id === appointment.petId);
          if (pet && pet.ownerId) {
              navigate(`/owners/${pet.ownerId}`);
              onClose();
          }
      }
  };

  const getStatusColor = (status?: string) => {
    switch (status) {
      case "受付予約": return "bg-blue-50 text-blue-700 border-blue-200";
      case "受付済": return "bg-orange-50 text-orange-700 border-orange-200";
      case "診療中": return "bg-green-50 text-green-700 border-green-200";
      case "会計待ち": return "bg-purple-50 text-purple-700 border-purple-200";
      case "会計済": return "bg-gray-100 text-gray-700 border-gray-200";
      default: return "bg-gray-100 text-gray-700 border-gray-200";
    }
  };

  const renderActions = () => {
    // Common Actions
    const ownerDetailBtn = appointment.petId && (
         <Button variant="ghost" onClick={handleOpenOwnerDetail} className="text-[#37352F]">
            <User className="size-4 mr-2" />
            飼主詳細
         </Button>
    );

    // Status Specific Actions
    if (currentStatus === "受付予約") {
        return (
            <>
                {onCancel && (
                    <Button variant="ghost" onClick={() => onCancel(appointment)} className="text-red-500 hover:text-red-600 hover:bg-red-50">
                        <Trash2 className="size-4 mr-2" />
                        取消
                    </Button>
                )}
                {onEdit && (
                    <Button variant="outline" onClick={() => onEdit(appointment)}>
                        <Pencil className="size-4 mr-2" />
                        編集
                    </Button>
                )}
                {ownerDetailBtn}
                {onConfirm && (
                    <Button onClick={onConfirm} className="bg-blue-600 hover:bg-blue-700 text-white">
                        受付済にする
                    </Button>
                )}
            </>
        );
    }

    if (currentStatus === "受付済") {
        return (
            <>
                {ownerDetailBtn}
                {isMedical ? (
                    <div className="flex flex-col items-end gap-2">
                        <Button 
                            onClick={() => {
                                if (onConfirm) onConfirm();
                                handleCreateMedicalRecord();
                            }} 
                            className="bg-blue-600 hover:bg-blue-700 text-white w-full sm:w-auto"
                        >
                            <FileText className="size-4 mr-2" />
                            カルテ作成
                        </Button>
                        <span className="text-[10px] text-muted-foreground">※カルテ作成と同時に「診療中」へ移動します</span>
                    </div>
                ) : (
                    onConfirm && (
                        <Button onClick={onConfirm} className="bg-blue-600 hover:bg-blue-700 text-white">
                            診察を開始する
                        </Button>
                    )
                )}
            </>
        );
    }

    if (currentStatus === "診療中") {
        return (
            <>
                {ownerDetailBtn}
                {onConfirm && (
                    <Button variant="outline" onClick={onConfirm} className="text-red-600 border-red-200 hover:bg-red-50">
                        診察を終了する
                    </Button>
                )}
                {isMedical && (
                     <Button onClick={handleCreateMedicalRecord} className="bg-[#37352F] text-white hover:bg-[#37352F]/90">
                        <FileText className="size-4 mr-2" />
                        カルテ入力
                     </Button>
                 )}
                 {isTrimming && (
                     <Button onClick={handleCreateTrimming} className="bg-orange-600 text-white hover:bg-orange-700">
                        <Scissors className="size-4 mr-2" />
                        施術記録
                     </Button>
                 )}
            </>
        );
    }

    if (currentStatus === "会計待ち") {
        return (
            <>
                {ownerDetailBtn}
                <Button 
                    onClick={handleCreateAccounting} 
                    className="bg-emerald-600 text-white hover:bg-emerald-700"
                    disabled={appointment.nextAppointment === "精算未確認"}
                >
                    <CreditCard className="size-4 mr-2" />
                    会計へ進む
                </Button>
            </>
        );
    }

    if (currentStatus === "会計済") {
        return (
            <>
                {ownerDetailBtn}
                {onConfirm && (
                    <Button onClick={onConfirm} className="bg-gray-800 text-white hover:bg-gray-900">
                        完了/リストから削除
                    </Button>
                )}
            </>
        );
    }

    // Default Fallback
    return (
        <>
            {ownerDetailBtn}
             {isMedical && (
                 <Button onClick={handleCreateMedicalRecord} className="bg-[#37352F] text-white hover:bg-[#37352F]/90">
                    <FileText className="size-4 mr-2" />
                    カルテ確認
                 </Button>
             )}
             {isTrimming && (
                 <Button onClick={handleCreateTrimming} className="bg-orange-600 text-white hover:bg-orange-700">
                    <Scissors className="size-4 mr-2" />
                    トリミング
                 </Button>
             )}
             {isHospitalization && (
                 <Button onClick={handleCreateHospitalization} className="bg-indigo-600 text-white hover:bg-indigo-700">
                    <Bed className="size-4 mr-2" />
                    入院登録
                 </Button>
             )}
             {onConfirm && (
                <Button onClick={onConfirm} className="bg-blue-600 hover:bg-blue-700 text-white">
                    ステータス変更
                </Button>
             )}
        </>
    );
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[500px] p-0 gap-0 overflow-hidden bg-white">
        <DialogHeader className="p-5 pb-4 border-b border-gray-100">
          <div className="flex items-center justify-between gap-4">
             <div className="flex items-center gap-2">
                <span className={`flex h-8 w-8 items-center justify-center rounded-full ${appointment.visitType === '初診' ? 'bg-blue-100 text-blue-600' : 'bg-slate-100 text-slate-600'}`}>
                    {appointment.visitType === '初診' ? '初' : '再'}
                </span>
                <div className="flex flex-col">
                    <DialogTitle className="text-lg font-bold text-gray-900">
                        {appointment.serviceType}
                    </DialogTitle>
                    <span className="text-xs text-gray-500 font-normal">ID: {appointment.id}</span>
                </div>
             </div>
             {currentStatus && (
                <Badge variant="outline" className={`${getStatusColor(currentStatus)} px-3 py-1 text-sm font-medium border`}>
                    {currentStatus}
                </Badge>
             )}
          </div>
          <DialogDescription className="sr-only">
            予約の詳細情報
          </DialogDescription>
        </DialogHeader>

        <div className="p-5 space-y-6">
            {/* Time */}
            <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                <Clock className="size-5 text-gray-500" />
                <div className="font-mono text-xl font-medium text-gray-800">
                    {appointment.time}
                </div>
                {appointment.nextAppointment && (
                     <Badge 
                        variant={appointment.nextAppointment === "精算未確認" ? "destructive" : "secondary"}
                        className="ml-auto"
                      >
                        {appointment.nextAppointment === "精算未確認" && <AlertCircle className="size-3 mr-1" />}
                        {appointment.nextAppointment}
                      </Badge>
                )}
            </div>

            {/* Patient Info */}
            <div className="space-y-3">
                <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">患者情報</h3>
                <div className="flex items-center justify-between border-b border-gray-100 pb-2">
                    <div className="flex items-center gap-2 text-gray-600">
                        <Dog className="size-4" />
                        <span className="text-sm">ペット</span>
                    </div>
                    <div className="text-right">
                        <div className="font-bold text-base">{appointment.petName}</div>
                        <div className="text-sm text-gray-500">{appointment.petType}</div>
                    </div>
                </div>
                <div className="flex items-center justify-between border-b border-gray-100 pb-2">
                    <div className="flex items-center gap-2 text-gray-600">
                        <User className="size-4" />
                        <span className="text-sm">飼い主</span>
                    </div>
                    <span className="font-medium">{appointment.ownerName}</span>
                </div>
            </div>

            {/* Medical Details */}
            <div className="space-y-3">
                <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">診療詳細</h3>
                <div className="flex items-center justify-between border-b border-gray-100 pb-2">
                    <div className="flex items-center gap-2 text-gray-600">
                        <Stethoscope className="size-4" />
                        <span className="text-sm">担当医</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <span className="font-medium">{appointment.doctor || "未定"}</span>
                        {appointment.isDesignated && <Badge variant="outline" className="text-sm h-6 bg-orange-50 text-orange-700 border-orange-200">指名</Badge>}
                    </div>
                </div>
            </div>
        </div>

        <DialogFooter className="p-4 bg-gray-50 flex flex-wrap justify-end items-center gap-4">
             <div className="flex flex-wrap gap-2 justify-end">
                 {renderActions()}
             </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
