import { useMemo } from "react";
import { useGetMasterItems } from "@/hooks/use-master-items";

/**
 * Hook that provides validation for active staff names.
 * Returns a Set of valid staff names for efficient lookup.
 */
export function useStaffValidation() {
  const { data: staffItems } = useGetMasterItems("staff");

  const validStaffNames = useMemo(() => {
    return new Set(staffItems.flatMap((item) => item.status === "active" ? [item.name] : []));
  }, [staffItems]);

  const isValidStaff = (name: string): boolean => {
    return validStaffNames.has(name);
  };

  return { isValidStaff, validStaffNames };
}
