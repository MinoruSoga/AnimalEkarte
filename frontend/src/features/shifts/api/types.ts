/**
 * Shifts API types
 * Backend types: {@link ShiftEntry} from models.ts (tygo generated)
 */
import type { ShiftEntry } from "@/types/generated/models";

/** Backend response — ShiftEntry + API が追加で返す staff_name */
export type BackendShift = ShiftEntry & {
  staff_name?: string;
};
