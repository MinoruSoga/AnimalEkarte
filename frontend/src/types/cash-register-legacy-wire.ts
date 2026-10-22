/**
 * レジ締め API の移行 HTTP DTO（非永続）。
 *
 * 保存 model（生成 models.ts）ではなく、現行 handler が返す JSON の列をそのまま写す。
 * Go の json tag に合わせ、omitempty は省略可（?）で表す。
 * cash-register feature と shared hooks/lib が同じ wire を読むため共有層に置く
 * （feature 間 import は禁止）。レジ締めの wire 契約を生成型へ移すときに、
 * この adapter ごと置き換える（EMR-527）。
 */

/** @see backend/internal/billing/cash_register_response.go cashRegisterCloseResponse */
export interface LegacyCashRegisterCloseResponse {
  id: number;
  clinic_id: number;
  close_date: string;
  /** "am" | "pm" | "emg"（Go は素の string）。 */
  period: string;
  theoretical_cash: number;
  actual_cash: number;
  cash_difference: number;
  /** json.RawMessage（category_breakdown JSONB をそのまま返す）。 */
  category_breakdown: unknown;
  memo: string;
  closed_by?: number;
  closed_at: string;
  created_at: string;
  updated_at: string;
  /**
   * Go の cashRegisterCloseResponse には無い列（常に undefined）。
   * 既存 transform（closedByStaffName）の型を保つためだけに残す。要確認（EMR-527）。
   */
  closed_by_staff?: { name: string };
}
