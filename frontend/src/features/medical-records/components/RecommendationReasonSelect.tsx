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

type Props =
  | {
      mode: "edit";
      medicalRecordId: string;
      value: RecommendationReason | null;
      disabled?: boolean;
    }
  | {
      mode: "create";
      value: RecommendationReason | null;
      onChange: (value: RecommendationReason | null) => void;
      disabled?: boolean;
    };

export function RecommendationReasonSelect(props: Props) {
  const { mutate, isPending } = useUpdateRecommendationReason(
    props.mode === "edit" ? props.medicalRecordId : "",
  );

  const handleValueChange = (selected: string) => {
    if (props.mode === "edit") {
      if (selected === "_clear") {
        mutate({ reason: null });
      } else {
        mutate({ reason: selected as RecommendationReason });
      }
    } else {
      props.onChange(selected === "_clear" ? null : (selected as RecommendationReason));
    }
  };

  const { value, disabled = false } = props;

  return (
    <div className="flex flex-col gap-2">
      <label className={STYLE.formLabel}>推奨理由</label>
      <Select
        value={value ?? ""}
        onValueChange={handleValueChange}
        disabled={disabled || (props.mode === "edit" ? isPending : false)}
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
          {value !== null ? <SelectItem value="_clear">クリア</SelectItem> : null}
        </SelectContent>
      </Select>
    </div>
  );
}
