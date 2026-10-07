// External
import { memo } from "react";

// Internal
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { Textarea } from "@/components/ui/textarea";
import { H_STYLES } from "../lib/styles";

// Types
import type { LucideIcon } from "lucide-react";

interface HospitalizationNoteCardProps {
  id?: string;
  title: string;
  icon: LucideIcon;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export const HospitalizationNoteCard = memo(function HospitalizationNoteCard({
  id,
  title,
  icon: Icon,
  value,
  onChange,
  placeholder,
}: HospitalizationNoteCardProps) {
  return (
    <div className={STYLE.formCard}>
      <h2
        id={id ? `${id}-label` : undefined}
        className={`${H_STYLES.form.cardHeading} mb-3 flex items-center gap-2`}
      >
        <Icon className={`${ICON.action} ${C.text60}`} />
        {title}
      </h2>
      <Textarea
        id={id}
        aria-labelledby={id ? `${id}-label` : undefined}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={`min-h-50 resize-none ${C.bgWhite} ${C.borderMedium} ${C.focusVisibleRingActionPrimary}`}
      />
    </div>
  );
});
