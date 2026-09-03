import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

export interface DiscountSuggestion {
  type: "campaign" | "owner";
  campaign_id?: number;
  name: string;
  discount_type: "rate" | "amount";
  discount_value: number;
  amount: number;
}

interface BackendDiscountSuggestionsResponse {
  suggestions: DiscountSuggestion[];
}

const getDiscountSuggestions = async (itemId: string): Promise<DiscountSuggestion[]> => {
  const { data } = await axios.get<BackendDiscountSuggestionsResponse>(
    `/v1/billing-items/${itemId}/discount-suggestions`,
  );
  return data.suggestions ?? [];
};

/** #81 Q-I: 明細に適用可能な割引候補を取得するフック。enabled=true になった時点で fetch する。 */
export const useGetBillingItemDiscountSuggestions = (
  itemId: string | undefined,
  enabled: boolean,
) =>
  useQuery({
    queryKey: queryKeys.billingItemDiscountSuggestions(itemId!),
    queryFn: () => getDiscountSuggestions(itemId!),
    enabled: enabled && itemId !== undefined,
    staleTime: QUERY_STALE_TIMES.SHORT,
  });
