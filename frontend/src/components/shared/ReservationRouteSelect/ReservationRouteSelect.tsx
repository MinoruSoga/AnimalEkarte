import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { STYLE } from "@/lib/design-tokens";
import {
  RESERVATION_ROUTE_LABELS,
  RESERVATION_ROUTE_VALUES,
  type ReservationRoute,
} from "@/types/reservation-route";
import { useUpdateReservationRoute } from "./hooks/use-update-reservation-route";

interface ReservationRouteSelectProps {
  reservationId: string;
  value: ReservationRoute | null;
  disabled?: boolean;
}

export function ReservationRouteSelect({
  reservationId,
  value,
  disabled = false,
}: ReservationRouteSelectProps) {
  const { mutate, isPending } = useUpdateReservationRoute(reservationId);

  const handleValueChange = (selected: string) => {
    if (selected === "_clear") {
      mutate({ route: "" });
    } else {
      mutate({ route: selected as ReservationRoute });
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <label className={STYLE.formLabel}>予約経路</label>
      <Select
        value={value ?? ""}
        onValueChange={handleValueChange}
        disabled={disabled || isPending}
      >
        <SelectTrigger
          className="w-[160px] h-9 text-sm"
          aria-label="予約経路を選択"
          data-testid="reservation-route-trigger"
        >
          <SelectValue placeholder="未選択" />
        </SelectTrigger>
        <SelectContent>
          {RESERVATION_ROUTE_VALUES.map((v) => (
            <SelectItem key={v} value={v}>
              {RESERVATION_ROUTE_LABELS[v]}
            </SelectItem>
          ))}
          {value !== null ? <SelectItem value="_clear">クリア</SelectItem> : null}
        </SelectContent>
      </Select>
    </div>
  );
}
