import type { ReactNode } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import { C, ICON } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";

interface CalendarNavToolbarProps {
  /** ナビゲーション対象の現在値表示（見出し要素など、呼び出し側が完全に制御する） */
  label: ReactNode;
  onPrev: () => void;
  onNext: () => void;
  /** 指定時のみ「今日」ボタンを表示する */
  onToday?: () => void;
  todayLabel?: string;
  /** prev/next はアイコンのみのボタンのため、アクセシブルネームとして必須にする */
  prevAriaLabel: string;
  nextAriaLabel: string;
  prevDisabled?: boolean;
  nextDisabled?: boolean;
  variant?: "ghost" | "outline";
  /**
   * sm: コンパクトなラベル（簡易版・spread レイアウトで使用）
   * lg: 標準ラベル（今日ボタンありのクラスタ型で使用）
   * auto: 標準ラベル + Button の既定スタイル
   * ナビゲーションボタンはいずれも min 44x44px を保証する
   */
  size?: "sm" | "lg" | "auto";
  /**
   * clustered: prev+today?+next を枠線+シャドウ付きピルにまとめ、label はその横に並べる
   * spread: prev - label - next を直接の flex 子要素として justify-between で配置（today ボタンなし）
   * inline: prev - label - next を直接の flex 子要素として gap-2 で配置（today ボタンなし）
   */
  layout?: "clustered" | "spread" | "inline";
  className?: string;
}

/**
 * 予約管理カレンダー等の prev/today/next ナビゲーションクラスターの共通実装。
 * ReservationManagementCalendar / ReservationTypeAvailableSlotsCalendar (today付き・clustered) と
 * DailyDateNav (spread・today無し) / ShiftCalendar (inline・today無し) の4系統を統合。
 */
export function CalendarNavToolbar({
  label,
  onPrev,
  onNext,
  onToday,
  todayLabel = "今日",
  prevAriaLabel,
  nextAriaLabel,
  prevDisabled,
  nextDisabled,
  variant = "ghost",
  size = "lg",
  layout = "clustered",
  className,
}: CalendarNavToolbarProps) {
  const iconBtnClass = "size-11 min-h-11 min-w-11";
  const todayBtnClass = size === "sm" ? "h-11 px-3 text-sm" : "h-11 px-4 text-base font-medium";

  const prevBtn = (
    <Button
      variant={variant}
      size="icon"
      className={iconBtnClass}
      onClick={onPrev}
      disabled={prevDisabled}
      aria-label={prevAriaLabel}
    >
      <ChevronLeft className={ICON.page} />
    </Button>
  );
  const nextBtn = (
    <Button
      variant={variant}
      size="icon"
      className={iconBtnClass}
      onClick={onNext}
      disabled={nextDisabled}
      aria-label={nextAriaLabel}
    >
      <ChevronRight className={ICON.page} />
    </Button>
  );
  const todayBtn = onToday ? (
    <Button variant={variant} size="sm" className={todayBtnClass} onClick={onToday}>
      {todayLabel}
    </Button>
  ) : null;

  if (layout === "spread") {
    return (
      <div className={cn("flex items-center justify-between", className)}>
        {prevBtn}
        {label}
        {nextBtn}
      </div>
    );
  }

  if (layout === "inline") {
    return (
      <div className={cn("flex items-center gap-2", className)}>
        {prevBtn}
        {label}
        {nextBtn}
      </div>
    );
  }

  return (
    <div className={cn("flex items-center gap-4", className)}>
      <div
        className={cn("flex items-center", C.bgWhite, "rounded-md border", C.borderMedium, "p-1")}
      >
        {prevBtn}
        {todayBtn}
        {nextBtn}
      </div>
      {label}
    </div>
  );
}
