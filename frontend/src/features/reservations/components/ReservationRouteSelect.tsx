import { useEffect, useState } from "react";
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
} from "../constants/reservation-route";
import { useUpdateReservationRoute } from "../api/update-reservation-route";

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
  const [displayValue, setDisplayValue] = useState<ReservationRoute | null>(value);
  // sync internal state when prop changes (modal reopened with new data)
  useEffect(() => {
    setDisplayValue(value);
  }, [value]);

  const { mutate, isPending } = useUpdateReservationRoute(reservationId);

  const handleValueChange = (selected: string) => {
    if (selected === "_clear") {
      setDisplayValue(null);
      mutate({ route: "" });
    } else {
      const route = selected as ReservationRoute;
      setDisplayValue(route);
      mutate({ route });
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <label className={STYLE.formLabel}>予約経路</label>
      <Select
        value={displayValue ?? ""}
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
          {displayValue !== null ? (
            <SelectItem value="_clear">クリア</SelectItem>
          ) : null}
        </SelectContent>
      </Select>
    </div>
  );
}
