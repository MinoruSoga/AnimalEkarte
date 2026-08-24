import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { PaymentCard, type PaymentSplitDraft } from "./PaymentCard";

function renderCard(splits: PaymentSplitDraft[], billingAmount: number, onSplitsChange = vi.fn()) {
  render(
    <PaymentCard
      billingAmount={billingAmount}
      paymentSplits={splits}
      onSplitsChange={onSplitsChange}
      isCompleted={false}
      canEdit
      canCreate
      isEditMode={false}
    />,
  );
  return { onSplitsChange };
}

describe("PaymentCard 移行負額", () => {
  it("確定済みの負の請求金額を符号のまま表示する", () => {
    render(
      <PaymentCard
        billingAmount={-3000}
        paymentSplits={[{ method: "cash", amount: "-3000", receivedAmount: "0" }]}
        onSplitsChange={vi.fn()}
        isCompleted
        canEdit={false}
        canCreate={false}
        isEditMode={false}
      />,
    );
    expect(screen.getByText("今回の請求金額")).toBeInTheDocument();
    expect(screen.getAllByText("¥-3,000").length).toBeGreaterThanOrEqual(2);
  });

  it("負の請求で split が足りないときは差額を符号のまま出す", () => {
    render(
      <PaymentCard
        billingAmount={-3000}
        paymentSplits={[{ method: "cash", amount: "0", receivedAmount: "0" }]}
        onSplitsChange={vi.fn()}
        isCompleted={false}
        canEdit
        canCreate
        isEditMode={false}
      />,
    );
    expect(screen.getByText("差額 ¥-3,000")).toBeInTheDocument();
    expect(screen.queryByText(/超過/)).not.toBeInTheDocument();
  });
});

describe("PaymentCard お釣り整合 (#182 ②: 現行整合厳格を維持)", () => {
  it("お釣り = お預かり金額 − 請求金額 を表示する（整合: change == received − amount）", () => {
    renderCard([{ method: "cash", amount: "1000", receivedAmount: "1500" }], 1000);
    // お釣り 500 円が表示される
    expect(screen.getByText("¥500")).toBeInTheDocument();
  });

  it("お預かり < 請求 の場合はお釣りが負値（不足）として表示される", () => {
    renderCard([{ method: "cash", amount: "1000", receivedAmount: "800" }], 1000);
    // 表示は ¥ の後に符号（toLocaleString は負号を数値前に付ける）
    expect(screen.getByText("¥-200")).toBeInTheDocument();
  });

  it("「丁度」ボタンはお預かり金額を請求額に揃える（お釣り0・整合維持）", async () => {
    const user = userEvent.setup();
    const { onSplitsChange } = renderCard(
      [{ method: "cash", amount: "1000", receivedAmount: "" }],
      1000,
    );
    await user.click(screen.getByRole("button", { name: "丁度" }));
    // received を amount(1000) に設定 → change は received-amount=0 に再計算される
    expect(onSplitsChange).toHaveBeenCalledWith([
      { method: "cash", amount: "1000", receivedAmount: "1000" },
    ]);
  });
});

describe("PaymentCard お釣り直接上書き (#188)", () => {
  it("「手動修正」でお釣り入力欄が現れ、onSplitsChange に changeOverride=true と派生初期値が渡る", async () => {
    const user = userEvent.setup();
    const { onSplitsChange } = renderCard(
      [{ method: "cash", amount: "5000", receivedAmount: "6000" }],
      5000,
    );
    await user.click(screen.getByRole("button", { name: "手動修正" }));
    // 上書き ON 時は現在の派生値 max(0, 6000-5000)=1000 を初期値に置く
    expect(onSplitsChange).toHaveBeenCalledWith([
      { method: "cash", amount: "5000", receivedAmount: "6000", changeOverride: true, changeAmount: "1000" },
    ]);
  });

  it("上書きモードでは派生お釣り表示の代わりに編集可能な入力欄を表示する", () => {
    renderCard(
      [{ method: "cash", amount: "5000", receivedAmount: "6000", changeOverride: true, changeAmount: "500" }],
      5000,
    );
    expect(screen.getByRole("button", { name: "自動計算に戻す" })).toBeInTheDocument();
    expect(screen.getByDisplayValue("500")).toBeInTheDocument();
  });

  it("「自動計算に戻す」で上書きフィールドが除去され、基本ドラフト3項目に戻る", async () => {
    const user = userEvent.setup();
    const { onSplitsChange } = renderCard(
      [{ method: "cash", amount: "5000", receivedAmount: "6000", changeOverride: true, changeAmount: "500" }],
      5000,
    );
    await user.click(screen.getByRole("button", { name: "自動計算に戻す" }));
    expect(onSplitsChange).toHaveBeenCalledWith([
      { method: "cash", amount: "5000", receivedAmount: "6000" },
    ]);
  });

  it("上書き中でも預り金 < 請求 なら確定不可（下限ガード received >= amount を維持）", () => {
    renderCard(
      [{ method: "cash", amount: "5000", receivedAmount: "4000", changeOverride: true, changeAmount: "0" }],
      5000,
    );
    expect(screen.getByRole("button", { name: /会計を確定する/ })).toBeDisabled();
  });
});

describe("PaymentCard クレジット確定前バリデーション (#182 ③: 確定前編集のみ)", () => {
  it("カード金額が0（請求と不一致）のとき確定ボタンが無効", () => {
    renderCard([{ method: "credit_card", amount: "0", receivedAmount: "" }], 1000);
    expect(screen.getByRole("button", { name: /会計を確定する/ })).toBeDisabled();
  });

  it("カード金額が請求額と一致すれば確定ボタンが有効", () => {
    renderCard([{ method: "credit_card", amount: "1000", receivedAmount: "" }], 1000);
    expect(screen.getByRole("button", { name: /会計を確定する/ })).toBeEnabled();
  });

  it("カード金額が請求超過のとき超過メッセージを表示し確定不可", () => {
    renderCard([{ method: "credit_card", amount: "1500", receivedAmount: "" }], 1000);
    expect(screen.getByText("¥500 超過")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /会計を確定する/ })).toBeDisabled();
  });
});

describe("PaymentCard permission and labels", () => {
  it("新規作成ではcreate権限だけでも編集UIと固有labelを表示する", () => {
    render(
      <PaymentCard
        billingAmount={1000}
        paymentSplits={[{ method: "cash", amount: "1000", receivedAmount: "1000" }]}
        onSplitsChange={vi.fn()}
        isCompleted={false}
        canEdit={false}
        canCreate
        isEditMode={false}
      />,
    );

    expect(screen.getByRole("spinbutton", { name: "支払1の金額" })).toBeInTheDocument();
    expect(screen.getByRole("spinbutton", { name: "支払1のお預かり金額" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "会計を確定する" })).toBeEnabled();
  });

  it("既存会計でedit権限がなければ入力・submitを表示しない", () => {
    render(
      <PaymentCard
        billingAmount={1000}
        paymentSplits={[{ method: "cash", amount: "1000", receivedAmount: "1000" }]}
        onSplitsChange={vi.fn()}
        isCompleted={false}
        canEdit={false}
        canCreate
        isEditMode
      />,
    );

    expect(screen.queryByRole("spinbutton")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "会計を確定する" })).not.toBeInTheDocument();
    expect(screen.getAllByText("¥1,000")).toHaveLength(2);
  });

  it("手動修正buttonは44px以上の操作領域を持つ", () => {
    renderCard([{ method: "cash", amount: "1000", receivedAmount: "1000" }], 1000);
    expect(screen.getByRole("button", { name: "手動修正" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });
});
