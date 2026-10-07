// External
import { memo } from "react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { Textarea } from "@/components/ui/textarea";

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
    <div className={`${C.bgWhite} rounded-lg border ${C.borderMedium} p-6`}>
      <h2
        id={id ? `${id}-label` : undefined}
        className={`text-heading-3 font-bold mb-3 flex items-center gap-2 ${C.text}`}
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
        className={`min-h-[200px] resize-none ${C.bgWhite} ${C.borderMedium} ${C.focusVisibleRingActionPrimary}`}
      />
    </div>
  );
});
