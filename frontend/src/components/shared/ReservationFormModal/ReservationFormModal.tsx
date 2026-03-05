// React/Framework
import { useState, useEffect } from "react";

// External
import { User, Calendar, CalendarCheck, PawPrint, X, Search } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { usePetInfo } from "@/hooks/use-pet";
import { usePetSelection } from "@/hooks/use-pet-selection";

// Relative
import { PatientSelectionTable } from "./PatientSelectionTable";
import { ReservationFormFields } from "./ReservationFormFields";

// Types
import type { Pet, ReservationAppointment } from "@/types";

interface ReservationFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: Partial<ReservationAppointment>, selectedPets: Pet[]) => void;
  initialData: Partial<ReservationAppointment> | null;
}

function PetLoader({
  petId,
  onPetLoaded,
}: {
  petId: string;
  onPetLoaded: (pet: Pet) => void;
}) {
  const { pet } = usePetInfo(petId);

  useEffect(() => {
    if (pet) {
      onPetLoaded(pet);
    }
  }, [pet, onPetLoaded]);

  return null;
}

function StepIndicator({ step, label, active }: { step: number; label: string; active: boolean }) {
  return (
    <div className={`flex items-center gap-1.5 text-xs ${active ? "text-blue-600" : "text-[#37352F]/30"}`}>
      <span
        className={`w-5 h-5 rounded-full flex items-center justify-center text-[11px] font-bold transition-colors ${
          active ? "bg-blue-600 text-white" : "bg-[#37352F]/10 text-[#37352F]/30"
        }`}
      >
        {step}
      </span>
      {label}
    </div>
  );
}

function SelectedPetChip({ pet, onRemove }: { pet: Pet; onRemove: () => void }) {
  return (
    <div className="flex items-center gap-2 bg-white p-2 rounded-lg border border-[rgba(55,53,47,0.12)] shadow-sm">
      <PawPrint className="h-4 w-4 text-[#37352F]/60 flex-shrink-0" />
      <span className="text-sm font-bold text-[#37352F]">{pet.name}</span>
      <Badge variant="outline" className="text-[11px] font-normal text-[#37352F]/60 bg-[#F7F6F3] border-[rgba(55,53,47,0.12)] h-5">
        {pet.species}
      </Badge>
      <span className="text-[11px] text-[#37352F]/60 ml-auto">
        No. {pet.ownerId} {pet.ownerName}
      </span>
      <button
        onClick={onRemove}
        className="ml-1 p-1 hover:bg-red-50 rounded transition-colors"
      >
        <X className="h-4 w-4 text-red-600 hover:text-red-700" />
      </button>
    </div>
  );
}

export const ReservationFormModal = ({
  isOpen,
  onClose,
  onSave,
  initialData,
}: ReservationFormModalProps) => {
  const [formData, setFormData] = useState<Partial<ReservationAppointment>>({});
  const [pendingPetId, setPendingPetId] = useState<string | null>(null);
  const [mobilePanel, setMobilePanel] = useState<"search" | "form">("search");

  const {
    selectedPets,
    setSelectedPets,
    togglePetSelection,
  } = usePetSelection([], "multiple-same-owner");

  const [prevIsOpen, setPrevIsOpen] = useState(false);

  if (isOpen !== prevIsOpen) {
    setPrevIsOpen(isOpen);
    if (isOpen) {
      if (initialData) {
        setFormData({ ...initialData });
        if (initialData.petId) {
          setPendingPetId(initialData.petId);
        } else {
          setSelectedPets([]);
          setPendingPetId(null);
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
        setPendingPetId(null);
      }
    }
  }

  const handlePetLoaded = (pet: Pet) => {
    setSelectedPets([pet]);
    setPendingPetId(null);
  };

  const handleSave = () => {
    onSave(formData, selectedPets);
  };

  const handlePetSelect = (pet: Pet) => {
    togglePetSelection(pet);
  };

  const isEditMode = initialData && initialData.id;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      {/* Load pet data from API when petId is provided */}
      {pendingPetId && (
        <PetLoader petId={pendingPetId} onPetLoaded={handlePetLoaded} />
      )}

      <DialogContent className="w-[98%] sm:max-w-[1200px] h-[90vh] flex flex-col p-0 gap-0 bg-white overflow-hidden rounded-xl">
        <DialogHeader className="p-4 border-b shrink-0 h-auto flex flex-col gap-3 space-y-0">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {isEditMode ? (
                <CalendarCheck className="h-5 w-5 text-amber-600" />
              ) : (
                <Calendar className="h-5 w-5 text-blue-600" />
              )}
              <DialogTitle className="text-sm font-bold text-[#37352F]">
                {isEditMode ? "予約編集" : "新規予約作成"}
              </DialogTitle>
            </div>
            <DialogDescription className="sr-only">
              左側のリストからペットを選択し、右側のフォームで予約情報を入力してください
            </DialogDescription>
          </div>
          <div className="flex items-center gap-6">
            <StepIndicator step={1} label="患者選択" active={mobilePanel === "search"} />
            <StepIndicator step={2} label="予約情報" active={mobilePanel === "form"} />
          </div>
          {/* Mobile Tab Bar */}
          <div className="flex gap-2 lg:hidden border-t pt-3 -mx-4 px-4">
            <Button
              size="sm"
              variant={mobilePanel === "search" ? "default" : "outline"}
              onClick={() => setMobilePanel("search")}
              className="flex-1 h-9 text-sm"
            >
              患者選択
            </Button>
            <Button
              size="sm"
              variant={mobilePanel === "form" ? "default" : "outline"}
              onClick={() => setMobilePanel("form")}
              className="flex-1 h-9 text-sm"
            >
              予約情報
            </Button>
          </div>
        </DialogHeader>

        <div className="flex flex-col lg:flex-row flex-1 overflow-hidden">
          {/* Left Panel: Patient Selection Table */}
          <div
            className={cn(
              "w-full lg:w-7/12 border-b lg:border-b-0 lg:border-r bg-[#FAFAF8] p-4 flex flex-col overflow-hidden min-h-[300px] lg:min-h-auto flex-1",
              mobilePanel !== "search" && "hidden lg:flex"
            )}
          >
            <div className="mb-3 flex items-center gap-2 shrink-0">
              <Search className="h-4 w-4 text-[#37352F]/60" />
              <Label className="text-sm font-bold text-[#37352F]">患者検索</Label>
            </div>
            <div className="flex-1 overflow-hidden flex flex-col min-h-0">
              <PatientSelectionTable
                onSelect={handlePetSelect}
                selectedPets={selectedPets}
              />
            </div>
          </div>

          {/* Right Panel: Reservation Form */}
          <div
            className={cn(
              "w-full lg:w-5/12 bg-white flex flex-col overflow-hidden min-h-[300px] lg:min-h-auto flex-1",
              mobilePanel !== "form" && "hidden lg:flex"
            )}
          >
            <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
              {/* Selected Patient Summary (Top of Form) */}
              <div className={`rounded-lg border p-3 transition-colors ${selectedPets.length > 0 ? "bg-gradient-to-r from-blue-50/50 to-cyan-50/50 border-blue-100" : "bg-[#F7F6F3] border-[rgba(55,53,47,0.12)]"}`}>
                <Label className="text-[12px] text-[#37352F]/40 font-bold tracking-widest uppercase block mb-3">
                  予約対象（選択中）
                </Label>

                {selectedPets.length > 0 ? (
                  <div className="flex flex-col gap-2">
                    {selectedPets.map(pet => (
                      <SelectedPetChip
                        key={pet.id}
                        pet={pet}
                        onRemove={() => {
                          setSelectedPets(selectedPets.filter(p => p.id !== pet.id));
                        }}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center h-20 text-center">
                    <PawPrint className="h-6 w-6 text-[#37352F]/10 mb-2" />
                    <div className="text-[12px] text-[#37352F]/40">
                      左側から患者を選択してください
                    </div>
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

        <DialogFooter className="p-4 border-t bg-white shrink-0 h-14 flex items-center justify-between gap-2">
          <div className="flex items-center gap-1.5">
            <PawPrint className="h-4 w-4 text-[#37352F]/60" />
            <span className="text-sm text-[#37352F]/60">
              {selectedPets.length}頭 選択中
            </span>
          </div>
          <div className="flex gap-2">
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
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
