import { memo, useMemo, type ReactNode } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";

import {
  DOCUMENT_SECTION_KEYS,
  DOCUMENT_SECTION_LABELS,
  type DocumentSectionKey,
} from "@/config/accounting-document-sections";

import { C, LAYOUT, STYLE } from "@/lib/design-tokens";

const PROP_INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-xxs ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder} focus-visible:ring-2 ${C.focusRingAccent40}`;

export const PropertyRow = memo(function PropertyRow({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  return (
    <div className={STYLE.propertyRow}>
      <div
        className={`${LAYOUT.propertyRow.labelW} shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}
      >
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
});

export function ClinicTextProperty({
  label,
  value,
  onChange,
  placeholder,
  type = "text",
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  type?: string;
}) {
  return (
    <PropertyRow label={label}>
      <input
        type={type}
        aria-label={label}
        className={PROP_INPUT_CLASS}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
      />
    </PropertyRow>
  );
}

export function ClinicTaxRateProperty({
  label,
  value,
  onChange,
}: {
  label: string;
  value: number;
  onChange: (value: number) => void;
}) {
  return (
    <PropertyRow label={label}>
      <div className="flex items-center gap-1.5">
        <input
          type="number"
          min={0}
          max={100}
          step={1}
          aria-label={label}
          className={`${PROP_INPUT_CLASS} w-20`}
          value={Math.round(value * 100)}
          onChange={(e) => onChange(Number(e.target.value) / 100)}
        />
        <span className={`text-sm ${C.text50}`}>%</span>
      </div>
    </PropertyRow>
  );
}

export function ClinicBooleanProperty({
  label,
  value,
  onChange,
}: {
  label: string;
  value: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <PropertyRow label={label}>
      <button
        type="button"
        onClick={() => onChange(!value)}
        className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
      >
        <StatusPill status={value ? "active" : "inactive"} />
      </button>
    </PropertyRow>
  );
}

export function ClinicTextareaProperty({
  label,
  value,
  onChange,
  placeholder,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
}) {
  return (
    <PropertyRow label={label}>
      <textarea
        className={`${PROP_INPUT_CLASS} min-h-16 resize-y`}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        maxLength={500}
      />
    </PropertyRow>
  );
}

// #190: セクション表示順の変更UI
export function SectionOrderProperty({
  order,
  onChange,
}: {
  order: string[];
  onChange: (order: string[]) => void;
}) {
  // 有効なキーのみ抽出し、漏れたキーを末尾に補完（AccountingDocument と同ロジック）
  const effectiveOrder = useMemo<DocumentSectionKey[]>(() => {
    const validKeys = order.filter((k): k is DocumentSectionKey =>
      (DOCUMENT_SECTION_KEYS as readonly string[]).includes(k),
    );
    if (validKeys.length === 0) return [...DOCUMENT_SECTION_KEYS];
    const missing = DOCUMENT_SECTION_KEYS.filter((k) => !validKeys.includes(k));
    return [...validKeys, ...missing];
  }, [order]);

  function move(index: number, direction: -1 | 1) {
    const next = [...effectiveOrder];
    const target = index + direction;
    if (target < 0 || target >= next.length) return;
    [next[index], next[target]] = [next[target], next[index]];
    onChange(next);
  }

  return (
    <div className={`${STYLE.propertyRow} flex-col items-start gap-1`}>
      <div className={`text-sm ${C.text65} select-none`}>セクション順</div>
      <div className="w-full space-y-0.5 pl-1">
        {effectiveOrder.map((key, i) => (
          <div
            key={key}
            className={`flex items-center justify-between text-xs ${C.text} py-0.5 px-1 rounded ${C.hoverBgLight}`}
          >
            <span>{DOCUMENT_SECTION_LABELS[key]}</span>
            <div className="flex items-center gap-0.5">
              <button
                type="button"
                onClick={() => move(i, -1)}
                disabled={i === 0}
                className={`p-0.5 rounded ${C.hoverBgLight} disabled:opacity-30 cursor-pointer disabled:cursor-default`}
                aria-label="上に移動"
              >
                <ChevronUp className="size-3" />
              </button>
              <button
                type="button"
                onClick={() => move(i, 1)}
                disabled={i === effectiveOrder.length - 1}
                className={`p-0.5 rounded ${C.hoverBgLight} disabled:opacity-30 cursor-pointer disabled:cursor-default`}
                aria-label="下に移動"
              >
                <ChevronDown className="size-3" />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

const STATUS_CONFIG = {
  active: {
    dot: `${C.bgBrandDot}`,
    label: "有効",
    bg: `${C.bgBrandLight}`,
    text: `${C.textBrandDark}`,
  },
  inactive: {
    dot: C.bgPrimary10,
    label: "無効",
    bg: `${C.bgInactive}`,
    text: C.text60,
  },
} as const;

export function StatusPill({ status }: { status: "active" | "inactive" }) {
  const cfg = STATUS_CONFIG[status];
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-sm text-xs ${cfg.bg} ${cfg.text}`}
    >
      <span className={`size-[7px] rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  );
}
