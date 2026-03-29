import { C, LAYOUT } from "@/lib/design-tokens";

interface SidePeekTitleInputProps {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  autoFocus?: boolean;
  /** Called when the user presses Enter in the title field */
  onSave?: () => void;
}

export function SidePeekTitleInput({
  value,
  onChange,
  placeholder = "無題",
  autoFocus = true,
  onSave,
}: SidePeekTitleInputProps) {
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      onSave?.();
    }
  };

  return (
    <div className="pb-1 mb-4">
      <input
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
      />
    </div>
  );
}
