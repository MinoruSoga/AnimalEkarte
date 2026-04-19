export { useGetStaffs } from "@/hooks/use-staffs";
export type { StaffItem } from "@/hooks/use-staffs";

import type { StaffItem } from "@/hooks/use-staffs";

/** staffId → スタッフ名のMapを構築する */
export function buildStaffMap(staffs: StaffItem[]): Map<string, string> {
  return new Map(staffs.map((s) => [s.id, s.name]));
}
