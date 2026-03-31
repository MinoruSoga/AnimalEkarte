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
}

export function SidePeekTitleInput({
  id,
  value,
  onChange,
  placeholder = "無題",
  autoFocus = true,
  onSave,
  error,
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
        className={`w-full bg-transparent ${C.text} placeholder:text-[rgba(55,53,47,0.15)] outline-none border-none p-0`}
        style={{
          fontSize: LAYOUT.pageTitle.fontSize,
          fontWeight: LAYOUT.pageTitle.fontWeight,
          lineHeight: LAYOUT.pageTitle.lineHeight,
        }}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        autoFocus={autoFocus}
        aria-invalid={!!error}
      />
      {error ? <FormFieldError message={error} /> : null}
    </div>
  );
}
