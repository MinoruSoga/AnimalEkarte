import { C, PALETTE } from "@/lib/design-tokens";
import type { LineSendType } from "../api/send-line-message";

const TYPE_OPTIONS: { value: LineSendType; label: string }[] = [
  { value: "text", label: "テキスト送信" },
  { value: "pdf_url", label: "PDF送付" },
  { value: "image_url", label: "画像送付" },
];

interface LineSendTypeSelectorProps {
  value: LineSendType;
  onChange: (v: LineSendType) => void;
  disabled?: boolean;
}

export function LineSendTypeSelector({
  value,
  onChange,
  disabled = false,
}: LineSendTypeSelectorProps) {
  return (
    <div className="flex gap-4">
      {TYPE_OPTIONS.map((opt) => (
        <label
          key={opt.value}
          className={`flex items-center gap-2 cursor-pointer text-sm ${C.text70} select-none`}
        >
          <input
            type="radio"
            name="line_send_type"
            value={opt.value}
            checked={value === opt.value}
            onChange={() => onChange(opt.value)}
            disabled={disabled}
            style={{ accentColor: PALETTE.actionPrimary }}
          />
          {opt.label}
        </label>
      ))}
    </div>
  );
}
