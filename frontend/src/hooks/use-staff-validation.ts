import { useMemo } from "react";
import { useGetStaffs } from "@/hooks/use-staffs";

/**
 * Hook that provides validation for active staff names.
 * Shares the raw `/v1/masters/staffs` cache with useGetStaffs (STG P0-3).
 */
export function useStaffValidation() {
  const { data: staffItems } = useGetStaffs();

  // EMR-412 R5: マスタ未取得（loading/error）の間は「無効」と断定しない。
  // 取得済みのときだけアクティブ氏名集合で判定する。
  const staffNames = useMemo(() => {
    if (staffItems === undefined) return undefined;
    return new Set(staffItems.flatMap((item) => (item.isActive ? [item.name] : [])));
  }, [staffItems]);

  const isValidStaff = (name: string): boolean => {
    // 未取得中は検証不能 → 無効マークを出さず氏名のまま見せる。
    if (staffNames === undefined) return true;
    return staffNames.has(name);
  };

  return { isValidStaff, validStaffNames: staffNames };
}
