import { memo } from "react";

import { cn } from "@/components/ui/utils";
import { Textarea } from "@/components/ui/textarea";
import { C, STYLE } from "@/lib/design-tokens";

interface CharCountTextareaProps {
  value: string;
  onChange: (value: string) => void;
  maxLength?: number;
  placeholder?: string;
  className?: string;
  textareaClassName?: string;
  disabled?: boolean;
  id?: string;
  name?: string;
}

/**
 * Textarea with a character count indicator shown below.
 * Displays "X / maxLength 文字" and turns red when the limit is reached.
 */
export const CharCountTextarea = memo(function CharCountTextarea({
  value,
  onChange,
  maxLength = 500,
  placeholder,
  className,
  textareaClassName,
  disabled,
  id,
  name,
}: CharCountTextareaProps) {
  const count = value.length;
  const isOver = count >= maxLength;

  return (
    <div className={cn("flex flex-col gap-1", className)}>
      <Textarea
        id={id}
        name={name}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        maxLength={maxLength}
        placeholder={placeholder}
        disabled={disabled}
        className={cn(STYLE.textarea, "flex-1", `focus-visible:ring-2 ${C.focusRingAccent40}`, textareaClassName)}
      />
      <p
        className={cn(
          "text-xs text-right select-none",
          isOver ? C.danger : C.text40,
        )}
        aria-live="polite"
        aria-atomic="true"
      >
        {count} / {maxLength} 文字
      </p>
    </div>
  );
});
