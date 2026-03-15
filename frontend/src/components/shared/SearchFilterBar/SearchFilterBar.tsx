import { Search } from "lucide-react";
import { ReactNode } from "react";
import { STYLE } from "@/lib/design-tokens";

interface SearchFilterBarProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  placeholder?: string;
  children?: ReactNode; // Additional filters
  className?: string;
  count?: number; // Result count
}

export function SearchFilterBar({
  searchTerm,
  onSearchChange,
  placeholder = "検索...",
  children,
  className = "",
  count,
}: SearchFilterBarProps) {
  return (
    <div className={`flex items-center gap-3 ${className}`}>
      {count !== undefined ? (
        <span className={STYLE.searchCount}>
          {count} 件
        </span>
      ) : null}
      {children}
      <div className="relative flex-1 max-w-md">
        <Search className={STYLE.searchIcon} />
        <input
          type="text"
          placeholder={placeholder}
          value={searchTerm}
          onChange={(e) => onSearchChange(e.target.value)}
          className={STYLE.searchInput}
        />
      </div>
    </div>
  );
}
