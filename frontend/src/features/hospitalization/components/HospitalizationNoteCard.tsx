// Internal
import { Textarea } from "@/components/ui/textarea";

// Relative
import { H_STYLES } from "../styles";

// Types
import type { LucideIcon } from "lucide-react";

interface HospitalizationNoteCardProps {
  title: string;
  icon: LucideIcon;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export const HospitalizationNoteCard = ({ title, icon: Icon, value, onChange, placeholder }: HospitalizationNoteCardProps) => {
  return (
    <div className={`bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] ${H_STYLES.padding.box} h-full`}>
      <h2 className={`${H_STYLES.text.base} font-bold mb-3 flex items-center gap-2 text-[#37352F]`}>
        <Icon className="h-4 w-4 text-[#37352F]/60" />
        {title}
      </h2>
      <Textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={`min-h-[200px] ${H_STYLES.text.base} resize-none bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]`}
      />
    </div>
  );
};
