import { SelectItem } from "@/components/ui/select";

export const DAY_OF_WEEK_LABELS: Record<number, string> = {
  0: "日",
  1: "月",
  2: "火",
  3: "水",
  4: "木",
  5: "金",
  6: "土",
};

export const TIME_OPTIONS: string[] = [];
for (let h = 0; h < 24; h++) {
  for (const m of [0, 15, 30, 45]) {
    TIME_OPTIONS.push(`${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`);
  }
}

export const TIME_SELECT_ITEMS = TIME_OPTIONS.map((time) => (
  <SelectItem key={time} value={time}>{time}</SelectItem>
));
