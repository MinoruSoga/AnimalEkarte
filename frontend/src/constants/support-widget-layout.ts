/**
 * 右下常駐のサポート起動ボタン（features/support の SupportWidget）が
 * コンテンツを覆わないよう、画面下端に届くスクロール領域・パネルへ付ける
 * 下余白を連動させる共有定数。
 *
 * 余白はスクロール可能な末尾スペースとして内側に付与する（シェルの main に
 * 固定 padding を持たせると全ページ下端に恒常的な帯ができるため）。
 * SidePeek パネルはフッターをパネル下端に固定するためこの余白は付けず、
 * 代わりに globals.css が [data-side-peek] 存在時にボタンをパネルの左へ退避する。
 */
export const SUPPORT_WIDGET_LAYOUT = {
  /** 起動ボタンの固定位置（右・下 16px） */
  position: "fixed bottom-4 right-4",
  /** 起動ボタンのサイズ（48px）。変更時は footprintPx の算術を Layout.test.tsx が検証する */
  buttonSizeClass: "size-12",
  /** 起動ボタンが画面下端から占有する高さ = 16px(bottom-4) + 48px(buttonSizeClass) */
  footprintPx: 64,
  /** スクロール末端で footprintPx 以上の余白を確保し、ページ下端のコンテンツ・操作をボタンが覆わないようにする */
  scrollBottomClearance: "pb-20",
  scrollBottomClearancePx: 80,
} as const;
