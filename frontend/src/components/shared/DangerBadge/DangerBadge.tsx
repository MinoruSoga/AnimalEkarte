import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { C } from "@/lib/design-tokens";

/**
 * ペット危険度バッジの段階。
 * 高 = 赤「⚠ 危険」/ 中 = 黄「⚠ 注意」。低・未設定は描画しない。
 */
export type DangerBadgePetLevel = "high" | "medium";

const PET_BADGE_STYLE: Record<
  DangerBadgePetLevel,
  { label: string; reasonLabel: string; badgeClass: string; headingClass: string }
> = {
  high: {
    label: "⚠ 危険",
    reasonLabel: "危険理由",
    badgeClass: `${C.bgDanger10} ${C.danger} ${C.borderDanger20}`,
    headingClass: C.danger,
  },
  medium: {
    label: "⚠ 注意",
    reasonLabel: "注意理由",
    badgeClass: `${C.bgNotice} ${C.textNotice} ${C.borderNotice}`,
    headingClass: C.textNotice,
  },
};

/**
 * wire 値 ("high" / "medium" / "low") と画面表示値 ("高" / "中" / "低") の
 * 両方からバッジ段階を導く。低・未設定・未知値は undefined（fail-closed で非表示）。
 */
function toDangerBadgePetLevel(
  dangerLevel: string | null | undefined,
): DangerBadgePetLevel | undefined {
  switch (dangerLevel) {
    case "高":
    case "high":
      return "high";
    case "中":
    case "medium":
      return "medium";
    default:
      return undefined;
  }
}

export interface PetDangerBadgeProps {
  variant: "pet";
  /**
   * 危険度。wire 値 ("high" 等) と画面表示値 ("高" 等) の両方を受け付ける。
   * 低・未設定・未知値なら何も描画しない。
   */
  level?: string | null;
  /** aria-label と Popover 見出しに使う対象名（ペット名）。 */
  subjectName: string;
  /** 危険理由。未設定・空白のみは「理由未登録」を表示する。 */
  reason?: string | null;
  /**
   * カード全体が click / drag 対象の場面で true。
   * trigger と content 双方の pointerdown・click 伝播を止める。
   */
  stopPropagation?: boolean;
}

export interface OwnerDangerBadgeProps {
  variant: "owner";
  /** カード全体が click / drag 対象の場面で true。 */
  stopPropagation?: boolean;
}

export type DangerBadgeProps = PetDangerBadgeProps | OwnerDangerBadgeProps;

/**
 * スタッフ向け危険マークの共有バッジ。
 * - variant="pet": 危険度 高/中 の Popover 付きバッジ（⚠ + 文言で色に依存しない識別）
 * - variant="owner": 飼主 is_dangerous の静的バッジ「⚠ 危険人物」（理由は契約上存在しない）
 */
export function DangerBadge(props: DangerBadgeProps) {
  const propagationGuard = props.stopPropagation
    ? {
        onPointerDown: (event: { stopPropagation(): void }) => event.stopPropagation(),
        onClick: (event: { stopPropagation(): void }) => event.stopPropagation(),
      }
    : {};

  if (props.variant === "owner") {
    return (
      <span
        className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold ${C.bgDanger10} ${C.danger} ${C.borderDanger20}`}
        {...propagationGuard}
      >
        ⚠ 危険人物
      </span>
    );
  }

  const level = toDangerBadgePetLevel(props.level);
  if (level === undefined) return null;

  const style = PET_BADGE_STYLE[level];
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label={`${props.subjectName}の${style.reasonLabel}を表示`}
          className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold ${style.badgeClass} outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
          {...propagationGuard}
        >
          {style.label}
        </button>
      </PopoverTrigger>
      <PopoverContent
        align="start"
        aria-label={`${props.subjectName}の${style.reasonLabel}`}
        onOpenAutoFocus={(event) => event.preventDefault()}
        className="w-64"
        {...propagationGuard}
      >
        <p className={`text-sm font-semibold ${style.headingClass}`}>{style.reasonLabel}</p>
        <p className={`mt-1 whitespace-pre-wrap break-words text-sm ${C.textInkSecondary}`}>
          {props.reason?.trim() || "理由未登録"}
        </p>
      </PopoverContent>
    </Popover>
  );
}
