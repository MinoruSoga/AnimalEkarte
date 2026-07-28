import { useId, type ReactNode } from "react";

import { C } from "@/lib/design-tokens";

interface ClinicalBriefingPanelProps {
  title: string;
  description?: string;
  count?: string;
  areaClassName: string;
  bodyClassName?: string;
  bodyTestId?: string;
  children: ReactNode;
}

/** 固定1画面の各区画。見出しを残し、内容だけを個別スクロールさせる。 */
export function ClinicalBriefingPanel({
  title,
  description,
  count,
  areaClassName,
  bodyClassName = "p-2",
  bodyTestId,
  children,
}: ClinicalBriefingPanelProps) {
  const headingId = useId();

  return (
    <section
      aria-labelledby={headingId}
      className={`${areaClassName} flex min-h-0 min-w-0 flex-col overflow-hidden rounded-sm border shadow-level1 ${C.borderLight} ${C.bgWhite}`}
    >
      <header
        className={`flex shrink-0 items-start justify-between gap-2 border-b px-2 py-1 ${C.borderLight} ${C.bgPage}`}
      >
        <div className="flex min-w-0 items-baseline gap-2">
          <h2 id={headingId} className={`text-xl leading-snug font-semibold ${C.text}`}>
            {title}
          </h2>
          {description ? (
            <p className={`truncate text-xs leading-snug ${C.text50} max-[840px]:hidden`}>
              {description}
            </p>
          ) : null}
        </div>
        {count ? (
          <span className={`shrink-0 text-2xs font-semibold tabular-nums ${C.text60}`}>
            {count}
          </span>
        ) : null}
      </header>
      <div
        data-owner-report-scroll=""
        data-testid={bodyTestId}
        className={`min-h-0 flex-1 overflow-auto overscroll-contain ${bodyClassName}`}
        tabIndex={0}
      >
        {children}
      </div>
    </section>
  );
}
