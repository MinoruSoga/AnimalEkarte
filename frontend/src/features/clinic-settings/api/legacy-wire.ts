/**
 * 医院マスタ API の移行 HTTP DTO（非永続）。
 *
 * 保存 model（生成 Clinic）ではなく、現行 handler が返す JSON の列をそのまま写す。
 * 税率は Go の応答 struct で float64 に変換済みのため number で受ける。
 * 新しい医院設定の契約へ切り替えるときに、この adapter ごと置き換える（EMR-527）。
 */

/** @see backend/internal/clinic/clinic_response.go ClinicResponse */
export interface LegacyClinicResponse {
  id: number;
  company_id: number;
  name: string;
  postal_code: string;
  address: string;
  phone_number: string;
  fax_number: string;
  registration_number: string;
  director_name: string;
  email: string;
  website: string;
  logo_url: string;
  is_active: boolean;
  standard_tax_rate: number;
  reduced_tax_rate: number;
  accounting_document_show_logo: boolean;
  accounting_document_show_registration_warning: boolean;
  accounting_document_show_item_category: boolean;
  accounting_document_footer_note: string;
  // #190: セクション表示/非表示トグルと表示順
  accounting_document_show_clinic_header: boolean;
  accounting_document_show_owner_pet_info: boolean;
  accounting_document_show_items_table: boolean;
  accounting_document_show_payment_summary: boolean;
  accounting_document_section_order: string[];
  created_at: string;
  updated_at: string;
}
