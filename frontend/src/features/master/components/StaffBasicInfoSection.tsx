import { useCallback, useMemo, type ChangeEvent, type Dispatch, type SetStateAction } from "react";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { C, STYLE } from "@/lib/design-tokens";

import type { Staff } from "../api/staffs";
import type { Occupation } from "../api/occupations";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import type { StaffFormData } from "../lib/staff-side-panel-model";

interface StaffBasicInfoSectionProps {
  item: Staff | null;
  isNew: boolean;
  formData: StaffFormData;
  setFormDataDirty: Dispatch<SetStateAction<StaffFormData>>;
  allOccupations: Occupation[];
}

export function StaffBasicInfoSection({
  item,
  isNew,
  formData,
  setFormDataDirty,
  allOccupations,
}: StaffBasicInfoSectionProps) {
  const occupationSelectItems = useMemo(
    () =>
      allOccupations
        .filter((occupation) => occupation.isActive)
        .map((occupation) => (
          <SelectItem key={occupation.id} value={occupation.id}>
            {occupation.name}
          </SelectItem>
        )),
    [allOccupations],
  );

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleOccupationChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, jobTitleId: value }));
    },
    [setFormDataDirty],
  );

  const handleLicenseNumberChange = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => {
      setFormDataDirty((prev) => ({ ...prev, licenseNumber: event.target.value }));
    },
    [setFormDataDirty],
  );

  const handleEmailChange = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => {
      setFormDataDirty((prev) => ({ ...prev, email: event.target.value }));
    },
    [setFormDataDirty],
  );

  const handlePasswordChange = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => {
      setFormDataDirty((prev) => ({ ...prev, password: event.target.value }));
    },
    [setFormDataDirty],
  );

  return (
    <>
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />

      <PropertyRow label="職種">
        <Select value={formData.jobTitleId ?? undefined} onValueChange={handleOccupationChange}>
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>{occupationSelectItems}</SelectContent>
        </Select>
      </PropertyRow>

      <PropertyRow label="資格番号">
        <input
          type="text"
          aria-label="資格番号"
          className={MASTER_INPUT_CLASS}
          value={formData.licenseNumber}
          onChange={handleLicenseNumberChange}
          placeholder="空"
        />
      </PropertyRow>

      {isNew ? (
        <>
          <PropertyRow label="メールアドレス">
            <input
              type="email"
              aria-label="メールアドレス"
              className={MASTER_INPUT_CLASS}
              value={formData.email}
              onChange={handleEmailChange}
              placeholder="例: staff@clinic.com"
            />
          </PropertyRow>
          <PropertyRow label="パスワード">
            <input
              type="password"
              aria-label="パスワード"
              className={MASTER_INPUT_CLASS}
              value={formData.password}
              onChange={handlePasswordChange}
              placeholder="8文字以上"
            />
          </PropertyRow>
        </>
      ) : (
        <>
          <PropertyRow label="メールアドレス">
            <span className={`text-sm ${C.text65}`}>{item?.email || "未設定"}</span>
          </PropertyRow>
          <PropertyRow label="パスワード">
            <input
              type="password"
              aria-label="パスワード"
              className={MASTER_INPUT_CLASS}
              value={formData.password}
              onChange={handlePasswordChange}
              placeholder="変更する場合のみ入力"
            />
          </PropertyRow>
        </>
      )}
    </>
  );
}
