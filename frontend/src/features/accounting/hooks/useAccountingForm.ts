import { useState, useEffect, useMemo, useRef } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { format } from "date-fns";
import { toast } from "sonner";
import { AccountingRecord } from "@/types";
import { AccountingItem } from "../types";
import { MOCK_ACCOUNTING_RECORDS, MOCK_PETS } from "@/config/mock-data";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { findPetByRecord } from "@/utils/pet-matching";

// Helper to calculate amount from accounting items
function calculateAmountFromItems(items: AccountingItem[]): number {
  const subtotal = items.reduce((sum, item) => sum + (item.unitPrice * item.quantity), 0);
  const tax = Math.floor(subtotal * 0.1);
  return subtotal + tax;
}


export function useAccountingForm(id?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  // Compute initial values from location state and params
  const locationState = location.state as { accountingItems?: AccountingItem[] } | null;
  const initialData = useMemo(() => {
    const baseData: Partial<AccountingRecord> = {
      status: "未収",
      method: "-",
      ownerName: "",
      petName: "",
      date: format(new Date(), "yyyy/MM/dd HH:mm"),
      amount: 0,
      note: ""
    };

    // Handle passed accounting items from Medical Record
    if (locationState?.accountingItems) {
      baseData.amount = calculateAmountFromItems(locationState.accountingItems);
    }

    // Handle edit mode
    if (isEdit && id) {
      const record = MOCK_ACCOUNTING_RECORDS.find(r => r.id === id);
      if (record) {
        return record;
      }
    }

    return baseData;
  }, [id, isEdit, locationState]);

  // Compute initial pet selection
  const initialPet = useMemo(() => {
    if (isEdit && id) {
      const record = MOCK_ACCOUNTING_RECORDS.find(r => r.id === id);
      if (record) {
        return findPetByRecord(record.petName, record.ownerName);
      }
    } else if (petId) {
      return MOCK_PETS.find(p => p.id === petId);
    }
    return undefined;
  }, [id, isEdit, petId]);

  const petSelection = usePetSelection();
  const { selectedPets, setSelectedPets } = petSelection;

  const [formData, setFormData] = useState<Partial<AccountingRecord>>(initialData);

  // Track if we've done initial setup
  const initializedRef = useRef(false);

  // Initialize pet selection and handle navigation (only once)
  useEffect(() => {
    if (initializedRef.current) return;
    initializedRef.current = true;

    if (initialPet) {
      setSelectedPets([initialPet]);
    } else if (!isEdit && petId) {
      // Pet ID provided but not found - redirect
      navigate("/accounting/select-pet");
    }
  }, [initialPet, isEdit, petId, setSelectedPets, navigate]);

  // Sync form data with selected pet (derived state pattern)
  const formDataWithPet = useMemo(() => {
    if (selectedPets.length > 0) {
      const pet = selectedPets[0];
      return {
        ...formData,
        ownerName: pet.ownerName,
        petName: pet.name,
      };
    }
    return formData;
  }, [formData, selectedPets]);

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
    formData: formDataWithPet,
    setFormData,
    petSelection,
    handleSave,
    isEdit
  };
}
