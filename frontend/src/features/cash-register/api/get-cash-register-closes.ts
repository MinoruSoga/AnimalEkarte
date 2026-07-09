// R-F2-S11: shared hook (@/hooks/use-cash-register-closes) へ昇格。
// accounting から cash-register feature への直接 import を避けるための re-export。
export {
  getCashRegisterCloses,
  useGetCashRegisterCloses,
  type GetCashRegisterClosesParams,
  type CashRegisterClosesResponse,
  type CashRegisterClose,
} from "@/hooks/use-cash-register-closes";
