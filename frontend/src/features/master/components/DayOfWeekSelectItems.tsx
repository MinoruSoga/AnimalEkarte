import { SelectItem } from "@/components/ui/select";
import { DAY_OF_WEEK_LABELS } from "@/constants/day-of-week";

/**
 * 曜日選択用 SelectItem 一覧（value="0"〜"6", 0=日曜〜6=土曜）。
 * FE6-7: ReservationTypeAvailableSlotsSection / ReservationTypeUnavailableTimesSection
 * に同一内容が2複製されていたため共有定数化。
 */
export const DAY_OF_WEEK_SELECT_ITEMS = Object.entries(DAY_OF_WEEK_LABELS).map(([value, label]) => (
  <SelectItem key={value} value={value}>
    {label}曜日
  </SelectItem>
));
