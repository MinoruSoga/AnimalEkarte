import { ReactNode } from 'react';
import { Tabs as ShadcnTabs, TabsList as ShadcnTabsList, TabsTrigger as ShadcnTabsTrigger, TabsContent as ShadcnTabsContent } from '@/components/ui/tabs';
import { C } from '@/lib/design-tokens';

/**
 * Unified Tab Interface
 * Supports both shadcn/ui Tabs and custom button-based implementations
 * Applies consistent Notion-like design tokens across all variations
 */

interface TabItem {
  value: string;
  label: ReactNode;
}

type TabVariant = 'shadcn' | 'button'; // 'button' for custom bottom-line style

interface UnifiedTabsProps {
  items: readonly TabItem[];
  value: string;
  onValueChange: (value: string) => void;
  children?: ReactNode;
  variant?: TabVariant;
  className?: string;
  listClassName?: string;
  triggerClassName?: string;
}

/**
 * UnifiedTabs: Single entry point for all tab UI patterns in the application
 *
 * Supports:
 * - shadcn/ui Tabs (via wrapper) — used by Aggregation, Hospitalization
 * - Radix TabsPrimitive direct — used by Master settings
 * - Custom button-based (bottom-line indicator) — used by MedicalRecordForm
 *
 * All variants apply consistent design tokens:
 * - TabsList: C.bgLight (subtle gray background)
 * - TabsTrigger/Button active: C.text + C.dataActiveBorderB (1px Notion Primary bottom border)
 * - Inactive: C.text50 or C.text60
 */
export function UnifiedTabs({
  items,
  value,
  onValueChange,
  children,
  variant = 'shadcn',
  className = '',
  listClassName = '',
  triggerClassName = '',
}: UnifiedTabsProps) {
  if (variant === 'button') {
    // Custom button-based tabs with bottom-line indicator (for MedicalRecordForm style)
    return (
      <div className={className}>
        <div className={`flex border-b ${C.borderLight}`}>
          {items.map((item) => (
            <button
              key={item.value}
              type="button"
              onClick={() => onValueChange(item.value)}
              className={`flex items-center px-4 h-11 text-base font-medium transition-colors relative whitespace-nowrap ${
                value === item.value
                  ? `${C.text} ${C.bgLight}`
                  : `${C.text60} hover:${C.text}`
              } ${triggerClassName}`}
            >
              {item.label}
              {value === item.value ? (
                <div className={`absolute bottom-0 left-0 w-full h-[2px] ${C.bgPrimary}`} />
              ) : null}
            </button>
          ))}
        </div>
        {children}
      </div>
    );
  }

  // Default: shadcn/ui Tabs (unified wrapper around Radix TabsPrimitive)
  // Applies consistent design tokens to TabsList and TabsTrigger
  return (
    <ShadcnTabs value={value} onValueChange={onValueChange} className={className}>
      <ShadcnTabsList className={`${C.bgLight} ${listClassName}`}>
        {items.map((item) => (
          <ShadcnTabsTrigger
            key={item.value}
            value={item.value}
            className={`rounded-[10px] ${C.dataActiveBorderB} ${C.dataActiveText} ${triggerClassName}`}
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

export default UnifiedTabs;
