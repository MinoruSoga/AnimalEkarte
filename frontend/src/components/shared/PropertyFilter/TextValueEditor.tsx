import { memo, useCallback, useState } from "react";
import { Check } from "lucide-react";

import { C, ICON } from "@/lib/design-tokens";

/**
 * EMR-245: 表示列テキストフィルタ（type: "text"）の値エディタ。
 * DateValueEditor と同じ onApply(value, displayValue) 契約。
 * 確定は Enter / 適用ボタンのみ — キー入力毎に確定しない
 * （EMR-247 の検索バー確定方式と同じ運用。列フィルタも確定毎に
 * 一覧+件数の DB クエリが走るため連打を防ぐ）。
 */
interface TextValueEditorProps {
  /** 対象プロパティのラベル（aria-label / placeholder に使用） */
  label: string;
  /** 編集中の既存値（ルール行からの再編集時） */
  currentValue?: string;
  onApply: (value: string, displayValue: string) => void;
}

export const TextValueEditor = memo(function TextValueEditor({
  label,
  currentValue = "",
  onApply,
}: TextValueEditorProps) {
  const [draft, setDraft] = useState(currentValue);
  const trimmed = draft.trim();

  // 空白のみの入力は確定しない（BE 側も正規化後に空なら 0 件になるため、
  // 全件消失に見える空フィルタを UI 側で作らない）
  const apply = useCallback(() => {
    if (trimmed === "") return;
    onApply(trimmed, trimmed);
  }, [trimmed, onApply]);

  return (
    <div className="flex items-center gap-1 p-1">
      <input
        type="text"
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={(e) => {
          // IME 変換中の確定 Enter / キーリピートでは適用しない
          // （検索バーと同じ isComposing + keyCode 229 判定）。
          const isComposing = e.nativeEvent.isComposing || e.nativeEvent.keyCode === 229;
          if (e.key === "Enter" && !e.repeat && !isComposing) apply();
        }}
        placeholder={`${label}を入力...`}
        aria-label={`${label}を入力`}
        className={`h-11 w-full px-2 text-base ${C.text} ${C.textPlaceholder} ${C.bgPage} border border-transparent rounded-xs outline-none transition-colors ${C.hoverBgPageDark} focus:bg-white ${C.focusBorderLight} focus-visible:ring-2 ${C.focusRingAccent40}`}
        autoFocus
      />
      <button
        type="button"
        onClick={apply}
        disabled={trimmed === ""}
        className={`flex min-h-11 min-w-11 shrink-0 items-center justify-center rounded-sm ${C.text40} ${C.hoverText80} ${C.hoverBgMedium} transition-colors disabled:opacity-40`}
        aria-label={`${label}フィルタを適用`}
      >
        <Check className={ICON.smXs} />
      </button>
    </div>
  );
});
