/**
 * Backend ↔ Frontend 型変換
 * バックエンドは snake_case、フロントエンドは camelCase
 */
import { z } from "zod";
import type { AuthUser, AuthClinic } from "../types";

const clinicMembershipSchema = z.object({
  clinic_id: z.string(),
  clinic_name: z.string(),
  is_main: z.boolean(),
});

const meClinicInfoSchema = z.object({
  id: z.string(),
  name: z.string(),
  postal_code: z.string().default(""),
  address: z.string().default(""),
  phone_number: z.string().default(""),
  fax_number: z.string().default(""),
  registration_number: z.string().default(""),
  director_name: z.string().default(""),
  email: z.string().default(""),
  website: z.string().default(""),
  logo_url: z.string().nullable().default(null),
  // BUG-367: インボイス帳票用
  standard_tax_rate: z.number().default(0.1),
  reduced_tax_rate: z.number().default(0.08),
  accounting_document_show_logo: z.boolean().default(false),
  accounting_document_show_registration_warning: z.boolean().default(true),
  accounting_document_show_item_category: z.boolean().default(true),
  accounting_document_footer_note: z.string().default(""),
  // #190: セクション表示/非表示トグルと表示順
  accounting_document_show_clinic_header: z.boolean().default(true),
  accounting_document_show_owner_pet_info: z.boolean().default(true),
  accounting_document_show_items_table: z.boolean().default(true),
  accounting_document_show_payment_summary: z.boolean().default(true),
  accounting_document_section_order: z.array(z.string()).default([]),
});

const resourcePermissionSchema = z.object({
  view: z.boolean(),
  create: z.boolean(),
  edit: z.boolean(),
  delete: z.boolean(),
});

const backendMeResponseSchema = z.object({
  id: z.string(),
  email: z.string(),
  display_name: z.string(),
  is_system_admin: z.boolean().default(false),
  // occupation は職種マスタ名
  occupation: z.string().nullable().optional(),
  avatar_url: z.string().nullable().optional(),
  main_clinic_id: z.string(),
  // clinic は /me レスポンスのメイン医院詳細。未所属の場合は null
  clinic: meClinicInfoSchema.nullable().optional(),
  // BE MeResponse.Clinics は `json:"clinics,omitempty"`（未所属時は省略され得る）。
  // FE5-1: 必須配列のままだと所属クリニック 0 件のスタッフで /me parse が throw していた。
  clinics: z.array(clinicMembershipSchema).default([]),
  // permissions: resource → CRUD（BEがUNION計算済みのフラット構造）
  permissions: z.record(z.string(), resourcePermissionSchema),
});

export type BackendMeResponse = z.infer<typeof backendMeResponseSchema>;

function mapMeClinicInfo(raw: z.infer<typeof meClinicInfoSchema>): AuthClinic {
  return {
    id: raw.id,
    name: raw.name,
    postalCode: raw.postal_code,
    address: raw.address,
    phoneNumber: raw.phone_number,
    faxNumber: raw.fax_number,
    registrationNumber: raw.registration_number,
    directorName: raw.director_name,
    email: raw.email,
    website: raw.website,
    logoUrl: raw.logo_url,
    standardTaxRate: raw.standard_tax_rate,
    reducedTaxRate: raw.reduced_tax_rate,
    accountingDocumentShowLogo: raw.accounting_document_show_logo,
    accountingDocumentShowRegistrationWarning: raw.accounting_document_show_registration_warning,
    accountingDocumentShowItemCategory: raw.accounting_document_show_item_category,
    accountingDocumentFooterNote: raw.accounting_document_footer_note,
    accountingDocumentShowClinicHeader: raw.accounting_document_show_clinic_header,
    accountingDocumentShowOwnerPetInfo: raw.accounting_document_show_owner_pet_info,
    accountingDocumentShowItemsTable: raw.accounting_document_show_items_table,
    accountingDocumentShowPaymentSummary: raw.accounting_document_show_payment_summary,
    accountingDocumentSectionOrder: raw.accounting_document_section_order,
  };
}

export function mapMeToAuthUser(raw: unknown): AuthUser {
  const me = backendMeResponseSchema.parse(raw);
  return {
    id: me.id,
    email: me.email,
    displayName: me.display_name,
    isSystemAdmin: me.is_system_admin,
    avatarUrl: me.avatar_url ?? null,
    mainClinicId: me.main_clinic_id,
    clinic: me.clinic ? mapMeClinicInfo(me.clinic) : null,
    clinics: me.clinics.map((c) => ({
      clinicId: c.clinic_id,
      clinicName: c.clinic_name,
      isMain: c.is_main,
    })),
    permissions: me.permissions,
  };
}
