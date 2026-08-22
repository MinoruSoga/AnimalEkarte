import { describe, it, expect } from "vitest";

import {
  buildPaymentSplitRequests,
  createInitialPaymentSplits,
  getCorrectableCardPayments,
  hasInvalidChangeOverride,
  isPaymentSubmitDisabled,
  isPostCloseSubmitBlocked,
} from "./accounting-detail-model";
import type { PaymentSplitDraft } from "./PaymentCard";
import type { Accounting } from "../types";

describe("isPostCloseSubmitBlocked (BUG-004)", () => {
  it("未締めなら理由なしでもブロックしない", () => {
    expect(isPostCloseSubmitBlocked({
      isScheduledDateClosed: false,
      canPostCloseEdit: true,
      postCloseReason: "",
    })).toBe(false);
  });

  it("締め済みで権限なしは確定をブロックする", () => {
    expect(isPostCloseSubmitBlocked({
      isScheduledDateClosed: true,
      canPostCloseEdit: false,
      postCloseReason: "理由あり",
    })).toBe(true);
  });

  it("締め済みで権限ありでも理由が空ならブロックする", () => {
    expect(isPostCloseSubmitBlocked({
      isScheduledDateClosed: true,
      canPostCloseEdit: true,
      postCloseReason: "  ",
    })).toBe(true);
  });

  it("締め済み・権限あり・理由ありは通す", () => {
    expect(isPostCloseSubmitBlocked({
      isScheduledDateClosed: true,
      canPostCloseEdit: true,
      postCloseReason: "当日締め後の追加精算",
    })).toBe(false);
  });
});

describe("buildPaymentSplitRequests (#188 お釣り上書き payload 契約)", () => {
  it("現金・上書きなし: change を received-amount から派生し change_override=false", () => {
    const splits: PaymentSplitDraft[] = [{ method: "cash", amount: "5000", receivedAmount: "6000" }];
    expect(buildPaymentSplitRequests(splits)).toEqual([
      { method: "cash", amount: 5000, received_amount: 6000, change_amount: 1000, change_override: false },
    ]);
  });

  it("現金・上書きあり: 入力 changeAmount を verbatim 送出し change_override=true（received-amount と不一致でも）", () => {
    const splits: PaymentSplitDraft[] = [
      { method: "cash", amount: "5000", receivedAmount: "6000", changeOverride: true, changeAmount: "500" },
    ];
    expect(buildPaymentSplitRequests(splits)).toEqual([
      { method: "cash", amount: 5000, received_amount: 6000, change_amount: 500, change_override: true },
    ]);
  });

  it("上書きモードで負値入力は 0 にクランプする（誤入力耐性）", () => {
    const splits: PaymentSplitDraft[] = [
      { method: "cash", amount: "5000", receivedAmount: "6000", changeOverride: true, changeAmount: "-300" },
    ];
    expect(buildPaymentSplitRequests(splits)[0].change_amount).toBe(0);
  });

  it("非現金: received/change を 0 に固め change_override=false", () => {
    const splits: PaymentSplitDraft[] = [{ method: "credit_card", amount: "5000", receivedAmount: "" }];
    expect(buildPaymentSplitRequests(splits)).toEqual([
      { method: "credit_card", amount: 5000, received_amount: 0, change_amount: 0, change_override: false },
    ]);
  });

  it("金額未入力/0 の行は除外する", () => {
    const splits: PaymentSplitDraft[] = [
      { method: "cash", amount: "", receivedAmount: "" },
      { method: "cash", amount: "0", receivedAmount: "" },
      { method: "credit_card", amount: "3000", receivedAmount: "" },
    ];
    const result = buildPaymentSplitRequests(splits);
    expect(result).toHaveLength(1);
    expect(result[0].method).toBe("credit_card");
  });
});

describe("createInitialPaymentSplits (#188 再編集時の上書き状態復元)", () => {
  it("保存お釣りが派生値と一致: 上書き OFF で復元する", () => {
    const acc = {
      paymentSplits: [{ method: "cash", amount: 5000, receivedAmount: 6000, changeAmount: 1000 }],
    } as unknown as Accounting;
    const result = createInitialPaymentSplits(acc);
    expect(result[0].changeOverride).toBeUndefined();
  });

  it("保存お釣りが派生値と不一致: 上書き ON で復元し changeAmount を維持する（再保存での silent 復元を防ぐ）", () => {
    const acc = {
      paymentSplits: [{ method: "cash", amount: 5000, receivedAmount: 6000, changeAmount: 500 }],
    } as unknown as Accounting;
    const result = createInitialPaymentSplits(acc);
    expect(result[0].changeOverride).toBe(true);
    expect(result[0].changeAmount).toBe("500");
  });

  it("payment フォールバック経路（単一支払い backward compat）でも上書き状態を復元する", () => {
    const acc = {
      paymentSplits: [],
      payment: { method: "cash", billingAmount: 5000, receivedAmount: 6000, changeAmount: 500 },
    } as unknown as Accounting;
    const result = createInitialPaymentSplits(acc);
    expect(result[0].changeOverride).toBe(true);
    expect(result[0].changeAmount).toBe("500");
  });
});

describe("getCorrectableCardPayments (#189 確定後クレジット訂正の対象判定)", () => {
  it("確定済み + クレジット split: 対象として金額付きで返す", () => {
    const acc = {
      status: "completed",
      paymentSplits: [{ method: "credit_card", amount: 12000, receivedAmount: 0, changeAmount: 0 }],
    } as unknown as Accounting;
    expect(getCorrectableCardPayments(acc)).toEqual([{ method: "credit_card", amount: 12000 }]);
  });

  it("確定済み + 電子マネー split: 対象に含む", () => {
    const acc = {
      status: "completed",
      paymentSplits: [{ method: "electronic_money", amount: 800, receivedAmount: 0, changeAmount: 0 }],
    } as unknown as Accounting;
    expect(getCorrectableCardPayments(acc)).toEqual([{ method: "electronic_money", amount: 800 }]);
  });

  it("確定済み + 混在(現金+カード): カードのみ対象", () => {
    const acc = {
      status: "completed",
      paymentSplits: [
        { method: "cash", amount: 5000, receivedAmount: 5000, changeAmount: 0 },
        { method: "credit_card", amount: 5000, receivedAmount: 0, changeAmount: 0 },
      ],
    } as unknown as Accounting;
    expect(getCorrectableCardPayments(acc)).toEqual([{ method: "credit_card", amount: 5000 }]);
  });

  it("確定済み + 現金のみ: 対象なし(空配列=導線を出さない)", () => {
    const acc = {
      status: "completed",
      paymentSplits: [{ method: "cash", amount: 5000, receivedAmount: 5000, changeAmount: 0 }],
    } as unknown as Accounting;
    expect(getCorrectableCardPayments(acc)).toEqual([]);
  });

  it("確定済み + 銀行振込のみ: 対象なし", () => {
    const acc = {
      status: "completed",
      paymentSplits: [{ method: "bank_transfer", amount: 5000, receivedAmount: 0, changeAmount: 0 }],
    } as unknown as Accounting;
    expect(getCorrectableCardPayments(acc)).toEqual([]);
  });

  it("確定前(waiting): カードでも対象なし", () => {
    const acc = {
      status: "waiting",
      paymentSplits: [{ method: "credit_card", amount: 12000, receivedAmount: 0, changeAmount: 0 }],
    } as unknown as Accounting;
    expect(getCorrectableCardPayments(acc)).toEqual([]);
  });

  it("split 無し + 単一 payment がカード: payment.billingAmount を金額として返す", () => {
    const acc = {
      status: "completed",
      paymentSplits: undefined,
      payment: { method: "credit_card", billingAmount: 9800 },
    } as unknown as Accounting;
    expect(getCorrectableCardPayments(acc)).toEqual([{ method: "credit_card", amount: 9800 }]);
  });

  it("null: 空配列", () => {
    expect(getCorrectableCardPayments(null)).toEqual([]);
  });
});

describe("isPaymentSubmitDisabled / hasInvalidChangeOverride (BUG-010)", () => {
  const cash = (changeAmount: string): PaymentSplitDraft => ({
    method: "cash",
    amount: "1000",
    receivedAmount: "2000",
    changeOverride: true,
    changeAmount,
  });

  it("お釣りが負なら保存不可", () => {
    expect(isPaymentSubmitDisabled(1000, [cash("-1")])).toBe(true);
    expect(hasInvalidChangeOverride([cash("-1")])).toBe(true);
  });

  it("お釣り0以上なら保存可", () => {
    expect(isPaymentSubmitDisabled(1000, [cash("0")])).toBe(false);
    expect(hasInvalidChangeOverride([cash("0")])).toBe(false);
  });
});
