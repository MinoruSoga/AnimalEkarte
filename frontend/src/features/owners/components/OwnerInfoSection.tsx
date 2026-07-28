import { memo } from "react";
import type { ClinicMembership } from "@/types/auth";
import type { MembershipTypeLabel, OwnerData } from "../types";
import { OwnerAddressFields } from "./OwnerAddressFields";
import { OwnerBasicFields } from "./OwnerBasicFields";
import { OwnerMembershipFields } from "./OwnerMembershipFields";

interface OwnerInfoSectionProps {
  ownerData: OwnerData;
  fieldErrors: Record<string, string>;
  isEdit: boolean;
  canEditDiscount: boolean;
  /** #84: 登録先医院の選択肢（ユーザー所属医院）。2件以上かつ新規登録時のみセレクトを表示 */
  clinicOptions?: ClinicMembership[];
  /** #84: 未選択時に表示する現在の医院ID */
  currentClinicId?: string | null;
  onChange: (field: string, value: string | boolean | number | null | undefined) => void;
  onClearError: (field: string) => void;
  onMembershipChange: (type: MembershipTypeLabel) => void;
  onPostalCodeLookup: (postalCodeField: string, addressField: string) => void;
}

export const OwnerInfoSection = memo(function OwnerInfoSection({
  ownerData,
  fieldErrors,
  isEdit,
  canEditDiscount,
  clinicOptions,
  currentClinicId,
  onChange,
  onClearError,
  onMembershipChange,
  onPostalCodeLookup,
}: OwnerInfoSectionProps) {
  return (
    <div className="grid w-full grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <OwnerBasicFields
        ownerData={ownerData}
        fieldErrors={fieldErrors}
        isEdit={isEdit}
        clinicOptions={clinicOptions}
        currentClinicId={currentClinicId}
        onChange={onChange}
        onClearError={onClearError}
      />
      <OwnerMembershipFields
        ownerData={ownerData}
        fieldErrors={fieldErrors}
        canEditDiscount={canEditDiscount}
        onChange={onChange}
        onClearError={onClearError}
        onMembershipChange={onMembershipChange}
      />
      <OwnerAddressFields
        ownerData={ownerData}
        fieldErrors={fieldErrors}
        onChange={onChange}
        onClearError={onClearError}
        onPostalCodeLookup={onPostalCodeLookup}
      />
    </div>
  );
});
