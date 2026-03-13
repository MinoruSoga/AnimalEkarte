import { Search, X } from "lucide-react";
import { Input } from "@/components/ui/input";
import { cn } from "@/components/ui/utils";
import { STYLE } from "@/lib/design-tokens";

interface SearchBoxProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

export function SearchBox({
  value,
  onChange,
  placeholder = "検索...",
  className,
}: SearchBoxProps) {
  return (
    <div className={cn("relative", className)}>
      <Search className={`${STYLE.searchIcon} pointer-events-none`} />
      <Input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={cn(
          "pl-8 text-sm",
          value ? "pr-8" : "pr-3",
        )}
      />
      {value && (
        <button
          type="button"
          onClick={() => onChange("")}
          className="absolute right-2.5 top-1/2 -translate-y-1/2 flex items-center justify-center size-4 rounded-full text-[#37352F]/40 hover:text-[#37352F]/70 hover:bg-[rgba(55,53,47,0.08)] transition-colors"
          aria-label="クリア"
        >
          <X className="size-3" />
        </button>
      )}
    </div>
  );
}
