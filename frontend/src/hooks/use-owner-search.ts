import { useMutation } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { transformOwner, type OwnerApiResponse } from "@/lib/transforms/owner";

export const OWNER_SEARCH_LIMIT = 100;

export interface OwnerSummary {
  id: string;
  name: string;
  phone: string;
  address: string;
  discountRate: number;
  membershipType: string;
}

interface OwnerSearchResponse {
  data: OwnerApiResponse[];
  total?: number;
}

function toOwnerSummary(o: OwnerApiResponse): OwnerSummary {
  const owner = transformOwner(o);
  return {
    id: owner.id,
    name: owner.ownerName,
    phone: owner.phone,
    address: [owner.address1, owner.address2].filter(Boolean).join(" "),
    discountRate: owner.discountRate,
    membershipType: owner.membershipType,
  };
}

export interface OwnerSearchResult {
  owners: OwnerSummary[];
  /** total が返却件数を超える＝先頭100件に打ち切られている */
  isTruncated: boolean;
}

/**
 * 飼主の横断検索（GET /v1/owners, page=1, limit=100）。
 * 検索ボタン等の命令的トリガーで呼ぶため mutation として実装する
 * （useQuery の宣言的 fetch ではなく mutateAsync で呼び出す）。
 */
export function useOwnerSearch() {
  return useMutation({
    mutationFn: async (searchTerm: string): Promise<OwnerSearchResult> => {
      const { data } = await axios.get<OwnerSearchResponse>("/v1/owners", {
        params: { search: searchTerm, page: 1, limit: OWNER_SEARCH_LIMIT },
      });
      const owners = (data.data ?? []).map(toOwnerSummary);
      return {
        owners,
        isTruncated: typeof data.total === "number" && data.total > owners.length,
      };
    },
  });
}
