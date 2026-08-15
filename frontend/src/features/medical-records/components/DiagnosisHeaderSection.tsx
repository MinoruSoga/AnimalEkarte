import type { ReactNode } from "react";
import { C } from "@/lib/design-tokens";
import { cn } from "@/components/ui/utils";

interface DiagnosisHeaderSectionProps {
  icon: ReactNode;
  title: string;
  controls?: ReactNode;
  children: ReactNode;
  className?: string;
}

/** FE8-3: DiagnosisHeader 3列共通の subgrid セクション（row1=タイトル / row2=補助 / row3=ボックス） */
export function DiagnosisHeaderSection({
  icon,
  title,
  controls,
  children,
  className,
}: DiagnosisHeaderSectionProps) {
  return (
    <section className={cn("grid grid-rows-subgrid row-span-3 min-h-0", className)}>
      <h3 className={`text-sm font-bold ${C.text} flex items-center gap-2 min-h-0`}>
        {icon}
        {title}
      </h3>
      <div className="min-h-0">{controls ?? null}</div>
      <div className="min-h-0 flex flex-col h-full">{children}</div>
    </section>
  );
}
