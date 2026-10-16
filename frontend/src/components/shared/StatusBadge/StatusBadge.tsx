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
      className={`text-2xs px-2 h-7 font-semibold border ${colorClass} ${className}`}
    >
      {children}
    </Badge>
  );
}
