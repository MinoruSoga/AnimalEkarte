import { useState, useEffect, useMemo, useRef } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { format } from "date-fns";
import { toast } from "sonner";
import type { AccountingRecord } from "@/types";
import type { AccountingItem } from "../types";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetAccounting } from "../api/get-accounting";

// Helper to calculate amount from accounting items
function calculateAmountFromItems(items: AccountingItem[]): number {
  const subtotal = items.reduce((sum, item) => sum + item.unitPrice * item.quantity, 0);
  const tax = Math.floor(subtotal * 0.1);
  return subtotal + tax;
}

export function useAccountingForm(id?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  const locationState = location.state as { accountingItems?: AccountingItem[] } | null;

  // 編集モード時は API からデータ取得
  const { data: existingRecord, isLoading: isLoadingRecord } = useGetAccounting(id ?? "");

  // 新規作成時のデフォルト値
  const newRecordDefault = useMemo<Partial<AccountingRecord>>(() => {
    const base: Partial<AccountingRecord> = {
      status: "未収",
      method: "-",
      ownerName: "",
      petName: "",
      date: format(new Date(), "yyyy/MM/dd HH:mm"),
      amount: 0,
      note: "",
    };
    if (locationState?.accountingItems) {
      base.amount = calculateAmountFromItems(locationState.accountingItems);
    }
    return base;
  }, [locationState]);

  const [formData, setFormData] = useState<Partial<AccountingRecord>>(newRecordDefault);

  // 表示用データ: 編集モードかつ API データ取得済みなら API データ優先（derived state パターン）
  const baseFormData: Partial<AccountingRecord> =
    isEdit && existingRecord ? existingRecord : formData;

  const petSelection = usePetSelection();
  const { setSelectedPets } = petSelection;
  const initializedRef = useRef(false);

  // 新規作成モード: petId が URL にある場合、選択済みペットがなければリダイレクト
  useEffect(() => {
    if (!isEdit && petId && !initializedRef.current) {
      initializedRef.current = true;
      if (petSelection.selectedPets.length === 0) {
        navigate("/accounting/select-pet");
      }
    }
  }, [isEdit, petId, petSelection.selectedPets.length, navigate, setSelectedPets]);

  // 選択済みペット情報を baseFormData に同期
  const formDataWithPet = useMemo(() => {
    if (petSelection.selectedPets.length > 0) {
      const pet = petSelection.selectedPets[0];
      return {
        ...baseFormData,
        ownerName: pet.ownerName,
        petName: pet.name,
      };
    }
    return baseFormData;
  }, [baseFormData, petSelection.selectedPets]);

  const handleSave = () => {
    toast.success(isEdit ? "会計情報を更新しました" : "会計情報を登録しました");
    setTimeout(() => {
      const fromPath = (location.state as { from?: string } | null)?.from;
      if (fromPath) {
        navigate(fromPath);
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
    isEdit,
    isLoadingRecord,
  };
}
