import { memo, type ReactNode, type Dispatch, type SetStateAction } from "react";
import { Building2, Percent, Plus, X } from "lucide-react";

import { TableCell } from "@/components/ui/table";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, LAYOUT, STYLE } from "@/lib/design-tokens";
import { ResourceHospitalSettings } from "@/types/generated/models";
import type { Clinic } from "../api/clinics";

const COLUMNS = [
  { header: "院名" },
  { header: "電話番号", className: "w-[150px]" },
  { header: "メール" },
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
];

const PROP_INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`;

export interface ClinicFormData {
  name: string;
  postal_code: string;
  address: string;
  phone_number: string;
  fax_number: string;
  registration_number: string;
  director_name: string;
  email: string;
  website: string;
  is_active: boolean;
  standard_tax_rate: number;
  reduced_tax_rate: number;
}

export const DEFAULT_CLINIC_FORM_DATA: ClinicFormData = {
  name: "",
  postal_code: "",
  address: "",
  phone_number: "",
  fax_number: "",
  registration_number: "",
  director_name: "",
  email: "",
  website: "",
  is_active: true,
  standard_tax_rate: 0.1,
  reduced_tax_rate: 0.08,
};

export function clinicToFormData(item: Clinic): ClinicFormData {
  return {
    name: item.name,
    postal_code: item.postalCode,
    address: item.address,
    phone_number: item.phoneNumber,
    fax_number: item.faxNumber,
    registration_number: item.registrationNumber,
    director_name: item.directorName,
    email: item.email,
    website: item.website,
    is_active: item.isActive,
    standard_tax_rate: item.standardTaxRate ?? 0.1,
    reduced_tax_rate: item.reducedTaxRate ?? 0.08,
  };
}

interface ClinicMasterListProps {
  canCreate: boolean;
  canEdit: boolean;
  items: Clinic[];
  searchTerm: string;
  onSearchChange: (term: string) => void;
  onBack: () => void;
  onCreate: () => void;
  onEdit: (item: Clinic) => void;
}

export function ClinicMasterList({
  canCreate,
  canEdit,
  items,
  searchTerm,
  onSearchChange,
  onBack,
  onCreate,
  onEdit,
}: ClinicMasterListProps) {
  return (
    <PageLayout
      title="医院マスタ"
      resource={ResourceHospitalSettings}
      icon={<Building2 className={`${ICON.page} ${C.text}`} />}
      onBack={onBack}
      headerAction={
        canCreate ? (
          <PrimaryButton onClick={onCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <NotionFilter
          properties={[]}
          activeFilters={[]}
          onFilterChange={() => {}}
          searchTerm={searchTerm}
          onSearchChange={onSearchChange}
          searchPlaceholder="院名、電話番号、メールで検索..."
          count={items.length}
        />

        <DataTable
          columns={COLUMNS}
          data={items}
          emptyMessage="医院が登録されていません"
          renderRow={(item) => (
            <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
              <TableCell className={`font-medium text-sm ${C.text} py-2.5`}>
                {item.name}
              </TableCell>
              <TableCell className={`font-mono text-sm ${C.text80} py-2.5`}>
                {item.phoneNumber || "-"}
              </TableCell>
              <TableCell className={`text-sm ${C.text80} py-2.5`}>
                {item.email || "-"}
              </TableCell>
              <TableCell className="text-right py-2.5">
                <span className="inline-flex items-center gap-1.5">
                  <span className={`size-[7px] rounded-full ${item.isActive ? C.bgAccent : C.bgPrimary10}`} />
                  <span className={`text-sm ${item.isActive ? C.text65 : C.text35}`}>
                    {item.isActive ? "有効" : "無効"}
                  </span>
                </span>
              </TableCell>
            </DataTableRow>
          )}
        />
        {canCreate ? (
          <button
            type="button"
            onClick={onCreate}
            className={`flex items-center gap-1.5 w-full px-3 py-2 text-sm ${C.text40} ${C.hoverText60} ${C.hoverBgLight} transition-colors rounded`}
          >
            <Plus className={ICON.xs} />
            新しい医院を追加...
          </button>
        ) : null}
      </div>
    </PageLayout>
  );
}

interface ClinicMasterSidePanelProps {
  selectedItem: Clinic | null;
  formData: ClinicFormData;
  setFormData: Dispatch<SetStateAction<ClinicFormData>>;
  formAction: (payload: FormData) => void;
  nameError?: string;
  canEdit: boolean;
  canDelete: boolean;
  onClose: () => void;
  onDeleteClick: (item: Clinic) => void;
}

export function ClinicMasterSidePanel({
  selectedItem,
  formData,
  setFormData,
  formAction,
  nameError,
  canEdit,
  canDelete,
  onClose,
  onDeleteClick,
}: ClinicMasterSidePanelProps) {
  return (
    <div className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 flex flex-col`}>
      <div className={STYLE.sidePeekToolbar}>
        <span className={`text-xs ${C.text35} pl-1 select-none`}>
          {selectedItem ? "編集" : "新規作成"}
        </span>
        <div className="flex items-center gap-1">
          {selectedItem && canDelete ? (
            <DeleteIconButton onClick={() => onDeleteClick(selectedItem)} />
          ) : null}
          <button
            type="button"
            onClick={onClose}
            className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
          >
            <X className={ICON.action} />
          </button>
        </div>
      </div>

      <form action={formAction} className="flex-1 flex flex-col min-h-0">
        <fieldset disabled={!canEdit} className="border-0 p-0 m-0 min-w-0">
          <div className={STYLE.sidePeekBody}>
            <div className="px-16 pb-8">
              <div className="pt-4 pb-2">
                <div className={STYLE.pageIcon}>
                  <Building2 className={LAYOUT.pageIcon.innerIcon} />
                </div>
              </div>

              <div className="pb-1 mb-4">
                <input
                  type="text"
                  className={`w-full bg-transparent ${C.text} ${C.textPlaceholderFaint} outline-none border-none p-0`}
                  style={{
                    fontSize: LAYOUT.pageTitle.fontSize,
                    fontWeight: LAYOUT.pageTitle.fontWeight,
                    lineHeight: LAYOUT.pageTitle.lineHeight,
                  }}
                  value={formData.name}
                  onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
                  placeholder="無題"
                />
                <FormFieldError message={nameError} />
              </div>

              <div className={`${STYLE.sectionDivider} mb-1`} />

              <div className="py-1">
                <PropertyRow label="ステータス">
                  <button
                    type="button"
                    onClick={() =>
                      setFormData((prev) => ({ ...prev, is_active: !prev.is_active }))
                    }
                    className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
                  >
                    <NotionStatusPill status={formData.is_active ? "active" : "inactive"} />
                  </button>
                </PropertyRow>

                <ClinicTextProperty
                  label="郵便番号"
                  value={formData.postal_code}
                  placeholder="例: 150-0001"
                  onChange={(postalCode) =>
                    setFormData((prev) => ({ ...prev, postal_code: postalCode }))
                  }
                />
                <ClinicTextProperty
                  label="住所"
                  value={formData.address}
                  placeholder="例: 東京都渋谷区..."
                  onChange={(address) => setFormData((prev) => ({ ...prev, address }))}
                />
                <ClinicTextProperty
                  label="電話番号"
                  value={formData.phone_number}
                  placeholder="例: 03-1234-5678"
                  onChange={(phoneNumber) =>
                    setFormData((prev) => ({ ...prev, phone_number: phoneNumber }))
                  }
                />
                <ClinicTextProperty
                  label="FAX番号"
                  value={formData.fax_number}
                  placeholder="例: 03-1234-5679"
                  onChange={(faxNumber) =>
                    setFormData((prev) => ({ ...prev, fax_number: faxNumber }))
                  }
                />
                <ClinicTextProperty
                  label="登録番号"
                  value={formData.registration_number}
                  placeholder="例: 東京都獣医師会 第12345号"
                  onChange={(registrationNumber) =>
                    setFormData((prev) => ({
                      ...prev,
                      registration_number: registrationNumber,
                    }))
                  }
                />
                <ClinicTextProperty
                  label="院長名"
                  value={formData.director_name}
                  placeholder="例: 山田 太郎"
                  onChange={(directorName) =>
                    setFormData((prev) => ({ ...prev, director_name: directorName }))
                  }
                />
                <ClinicTextProperty
                  label="メール"
                  type="email"
                  value={formData.email}
                  placeholder="例: info@clinic.com"
                  onChange={(email) => setFormData((prev) => ({ ...prev, email }))}
                />
                <ClinicTextProperty
                  label="Webサイト"
                  value={formData.website}
                  placeholder="例: https://example.com"
                  onChange={(website) => setFormData((prev) => ({ ...prev, website }))}
                />

                <div className={`${STYLE.sectionDivider} my-2`} />
                <div className={`flex items-center gap-1.5 py-1.5 text-xs ${C.text45} select-none`}>
                  <Percent className={ICON.xs} />
                  消費税率
                </div>

                <ClinicTaxRateProperty
                  label="通常課税"
                  value={formData.standard_tax_rate}
                  onChange={(standardTaxRate) =>
                    setFormData((prev) => ({
                      ...prev,
                      standard_tax_rate: standardTaxRate,
                    }))
                  }
                />
                <ClinicTaxRateProperty
                  label="軽減税率"
                  value={formData.reduced_tax_rate}
                  onChange={(reducedTaxRate) =>
                    setFormData((prev) => ({
                      ...prev,
                      reduced_tax_rate: reducedTaxRate,
                    }))
                  }
                />
              </div>
            </div>
          </div>

          <div className={STYLE.sidePeekFooter}>
            <button type="button" onClick={onClose} className={STYLE.sidePeekCancelBtn}>
              キャンセル
            </button>
            {canEdit ? (
              <SubmitButton className={STYLE.sidePeekSaveBtn}>保存</SubmitButton>
            ) : null}
          </div>
        </fieldset>
      </form>
    </div>
  );
}

interface ClinicDeleteDialogProps {
  pendingDelete: Clinic | null;
  isPending: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

export function ClinicDeleteDialog({
  pendingDelete,
  isPending,
  onClose,
  onConfirm,
}: ClinicDeleteDialogProps) {
  return (
    <ConfirmDialog
      open={pendingDelete !== null}
      onClose={onClose}
      title="医院を削除しますか？"
      description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
      confirmLabel="削除"
      variant="destructive"
      onConfirm={onConfirm}
      isPending={isPending}
    />
  );
}

const PropertyRow = memo(function PropertyRow({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  return (
    <div className={STYLE.propertyRow}>
      <div className={`${LAYOUT.propertyRow.labelW} shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
});

function ClinicTextProperty({
  label,
  value,
  onChange,
  placeholder,
  type = "text",
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  type?: string;
}) {
  return (
    <PropertyRow label={label}>
      <input
        type={type}
        className={PROP_INPUT_CLASS}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
      />
    </PropertyRow>
  );
}

function ClinicTaxRateProperty({
  label,
  value,
  onChange,
}: {
  label: string;
  value: number;
  onChange: (value: number) => void;
}) {
  return (
    <PropertyRow label={label}>
      <div className="flex items-center gap-1.5">
        <input
          type="number"
          min={0}
          max={100}
          step={1}
          className={`${PROP_INPUT_CLASS} w-20`}
          value={Math.round(value * 100)}
          onChange={(e) => onChange(Number(e.target.value) / 100)}
        />
        <span className={`text-sm ${C.text50}`}>%</span>
      </div>
    </PropertyRow>
  );
}

const STATUS_CONFIG = {
  active: {
    dot: `${C.bgAccent}`,
    label: "有効",
    bg: `${C.bgAccentLight}`,
    text: `${C.textAccentDark}`,
  },
  inactive: {
    dot: C.bgPrimary10,
    label: "無効",
    bg: `${C.bgInactive}`,
    text: C.text60,
  },
} as const;

function NotionStatusPill({ status }: { status: "active" | "inactive" }) {
  const cfg = STATUS_CONFIG[status];
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-sm text-xs ${cfg.bg} ${cfg.text}`}
    >
      <span className={`size-[7px] rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  );
}
