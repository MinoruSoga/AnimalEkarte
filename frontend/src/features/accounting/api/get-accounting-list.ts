import type { Accounting } from "../types";
import { MOCK_ACCOUNTING_LIST } from "./mock-data";

/**
 * Get all accounting records
 * Currently returns mock data, will be replaced with API call
 */
export const getAccountingList = async (): Promise<Accounting[]> => {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 100));
  return MOCK_ACCOUNTING_LIST;
};

/**
 * Get accounting records synchronously (for hooks that don't need async)
 */
export const getAccountingListSync = (): Accounting[] => {
  return MOCK_ACCOUNTING_LIST;
};
