/**
 * 右下常駐のサポート起動ボタン（features/support の SupportWidget）と、
 * 認証済みシェル（components/shared/Layout）の main 下余白を連動させる共有定数。
 */
export const SUPPORT_WIDGET_LAYOUT = {
  /** 起動ボタンの固定位置（右・下 16px） */
  position: "fixed bottom-4 right-4",
  /** 起動ボタンのサイズ（48px）。変更時は footprintPx の算術を Layout.test.tsx が検証する */
  buttonSizeClass: "size-12",
  /** 起動ボタンが画面下端から占有する高さ = 16px(bottom-4) + 48px(buttonSizeClass) */
  footprintPx: 64,
  /** 認証済みシェル main の下余白。footprintPx 以上を確保し、ページ下端や SidePeek フッタの操作をボタンが覆わないようにする */
  shellBottomClearance: "pb-20",
  shellBottomClearancePx: 80,
} as const;
