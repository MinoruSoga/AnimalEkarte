import Cake from "lucide-react/dist/esm/icons/cake";

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { BADGE, C, ICON } from "@/lib/design-tokens";
import { isBirthMonthJST } from "@/lib/jst-date";

interface BirthdayBadgeProps {
  /**
   * 生年月日（YYYY-MM-DD）。誕生月でない・欠損・形式不正・暦上無効なら
   * 何も描画しない（isBirthMonthJST が fail-closed）。
   */
  birthDate?: string | null;
  /** Popover 見出し / tooltip に使う対象名（ペット名）。 */
  subjectName: string;
  /**
   * カード全体が click / drag 対象の場面で true。
   * trigger と content 双方の pointerdown・click 伝播を止める。
   * （variant="static" では無効 — 祖先の button 内に interactive 要素を置けないため）
   */
  stopPropagation?: boolean;
  /**
   * "popover"（既定）: ボタン + Popover で生年月日をクリック開示する。
   * "static": 祖先が button / drag ハンドルの場面用。button を入れ子にできないため
   *   `role="img"` + `title`（生年月日 tooltip）の静的アイコンのみを描画する。
   */
  variant?: "popover" | "static";
}

/**
 * 誕生月ペットの共有バッジ（EMR-301）。DangerBadge と同じ
 * button + Popover + stopPropagation + design token の型に揃える。
 * - 表示条件: isBirthMonthJST(birthDate)。死亡ペットへの抑制は呼び出し側責務。
 * - アイコンのみ描画し、生年月日はクリック開示する。
 */
export function BirthdayBadge({
  birthDate,
  subjectName,
  stopPropagation,
  variant = "popover",
}: BirthdayBadgeProps) {
  const propagationGuard =
    stopPropagation && variant !== "static"
      ? {
          onPointerDown: (event: { stopPropagation(): void }) => event.stopPropagation(),
          onClick: (event: { stopPropagation(): void }) => event.stopPropagation(),
        }
      : {};

  if (!isBirthMonthJST(birthDate)) return null;

  if (variant === "static") {
    return (
      <span
        role="img"
        aria-label="今月誕生日"
        title={`${subjectName}の生年月日: ${birthDate}`}
        className={`inline-flex items-center rounded px-1.5 py-0.5 leading-none ${BADGE.purple}`}
      >
        <Cake className={ICON.sm} aria-hidden="true" />
      </span>
    );
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label="今月誕生日"
          className={`inline-flex items-center rounded px-1.5 py-0.5 leading-none ${BADGE.purple} outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
          {...propagationGuard}
        >
          <Cake className={ICON.sm} aria-hidden="true" />
        </button>
      </PopoverTrigger>
      <PopoverContent
        align="start"
        aria-label="今月誕生日"
        onOpenAutoFocus={(event) => event.preventDefault()}
        className="w-64"
        {...propagationGuard}
      >
        <p className={`text-sm font-semibold ${C.textStatusPurple}`}>
          <span className="inline-flex items-center gap-1">
            <Cake className={ICON.sm} aria-hidden="true" />
            {subjectName}は今月誕生日
          </span>
        </p>
        <p className={`mt-1 text-sm ${C.textInkSecondary}`}>生年月日: {birthDate}</p>
      </PopoverContent>
    </Popover>
  );
}
