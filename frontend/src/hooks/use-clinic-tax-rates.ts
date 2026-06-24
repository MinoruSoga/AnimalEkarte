import { useAuth } from "@/hooks/use-auth";

/** 消費税率の法定既定値（病院マスタ未設定時のフォールバック）。 */
const DEFAULT_STANDARD_TAX_RATE = 0.1;
const DEFAULT_REDUCED_TAX_RATE = 0.08;

export interface ClinicTaxRates {
  /** 標準税率（例: 0.1）。 */
  standardTaxRate: number;
  /** 軽減税率（例: 0.08）。 */
  reducedTaxRate: number;
}

/**
 * 病院マスタ設定の消費税率を返す共有フック（#179 ②）。
 *
 * 明細兼領収書（AccountingDocument）と同一の正本（`user.clinic` の
 * standardTaxRate / reducedTaxRate）を参照し、月次集計レポートを含む
 * 全帳票で税率表記を一貫させる。病院マスタの税率を変更すれば、
 * 明細兼領収書・月次集計レポートの双方へ自動反映される。
 *
 * 未設定時は法定既定（標準 10% / 軽減 8%）へフォールバックする。
 * メンバーシップ（user.clinics）は税率を保持しないため、税率の正本は
 * メイン医院詳細（user.clinic）に限られる。
 */
export function useClinicTaxRates(): ClinicTaxRates {
  const { user } = useAuth();
  return {
    standardTaxRate: user?.clinic?.standardTaxRate ?? DEFAULT_STANDARD_TAX_RATE,
    reducedTaxRate: user?.clinic?.reducedTaxRate ?? DEFAULT_REDUCED_TAX_RATE,
  };
}

/** 税率（0.1 等）を表示用パーセント文字列（"10%"）へ整形する。 */
export function formatTaxRatePercent(rate: number): string {
  return `${Math.round(rate * 100)}%`;
}
