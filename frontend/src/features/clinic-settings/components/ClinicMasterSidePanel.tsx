import type { Dispatch, SetStateAction } from "react";
import { Building2, FileText, Percent, X } from "lucide-react";

import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, LAYOUT, STYLE } from "@/lib/design-tokens";
import type { Clinic } from "../api/clinics";
import type { ClinicFormData } from "../lib/clinic-master-settings-model";
import {
  ClinicBooleanProperty,
  ClinicTaxRateProperty,
  ClinicTextProperty,
  ClinicTextareaProperty,
  PropertyRow,
  SectionOrderProperty,
  StatusPill,
} from "./ClinicMasterSidePanelProperties";

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
                  className={`w-full bg-transparent ${C.text} ${C.textPlaceholderFaint} outline-none border-none p-0 focus-visible:ring-2 ${C.focusRingAccent40}`}
                  style={{
                    fontSize: LAYOUT.pageTitle.fontSize,
                    fontWeight: LAYOUT.pageTitle.fontWeight,
                    lineHeight: LAYOUT.pageTitle.lineHeight,
                    letterSpacing: LAYOUT.pageTitle.letterSpacing,
                  }}
                  value={formData.name}
                  onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
                  placeholder="無題"
                  aria-label="無題"
                />
                <FormFieldError message={nameError} />
              </div>

              <div className={`${STYLE.sectionDivider} mb-1`} />

              <div className="py-1">
                <PropertyRow label="ステータス">
                  <button
                    type="button"
                    onClick={() => setFormData((prev) => ({ ...prev, is_active: !prev.is_active }))}
                    className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
                  >
                    <StatusPill status={formData.is_active ? "active" : "inactive"} />
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

                <div className={`${STYLE.sectionDivider} my-2`} />
                <div className={`flex items-center gap-1.5 py-1.5 text-xs ${C.text45} select-none`}>
                  <FileText className={ICON.xs} />
                  明細兼領収書
                </div>

                <ClinicBooleanProperty
                  label="ロゴ表示"
                  value={formData.accounting_document_show_logo}
                  onChange={(value) =>
                    setFormData((prev) => ({
                      ...prev,
                      accounting_document_show_logo: value,
                    }))
                  }
                />
                <ClinicBooleanProperty
                  label="登録番号警告"
                  value={formData.accounting_document_show_registration_warning}
                  onChange={(value) =>
                    setFormData((prev) => ({
                      ...prev,
                      accounting_document_show_registration_warning: value,
                    }))
                  }
                />
                <ClinicBooleanProperty
                  label="項目カテゴリ"
                  value={formData.accounting_document_show_item_category}
                  onChange={(value) =>
                    setFormData((prev) => ({
                      ...prev,
                      accounting_document_show_item_category: value,
                    }))
                  }
                />
                {/* #190: セクション表示トグル */}
                <ClinicBooleanProperty
                  label="病院情報ヘッダー"
                  value={formData.accounting_document_show_clinic_header}
                  onChange={(value) =>
                    setFormData((prev) => ({
                      ...prev,
                      accounting_document_show_clinic_header: value,
                    }))
                  }
                />
                <ClinicBooleanProperty
                  label="飼主・ペット情報"
                  value={formData.accounting_document_show_owner_pet_info}
                  onChange={(value) =>
                    setFormData((prev) => ({
                      ...prev,
                      accounting_document_show_owner_pet_info: value,
                    }))
                  }
                />
                <ClinicBooleanProperty
                  label="明細テーブル"
                  value={formData.accounting_document_show_items_table}
                  onChange={(value) =>
                    setFormData((prev) => ({
                      ...prev,
                      accounting_document_show_items_table: value,
                    }))
                  }
                />
                <ClinicBooleanProperty
                  label="お会計サマリー"
                  value={formData.accounting_document_show_payment_summary}
                  onChange={(value) =>
                    setFormData((prev) => ({
                      ...prev,
                      accounting_document_show_payment_summary: value,
                    }))
                  }
                />
                {/* #190: セクション表示順 */}
                <SectionOrderProperty
                  order={formData.accounting_document_section_order}
                  onChange={(order) =>
                    setFormData((prev) => ({
                      ...prev,
                      accounting_document_section_order: order,
                    }))
                  }
                />
                <ClinicTextareaProperty
                  label="フッター"
                  value={formData.accounting_document_footer_note}
                  placeholder="例: ご来院ありがとうございました。"
                  onChange={(value) =>
                    setFormData((prev) => ({
                      ...prev,
                      accounting_document_footer_note: value,
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
            {canEdit ? <SubmitButton className="h-9 px-4">保存</SubmitButton> : null}
          </div>
        </fieldset>
      </form>
    </div>
  );
}
