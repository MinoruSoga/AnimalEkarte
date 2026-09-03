/**
 * FE5-1: BE `MeResponse.Clinics` は `json:"clinics,omitempty"` で省略可能だが、
 * FE スキーマは `clinics: z.array(...)` 必須だったため、所属クリニック 0 件のスタッフで
 * `/me` の parse が throw し認証フローが落ちる回帰バグ（FE-refactor.md §4.1 M1）。
 */
import { describe, it, expect } from "vitest";
import { mapMeToAuthUser } from "./transforms";

function buildRawMe(overrides: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: "user-1",
    email: "staff@example.com",
    display_name: "田中 太郎",
    is_system_admin: false,
    main_clinic_id: "clinic-a",
    clinic: null,
    permissions: {},
    ...overrides,
  };
}

describe("FE5-1: mapMeToAuthUser — clinics omitempty 耐性", () => {
  it("clinics フィールドが未定義（BE omitempty）でも throw せず空配列になる", () => {
    const raw = buildRawMe();
    expect("clinics" in raw).toBe(false);

    const user = mapMeToAuthUser(raw);

    expect(user.clinics).toEqual([]);
  });

  it("clinics が通常通り渡された場合は従来通りマップされる", () => {
    const raw = buildRawMe({
      clinics: [{ clinic_id: "clinic-a", clinic_name: "八王子院", is_main: true }],
    });

    const user = mapMeToAuthUser(raw);

    expect(user.clinics).toEqual([{ clinicId: "clinic-a", clinicName: "八王子院", isMain: true }]);
  });
});

/**
 * FE-R3-2: meClinicInfoSchema の帳票設定 9 フィールドは BE (`auth_response.go`) が
 * 非ポインタ・omitempty なしで常に送信する。`.default()` による silent フォールバックは
 * BE のリネームを隠蔽し、帳票からロゴ/警告文言が黙って消える逆失敗モードを生む
 * （例: accounting_document_show_registration_warning が silent に true 既定へ）。
 */
function buildRawClinic(overrides: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: "clinic-a",
    name: "八王子院",
    postal_code: "192-0000",
    address: "東京都八王子市",
    phone_number: "042-000-0000",
    fax_number: "042-000-0001",
    registration_number: "T1234567890123",
    director_name: "院長 太郎",
    email: "clinic@example.com",
    website: "https://example.com",
    logo_url: null,
    standard_tax_rate: 0.1,
    reduced_tax_rate: 0.08,
    accounting_document_show_logo: true,
    accounting_document_show_registration_warning: true,
    accounting_document_show_item_category: true,
    accounting_document_footer_note: "footer",
    accounting_document_show_clinic_header: true,
    accounting_document_show_owner_pet_info: true,
    accounting_document_show_items_table: true,
    accounting_document_show_payment_summary: true,
    accounting_document_section_order: ["items", "summary"],
    ...overrides,
  };
}

describe("FE-R3-2: meClinicInfoSchema — 帳票設定9フィールドの必須化", () => {
  it.each([
    "accounting_document_show_logo",
    "accounting_document_show_registration_warning",
    "accounting_document_show_item_category",
    "accounting_document_footer_note",
    "accounting_document_show_clinic_header",
    "accounting_document_show_owner_pet_info",
    "accounting_document_show_items_table",
    "accounting_document_show_payment_summary",
    "accounting_document_section_order",
  ])("%s が欠落した clinic payload は parse が throw する（silent default 禁止）", (field) => {
    const clinic = buildRawClinic();
    delete clinic[field];
    const raw = buildRawMe({ clinic });

    expect(() => mapMeToAuthUser(raw)).toThrow();
  });

  it("帳票9フィールドが全て揃った payload は従来通りマップされる", () => {
    const raw = buildRawMe({ clinic: buildRawClinic() });

    const user = mapMeToAuthUser(raw);

    expect(user.clinic).toEqual(
      expect.objectContaining({
        accountingDocumentShowLogo: true,
        accountingDocumentShowRegistrationWarning: true,
        accountingDocumentShowItemCategory: true,
        accountingDocumentFooterNote: "footer",
        accountingDocumentShowClinicHeader: true,
        accountingDocumentShowOwnerPetInfo: true,
        accountingDocumentShowItemsTable: true,
        accountingDocumentShowPaymentSummary: true,
        accountingDocumentSectionOrder: ["items", "summary"],
      }),
    );
  });

  it("clinics: z.array(...).default([]) は不変 — clinics 未定義でも throw しない（FE5-1回帰防止）", () => {
    const raw = buildRawMe({ clinic: buildRawClinic() });
    expect("clinics" in raw).toBe(false);

    expect(() => mapMeToAuthUser(raw)).not.toThrow();
  });
});
