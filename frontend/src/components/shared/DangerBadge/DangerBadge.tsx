import { OctagonAlert, TriangleAlert, type LucideIcon } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { BADGE, C, ICON } from "@/lib/design-tokens";

/**
 * ペット特記マークの段階（EMR-231）。
 * 高 = 赤い八角形アイコン / 中 = 黄い三角形アイコン。低・未設定は描画しない。
 * 文言は一切出さず、アイコンの形と色だけでスタッフが識別する
 * （色だけに依存しないよう段階ごとに形も分ける）。
 */
type DangerBadgePetLevel = "high" | "medium";

const PET_BADGE_STYLE: Record<
  DangerBadgePetLevel,
  { Icon: LucideIcon; badgeClass: string; headingClass: string; levelText: string }
> = {
  high: {
    Icon: OctagonAlert,
    badgeClass: `${C.bgDanger10} ${C.danger} ${C.borderDanger20}`,
    headingClass: C.danger,
    levelText: "高",
  },
  medium: {
    Icon: TriangleAlert,
    badgeClass: BADGE.yellow,
    headingClass: C.textBadgeYellow,
    levelText: "中",
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

interface PetDangerBadgeProps {
  variant: "pet";
  /**
   * 特記レベル。wire 値 ("high" 等) と画面表示値 ("高" 等) の両方を受け付ける。
   * 低・未設定・未知値なら何も描画しない。
   */
  level?: string | null;
  /** aria-label と Popover 見出しに使う対象名（ペット名）。 */
  subjectName: string;
  /** スタッフ向け補足メモ（ユーザー入力）。未設定・空白のみは「内容未登録」を表示する。 */
  reason?: string | null;
  /**
   * カード全体が click / drag 対象の場面で true。
   * trigger と content 双方の pointerdown・click 伝播を止める。
   */
  stopPropagation?: boolean;
}

interface OwnerDangerBadgeProps {
  variant: "owner";
  /** カード全体が click / drag 対象の場面で true。 */
  stopPropagation?: boolean;
}

type DangerBadgeProps = PetDangerBadgeProps | OwnerDangerBadgeProps;

/**
 * スタッフ向け特記マークの共有バッジ（EMR-231）。
 * 来院者に意味が伝わる文言は画面・aria-label・title のいずれにも出さない。
 * - variant="pet": 高/中 の Popover 付きアイコンバッジ。補足メモはクリック開示。
 * - variant="owner": 飼主 is_dangerous の静的アイコンマーク（補足メモは契約上存在しない）
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
        role="img"
        aria-label="特記"
        className={`inline-flex items-center rounded px-1.5 py-0.5 leading-none ${C.bgDanger10} ${C.danger} ${C.borderDanger20}`}
        {...propagationGuard}
      >
        <OctagonAlert className={ICON.sm} aria-hidden="true" />
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
          aria-label={`${props.subjectName}の詳細を表示`}
          className={`inline-flex items-center rounded px-1.5 py-0.5 leading-none ${style.badgeClass} outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
          {...propagationGuard}
        >
          <style.Icon className={ICON.sm} aria-hidden="true" />
        </button>
      </PopoverTrigger>
      <PopoverContent
        align="start"
        aria-label={`${props.subjectName}の詳細`}
        onOpenAutoFocus={(event) => event.preventDefault()}
        className="w-64"
        {...propagationGuard}
      >
        <p className={`text-sm font-semibold ${style.headingClass}`}>
          <span className="inline-flex items-center gap-1">
            <style.Icon className={ICON.sm} aria-hidden="true" />
            特記レベル: {style.levelText}
          </span>
        </p>
        <p className={`mt-1 whitespace-pre-wrap break-words text-sm ${C.textInkSecondary}`}>
          {props.reason?.trim() || "内容未登録"}
        </p>
      </PopoverContent>
    </Popover>
  );
}
