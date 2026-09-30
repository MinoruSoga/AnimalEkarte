import { Info } from "lucide-react";
import { Tooltip } from "@/components/ui/tooltip";
import { C, ICON } from "@/lib/design-tokens";
import { cn } from "@/components/ui/utils";

// ─────────────────────────────────────────────────
// ⓘ フィールド説明アイコン
//
// マスタ登録・編集フォームの各項目に「何の項目か」を
// 説明するヘルプアイコン。ホバーとキーボードフォーカス
// の両方で Tooltip を表示する。
// <label> 内に置くとスクリーンリーダーの名前計算に
// 混入するため、ラベル要素の外（行の右端など）に
// 兄弟要素として配置して使うこと。
// ─────────────────────────────────────────────────

interface FieldHelpProps {
  /** 説明対象のラベル名。ボタンの aria-label「{label}の説明」に使う */
  label: string;
  /** ツールチップの説明文（\n で改行可） */
  content: string;
  className?: string;
}

export function FieldHelp({ label, content, className }: FieldHelpProps) {
  return (
    <Tooltip content={content} className={cn("shrink-0 self-center", className)}>
      <button
        type="button"
        aria-label={`${label}の説明`}
        className={`inline-flex size-6 items-center justify-center rounded-xxs ${C.text40} ${C.hoverBgLight} transition-colors cursor-help outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
      >
        <Info className={ICON.sm} aria-hidden="true" />
      </button>
    </Tooltip>
  );
}
