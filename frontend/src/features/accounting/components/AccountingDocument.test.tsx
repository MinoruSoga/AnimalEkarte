import { render, screen } from "@testing-library/react";
import {
  BillingStatusCompleted,
  ItemCategoryExamination,
  PaymentMethodCash,
  TaxTypeExcluded,
} from "@/types/generated/models";
import type { Accounting, PaymentInfo } from "../types";
import { AccountingDocument, type ClinicInfo } from "./AccountingDocument";

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
