import { useMemo, useState } from "react";
import { Check, ChevronDown } from "lucide-react";

import { cn } from "@/lib/utils";
import { C, PALETTE, Z_CLASS } from "@/lib/design-tokens";
import { normalizedIncludes } from "@/lib/normalize-kana";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

export interface SearchableSelectOption {
  value: string;
  label: string;
  /** 検索対象に追加したい別名・かな等。label に加えてマッチさせる。 */
  keywords?: string[];
  /** 選択不可(非アクティブ等)。表示はするが選択させない。 */
  disabled?: boolean;
}

export interface SearchableSelectGroup {
  label: string;
  options: SearchableSelectOption[];
}

interface SearchableSelectProps {
  value: string;
  onValueChange: (value: string) => void;
  /** フラットな選択肢。groups と排他。 */
  options?: SearchableSelectOption[];
  /** グループ階層の選択肢。options と排他。 */
  groups?: SearchableSelectGroup[];
  placeholder?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  disabled?: boolean;
  /** トリガーに付与する className。 */
  className?: string;
  /** ポップオーバー内容(リスト)に付与する className。 */
  contentClassName?: string;
  triggerTestId?: string;
  /** トリガー要素の id。Label の htmlFor 連携用。 */
  id?: string;
  /** id/htmlFor を使わないfilter toolbar等でのトリガー名。 */
  ariaLabel?: string;
  ariaInvalid?: boolean;
  /** 近傍エラー等との関連付け。FormFieldError の id を渡す。 */
  ariaDescribedBy?: string;
}

function flattenOptions(
  options?: SearchableSelectOption[],
  groups?: SearchableSelectGroup[],
): SearchableSelectOption[] {
  if (groups) return groups.flatMap((g) => g.options);
  return options ?? [];
}

/**
 * 検索可能なセレクトボックス(Combobox)。
 * shadcn の Popover + Command(cmdk)で構成。候補が多い(10件以上)データ駆動
 * セレクトの置換用。Select と同じく value/onValueChange で制御する。
 *
 * cmdk の既定フィルタは CommandItem の value 属性で照合するため、value には
 * 一意の opt.value を渡しつつ keywords(label + 任意の別名)で検索一致させる。
 */
export function SearchableSelect({
  value,
  onValueChange,
  options,
  groups,
  placeholder = "選択してください",
  searchPlaceholder = "検索...",
  emptyMessage = "該当する候補が見つかりません。",
  disabled = false,
  className,
  contentClassName,
  triggerTestId,
  id,
  ariaLabel,
  ariaInvalid = false,
  ariaDescribedBy,
}: SearchableSelectProps) {
  const [open, setOpen] = useState(false);

  const selectedLabel = useMemo(() => {
    const all = flattenOptions(options, groups);
    return all.find((o) => o.value === value)?.label ?? "";
  }, [options, groups, value]);

  const handleSelect = (next: string) => {
    onValueChange(next);
    setOpen(false);
  };

  const renderItem = (opt: SearchableSelectOption, indentClassName?: string) => {
    const isSelected = opt.value === value;
    return (
      <CommandItem
        key={opt.value}
        value={opt.value}
        keywords={[opt.label, ...(opt.keywords ?? [])]}
        disabled={opt.disabled}
        onSelect={() => handleSelect(opt.value)}
        className={cn("cursor-pointer", indentClassName)}
      >
        <span className={cn("flex-1 text-sm", C.text)}>{opt.label}</span>
        {isSelected ? <Check className={cn("size-4", C.text)} /> : null}
      </CommandItem>
    );
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        type="button"
        role="combobox"
        id={id}
        aria-label={ariaLabel}
        aria-expanded={open}
        aria-invalid={ariaInvalid}
        aria-describedby={ariaDescribedBy}
        disabled={disabled}
        data-testid={triggerTestId}
        className={cn(
          "flex h-11 min-w-11 w-full items-center justify-between gap-2 rounded-xs border bg-white px-3 py-2 text-sm whitespace-nowrap transition-colors outline-none",
          C.borderMedium,
          `${PALETTE.hoverBgInput} focus:bg-white ${PALETTE.focusBorderLegacyAccent} ${PALETTE.focusRingActionPrimary}`,
          "disabled:cursor-not-allowed disabled:opacity-50",
          ariaInvalid && C.borderDanger,
          className,
        )}
      >
        <span className={cn("line-clamp-1 text-left", selectedLabel ? C.text : C.text40)}>
          {selectedLabel || placeholder}
        </span>
        <ChevronDown className={cn("size-4 shrink-0 opacity-50", C.text)} />
      </PopoverTrigger>
      <PopoverContent
        align="start"
        className={cn(`${Z_CLASS.overlay} w-[var(--radix-popover-trigger-width)] p-0`, contentClassName)}
      >
        <Command
          filter={(_value, search, keywords) => {
            const haystack = keywords?.join(" ") ?? "";
            return normalizedIncludes(haystack, search) ? 1 : 0;
          }}
        >
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList className="max-h-[280px]">
            <CommandEmpty className={cn("py-6 text-center text-sm", C.text60)}>
              {emptyMessage}
            </CommandEmpty>
            {groups
              ? groups.map((group) => (
                  <CommandGroup key={group.label} heading={group.label}>
                    {group.options.map((opt) => renderItem(opt, "pl-4"))}
                  </CommandGroup>
                ))
              : (options ?? []).map((opt) => renderItem(opt))}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
