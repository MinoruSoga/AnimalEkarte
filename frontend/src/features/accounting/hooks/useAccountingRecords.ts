import { useMemo } from "react";
import type { AccountingRecord } from "@/types";
import type { Accounting, PaymentMethod } from "../types";
import { useGetAccountings } from "../api/get-accountings";

function toListAccounting(r: AccountingRecord): Accounting {
  const isCompleted = r.status === "回収済";
  const method: PaymentMethod =
    r.method === "現金"
      ? "cash"
      : r.method === "クレジットカード"
        ? "credit_card"
        : r.method === "電子マネー"
          ? "electronic_money"
          : "cash";

  return {
    id: r.id,
    ownerId: "",
    ownerName: r.ownerName,
    petId: "",
    petName: r.petName,
    status: isCompleted ? "completed" : r.status === "キャンセル" ? "canceled" : "waiting",
    scheduledDate: r.date,
    items: [],
    payment: isCompleted
      ? {
          subtotal: 0,
          taxTotal: 0,
          totalAmount: r.amount,
          insuranceAmount: 0,
          discountAmount: 0,
          billingAmount: r.amount,
          receivedAmount: r.amount,
          changeAmount: 0,
          method,
        }
      : undefined,
    memo: r.note,
  };
}

export function useAccountingRecords(searchTerm: string) {
  const { data: apiRecords = [], isLoading, isError } = useGetAccountings();

  const filteredRecords = useMemo(() => {
    const list: Accounting[] = apiRecords.map(toListAccounting);
    if (!searchTerm) return list;
    const lowerTerm = searchTerm.toLowerCase();
    return list.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm),
    );
  }, [apiRecords, searchTerm]);

  return { data: filteredRecords, isLoading, isError };
}
