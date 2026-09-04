import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { PaymentCard, type PaymentSplitDraft } from "./PaymentCard";

// #127: 銀行振込 (bank_transfer) が会計 UI で選択でき、選択時に
// method="bank_transfer" がフォーム state へ反映されることを検証する。
// 支払方法セレクタは Radix Select ではなく Button グリッドのため決定的にテストできる。
function renderCard(overrides?: {
  splits?: PaymentSplitDraft[];
  onSplitsChange?: (s: PaymentSplitDraft[]) => void;
}) {
  const splits: PaymentSplitDraft[] = overrides?.splits ?? [
    { method: "cash", amount: "5000", receivedAmount: "5000" },
  ];
  const onSplitsChange = overrides?.onSplitsChange ?? vi.fn();
  render(
    <PaymentCard
      billingAmount={5000}
      paymentSplits={splits}
      onSplitsChange={onSplitsChange}
      isCompleted={false}
      canEdit={true}
      canCreate={true}
      isEditMode={true}
    />,
  );
  return { onSplitsChange };
}

describe("PaymentCard — 銀行振込 (#127)", () => {
  it("銀行振込が支払方法の選択肢として表示される", () => {
    renderCard();
    expect(screen.getByRole("button", { name: "銀行振込" })).toBeInTheDocument();
    // 既存の支払方法も維持されていること（回帰）。
    expect(screen.getByRole("button", { name: "現金" })).toBeInTheDocument();
    // FE5-25: PAYMENT_METHOD_LABELS_SHORT 廃止により "カード" → "クレジットカード" へ変更
    expect(screen.getByRole("button", { name: "クレジットカード" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "電子マネー" })).toBeInTheDocument();
  });

  it("銀行振込を選択すると onSplitsChange に method=bank_transfer が渡る", async () => {
    const user = userEvent.setup();
    const onSplitsChange = vi.fn();
    renderCard({ onSplitsChange });

    await user.click(screen.getByRole("button", { name: "銀行振込" }));

    expect(onSplitsChange).toHaveBeenCalledTimes(1);
    expect(onSplitsChange).toHaveBeenCalledWith([
      { method: "bank_transfer", amount: "5000", receivedAmount: "5000" },
    ]);
  });
});
