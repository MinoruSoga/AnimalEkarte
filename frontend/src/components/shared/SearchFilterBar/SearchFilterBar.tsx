import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { ReactNode } from "react";

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
    <div className={`flex flex-col sm:flex-row items-center gap-4 bg-[#fafafa] p-4 rounded-lg border border-[#e5e5e5] shadow-sm ${className}`}>
      {/* Left side: Count & Filters */}
      {(children || count !== undefined) && (
        <div className="flex items-center gap-4 w-full sm:w-auto overflow-x-auto pb-1 sm:pb-0">
          {count !== undefined && (
            <div className="text-sm text-muted-foreground whitespace-nowrap px-2">
              {count} 件
            </div>
          )}
          {children}
        </div>
      )}

      {/* Right side: Search Input */}
      <div className={`relative flex-1 w-full ${children || count !== undefined ? "sm:w-64" : ""}`}>
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <Input
          placeholder={placeholder}
          value={searchTerm}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-10 bg-[#fafafa] border-[#e5e5e5] h-10 text-sm w-full"
        />
      </div>
    </div>
  );
}
