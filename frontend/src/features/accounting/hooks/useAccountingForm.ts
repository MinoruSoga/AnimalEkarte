import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router-dom";
import { format } from "date-fns";
import { toast } from "sonner@2.0.3";
import { AccountingRecord } from "../../../types";
import { MOCK_ACCOUNTING_RECORDS, MOCK_PETS } from "../../../lib/constants";
import { usePetSelection } from "../../pets/hooks/usePetSelection";

export function useAccountingForm(id?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  const petSelection = usePetSelection();
  const { selectedPets, setSelectedPets } = petSelection;

  const [formData, setFormData] = useState<Partial<AccountingRecord>>({
    status: "未収",
    method: "-",
    ownerName: "",
    petName: "",
    date: format(new Date(), "yyyy/MM/dd HH:mm"),
    amount: 0,
    note: ""
  });

  useEffect(() => {
    // Handle passed accounting items from Medical Record
    if (location.state?.accountingItems) {
        const items = location.state.accountingItems as any[];
        // Calculate total consistent with MedicalRecordBillCheck (Subtotal * 1.1)
        // Assuming 10% tax for now as per MedicalRecordBillCheck implementation
        const subtotal = items.reduce((sum, item) => {
            return sum + (item.unitPrice * item.quantity);
        }, 0);
        
        const tax = Math.floor(subtotal * 0.1);
        const totalAmount = subtotal + tax;

        setFormData(prev => ({
            ...prev,
            amount: totalAmount,
            status: "未収" // Ensure status is reset or set correctly for new bill
        }));
    }

    if (isEdit && id) {
        const record = MOCK_ACCOUNTING_RECORDS.find(r => r.id === id);
        if (record) {
            setFormData(record);
            
            const normalize = (s: string) => s.replace(/[\s\u3000]/g, "");
            
            const pet = MOCK_PETS.find(p => {
                const pName = normalize(p.name);
                const rName = normalize(record.petName);
                const pOwner = normalize(p.ownerName);
                const rOwner = normalize(record.ownerName);
                
                return (pName.includes(rName) || rName.includes(pName)) && 
                       (pOwner.includes(rOwner) || rOwner.includes(pOwner));
            });

            if (pet) {
                setSelectedPets([pet]);
            }
        }
    } else {
        if (petId) {
            const foundPet = MOCK_PETS.find(p => p.id === petId);
            if (foundPet) {
                setSelectedPets([foundPet]);
            } else {
                navigate("/accounting/select-pet");
            }
        }
    }
  }, [id, isEdit, petId, setSelectedPets]);

  useEffect(() => {
    if (selectedPets.length > 0) {
      const pet = selectedPets[0];
      setFormData((prev) => ({
        ...prev,
        ownerName: pet.ownerName,
        petName: pet.name,
      }));
    }
  }, [selectedPets]);

  const handleSave = () => {
      toast.success(isEdit ? "会計情報を更新しました" : "会計情報を登録しました");
      setTimeout(() => {
          if (location.state?.from) {
              navigate(location.state.from);
          } else {
              navigate("/accounting");
          }
      }, 500);
  };

  return {
    formData,
    setFormData,
    petSelection,
    handleSave,
    isEdit
  };
}
