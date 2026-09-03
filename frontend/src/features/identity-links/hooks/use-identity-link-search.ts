import { keepPreviousData, useQuery } from "@tanstack/react-query";

import { queryKeys } from "@/lib/query-keys";

import { searchOwnersForLink, searchPetsForLink } from "../api/identity-links-api";

const MIN_QUERY_LENGTH = 1;

/** 入力中に前回のヒット一覧を保持し、確定した文字だけを検索する（チラつき防止）。 */
export function useOwnerSearchQuery(trimmedQuery: string) {
  return useQuery({
    queryKey: queryKeys.identityLinks.ownerSearch(trimmedQuery),
    queryFn: () => searchOwnersForLink(trimmedQuery),
    enabled: trimmedQuery.length >= MIN_QUERY_LENGTH,
    placeholderData: keepPreviousData,
  });
}

export function usePetSearchQuery(trimmedQuery: string) {
  return useQuery({
    queryKey: queryKeys.identityLinks.petSearch(trimmedQuery),
    queryFn: () => searchPetsForLink(trimmedQuery),
    enabled: trimmedQuery.length >= MIN_QUERY_LENGTH,
    placeholderData: keepPreviousData,
  });
}
