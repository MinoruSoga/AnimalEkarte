import { ReactNode } from "react";
import {
  Tabs as ShadcnTabs,
  TabsList as ShadcnTabsList,
  TabsTrigger as ShadcnTabsTrigger,
  TabsContent as ShadcnTabsContent,
} from "@/components/ui/tabs";
import { C } from "@/lib/design-tokens";

/**
 * UnifiedTabs — Notion-like bottom-line tab UI.
 * Single shadcn-based implementation for every tab UI in the application.
 *
 * The defaults below cancel out the pill-style base classes from
 * `components/ui/tabs.tsx` (rounded-xl / p-[3px] / data-[state=active]:bg-card
 * / flex-1) so every tab UI renders as a Notion-like bottom-line.
 *
 * Pages should NOT pass listClassName / triggerClassName for cosmetic tweaks —
 * they exist only for layout extensions (e.g. width). Visual style is
 * centralised here.
 */

interface TabItem {
  value: string;
  label: ReactNode;
}

interface UnifiedTabsProps {
  items: readonly TabItem[];
  value: string;
  onValueChange: (value: string) => void;
  children?: ReactNode;
  className?: string;
  listClassName?: string;
  triggerClassName?: string;
}

// Notion-like bottom-line defaults. Each token is chosen to defeat the
// corresponding pill default from components/ui/tabs.tsx via tailwind-merge.
// Background is transparent — only a thin hair-line under the strip remains.
// max-w-full + overflow-x-auto: all tabs stay reachable at narrow widths (BUG-458).
// Triggers keep min-h-11 (44px) via h-11 / h-9 stack; do not shrink hit targets.
const DEFAULT_LIST_CLASS = `flex h-11 max-w-full w-full items-center justify-start overflow-x-auto rounded-none p-0 border-b ${C.borderLight} bg-transparent gap-0`;
const DEFAULT_TRIGGER_CLASS = `h-9 min-h-11 flex-none shrink-0 rounded-none border-0 border-b-2 border-b-transparent bg-transparent px-4 ${C.text60} data-[state=active]:bg-transparent data-[state=active]:shadow-none ${C.dataActiveBorderB} ${C.dataActiveText}`;

export function UnifiedTabs({
  items,
  value,
  onValueChange,
  children,
  className = "",
  listClassName = "",
  triggerClassName = "",
}: UnifiedTabsProps) {
  return (
    <ShadcnTabs value={value} onValueChange={onValueChange} className={`min-w-0 ${className}`}>
      <ShadcnTabsList className={`${DEFAULT_LIST_CLASS} ${listClassName}`}>
        {items.map((item) => (
          <ShadcnTabsTrigger
            key={item.value}
            value={item.value}
            className={`${DEFAULT_TRIGGER_CLASS} ${triggerClassName}`}
          >
            {item.label}
          </ShadcnTabsTrigger>
        ))}
      </ShadcnTabsList>
      {children}
    </ShadcnTabs>
  );
}

/**
 * UnifiedTabsContent: Wrapper for tab content panels
 * Replaces direct TabsContent usage to maintain unified layer
 */
export function UnifiedTabsContent({
  value,
  children,
  className,
}: {
  value: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <ShadcnTabsContent value={value} className={className}>
      {children}
    </ShadcnTabsContent>
  );
}

/**
 * UnifiedTabsRoot: Headless Tabs Root provider.
 * Use when TabsList and TabsContent must live in different branches of the tree
 * (e.g. sticky header + scrollable content area). Pair with `UnifiedTabsList`
 * for the styled trigger row, and `UnifiedTabsContent` for content panels.
 */
export function UnifiedTabsRoot({
  value,
  onValueChange,
  children,
  className,
  ariaBusy = false,
}: {
  value: string;
  onValueChange: (value: string) => void;
  children: ReactNode;
  className?: string;
  ariaBusy?: boolean;
}) {
  return (
    <ShadcnTabs
      value={value}
      onValueChange={onValueChange}
      className={className}
      aria-busy={ariaBusy}
    >
      {children}
    </ShadcnTabs>
  );
}

/**
 * UnifiedTabsList: Standalone styled trigger row. Must be rendered inside
 * `UnifiedTabsRoot`. Shares the same Notion-like bottom-line defaults as
 * `UnifiedTabs`.
 */
export function UnifiedTabsList({
  items,
  listClassName = "",
  triggerClassName = "",
}: {
  items: readonly TabItem[];
  listClassName?: string;
  triggerClassName?: string;
}) {
  return (
    <ShadcnTabsList className={`${DEFAULT_LIST_CLASS} ${listClassName}`}>
      {items.map((item) => (
        <ShadcnTabsTrigger
          key={item.value}
          value={item.value}
          className={`${DEFAULT_TRIGGER_CLASS} ${triggerClassName}`}
        >
          {item.label}
        </ShadcnTabsTrigger>
      ))}
    </ShadcnTabsList>
  );
}
