/**
 * Backend ↔ Frontend 型変換
 * バックエンドは snake_case、フロントエンドは camelCase
 */
import { z } from "zod";
import type { AuthUser, AuthClinic } from "../types";
import type { MeResponse } from "@/types/generated/auth-responses";

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
  // FE-R3-2: BE (auth_response.go) は非ポインタ・omitempty なしで常に送信するため必須化。
  // .default() は BE のリネームを隠蔽し帳票からロゴ/警告文言が黙って消える逆失敗モードを生む。
  accounting_document_show_logo: z.boolean(),
  accounting_document_show_registration_warning: z.boolean(),
  accounting_document_show_item_category: z.boolean(),
  accounting_document_footer_note: z.string(),
  // #190: セクション表示/非表示トグルと表示順
  accounting_document_show_clinic_header: z.boolean(),
  accounting_document_show_owner_pet_info: z.boolean(),
  accounting_document_show_items_table: z.boolean(),
  accounting_document_show_payment_summary: z.boolean(),
  accounting_document_section_order: z.array(z.string()),
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

// FE6-5: BE/FE の完全な形状一致（`satisfies z.ZodType<MeResponse>`）は要求しない —
// このスキーマは意図的に BE の実際の wire 形状より寛容（occupation/clinic への
// .nullable() 許容は防御的パースであり、BE が将来 null を送るようになっても FE は
// 落ちない）。代わりに「BE が実際に約束する形状は必ずこのスキーマの入力として
// 受理される」ことを型チェックで固定する — M1（BE の optional 化に FE が追随し損ねて
// parse が throw する退行）のようなケースを検知するのが目的。
// MeResponse が backendMeResponseSchema の入力形状と非互換なら、この行が型エラーになる。
const _meResponseIsParsable: MeResponse extends z.input<typeof backendMeResponseSchema>
  ? true
  : never = true;
void _meResponseIsParsable;

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
