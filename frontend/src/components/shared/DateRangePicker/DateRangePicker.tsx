import { ICON } from "@/lib/design-tokens";
import { useCallback, useState } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { CalendarIcon } from "lucide-react";
import type { DateRange } from "react-day-picker";

import { cn } from "@/components/ui/utils";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

interface DateRangePickerProps {
  value: DateRange | undefined;
  onChange: (range: DateRange | undefined) => void;
  className?: string;
  placeholder?: string;
}

export function DateRangePicker({
  value,
  onChange,
  className,
  placeholder = "期間を選択",
}: DateRangePickerProps) {
  const [open, setOpen] = useState(false);

  const handleSelect = useCallback(
    (range: DateRange | undefined) => {
      onChange(range);
      // 両方選択されたら閉じる
      if (range?.from && range?.to) {
        setOpen(false);
      }
    },
    [onChange],
  );

  const handleClear = useCallback(() => {
    onChange(undefined);
    setOpen(false);
  }, [onChange]);

  const label = value?.from
    ? value.to
      ? `${format(value.from, "yyyy/MM/dd")} - ${format(value.to, "yyyy/MM/dd")}`
      : `${format(value.from, "yyyy/MM/dd")} -`
    : placeholder;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            "h-9 justify-start text-left text-sm font-normal bg-white",
            !value?.from ? "text-muted-foreground" : "",
            className,
          )}
        >
          <CalendarIcon className={`mr-2 ${ICON.action}`} />
          {label}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="range"
          selected={value}
          onSelect={handleSelect}
          numberOfMonths={2}
          locale={ja}
        />
        {value?.from ? (
          <div className="border-t p-2 flex justify-end">
            <Button
              variant="ghost"
              size="sm"
              className="text-xs"
              onClick={handleClear}
            >
              クリア
            </Button>
          </div>
        ) : null}
      </PopoverContent>
    </Popover>
  );
}
