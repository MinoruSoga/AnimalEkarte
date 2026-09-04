// R-F2-S11: shared hook (@/hooks/use-cash-register-closes) へ昇格。
// accounting から cash-register feature への直接 import を避けるための re-export。
export { useGetCashRegisterCloses, type CashRegisterClose } from "@/hooks/use-cash-register-closes";
