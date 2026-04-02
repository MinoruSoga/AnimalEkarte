// React/Framework
import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// External
import { Building2, X } from "lucide-react";
import { toast } from "sonner";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { useGetCompany, useUpdateCompany } from "@/features/master/api/company";
import type { UpdateCompanyRequest } from "@/features/master/api/company";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const PROP_INPUT_CLASS = `w-full bg-transparent text-base ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`;

// ─────────────────────────────────────────────────
// Form state
// ─────────────────────────────────────────────────

interface CompanyFormData {
  name: string;
  postal_code: string;
  address: string;
  phone_number: string;
  fax_number: string;
  email: string;
  website: string;
  director_name: string;
  registration_number: string;
  invoice_registration_number: string;
}

const DEFAULT_FORM_DATA: CompanyFormData = {
  name: "",
  postal_code: "",
  address: "",
  phone_number: "",
  fax_number: "",
  email: "",
  website: "",
  director_name: "",
  registration_number: "",
  invoice_registration_number: "",
};

// ─────────────────────────────────────────────────
// CompanySettings
// ─────────────────────────────────────────────────

export function CompanySettings() {
  const navigate = useNavigate();
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState<CompanyFormData>(DEFAULT_FORM_DATA);

  const { data: company, isLoading } = useGetCompany();
  const updateMutation = useUpdateCompany();

  useEffect(() => {
    if (company) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- 非同期サーバーデータでフォームを初期化するパターン。React 18 が自動バッチするため実害なし
      setFormData({
        name: company.name,
        postal_code: company.postalCode,
        address: company.address,
        phone_number: company.phoneNumber,
        fax_number: company.faxNumber,
        email: company.email,
        website: company.website,
        director_name: company.directorName,
        registration_number: company.registrationNumber,
        invoice_registration_number: company.invoiceRegistrationNumber,
      });
    }
  }, [company]);

  const handleEdit = useCallback(() => {
    setIsEditing(true);
  }, []);

  const handleCloseEdit = useCallback(() => {
    setIsEditing(false);
    if (company) {
      setFormData({
        name: company.name,
        postal_code: company.postalCode,
        address: company.address,
        phone_number: company.phoneNumber,
        fax_number: company.faxNumber,
        email: company.email,
        website: company.website,
        director_name: company.directorName,
        registration_number: company.registrationNumber,
        invoice_registration_number: company.invoiceRegistrationNumber,
      });
    }
  }, [company]);

  const handleSave = useCallback(() => {
    if (!formData.name.trim()) {
      toast.error("法人名は必須です");
      return;
    }

    const req: UpdateCompanyRequest = {
      name: formData.name,
      postal_code: formData.postal_code || undefined,
      address: formData.address || undefined,
      phone_number: formData.phone_number || undefined,
      fax_number: formData.fax_number || undefined,
      email: formData.email || undefined,
      website: formData.website || undefined,
      director_name: formData.director_name || undefined,
      registration_number: formData.registration_number || undefined,
      invoice_registration_number: formData.invoice_registration_number || undefined,
    };

    updateMutation.mutate(req, {
      onSuccess: () => {
        toast.success("法人情報を更新しました");
        setIsEditing(false);
      },
      onError: () => {
        toast.error("更新に失敗しました");
      },
    });
  }, [formData, updateMutation]);

  return (
    <div className="flex h-full">
      {/* ── Left: Detail View ── */}
      <div className="flex-1 min-w-0">
        <PageLayout
          title="法人情報"
          icon={<Building2 className={`${ICON.page} ${C.text}`} />}
          onBack={() => navigate(paths.settings.getHref())}
          headerAction={
            isLoading ? null : (
              <button
                type="button"
                onClick={handleEdit}
                className={`px-4 py-[7px] text-base ${C.text65} ${C.hoverBgLight} border ${C.borderMedium} rounded-[3px] transition-colors cursor-pointer`}
              >
                編集
              </button>
            )
          }
          maxWidth="max-w-3xl"
        >
          {isLoading ? (
            <div className={`px-6 py-12 text-center text-base ${C.text50}`}>読み込み中...</div>
          ) : company ? (
            <div className="px-6 pb-8">
              {/* Page icon */}
              <div className="pt-6 pb-2">
                <div className={STYLE.pageIcon}>
                  <Building2 className={LAYOUT.pageIcon.innerIcon} />
                </div>
              </div>

              {/* Page title */}
              <div className="pb-1 mb-4">
                <h2
                  className={C.text}
                  style={{
                    fontSize: LAYOUT.pageTitle.fontSize,
                    fontWeight: LAYOUT.pageTitle.fontWeight,
                    lineHeight: LAYOUT.pageTitle.lineHeight,
                  }}
                >
                  {company.name || "（法人名未設定）"}
                </h2>
              </div>

              {/* Separator */}
              <div className={`${STYLE.sectionDivider} mb-1`} />

              {/* Properties (read-only) */}
              <div className="py-1">
                <PropertyRow label="郵便番号">
                  <span className={`text-base ${company.postalCode ? C.text : C.text30}`}>
                    {company.postalCode || "空"}
                  </span>
                </PropertyRow>
                <PropertyRow label="住所">
                  <span className={`text-base ${company.address ? C.text : C.text30}`}>
                    {company.address || "空"}
                  </span>
                </PropertyRow>
                <PropertyRow label="電話番号">
                  <span className={`text-base font-mono ${company.phoneNumber ? C.text : C.text30}`}>
                    {company.phoneNumber || "空"}
                  </span>
                </PropertyRow>
                <PropertyRow label="FAX番号">
                  <span className={`text-base font-mono ${company.faxNumber ? C.text : C.text30}`}>
                    {company.faxNumber || "空"}
                  </span>
                </PropertyRow>
                <PropertyRow label="メールアドレス">
                  <span className={`text-base ${company.email ? C.text : C.text30}`}>
                    {company.email || "空"}
                  </span>
                </PropertyRow>
                <PropertyRow label="Webサイト">
                  <span className={`text-base ${company.website ? C.text : C.text30}`}>
                    {company.website || "空"}
                  </span>
                </PropertyRow>
                <PropertyRow label="代表者名">
                  <span className={`text-base ${company.directorName ? C.text : C.text30}`}>
                    {company.directorName || "空"}
                  </span>
                </PropertyRow>
                <PropertyRow label="法人番号">
                  <span className={`text-base ${company.registrationNumber ? C.text : C.text30}`}>
                    {company.registrationNumber || "空"}
                  </span>
                </PropertyRow>
                <PropertyRow label="インボイス番号">
                  <span className={`text-base font-mono ${company.invoiceRegistrationNumber ? C.text : C.text30}`}>
                    {company.invoiceRegistrationNumber || "空"}
                  </span>
                </PropertyRow>
              </div>
            </div>
          ) : (
            <div className={`px-6 py-12 text-center text-base ${C.text50}`}>
              法人情報を取得できませんでした
            </div>
          )}
        </PageLayout>
      </div>

      {/* ── Right: Side Peek (Edit) ── */}
      {isEditing ? (
        <div
          className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 flex flex-col`}
        >
          {/* Toolbar */}
          <div className={STYLE.sidePeekToolbar}>
            <span className={`text-base ${C.text35} pl-1 select-none`}>編集</span>
            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={handleCloseEdit}
                className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
              >
                <X className={ICON.action} />
              </button>
            </div>
          </div>

          {/* Body */}
          <div className={STYLE.sidePeekBody}>
            <div className="px-16 pb-8">
              {/* Page icon */}
              <div className="pt-4 pb-2">
                <div className={STYLE.pageIcon}>
                  <Building2 className={LAYOUT.pageIcon.innerIcon} />
                </div>
              </div>

              {/* Title input (name) */}
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
                  placeholder="法人名"
                />
              </div>

              {/* Separator */}
              <div className={`${STYLE.sectionDivider} mb-1`} />

              {/* Properties */}
              <div className="py-1">
                <PropertyRow label="郵便番号">
                  <input
                    type="text"
                    className={PROP_INPUT_CLASS}
                    value={formData.postal_code}
                    onChange={(e) => setFormData((prev) => ({ ...prev, postal_code: e.target.value }))}
                    placeholder="例: 150-0001"
                  />
                </PropertyRow>

                <PropertyRow label="住所">
                  <input
                    type="text"
                    className={PROP_INPUT_CLASS}
                    value={formData.address}
                    onChange={(e) => setFormData((prev) => ({ ...prev, address: e.target.value }))}
                    placeholder="例: 東京都渋谷区..."
                  />
                </PropertyRow>

                <PropertyRow label="電話番号">
                  <input
                    type="text"
                    className={PROP_INPUT_CLASS}
                    value={formData.phone_number}
                    onChange={(e) => setFormData((prev) => ({ ...prev, phone_number: e.target.value }))}
                    placeholder="例: 03-1234-5678"
                  />
                </PropertyRow>

                <PropertyRow label="FAX番号">
                  <input
                    type="text"
                    className={PROP_INPUT_CLASS}
                    value={formData.fax_number}
                    onChange={(e) => setFormData((prev) => ({ ...prev, fax_number: e.target.value }))}
                    placeholder="例: 03-1234-5679"
                  />
                </PropertyRow>

                <PropertyRow label="メールアドレス">
                  <input
                    type="email"
                    className={PROP_INPUT_CLASS}
                    value={formData.email}
                    onChange={(e) => setFormData((prev) => ({ ...prev, email: e.target.value }))}
                    placeholder="例: info@company.com"
                  />
                </PropertyRow>

                <PropertyRow label="Webサイト">
                  <input
                    type="text"
                    className={PROP_INPUT_CLASS}
                    value={formData.website}
                    onChange={(e) => setFormData((prev) => ({ ...prev, website: e.target.value }))}
                    placeholder="例: https://example.com"
                  />
                </PropertyRow>

                <PropertyRow label="代表者名">
                  <input
                    type="text"
                    className={PROP_INPUT_CLASS}
                    value={formData.director_name}
                    onChange={(e) => setFormData((prev) => ({ ...prev, director_name: e.target.value }))}
                    placeholder="例: 山田 太郎"
                  />
                </PropertyRow>

                <PropertyRow label="法人番号">
                  <input
                    type="text"
                    className={PROP_INPUT_CLASS}
                    value={formData.registration_number}
                    onChange={(e) =>
                      setFormData((prev) => ({ ...prev, registration_number: e.target.value }))
                    }
                    placeholder="例: 1234567890123"
                  />
                </PropertyRow>

                <PropertyRow label="インボイス番号">
                  <input
                    type="text"
                    className={PROP_INPUT_CLASS}
                    value={formData.invoice_registration_number}
                    onChange={(e) =>
                      setFormData((prev) => ({ ...prev, invoice_registration_number: e.target.value }))
                    }
                    placeholder="例: T1234567890123"
                  />
                </PropertyRow>
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className={STYLE.sidePeekFooter}>
            <button
              type="button"
              onClick={handleCloseEdit}
              className={STYLE.sidePeekCancelBtn}
            >
              キャンセル
            </button>
            <button
              type="button"
              onClick={handleSave}
              className={STYLE.sidePeekSaveBtn}
            >
              保存
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
