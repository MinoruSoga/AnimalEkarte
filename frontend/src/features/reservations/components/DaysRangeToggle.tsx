import { C } from "@/lib/design-tokens";

interface DaysRangeToggleProps {
  days: 5 | 7;
  onChange: (days: 5 | 7) => void;
}

const OPTIONS: { value: 5 | 7; label: string }[] = [
  { value: 5, label: "5日" },
  { value: 7, label: "7日" },
];

export function DaysRangeToggle({ days, onChange }: DaysRangeToggleProps) {
  return (
    <div
      className={`flex items-center ${C.bgWhite} rounded-md border ${C.borderMedium} p-0.5`}
      role="group"
      aria-label="表示日数切替"
    >
      {OPTIONS.map((opt) => {
        const isActive = days === opt.value;
        return (
          <button
            key={opt.value}
            type="button"
            data-testid={`days-toggle-${opt.value}`}
            aria-pressed={isActive}
            onClick={() => onChange(opt.value)}
            className={`h-11 min-w-11 px-3 rounded text-base font-medium transition-colors ${
              isActive
                ? `${C.bgActionPrimary} ${C.textOnActionPrimary}`
                : `${C.text60} ${C.hoverText}`
            }`}
          >
            {opt.label}
          </button>
        );
      })}
    </div>
  );
}
