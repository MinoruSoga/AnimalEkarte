import type { PaymentSplitRequest } from "../api/types";
import type { PaymentSplitDraft } from "../components/PaymentCard";
import type { Accounting, PaymentMethod } from "../types";

// #189: 確定後クレジット訂正の対象となるカード系支払い手段。
// 生成型 PaymentMethod は string 別名のため Extract が never に潰れる。リテラルユニオンで直接定義する。
export type CorrectableCardMethod = "credit_card" | "electronic_money";

export interface CorrectableCardPayment {
  method: CorrectableCardMethod;
  amount: number;
}

function isCorrectableCardMethod(method: PaymentMethod): method is CorrectableCardMethod {
  return method === "credit_card" || method === "electronic_money";
}

// getCorrectableCardPayments は確定後クレジット訂正(#189)の対象となるカード内訳を返す。
// 確定済み(status=completed)かつカード系支払いが存在する場合のみ非空配列を返す。
// payment_splits があればそれを優先し、無ければ単一 payment.method を対象とする。
// 確定前・現金のみ・銀行振込のみ等は空配列（=訂正導線を出さない）。
export function getCorrectableCardPayments(accounting: Accounting | null): CorrectableCardPayment[] {
  if (!accounting || accounting.status !== "completed") return [];
  const splits = accounting.paymentSplits;
  if (splits && splits.length > 0) {
    return splits.flatMap((s) =>
      isCorrectableCardMethod(s.method) ? [{ method: s.method, amount: s.amount }] : [],
    );
  }
  const payment = accounting.payment;
  if (payment && isCorrectableCardMethod(payment.method)) {
    return [{ method: payment.method, amount: payment.billingAmount }];
  }
  return [];
}

/** 締め済み日の新規/既存確定は理由入力か権限が無いと物理ブロックする（BUG-004）。 */
export function isPostCloseSubmitBlocked(args: {
  isScheduledDateClosed: boolean;
  canPostCloseEdit: boolean;
  postCloseReason: string;
}): boolean {
  if (!args.isScheduledDateClosed) return false;
  if (!args.canPostCloseEdit) return true;
  return args.postCloseReason.trim() === "";
}

export function isPaymentSubmitDisabled(
  billingAmount: number,
  paymentSplits: PaymentSplitDraft[],
): boolean {
  const splitTotal = paymentSplits.reduce((sum, s) => sum + parseInt(s.amount || "0", 10), 0);
  if (billingAmount - splitTotal !== 0) return true;
  return paymentSplits.some((s) => {
    if (!s.amount || parseInt(s.amount, 10) <= 0) return true;
    if (s.method === "cash") {
      if (parseInt(s.receivedAmount || "0", 10) < parseInt(s.amount, 10)) return true;
      if (s.changeOverride === true) {
        const change = Number(s.changeAmount);
        if (s.changeAmount !== undefined && s.changeAmount !== "" && (Number.isNaN(change) || change < 0)) {
          return true;
        }
      }
    }
    return false;
  });
}

export function hasInvalidChangeOverride(paymentSplits: PaymentSplitDraft[]): boolean {
  return paymentSplits.some((s) => {
    if (s.method !== "cash" || s.changeOverride !== true) return false;
    if (s.changeAmount === undefined || s.changeAmount === "") return false;
    const change = Number(s.changeAmount);
    return Number.isNaN(change) || change < 0;
  });
}

export interface AccountingFormState {
  success: boolean;
  timestamp: number;
}

// #188: 保存済み split から「お釣り直接上書き」状態を復元する。
// 現金 split の保存お釣りが派生値（received - amount）と異なる場合は上書きされていたとみなし、
// 再編集時も上書きモードを維持する（再保存で派生値へ silently 戻る回帰を防ぐ）。
function restoreChangeOverride(
  method: PaymentSplitDraft["method"],
  amount: number,
  receivedAmount: number,
  changeAmount: number,
): Partial<Pick<PaymentSplitDraft, "changeOverride" | "changeAmount">> {
  if (method !== "cash") return {};
  const derivedChange = Math.max(0, receivedAmount - amount);
  if (changeAmount === derivedChange) return {};
  return { changeOverride: true, changeAmount: changeAmount.toString() };
}

export function createInitialPaymentSplits(baseAccounting: Accounting | null): PaymentSplitDraft[] {
  const splits = baseAccounting?.paymentSplits;
  if (splits && splits.length > 0) {
    return splits.map((split) => ({
      method: split.method,
      amount: split.amount.toString(),
      receivedAmount: split.receivedAmount > 0 ? split.receivedAmount.toString() : "",
      ...restoreChangeOverride(split.method, split.amount, split.receivedAmount, split.changeAmount),
    }));
  }

  const payment = baseAccounting?.payment;
  if (payment) {
    return [{
      method: payment.method,
      amount: payment.billingAmount.toString(),
      receivedAmount: payment.receivedAmount > 0 ? payment.receivedAmount.toString() : "",
      ...restoreChangeOverride(payment.method, payment.billingAmount, payment.receivedAmount, payment.changeAmount),
    }];
  }

  return [{ method: "cash", amount: "", receivedAmount: "" }];
}

// buildPaymentSplitRequests は支払い内訳ドラフトを API リクエスト payload に変換する（#188 契約テスト対象）。
// - 金額未入力/0 の行は除外する。
// - 現金以外は received/change を 0 に固める。
// - 現金は通常 change = max(0, received - amount) を派生算出するが、
//   changeOverride=true（#188 お釣り直接上書き）の場合は入力された changeAmount を verbatim で送出し、
//   change_override フラグを立てて BE の整合検証緩和を要求する。
export function buildPaymentSplitRequests(splits: PaymentSplitDraft[]): PaymentSplitRequest[] {
  return splits
    .filter((split) => split.amount && parseInt(split.amount, 10) > 0)
    .map((split) => {
      const amount = parseInt(split.amount, 10);
      const isCash = split.method === "cash";
      const received = isCash ? parseInt(split.receivedAmount || "0", 10) : 0;
      const isOverride = isCash && split.changeOverride === true;
      const change = isCash
        ? isOverride
          ? Math.max(0, parseInt(split.changeAmount || "0", 10))
          : Math.max(0, received - amount)
        : 0;
      return {
        method: split.method,
        amount,
        received_amount: received,
        change_amount: change,
        change_override: isOverride,
      };
    });
}
