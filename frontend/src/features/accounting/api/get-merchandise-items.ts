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
  // EMR-530: 有限な非負 exact decimal 文字列（採択405・483の十進輸送、api.yaml #806）。
  tax_rate: string;
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
  // 課税行の既定税率は number wire（billing_item は EMR-530 で変更しない）のため
  // master の exact decimal 文字列を表示・伝播用に数値化する。契約外の字句は
  // 数値化せず undefined にする（0 扱いで捏造しない）。
  taxRate: number | undefined;
  isActive: boolean;
}

// parseExactDecimal は exact decimal 文字列を課税行wire（number）用に数値化する。
// binary64 で無損失に表現できない字句（"0.10000000000000001" 等）は丸めて伝播すると
// 誤課税になるため undefined（fail closed）にする。round-trip で同一の十進値のときだけ数値を返す。
function parseExactDecimal(text: string): number | undefined {
  if (!/^(0|[1-9][0-9]*)(\.[0-9]+)?$/.test(text)) return undefined;
  const asNumber = Number(text);
  return canonicalDecimal(String(asNumber)) === canonicalDecimal(text) ? asNumber : undefined;
}

function canonicalDecimal(text: string): string | null {
  if (!/^(0|[1-9][0-9]*)(\.[0-9]+)?$/.test(text)) return null;
  const [intPart, fracPart = ""] = text.split(".");
  const frac = fracPart.replace(/0+$/, "");
  if (frac === "") return intPart;
  return `${intPart}.${frac}`;
}

function transformMerchandiseItem(item: LegacyMerchandiseItemResponse): FrontendMerchandiseItem {
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
      const { data } = await axios.get<MerchandiseItemsPayload>("/v1/masters/merchandise-items");
      const items = Array.isArray(data) ? data : data.data;
      return items.map(transformMerchandiseItem);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
