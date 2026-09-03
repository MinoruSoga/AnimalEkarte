import { useRef } from "react";
import { Search, X } from "lucide-react";

import { Input } from "@/components/ui/input";
import { C, ICON } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";

interface ClearableSearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  inputClassName?: string;
  clearAriaLabel?: string;
}

/**
 * 検索アイコン + Input + 条件付きクリアボタンの共通パターン。
 * MasterSelectModal / TreatmentSearchDialog / ReservationTypePickerDialog から抽出。
 */
export function ClearableSearchInput({
  value,
  onChange,
  placeholder,
  className,
  inputClassName,
  clearAriaLabel = "検索をクリア",
}: ClearableSearchInputProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div className={cn("relative", className)}>
      <Search
        aria-hidden="true"
        className={cn("absolute left-2.5 top-1/2 -translate-y-1/2", ICON.action, C.text40)}
      />
      <Input
        ref={inputRef}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={cn("pl-9 h-11 text-sm bg-white", C.borderMedium, inputClassName)}
      />
      {value ? (
        <button
          type="button"
          onClick={() => {
            onChange("");
            // クリアするとボタン自体が DOM から消えるため、フォーカスを検索入力へ戻す
            // （そのまま放置すると focus が body に落ち、キーボード/SRユーザーがページ先頭からの
            // タブやり直しを強いられる）
            inputRef.current?.focus();
          }}
          aria-label={clearAriaLabel}
          className={cn("absolute right-2.5 top-1/2 -translate-y-1/2", C.text40, C.hoverText)}
        >
          <X aria-hidden="true" className={ICON.xs} />
        </button>
      ) : null}
    </div>
  );
}
