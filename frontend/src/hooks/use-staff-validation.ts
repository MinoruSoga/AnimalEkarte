import { useMemo } from "react";
import { useMasterItems } from "@/features/master";

/**
 * Hook that provides validation for active staff names.
 * Returns a Set of valid staff names for efficient lookup.
 */
export function useStaffValidation() {
  const { data: staffItems } = useMasterItems("staff");

  const validStaffNames = useMemo(() => {
    return new Set(
      staffItems
        .filter((item) => item.status === "active")
        .map((item) => item.name)
    );
  }, [staffItems]);

  const isValidStaff = (name: string): boolean => {
    return validStaffNames.has(name);
  };

  return { isValidStaff, validStaffNames };
}
