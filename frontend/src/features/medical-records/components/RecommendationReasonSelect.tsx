import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { STYLE } from "@/lib/design-tokens";
import {
  RECOMMENDATION_REASON_LABELS,
  RECOMMENDATION_REASON_VALUES,
  type RecommendationReason,
} from "../constants/recommendation-reason";
import { useUpdateRecommendationReason } from "../api/update-recommendation-reason";

interface RecommendationReasonSelectProps {
  medicalRecordId: string;
  value: RecommendationReason | null;
  disabled?: boolean;
}

export function RecommendationReasonSelect({
  medicalRecordId,
  value,
  disabled = false,
}: RecommendationReasonSelectProps) {
  const { mutate, isPending } = useUpdateRecommendationReason(medicalRecordId);

  const handleValueChange = (selected: string) => {
    if (selected === "_clear") {
      mutate({ reason: null });
    } else {
      mutate({ reason: selected as RecommendationReason });
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <label className={STYLE.formLabel}>推奨理由</label>
      <Select
        value={value ?? ""}
        onValueChange={handleValueChange}
        disabled={disabled || isPending}
      >
        <SelectTrigger
          className="w-[160px] h-9 text-sm"
          aria-label="推奨理由を選択"
          data-testid="recommendation-reason-trigger"
        >
          <SelectValue placeholder="未選択" />
        </SelectTrigger>
        <SelectContent>
          {RECOMMENDATION_REASON_VALUES.map((v) => (
            <SelectItem key={v} value={v}>
              {RECOMMENDATION_REASON_LABELS[v]}
            </SelectItem>
          ))}
          {value !== null ? (
            <SelectItem value="_clear">クリア</SelectItem>
          ) : null}
        </SelectContent>
      </Select>
    </div>
  );
}
