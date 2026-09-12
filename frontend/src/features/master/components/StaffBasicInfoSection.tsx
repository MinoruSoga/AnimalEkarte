import {
  useCallback,
  useMemo,
  useState,
  type ChangeEvent,
  type Dispatch,
  type SetStateAction,
} from "react";
import { Link } from "react-router";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { paths } from "@/config/paths";
import { useAuth } from "@/hooks/use-auth";
import { C, STYLE } from "@/lib/design-tokens";

import { useAttachStaffAccount, type Staff } from "../api/staffs";
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
  const { user } = useAuth();
  const attachAccount = useAttachStaffAccount();
  const [attachEmail, setAttachEmail] = useState("");
  const canAttachAccount = user?.isSystemAdmin === true && !isNew && Boolean(item) && !item?.email;

  const hasOccupationMaster = allOccupations.length > 0;

  const occupationSelectItems = useMemo(() => {
    const selectedId = formData.jobTitleId;
    const options = allOccupations.filter(
      (occupation) =>
        occupation.isActive || (selectedId !== null && occupation.id === selectedId),
    );

    return options.map((occupation) => (
      <SelectItem key={occupation.id} value={occupation.id}>
        {occupation.name}
      </SelectItem>
    ));
  }, [allOccupations, formData.jobTitleId]);

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

  const handleAttachEmailChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    setAttachEmail(event.target.value);
  }, []);

  const handleAttachAccount = useCallback(async () => {
    if (item === null) {
      return;
    }
    const result = await attachAccount.mutateAsync({ id: item.id, email: attachEmail });
    toast.success(result.message);
    setAttachEmail("");
  }, [attachAccount, attachEmail, item]);

  return (
    <>
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />

      <PropertyRow label="職種">
        {hasOccupationMaster ? (
          <Select value={formData.jobTitleId ?? undefined} onValueChange={handleOccupationChange}>
            <SelectTrigger className={STYLE.selectCompact}>
              <SelectValue placeholder="選択" />
            </SelectTrigger>
            <SelectContent>{occupationSelectItems}</SelectContent>
          </Select>
        ) : (
          <div className="flex flex-col gap-1">
            <p className={`text-sm ${C.text40}`}>職種が未登録です</p>
            <Link
              to={paths.settings.occupations.getHref()}
              className={`text-sm underline ${C.text}`}
            >
              職種マスタを開く
            </Link>
          </div>
        )}
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
          {canAttachAccount ? (
            <>
              <PropertyRow label="ログインアカウント">
                <input
                  type="email"
                  aria-label="追加するメールアドレス"
                  className={MASTER_INPUT_CLASS}
                  value={attachEmail}
                  onChange={handleAttachEmailChange}
                  placeholder="本人専用のメールアドレス"
                />
              </PropertyRow>
              <Button
                type="button"
                size="sm"
                disabled={attachAccount.isPending || attachEmail.trim() === ""}
                onClick={() => {
                  void handleAttachAccount();
                }}
              >
                ログインアカウントを追加
              </Button>
            </>
          ) : item?.email ? (
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
          ) : null}
        </>
      )}
    </>
  );
}
