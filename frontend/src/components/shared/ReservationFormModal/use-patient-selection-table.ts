import { useCallback, useMemo, useState } from "react";

import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useGetPets } from "@/hooks/use-pet";
import { SEARCH_DEBOUNCE_MS } from "@/hooks/use-pet-selection-page";
import type { Pet } from "@/types";

import {
  INITIAL_SEARCH_PARAMS,
  PAGE_SIZE,
  resolvePatientSelectionText,
  resolvePatientStatusText,
  type PatientSearchParams,
} from "./patient-selection-table-model";

interface UsePatientSelectionTableArgs {
  selectedPets: Pet[];
  includeDeceased: boolean;
}

export function usePatientSelectionTable({
  selectedPets,
  includeDeceased,
}: UsePatientSelectionTableArgs) {
  const [searchParams, setSearchParams] = useState<PatientSearchParams>(
    INITIAL_SEARCH_PARAMS,
  );
  const debouncedSearchParams = useDebouncedValue(
    searchParams,
    SEARCH_DEBOUNCE_MS,
  );
  const isSearchPending = searchParams !== debouncedSearchParams;
  const hasSearchConditions = useMemo(
    () =>
      Object.values(debouncedSearchParams).some((value) => value.trim() !== ""),
    [debouncedSearchParams],
  );

  const [pageBinding, setPageBinding] = useState<{
    page: number;
    params: PatientSearchParams;
  }>({ page: 1, params: INITIAL_SEARCH_PARAMS });
  const page =
    pageBinding.params === debouncedSearchParams ? pageBinding.page : 1;

  const {
    data: pets = [],
    total,
    page: rawResponsePage,
    limit: rawResponseLimit,
    error,
    isLoading,
    isPlaceholderData,
  } = useGetPets(
    debouncedSearchParams.ownerId || undefined,
    {
      ...(includeDeceased ? { includeDeceased: true } : {}),
      page,
      limit: PAGE_SIZE,
      ...(debouncedSearchParams.search
        ? { search: debouncedSearchParams.search }
        : {}),
      ...(debouncedSearchParams.species
        ? { species: debouncedSearchParams.species }
        : {}),
    },
    { enabled: hasSearchConditions, preservePreviousData: true },
  );

  const isBusy = Boolean(isLoading || isPlaceholderData || isSearchPending);
  const responsePage =
    typeof rawResponsePage === "number" && rawResponsePage > 0
      ? rawResponsePage
      : page;
  const responseLimit =
    typeof rawResponseLimit === "number" && rawResponseLimit > 0
      ? rawResponseLimit
      : PAGE_SIZE;
  const hasTotal = typeof total === "number";
  const totalCount = total ?? 0;
  const totalPages = Math.max(1, Math.ceil(totalCount / responseLimit));
  const startIndex =
    totalCount === 0 ? 0 : (responsePage - 1) * responseLimit + 1;
  const endIndex =
    totalCount === 0 ? 0 : Math.min(responsePage * responseLimit, totalCount);
  const rangeText = `${totalCount.toLocaleString()}件中 ${startIndex.toLocaleString()}-${endIndex.toLocaleString()}件`;
  const isCountTrustworthy =
    hasTotal && (totalCount > 0 ? startIndex <= totalCount : pets.length === 0);
  const showCachedRange =
    pets.length > 0 && Boolean(error || isBusy) && isCountTrustworthy;
  const statusText = resolvePatientStatusText({
    hasSearchConditions,
    showCachedRange,
    rangeText,
    error,
    isBusy,
    isCountTrustworthy,
    totalCount,
  });
  const isRangeShownByPagination = !error && totalPages > 1;

  const handlePageChange = useCallback(
    (nextPage: number) => {
      setPageBinding({
        page: Math.min(Math.max(1, nextPage), totalPages),
        params: debouncedSearchParams,
      });
    },
    [debouncedSearchParams, totalPages],
  );

  const handleTextFieldChange = useCallback(
    (key: "search" | "ownerId", value: string) => {
      setSearchParams((prev) => ({ ...prev, [key]: value }));
    },
    [],
  );

  const handleSpeciesChange = useCallback((species: string) => {
    setSearchParams((prev) => ({ ...prev, species }));
  }, []);

  const handleClear = useCallback(() => {
    setSearchParams(INITIAL_SEARCH_PARAMS);
    setPageBinding({ page: 1, params: INITIAL_SEARCH_PARAMS });
  }, []);

  const selectedPetIds = useMemo(
    () => new Set(selectedPets.map((p) => p.id)),
    [selectedPets],
  );
  const currentPageIds = useMemo(() => new Set(pets.map((p) => p.id)), [pets]);
  const canLocateSelectionOnPage = hasSearchConditions && pets.length > 0;
  const offPageSelectedCount = useMemo(
    () =>
      canLocateSelectionOnPage
        ? selectedPets.reduce(
            (count, pet) => (currentPageIds.has(pet.id) ? count : count + 1),
            0,
          )
        : 0,
    [canLocateSelectionOnPage, selectedPets, currentPageIds],
  );

  return {
    searchParams,
    pets,
    error,
    isBusy,
    isSearchPending,
    hasSearchConditions,
    responsePage,
    totalPages,
    totalCount,
    startIndex,
    endIndex,
    statusText,
    isRangeShownByPagination,
    selectedPetIds,
    selectionText: resolvePatientSelectionText(
      selectedPets.length,
      offPageSelectedCount,
    ),
    handlePageChange,
    handleTextFieldChange,
    handleSpeciesChange,
    handleClear,
  };
}
