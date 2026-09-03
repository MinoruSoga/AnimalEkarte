import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { MerchandiseItem } from "@/types/generated/models";

export interface FrontendMerchandiseItem {
  id: string;
  name: string;
  category: string;
  unitPrice: number;
  taxRate: number;
  isActive: boolean;
}

function transformMerchandiseItem(item: MerchandiseItem): FrontendMerchandiseItem {
  return {
    id: String(item.id ?? 0),
    name: item.name,
    category: item.category,
    unitPrice: item.unit_price,
    taxRate: item.tax_rate,
    isActive: item.is_active,
  };
}

export const useGetAllMerchandiseItems = () => {
  return useQuery({
    queryKey: queryKeys.accounting.merchandiseItems(),
    queryFn: async (): Promise<FrontendMerchandiseItem[]> => {
      const { data } = await axios.get<MerchandiseItem[] | { data: MerchandiseItem[] }>(
        "/v1/masters/merchandise-items",
      );
      const items = Array.isArray(data) ? data : data.data;
      return items.map(transformMerchandiseItem);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
