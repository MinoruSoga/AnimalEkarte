import { render, screen } from "@testing-library/react";
import {
  BillingStatusCompleted,
  ItemCategoryExamination,
  PaymentMethodCash,
  TaxTypeExcluded,
} from "@/types/generated/models";
import type { Accounting, PaymentInfo } from "../types";
import { AccountingDocument, type ClinicInfo } from "./AccountingDocument";

// R-F4: AccountingDocument は useClinicTaxRates() 経由で useAuth() に依存する。
// 税率自体はどのテストも検証しないため、フォールバック既定値（10%/8%）を使う。
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: null }),
}));

const ACCOUNTING: Accounting = {
  id: "A-001",
  clinicId: "1",
  ownerId: "10",
  ownerName: "山田 太郎",
  petId: "20",
  petName: "ポチ",
  petSpecies: "犬",
  status: BillingStatusCompleted,
  scheduledDate: "2026-06-25",
  items: [
    {
      id: "item-1",
      category: ItemCategoryExamination,
      name: "診察料",
      unitPrice: 1000,
      quantity: 1,
      discountRate: 0,
      discountAmount: 0,
      taxType: TaxTypeExcluded,
      taxRate: 0.1,
      taxAmount: 100,
      subtotal: 1000,
      isInsuranceApplicable: false,
      source: "manual",
    },
  ],
  totalRefundedAmount: 0,
};

const PAYMENT: PaymentInfo = {
  subtotal: 1000,
  taxTotal: 100,
  totalAmount: 1100,
  insuranceAmount: 0,
  discountAmount: 0,
  billingAmount: 1100,
  receivedAmount: 1100,
  changeAmount: 0,
  method: PaymentMethodCash,
};

function renderDocument(clinic: ClinicInfo) {
  render(<AccountingDocument accounting={ACCOUNTING} paymentInfo={PAYMENT} clinic={clinic} />);
}

describe("AccountingDocument 移行負額", () => {
  it("明細行の負額を符号のまま印字する", () => {
    render(
      <AccountingDocument
        accounting={{
          ...ACCOUNTING,
          items: [
            {
              ...ACCOUNTING.items[0],
              name: "赤伝",
              unitPrice: -3000,
              quantity: 1,
              discountAmount: 0,
              taxAmount: -300,
              subtotal: -3000,
            },
          ],
        }}
        paymentInfo={{
          ...PAYMENT,
          subtotal: -3000,
          taxTotal: -300,
          totalAmount: -3300,
          billingAmount: -3300,
          receivedAmount: 0,
          changeAmount: 0,
        }}
        clinic={{ name: "テスト病院" }}
      />,
    );
    expect(screen.getAllByText("¥-3,000").length).toBeGreaterThanOrEqual(2);
    expect(screen.getAllByText("¥-3,300").length).toBeGreaterThanOrEqual(1);
  });
});

describe("AccountingDocument #179 ① 帳票レイアウト設定", () => {
  it("登録番号未設定警告を病院設定で非表示にできる", () => {
    renderDocument({
      name: "テスト病院",
      accountingDocumentShowRegistrationWarning: false,
    });

    expect(screen.queryByText(/登録番号が未設定です/)).not.toBeInTheDocument();
  });

  it("項目カテゴリ表示とフッター文言を病院設定から切り替える", () => {
    renderDocument({
      name: "テスト病院",
      accountingDocumentShowItemCategory: false,
      accountingDocumentFooterNote: "ご来院ありがとうございました。",
    });

    expect(screen.queryByText("examination")).not.toBeInTheDocument();
    expect(screen.getByText("ご来院ありがとうございました。")).toBeInTheDocument();
  });

  it("ロゴ表示ONかつ安全な https URL でロゴ画像を描画する", () => {
    renderDocument({
      name: "テスト病院",
      accountingDocumentShowLogo: true,
      logoUrl: "https://example.com/logo.png",
    });

    expect(screen.getByRole("img", { name: "テスト病院ロゴ" })).toHaveAttribute(
      "src",
      "https://example.com/logo.png",
    );
  });

  it("javascript: スキームのロゴ URL は描画しない（XSS 防御）", () => {
    renderDocument({
      name: "テスト病院",
      accountingDocumentShowLogo: true,
      logoUrl: "javascript:alert(1)",
    });

    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });
});

describe("AccountingDocument #190 セクション表示/非表示と順序", () => {
  it("各フラグがデフォルト（未設定）のときすべてのセクションを表示する（後方互換）", () => {
    renderDocument({ name: "テスト病院" });

    expect(screen.getByRole("heading", { name: "明細兼領収書" })).toBeInTheDocument();
    expect(screen.getByText(/山田 太郎/)).toBeInTheDocument();
    expect(screen.getByText("診察料")).toBeInTheDocument();
    expect(screen.getByText("請求金額")).toBeInTheDocument();
  });

  it("accountingDocumentShowClinicHeader: false で病院情報ヘッダーを非表示にする", () => {
    renderDocument({ name: "テスト病院", accountingDocumentShowClinicHeader: false });

    expect(screen.queryByRole("heading", { name: "明細兼領収書" })).not.toBeInTheDocument();
  });

  it("accountingDocumentShowOwnerPetInfo: false で飼主・ペット情報を非表示にする", () => {
    renderDocument({ name: "テスト病院", accountingDocumentShowOwnerPetInfo: false });

    expect(screen.queryByText(/山田 太郎/)).not.toBeInTheDocument();
  });

  it("accountingDocumentShowItemsTable: false で明細テーブルを非表示にする", () => {
    renderDocument({ name: "テスト病院", accountingDocumentShowItemsTable: false });

    expect(screen.queryByText("診察料")).not.toBeInTheDocument();
  });

  it("accountingDocumentShowPaymentSummary: false でお会計サマリーを非表示にする", () => {
    renderDocument({ name: "テスト病院", accountingDocumentShowPaymentSummary: false });

    expect(screen.queryByText("請求金額")).not.toBeInTheDocument();
  });

  it("section_order が空配列のときデフォルト順(clinic_header→payment_summary)を維持する", () => {
    renderDocument({ name: "テスト病院", accountingDocumentSectionOrder: [] });

    const heading = screen.getByRole("heading", { name: "明細兼領収書" });
    const billing = screen.getByText("請求金額");
    // heading (clinic_header) が billing (payment_summary) より DOM 上で先に来る
    expect(
      heading.compareDocumentPosition(billing) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  it("section_order で payment_summary を先頭にするとDOMでも先に現れる", () => {
    renderDocument({
      name: "テスト病院",
      accountingDocumentSectionOrder: ["payment_summary", "clinic_header"],
    });

    const heading = screen.getByRole("heading", { name: "明細兼領収書" });
    const billing = screen.getByText("請求金額");
    // billing (payment_summary) が heading (clinic_header) より DOM 上で先に来る
    expect(
      billing.compareDocumentPosition(heading) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  it("section_order に未知のキーが混入しても既知キーのみ有効化される", () => {
    renderDocument({
      name: "テスト病院",
      accountingDocumentSectionOrder: ["unknown_key", "clinic_header"],
    });

    // unknown_key は無視され、clinic_header は正常に描画される
    expect(screen.getByRole("heading", { name: "明細兼領収書" })).toBeInTheDocument();
  });
});
