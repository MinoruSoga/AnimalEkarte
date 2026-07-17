import type { OwnerData } from "../types";

/** FE3-14: OwnerBasicFields / OwnerAddressFields / OwnerMembershipFields で共有する props 基底型。 */
export interface OwnerFieldSectionProps {
  ownerData: OwnerData;
  fieldErrors: Record<string, string>;
  onChange: (field: string, value: string | boolean | number | null | undefined) => void;
  onClearError: (field: string) => void;
}
