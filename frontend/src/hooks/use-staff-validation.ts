import { useMemo } from "react";
import { useGetStaffs } from "@/hooks/use-staffs";

/**
 * Hook that provides validation for active staff names.
 * Shares the raw `/v1/masters/staffs` cache with useGetStaffs (STG P0-3).
 */
export function useStaffValidation() {
  const { data: staffItems } = useGetStaffs();

  const validStaffNames = useMemo(() => {
    return new Set((staffItems ?? []).flatMap((item) => (item.isActive ? [item.name] : [])));
  }, [staffItems]);

  const isValidStaff = (name: string): boolean => {
    return validStaffNames.has(name);
  };

  return { isValidStaff, validStaffNames };
}
