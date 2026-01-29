export const APP_COLORS = {
  primary: "#37352F",
  borderLight: "rgba(55,53,47,0.09)",
  borderMedium: "rgba(55,53,47,0.16)",
  bgHover: "#F7F6F3",
  textMuted: "#37352F/60",
};

export const TABLE_STYLES = {
  row: "border-b-[rgba(55,53,47,0.09)] hover:bg-[#F7F6F3]/50 transition-colors cursor-pointer h-12",
  actionButton: "size-11 text-[#37352F]/60 hover:text-[#37352F]",
  cell: "text-sm text-[#37352F] py-2",
  cellMono: "font-mono text-sm text-[#37352F] py-2",
};

/**
 * タブレット最適化用のスタイルトークン
 * Apple HIG推奨: タップ領域最小 44x44px
 */
export const TABLET_STYLES = {
  touch: {
    /** 最小タップ領域: 44x44px */
    minTapSize: "min-h-[44px] min-w-[44px]",
    /** ボタン高さ: 44px */
    buttonHeight: "h-11",
    /** アイコンボタンサイズ: 44x44px */
    iconButtonSize: "size-11",
  },
  sidebar: {
    /** ドロワー幅 */
    drawerWidth: "w-[280px]",
    /** メニューアイテム高さ */
    menuItemHeight: "min-h-[44px]",
    /** メニューアイテムパディング */
    menuItemPadding: "px-4 py-3",
  },
  grid: {
    /** フォーム用グリッド: 段階的レスポンシブ */
    form: "grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4",
    /** カンバン用グリッド: タブレットで2-3列 */
    kanban: "grid grid-cols-2 md:grid-cols-3 lg:flex lg:gap-2",
  },
  breakpoints: {
    /** モバイル: < 768px */
    mobile: 768,
    /** タブレット: 768px - 1023px */
    tablet: 1024,
  },
};
