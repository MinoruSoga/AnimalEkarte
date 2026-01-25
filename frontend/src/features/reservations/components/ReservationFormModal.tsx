// React/Framework
import { useState, useEffect } from "react";

// External
import { User } from "lucide-react";

// Internal
import { Button } from "../../../components/ui/button";
import { Label } from "../../../components/ui/label";
import { Badge } from "../../../components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "../../../components/ui/dialog";
import { MOCK_PETS } from "../../../lib/constants";

// Relative
import { PatientSelectionTable } from "./PatientSelectionTable";
import { ReservationFormFields } from "./ReservationFormFields";
import { usePetSelection } from "../../pets/hooks/usePetSelection";

// Types
import type { Pet, ReservationAppointment } from "../../../types";

interface ReservationFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: Partial<ReservationAppointment>, selectedPets: Pet[]) => void;
  initialData: Partial<ReservationAppointment> | null;
}

export const ReservationFormModal = ({
  isOpen,
  onClose,
  onSave,
  initialData,
}: ReservationFormModalProps) => {
  const [formData, setFormData] = useState<Partial<ReservationAppointment>>({});
  
  const {
    selectedPets,
    setSelectedPets,
    togglePetSelection
  } = usePetSelection([], "multiple-same-owner");

  useEffect(() => {
    if (isOpen) {
      if (initialData) {
        setFormData({ ...initialData });
        if (initialData.petId) {
            const foundPet = MOCK_PETS.find((p) => p.id === initialData.petId);
            if (foundPet) {
              setSelectedPets([foundPet]);
            } else {
              setSelectedPets([]);
            }
        } else {
            setSelectedPets([]);
        }
      } else {
        const defaultStart = new Date();
        defaultStart.setHours(10, 0, 0, 0);
        const defaultEnd = new Date(defaultStart);
        defaultEnd.setHours(11, 0, 0, 0);

        setFormData({
          start: defaultStart,
          end: defaultEnd,
          visitType: "first",
          type: "診療",
          doctor: "医師A",
          isDesignated: false,
          status: "confirmed",
        });
        setSelectedPets([]);
      }
    }
  }, [isOpen, initialData, setSelectedPets]);

  const handleSave = () => {
    onSave(formData, selectedPets);
  };

  const handlePetSelect = (pet: Pet) => {
    togglePetSelection(pet);
  };

  const isEditMode = initialData && initialData.id;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="w-[98%] sm:max-w-[1200px] h-[90vh] flex flex-col p-0 gap-0 bg-white overflow-hidden">
        <DialogHeader className="p-4 border-b shrink-0 h-14 flex flex-row items-center justify-between space-y-0">
          <DialogTitle className="text-sm font-bold text-[#37352F]">
            {isEditMode ? "予約編集" : "新規予約作成"}
          </DialogTitle>
          <DialogDescription className="sr-only">
            左側のリストからペットを選択し、右側のフォームで予約情報を入力してください
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col lg:flex-row flex-1 overflow-hidden">
          {/* Left Panel: Patient Selection Table */}
          <div className="w-full lg:w-7/12 border-b lg:border-b-0 lg:border-r bg-[#F7F6F3] p-4 flex flex-col overflow-hidden min-h-[300px] lg:min-h-auto flex-1">
             <div className="mb-2 flex items-center justify-between shrink-0">
                <Label className="text-sm font-bold text-[#37352F] flex items-center gap-2">
                    <User className="h-3.5 w-3.5" />
                    患者選択
                </Label>
                <span className="text-sm text-muted-foreground">リストから選択してください</span>
             </div>
             <div className="flex-1 overflow-hidden flex flex-col min-h-0">
                <PatientSelectionTable 
                    onSelect={handlePetSelect} 
                    selectedPets={selectedPets}
                />
             </div>
          </div>

          {/* Right Panel: Reservation Form */}
          <div className="w-full lg:w-5/12 bg-white flex flex-col overflow-hidden min-h-[300px] lg:min-h-auto flex-1">
            <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
                 {/* Selected Patient Summary (Top of Form) */}
                <div className={`rounded-lg border p-3 transition-colors ${selectedPets.length > 0 ? "bg-blue-50/50 border-blue-100" : "bg-gray-50 border-gray-100 border-dashed"}`}>
                    <Label className="text-sm text-[#37352F]/60 font-bold uppercase tracking-wider block mb-2">
                        予約対象（選択中）
                    </Label>
                    
                    {selectedPets.length > 0 ? (
                        <div className="flex flex-col gap-2">
                            {selectedPets.map(pet => (
                                <div key={pet.id} className="flex items-center gap-2 bg-white p-2 rounded border border-blue-100 shadow-sm">
                                    <span className="text-sm font-bold text-[#37352F]">{pet.name}</span>
                                    <Badge variant="outline" className="text-sm font-normal text-muted-foreground bg-gray-50 h-6">
                                        {pet.species}
                                    </Badge>
                                    <span className="text-sm text-muted-foreground ml-auto">
                                        No. {pet.ownerId} {pet.ownerName}
                                    </span>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="flex items-center justify-center h-16 text-sm text-muted-foreground">
                            左側のリストから患者を選択してください
                        </div>
                    )}
                </div>

                {/* Form Fields */}
                <div className="space-y-4">
                    <Label className="text-sm font-bold text-[#37352F]">予約詳細</Label>
                    <ReservationFormFields formData={formData} onChange={setFormData} />
                </div>
            </div>
          </div>
        </div>

        <DialogFooter className="p-4 border-t bg-white shrink-0 h-16 flex items-center justify-end gap-2">
            <Button variant="outline" onClick={onClose} className="h-10 text-sm">
                キャンセル
            </Button>
            <Button
                onClick={handleSave}
                disabled={selectedPets.length === 0}
                className="bg-[#37352F] text-white hover:bg-[#37352F]/90 h-10 text-sm min-w-[100px]"
            >
                {isEditMode ? "更新する" : "予約を確定"}
            </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
