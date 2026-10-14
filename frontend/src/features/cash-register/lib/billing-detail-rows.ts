import type { CloseBillingDetail } from "../api/get-cash-register-preview";

/**
 * EMR-354: 個別会計明細を「会計1件=1行」の日次帳票バケットへ投影する pure helper。
 *
 * backend は split 明細行（1 会計 × split 数）を返す。画面（BillingDetailTable）と
 * 印刷（ClosePrintArea）はこの単一関数を描画源として共有し、行集約・合計・返金・
 * 純額の表示値を一致させる。
 *
 * - totalAmount: split billing_amount の和（GROSS）
 * - refundAmount: 全期間スカラが各行に重複しているため1回だけ取る
 * - netAmount: GROSS 合計 - 返金（netAmount の行和ではない）
 * - departments: category_amounts を日次7バケットへ合算。会計のいずれかの行で
 *   category_amounts が欠落（旧ペイロード / null）の場合は null（「記録なし」）。
 *   MIN(category) へ全額を寄せることはしない。
 * - 未知のカテゴリキーは用品他バケットへ畳み込む（サイレントに消さない）。
 */

/** 日次帳票の部門バケット（表示順固定）。goods/training/other は用品他にグルーピング。 */
export const DETAIL_DEPARTMENT_BUCKETS = [
  { key: "medical", label: "診療", categories: ["examination", "test", "procedure", "medicine"] },
  { key: "surgery", label: "外科", categories: ["surgery"] },
  { key: "rv", label: "RV", categories: ["vaccine"] },
  { key: "food", label: "フード", categories: ["food"] },
  { key: "trimming", label: "トリミング", categories: ["trimming"] },
  { key: "hotel", label: "ホテル", categories: ["hotel"] },
  { key: "goodsOther", label: "用品他", categories: ["goods", "training", "other"] },
] as const;

type DetailDepartmentKey = (typeof DETAIL_DEPARTMENT_BUCKETS)[number]["key"];

const CATEGORY_TO_BUCKET: Readonly<Record<string, DetailDepartmentKey>> =
  DETAIL_DEPARTMENT_BUCKETS.reduce(
    (acc, bucket) => {
      for (const cat of bucket.categories) acc[cat] = bucket.key;
      return acc;
    },
    {} as Record<string, DetailDepartmentKey>,
  );

function departmentBucketFor(category: string): DetailDepartmentKey {
  // 未知カテゴリは用品他へ畳み込む（金額を消さない）。想定外カテゴリの可視性は
  // 上部サマリーの「未分類・要確認」と併せて担保する。
  return CATEGORY_TO_BUCKET[category] ?? "goodsOther";
}

function zeroDepartments(): Record<DetailDepartmentKey, number> {
  return {
    medical: 0,
    surgery: 0,
    rv: 0,
    food: 0,
    trimming: 0,
    hotel: 0,
    goodsOther: 0,
  };
}

const RECEIPT_NO_WIDTH = 8;

function formatReceiptNo(billingId: number): string {
  return String(billingId).padStart(RECEIPT_NO_WIDTH, "0");
}

/** 投影後の1行（画面・印刷共通）。departments=null は「記録なし」。 */
export interface BillingDetailDisplayRow {
  billingId: number;
  paidAt: string;
  /** 領収番号（billingId の 8 桁ゼロ埋め） */
  receiptNo: string;
  departments: Record<DetailDepartmentKey, number> | null;
  /** 重複なし・出現順の支払方法名を "/" 結合した表示文字列 */
  paymentMethodLabel: string;
  /** Σ split billing_amount（GROSS） */
  totalAmount: number;
  /** 全期間返金スカラ（1回だけ） */
  refundAmount: number;
  /** totalAmount - refundAmount */
  netAmount: number;
  ownerName: string;
  petName: string;
  isHospitalization: boolean;
}

interface RowBuilder {
  billingId: number;
  paidAt: string;
  receiptNo: string;
  departments: Record<DetailDepartmentKey, number>;
  departmentsUnknown: boolean;
  paymentMethodNames: string[];
  totalAmount: number;
  refundAmount: number;
  ownerName: string;
  petName: string;
  isHospitalization: boolean;
}

export function projectBillingDetailRows(
  details: readonly CloseBillingDetail[],
): BillingDetailDisplayRow[] {
  const rowsById = new Map<number, RowBuilder>();
  const order: number[] = [];

  for (const d of details) {
    let row = rowsById.get(d.billingId);
    if (!row) {
      row = {
        billingId: d.billingId,
        paidAt: d.paidAt,
        receiptNo: formatReceiptNo(d.billingId),
        departments: zeroDepartments(),
        departmentsUnknown: false,
        paymentMethodNames: [],
        totalAmount: 0,
        refundAmount: d.refundAmount,
        ownerName: d.ownerName,
        petName: d.petName,
        isHospitalization: d.isHospitalization,
      };
      rowsById.set(d.billingId, row);
      order.push(d.billingId);
    }

    row.totalAmount += d.billingAmount;
    row.isHospitalization = row.isHospitalization || d.isHospitalization;
    if (d.paymentMethodName && !row.paymentMethodNames.includes(d.paymentMethodName)) {
      row.paymentMethodNames.push(d.paymentMethodName);
    }
    if (d.categoryAmounts == null) {
      row.departmentsUnknown = true;
    } else {
      for (const [category, amount] of Object.entries(d.categoryAmounts)) {
        row.departments[departmentBucketFor(category)] += amount;
      }
    }
  }

  return order.map((billingId) => {
    const row = rowsById.get(billingId)!;
    return {
      billingId: row.billingId,
      paidAt: row.paidAt,
      receiptNo: row.receiptNo,
      departments: row.departmentsUnknown ? null : row.departments,
      paymentMethodLabel: row.paymentMethodNames.join("/"),
      totalAmount: row.totalAmount,
      refundAmount: row.refundAmount,
      netAmount: row.totalAmount - row.refundAmount,
      ownerName: row.ownerName,
      petName: row.petName,
      isHospitalization: row.isHospitalization,
    };
  });
}

/** 印刷フッタ用の可視行再集計。departments=null は不明行を含み捏造を避ける。 */
export interface BillingDetailRowTotals {
  departments: Record<DetailDepartmentKey, number> | null;
  totalAmount: number;
  refundAmount: number;
  netAmount: number;
}

export function sumBillingDetailRowTotals(
  rows: readonly BillingDetailDisplayRow[],
): BillingDetailRowTotals {
  const departments = zeroDepartments();
  let departmentsUnknown = false;
  let totalAmount = 0;
  let refundAmount = 0;
  for (const row of rows) {
    totalAmount += row.totalAmount;
    refundAmount += row.refundAmount;
    if (row.departments === null) {
      departmentsUnknown = true;
      continue;
    }
    for (const bucket of DETAIL_DEPARTMENT_BUCKETS) {
      departments[bucket.key] += row.departments[bucket.key];
    }
  }
  return {
    departments: departmentsUnknown ? null : departments,
    totalAmount,
    refundAmount,
    netAmount: totalAmount - refundAmount,
  };
}
