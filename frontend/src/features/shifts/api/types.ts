/**
 * Shifts API types — EMR-514 固定API契約
 * 応答DTOは contract-types.ts の ShiftEntry（正十進文字列ID・version・interval_state）。
 */
export type {
  CreateShiftRequest,
  ShiftContractError,
  ShiftContractErrorCode,
  ShiftIntervalState,
  ShiftMutationIntent,
  ShiftMutationReceipt,
  ShiftOperationResult,
  ShiftProtectedBooking,
  ShiftWorkBreak,
  ShiftWorkBreakView,
} from "./contract-types";
export { SHIFT_CONTRACT_ERROR_CODES, isShiftContractErrorPayload } from "./contract-types";
