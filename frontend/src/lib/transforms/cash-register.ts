import type { CashRegisterClose as BackendCashRegisterClose } from "@/types/generated/models";

// BE 契約: period は "am" | "pm" | "emg" のみ（features/cash-register/constants.ts の
// CashRegisterPeriod と同一値域）。lib/transforms は feature に依存しないためここで narrow する。
export function transformCashRegisterClose(raw: BackendCashRegisterClose) {
  return {
    id: String(raw.id ?? 0),
    clinicId: String(raw.clinic_id ?? 0),
    closeDate: raw.close_date,
    period: raw.period as "am" | "pm" | "emg",
    theoreticalCash: raw.theoretical_cash,
    actualCash: raw.actual_cash,
    cashDifference: raw.cash_difference,
    categoryBreakdown: raw.category_breakdown as unknown,
    memo: raw.memo,
    closedBy: raw.closed_by ?? null,
    closedByStaffName: raw.closed_by_staff?.name ?? undefined,
    closedAt: raw.closed_at,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at,
  };
}

export type CashRegisterClose = ReturnType<typeof transformCashRegisterClose>;
