import type { ReactNode } from "react";

import { Checkbox } from "@/components/ui/checkbox";
import { C, STYLE } from "@/lib/design-tokens";

export interface StaffCheckboxItem {
  id: string;
  name: string;
}

interface StaffCheckboxSectionProps<T extends StaffCheckboxItem> {
  title: string;
  icon: ReactNode;
  items: T[];
  checkedIdSet: Set<string>;
  isDisabledUntilSaved: boolean;
  disabledMessage: string;
  emptyMessage: string;
  onToggle: (id: string, checked: boolean) => void;
  renderLeading?: (item: T) => ReactNode;
}

export function StaffCheckboxSection<T extends StaffCheckboxItem>({
  title,
  icon,
  items,
  checkedIdSet,
  isDisabledUntilSaved,
  disabledMessage,
  emptyMessage,
  onToggle,
  renderLeading,
}: StaffCheckboxSectionProps<T>) {
  return (
    <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-2">
        {icon}
        <p className={`text-xs font-medium ${C.text50}`}>{title}</p>
      </div>

      {isDisabledUntilSaved ? (
        <p className={`text-xs ${C.text50} pl-0.5`}>{disabledMessage}</p>
      ) : items.length === 0 ? (
        <p className={`text-xs ${C.text50} pl-0.5`}>{emptyMessage}</p>
      ) : (
        <div className="space-y-0.5">
          {items.map((item) => (
            <label
              key={item.id}
              className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer ${C.hoverBgLight} transition-colors`}
            >
              <Checkbox
                checked={checkedIdSet.has(item.id)}
                onCheckedChange={(checked) => onToggle(item.id, checked === true)}
              />
              {renderLeading?.(item)}
              <span className="text-sm">{item.name}</span>
            </label>
          ))}
        </div>
      )}
    </div>
  );
}
