import { describe, expect, it } from "vitest";

import type { PaymentMethod } from "../api/payment-method-master";
import { validatePaymentMethodForm } from "./payment-method-settings-model";

const existing: PaymentMethod[] = [
  {
    id: "1",
    name: "現金",
    isActive: true,
    displayOrder: 1,
  },
];

describe("validatePaymentMethodForm (BUG-029)", () => {
  it("名称未入力を拒否する", () => {
    expect(
      validatePaymentMethodForm(
        { name: "  ", isActive: true },
        { existing, editingId: null },
      ),
    ).toBe("名称は必須です");
  });

  it("新規で既存名と重複したら拒否する", () => {
    expect(
      validatePaymentMethodForm(
        { name: "現金", isActive: true },
        { existing, editingId: null },
      ),
    ).toBe("支払方法名「現金」は既に使用されています");
  });

  it("編集中の自身の名称は重複とみなさない", () => {
    expect(
      validatePaymentMethodForm(
        { name: "現金", isActive: true },
        { existing, editingId: "1" },
      ),
    ).toBeNull();
  });

  it("別名は許可する", () => {
    expect(
      validatePaymentMethodForm(
        { name: "電子マネー", isActive: true },
        { existing, editingId: null },
      ),
    ).toBeNull();
  });
});
