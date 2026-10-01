import { useState, type ReactNode, type KeyboardEvent } from "react";
import { ChevronDown } from "lucide-react";

import { C, ICON } from "@/lib/design-tokens";

/**
 * EMR-231: スタッフ向け内部メモ領域の折りたたみ開閉。
 * 来院者の前で項目が常時露出しないよう、普段は中立ラベルの行だけを出し、
 * 操作したときだけ内容を展開する。
 *
 * `role="button"` + Enter/Space + aria-expanded/aria-controls を使う。
 * ネイティブの <button> は `<fieldset disabled>`（詳細閲覧モード）内で
 * 無効化されてしまい閲覧側から開けなくなるため、ここでは使えない。
 */
interface StaffNoteDisclosureProps {
  /** 展開パネルの DOM id（aria-controls の参照先）。 */
  id: string;
  /** 先頭に置くマーク（呼び出し側が DangerBadge と同じ色付けで渡す）。 */
  icon: ReactNode;
  /** true のときスクリーンリーダーへ「設定あり」を補足し、内容が入っていることを色に依存せず伝える。 */
  flagged?: boolean;
  /**
   * false → true へ変わったタイミングで一度だけ自動展開する。
   * 折りたたみ内の項目にフィールドエラーが出たとき、エラーを見せるために使う。
   */
  autoOpen?: boolean;
  children: ReactNode;
}

export function StaffNoteDisclosure({
  id,
  icon,
  flagged = false,
  autoOpen = false,
  children,
}: StaffNoteDisclosureProps) {
  const [open, setOpen] = useState(false);
  const [prevAutoOpen, setPrevAutoOpen] = useState(autoOpen);

  // レンダー中の状態調整: autoOpen が false→true に変わった瞬間だけ展開する。
  // その後の手動開閉は妨げない（エラーが残っていても畳み直せる）。
  if (autoOpen !== prevAutoOpen) {
    setPrevAutoOpen(autoOpen);
    if (autoOpen) {
      setOpen(true);
    }
  }

  const toggle = () => setOpen((prev) => !prev);

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      toggle();
    }
  };

  return (
    <div className="space-y-1.5">
      <div
        role="button"
        tabIndex={0}
        aria-expanded={open}
        aria-controls={id}
        onClick={toggle}
        onKeyDown={handleKeyDown}
        className={`flex w-full cursor-pointer items-center gap-1.5 rounded px-1 -ml-1 py-1 text-sm ${C.text60} ${C.hoverBgPage} outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
      >
        {icon}
        <span>スタッフ向け特記</span>
        {flagged ? <span className="sr-only">設定あり</span> : null}
        <ChevronDown
          className={`${ICON.smXs} ml-auto ${C.text40} transition-transform ${open ? "rotate-180" : ""}`}
          aria-hidden="true"
        />
      </div>
      {open ? (
        <div id={id} className="space-y-1.5">
          {children}
        </div>
      ) : null}
    </div>
  );
}
