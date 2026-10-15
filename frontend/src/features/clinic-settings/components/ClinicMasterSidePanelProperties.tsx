import { memo, useMemo, type ReactNode } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";

import { FieldHelp } from "@/components/shared/FieldHelp";
import { PropertyToggleButton } from "@/components/shared/SidePeek";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

import {
  DOCUMENT_SECTION_KEYS,
  DOCUMENT_SECTION_LABELS,
  type DocumentSectionKey,
} from "@/config/accounting-document-sections";

import { C, ICON, LAYOUT, STYLE } from "@/lib/design-tokens";

const PROP_INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 min-h-11 rounded-xxs ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder} focus-visible:ring-2 ${C.focusRingAccent40}`;

export const PropertyRow = memo(function PropertyRow({
  label,
  description,
  children,
}: {
  label: string;
  /** 項目の説明文。指定すると行右端に ⓘ ツールチップを表示する */
  description?: string;
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
      {description ? <FieldHelp label={label} content={description} /> : null}
    </div>
  );
});

/** Renders a labeled controlled text input and emits each edited string through `onChange`. */
export function ClinicTextProperty({
  label,
  value,
  onChange,
  placeholder,
  type = "text",
  description,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  type?: string;
  description?: string;
}) {
  return (
    <PropertyRow label={label} description={description}>
      <Input
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

/** Edits a displayed percentage from 0 to 100 and emits the fractional rate to `onChange`. */
export function ClinicTaxRateProperty({
  label,
  value,
  onChange,
  description,
}: {
  label: string;
  value: number;
  onChange: (value: number) => void;
  description?: string;
}) {
  return (
    <PropertyRow label={label} description={description}>
      <div className="flex items-center gap-1.5">
        <NumberInput
          min={0}
          max={100}
          step={1}
          aria-label={label}
          className={`${PROP_INPUT_CLASS} w-20`}
          value={Math.round(value * 100)}
          onChange={(nextValue) => onChange(Number(nextValue) / 100)}
        />
        <span className={`text-sm ${C.text50}`}>%</span>
      </div>
    </PropertyRow>
  );
}

/** Renders a named boolean toggle whose pressed state reflects `value`; changes are emitted upward. */
export function ClinicBooleanProperty({
  label,
  value,
  onChange,
  description,
}: {
  label: string;
  value: boolean;
  onChange: (value: boolean) => void;
  description?: string;
}) {
  return (
    <PropertyRow label={label} description={description}>
      <PropertyToggleButton
        label={label}
        pressed={value}
        onPressedChange={onChange}
        className="px-0.5"
      >
        <StatusPill status={value ? "active" : "inactive"} />
      </PropertyToggleButton>
    </PropertyRow>
  );
}

/** Renders a controlled, labeled textarea with a 500-character limit. */
export function ClinicTextareaProperty({
  label,
  value,
  onChange,
  placeholder,
  description,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  description?: string;
}) {
  return (
    <PropertyRow label={label} description={description}>
      <Textarea
        aria-label={label}
        className={`${PROP_INPUT_CLASS} min-h-16 resize-y field-sizing-fixed`}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        maxLength={500}
      />
    </PropertyRow>
  );
}

// #190: セクション表示順の変更UI
/**
 * Edits the canonical accounting-document section order.
 * Unknown keys are removed and missing canonical keys appended before move results are emitted.
 */
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
      <div className={`flex items-center gap-1 text-sm ${C.text65} select-none`}>
        セクション順
        <FieldHelp
          label="セクション順"
          content="明細兼領収書内の各セクションの表示順です。上下ボタンで入れ替えます。"
        />
      </div>
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
                className={`min-h-11 min-w-11 flex items-center justify-center rounded ${C.hoverBgLight} disabled:opacity-30 cursor-pointer disabled:cursor-default`}
                aria-label="上に移動"
              >
                <ChevronUp className={ICON.smXs} aria-hidden="true" />
              </button>
              <button
                type="button"
                onClick={() => move(i, 1)}
                disabled={i === effectiveOrder.length - 1}
                className={`min-h-11 min-w-11 flex items-center justify-center rounded ${C.hoverBgLight} disabled:opacity-30 cursor-pointer disabled:cursor-default`}
                aria-label="下に移動"
              >
                <ChevronDown className={ICON.smXs} aria-hidden="true" />
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

/** Displays an active or inactive status label; it is not an interactive control. */
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
