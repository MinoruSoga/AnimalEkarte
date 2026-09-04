/**
 * BUG-027: LIFF 健康カードと LINE 予約で共通の顧客向けエラー chrome。
 * 背景・警告アイコン・見出し色/文言・本文階層・任意の再試行を一箇所に固定する。
 */
/* eslint-disable react-refresh/only-export-components -- title/theme は liff と line-reserve で共有 */
export const SHARED_ERROR_PAGE_TITLE = "エラーが発生しました";

/** line-reserve 側の teal / blue-gray 系を正とし、両アプリで同一クラスを使う */
export const DEFAULT_ERROR_PAGE_THEME: ErrorPageTheme = {
  bg: "bg-noah-teal-light",
  heading: "text-noah-teal-dark",
  body: "text-noah-text-sub text-sm",
  button: "bg-noah-teal",
  buttonHover: "hover:bg-noah-teal-dark",
};

export interface ErrorPageTheme {
  bg: string;
  heading: string;
  body: string;
  button: string;
  buttonHover: string;
}

interface ErrorPageProps {
  message?: string;
  /** 省略時は SHARED_ERROR_PAGE_TITLE */
  title?: string;
  /** 省略時は DEFAULT_ERROR_PAGE_THEME */
  theme?: ErrorPageTheme;
  /** 省略時は window.location.reload。showAction=false のとき無視 */
  onAction?: () => void;
  /** デフォルト: 再読み込み */
  actionLabel?: string;
  /**
   * false のときアクション非表示。
   * clinic_id 欠落など再試行で解消しない恒久エラー向け（BUG-014 / BUG-027）。
   */
  showAction?: boolean;
}

export function ErrorPage({
  message = "予期しないエラーが発生しました。",
  title = SHARED_ERROR_PAGE_TITLE,
  theme = DEFAULT_ERROR_PAGE_THEME,
  onAction,
  actionLabel = "再読み込み",
  showAction = true,
}: ErrorPageProps) {
  const handleAction = onAction ?? (() => window.location.reload());

  return (
    <div className={`min-h-screen flex items-center justify-center ${theme.bg}`}>
      <div className="max-w-md mx-auto px-4 text-center">
        <div className="text-6xl mb-4" aria-hidden="true">
          ⚠️
        </div>
        <h1 className={`text-xl font-bold ${theme.heading} mb-2`}>{title}</h1>
        <p className={`${theme.body} mb-6`} role="alert">
          {message}
        </p>
        {showAction ? (
          <button
            type="button"
            onClick={handleAction}
            className={`py-3 px-6 ${theme.button} text-white rounded-xl font-semibold ${theme.buttonHover}`}
          >
            {actionLabel}
          </button>
        ) : null}
      </div>
    </div>
  );
}
