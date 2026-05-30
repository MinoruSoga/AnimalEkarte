import { useMemo, useState } from "react";
import { Check, ChevronDown, X } from "lucide-react";

import { cn } from "@/lib/utils";
import { C } from "@/lib/design-tokens";
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
  /**
   * groups 指定時のみ有効。ポップオーバー上部にカテゴリ(グループ)チップを表示し、
   * クリックで該当グループのみに絞り込む。チップ表示時はポップオーバーを拡幅する。
   */
  groupFilter?: boolean;
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
  ariaInvalid?: boolean;
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
  groupFilter = false,
  placeholder = "選択してください",
  searchPlaceholder = "検索...",
  emptyMessage = "該当する候補が見つかりません。",
  disabled = false,
  className,
  contentClassName,
  triggerTestId,
  id,
  ariaInvalid = false,
}: SearchableSelectProps) {
  const [open, setOpen] = useState(false);
  const [activeGroup, setActiveGroup] = useState<string | null>(null);
  const showChips = groupFilter && groups !== undefined && groups.length > 0;

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
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) setActiveGroup(null);
      }}
    >
      <PopoverTrigger
        type="button"
        role="combobox"
        id={id}
        aria-expanded={open}
        aria-invalid={ariaInvalid}
        disabled={disabled}
        data-testid={triggerTestId}
        className={cn(
          "flex h-10 w-full items-center justify-between gap-2 rounded-md border bg-white px-3 py-2 text-sm whitespace-nowrap transition-colors outline-none",
          C.borderMedium,
          "hover:bg-[rgba(242,241,238,0.5)] focus:bg-white focus:border-[rgba(35,131,226,0.57)] focus:shadow-[0_0_0_1px_rgba(35,131,226,0.35)]",
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
        className={cn(
          "z-[9999] p-0",
          showChips
            ? "w-[360px] max-w-[calc(100vw-2rem)]"
            : "w-[var(--radix-popover-trigger-width)]",
          contentClassName,
        )}
      >
        <Command
          filter={(_value, search, keywords) => {
            const haystack = (keywords?.join(" ") ?? "").toLowerCase();
            return haystack.includes(search.toLowerCase()) ? 1 : 0;
          }}
        >
          <CommandInput placeholder={searchPlaceholder} />
          {showChips ? (
            <div
              className={cn(
                "flex items-center gap-1.5 overflow-x-auto border-b px-2 py-2 [&::-webkit-scrollbar]:hidden",
                C.borderMedium,
              )}
              style={{ scrollbarWidth: "none" }}
            >
              {activeGroup ? (
                <button
                  type="button"
                  onClick={() => setActiveGroup(null)}
                  className={cn(
                    "flex h-7 shrink-0 items-center gap-1 rounded-md px-2 text-xs whitespace-nowrap",
                    C.text60,
                    "hover:bg-[rgba(242,241,238,0.5)]",
                  )}
                >
                  <X className="size-3" />
                  解除
                </button>
              ) : null}
              {groups?.map((group) => {
                const isActive = activeGroup === group.label;
                return (
                  <button
                    key={group.label}
                    type="button"
                    onClick={() => setActiveGroup(isActive ? null : group.label)}
                    className={cn(
                      "h-7 shrink-0 rounded-md border px-2.5 text-xs whitespace-nowrap transition-colors",
                      isActive
                        ? cn(C.bgAccent, "border-transparent text-white")
                        : cn("bg-white", C.text, C.borderMedium, "hover:bg-[rgba(242,241,238,0.5)]"),
                    )}
                  >
                    {group.label}
                  </button>
                );
              })}
            </div>
          ) : null}
          <CommandList className="max-h-[280px]">
            <CommandEmpty className={cn("py-6 text-center text-sm", C.text60)}>
              {emptyMessage}
            </CommandEmpty>
            {groups
              ? groups
                  .filter((group) => activeGroup === null || group.label === activeGroup)
                  .map((group) => (
                    <CommandGroup
                      key={group.label}
                      heading={activeGroup === null ? group.label : undefined}
                    >
                      {group.options.map((opt) =>
                        renderItem(opt, activeGroup === null ? "pl-4" : undefined),
                      )}
                    </CommandGroup>
                  ))
              : (options ?? []).map((opt) => renderItem(opt))}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
