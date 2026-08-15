import { memo } from "react";
import { C, LAYOUT } from "@/lib/design-tokens";
import { FormFieldError } from "@/components/shared/FormFieldError";

interface SidePeekTitleInputProps {
  id?: string;
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  autoFocus?: boolean;
  /** Called when the user presses Enter in the title field */
  onSave?: () => void;
  /** BUG-083: inline error message displayed below the title field */
  error?: string;
  /**
   * Maximum number of characters allowed in the title field.
   * BUG-379: デフォルト 255 文字。マスタ名称の DB 上限と整合させる。
   */
  maxLength?: number;
}

export const SidePeekTitleInput = memo(function SidePeekTitleInput({
  id,
  value,
  onChange,
  placeholder = "無題",
  autoFocus = true,
  onSave,
  error,
  maxLength = 255,
}: SidePeekTitleInputProps) {
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      if (onSave) {
        // onSave が渡されている場合は手動で保存を呼び出し、
        // form のデフォルト送信（改行等）を防ぐ
        e.preventDefault();
        onSave();
      }
      // onSave が未定義の場合は <form action={fn}> のネイティブ送信に委ねる
    }
  };

  return (
    <div className="pb-1 mb-4">
      <input
        id={id}
        type="text"
        className={`w-full bg-transparent ${C.text} ${C.textPlaceholderFaint} outline-none border-none p-0`}
        style={{
          fontSize: LAYOUT.pageTitle.fontSize,
          fontWeight: LAYOUT.pageTitle.fontWeight,
          lineHeight: LAYOUT.pageTitle.lineHeight,
          letterSpacing: LAYOUT.pageTitle.letterSpacing,
        }}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        aria-label={placeholder}
        autoFocus={autoFocus}
        maxLength={maxLength}
        aria-invalid={!!error}
      />
      {error ? <FormFieldError message={error} /> : null}
    </div>
  );
});
