import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ItemCategory, TaxType } from "@/types";

/**
 * 物販品マスタ応答（非永続の移行 HTTP DTO・EMR-527）。保存 model（生成 MerchandiseItem）
 * ではなく handler の応答列を写す。master の merchandise-items.ts も同じ struct を写す
 * （feature 間 import は禁止のため各 feature に置く）。
 * @see backend/internal/inventory/merchandise_item_response.go merchandiseItemResponse
 */
interface LegacyMerchandiseItemResponse {
  id: number;
  clinic_id: number;
  name: string;
  category: ItemCategory;
  unit_price: number;
  tax_type: TaxType;
  tax_rate: number;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

/** 一覧は配列で返る。旧 { data } 包みも従来どおり受ける。 */
type MerchandiseItemsPayload =
  | LegacyMerchandiseItemResponse[]
  | { data: LegacyMerchandiseItemResponse[] };

export interface FrontendMerchandiseItem {
  id: string;
  name: string;
  category: string;
  unitPrice: number;
  taxType: TaxType;
  taxRate: number;
  isActive: boolean;
}

function transformMerchandiseItem(item: LegacyMerchandiseItemResponse): FrontendMerchandiseItem {
  return {
    id: String(item.id ?? 0),
    name: item.name,
    category: item.category,
    unitPrice: item.unit_price,
    // EMR-65: マスタ登録の税区分を欠落させない（excluded 一律潰しの回帰防止）
    taxType: item.tax_type,
    taxRate: item.tax_rate,
    isActive: item.is_active,
  };
}

export const useGetAllMerchandiseItems = () => {
  return useQuery({
    queryKey: queryKeys.accounting.merchandiseItems(),
    queryFn: async (): Promise<FrontendMerchandiseItem[]> => {
      const { data } = await axios.get<MerchandiseItemsPayload>("/v1/masters/merchandise-items");
      const items = Array.isArray(data) ? data : data.data;
      return items.map(transformMerchandiseItem);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
