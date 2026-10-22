import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { MerchandiseItem, TaxType } from "@/types/generated/models";

// EMR-530: master wire の tax_rate は 有限な非負 exact decimal 文字列に変更された。
// 生成modelの number 型（同期前）を Omit し、wire 契約どおりの文字列へ差し替える。
type MerchandiseItemExact = Omit<MerchandiseItem, "tax_rate"> & { tax_rate: string };

export interface FrontendMerchandiseItem {
  id: string;
  name: string;
  category: string;
  unitPrice: number;
  taxType: TaxType;
  // 課税行の既定税率は number wire（billing_item は EMR-530 で変更しない）のため
  // master の exact decimal 文字列を表示・伝播用に数値化する。契約外の字句は
  // 数値化せず undefined にする（0 扱いで捏造しない）。
  taxRate: number | undefined;
  isActive: boolean;
}

function parseExactDecimal(text: string): number | undefined {
  return /^(0|[1-9][0-9]*)(\.[0-9]+)?$/.test(text) ? Number(text) : undefined;
}

function transformMerchandiseItem(item: MerchandiseItemExact): FrontendMerchandiseItem {
  return {
    id: String(item.id ?? 0),
    name: item.name,
    category: item.category,
    unitPrice: item.unit_price,
    // EMR-65: マスタ登録の税区分を欠落させない（excluded 一律潰しの回帰防止）
    taxType: item.tax_type,
    taxRate: parseExactDecimal(item.tax_rate),
    isActive: item.is_active,
  };
}

export const useGetAllMerchandiseItems = () => {
  return useQuery({
    queryKey: queryKeys.accounting.merchandiseItems(),
    queryFn: async (): Promise<FrontendMerchandiseItem[]> => {
      const { data } = await axios.get<MerchandiseItemExact[] | { data: MerchandiseItemExact[] }>(
        "/v1/masters/merchandise-items",
      );
      const items = Array.isArray(data) ? data : data.data;
      return items.map(transformMerchandiseItem);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
