import { getStoredClinicId } from "@/lib/current-clinic";

/** 現在の選択クリニック ID を localStorage から取得する。 */
export function getClinicId(): string | null {
  return getStoredClinicId();
}
