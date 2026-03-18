import { Badge } from "@/components/ui/badge";

interface StatusBadgeProps {
  children: React.ReactNode;
  className?: string;
  colorClass?: string;
}

export function StatusBadge({ children, className = "", colorClass = "" }: StatusBadgeProps) {
  return (
    <Badge
      variant="outline"
      className={`text-base px-2 h-7 font-normal border ${colorClass} ${className}`}
    >
      {children}
    </Badge>
  );
}
