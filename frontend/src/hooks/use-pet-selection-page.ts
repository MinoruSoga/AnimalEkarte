import { useState, useMemo, useCallback } from "react";
import { useLocation, useNavigate } from "react-router";
import { useGetPets } from "@/hooks/use-pet";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import type { Pet } from "@/types";
import type { PetSelectionSearchParams } from "@/components/shared/PetSelection/PetSelectionSearchForm";

interface PetSelectionPageConfig {
  /** 選択後の遷移先パス (例: "/examinations/new") */
  selectPath: string;
  /** 戻るボタンの遷移先 (例: "/examinations") */
  backPath: string;
}

/**
 * backend が返した1ページ分。`items` は応答そのものであり、
 * FE 側で絞り込んではならない（BUG-451）。
 */
export interface PetSelectionResultPage {
  items: Pet[];
  totalCount: number;
  currentPage: number;
  totalPages: number;
  startIndex: number;
  endIndex: number;
  onPageChange: (page: number) => void;
}

const PAGE_SIZE = 20;

/** 入力1文字ごとに一覧APIを叩かないための待ち時間。 */
export const SEARCH_DEBOUNCE_MS = 300;

const INITIAL_SEARCH_PARAMS: PetSelectionSearchParams = {
  search: "",
  ownerId: "",
  species: "",
};

export function usePetSelectionPage(config: PetSelectionPageConfig) {
  const navigate = useNavigate();
  const location = useLocation();
  const [page, setPage] = useState(1);
  const [searchParams, setSearchParams] =
    useState<PetSelectionSearchParams>(INITIAL_SEARCH_PARAMS);

  // 入力停止後にだけ問い合わせる。「検索」ボタンは持たない（自動検索）。
  const debouncedSearchParams = useDebouncedValue(
    searchParams,
    SEARCH_DEBOUNCE_MS,
  );

  // 検索条件は全て backend の述語へ委譲する。ここで絞り込みを足すと、
  // 総件数(total)を根拠に「N件中 1-20件」と表示しながらその20件から
  // 黙って行を消すことになり、利用者は実在患者を未登録と誤認する。
  //
  // error / isLoading を破棄してはならない。破棄すると API 失敗が
  // 「該当0件」と区別できなくなり、利用者に嘘の検索結果を見せる。
  const {
    data: pets = [],
    total = 0,
    page: responsePage = page,
    limit: responseLimit = PAGE_SIZE,
    error,
    isLoading,
    isPlaceholderData,
  } = useGetPets(
    debouncedSearchParams.ownerId || undefined,
    {
      includeDeceased: true,
      page,
      limit: PAGE_SIZE,
      ...(debouncedSearchParams.search
        ? { search: debouncedSearchParams.search }
        : {}),
      ...(debouncedSearchParams.species
        ? { species: debouncedSearchParams.species }
        : {}),
    },
    { preservePreviousData: true },
  );

  const totalPages = Math.max(1, Math.ceil(total / responseLimit));

  const handlePageChange = useCallback(
    (nextPage: number) => {
      setPage(Math.min(Math.max(1, nextPage), totalPages));
    },
    [totalPages],
  );

  const updateSearchParams = useCallback((params: PetSelectionSearchParams) => {
    setSearchParams(params);
    setPage(1);
  }, []);

  const petPage = useMemo<PetSelectionResultPage>(() => {
    const startIndex = total === 0 ? 0 : (responsePage - 1) * responseLimit + 1;
    const endIndex = total === 0 ? 0 : Math.min(responsePage * responseLimit, total);

    return {
      items: pets,
      totalCount: total,
      currentPage: responsePage,
      totalPages,
      startIndex,
      endIndex,
      onPageChange: handlePageChange,
    };
  }, [
    handlePageChange,
    pets,
    responseLimit,
    responsePage,
    total,
    totalPages,
  ]);

  const handleClear = useCallback(() => {
    setSearchParams(INITIAL_SEARCH_PARAMS);
    setPage(1);
  }, []);

  const handleSelect = useCallback((pet: Pet) => {
    if (pet.status !== "生存") return;

    const nextParams = new URLSearchParams(location.search);
    nextParams.set("petId", pet.id);
    navigate(`${config.selectPath}?${nextParams.toString()}`, { state: location.state });
  }, [navigate, config.selectPath, location.search, location.state]);

  const handleBack = useCallback(() => {
    navigate(config.backPath);
  }, [navigate, config.backPath]);

  return {
    searchParams,
    setSearchParams: updateSearchParams,
    petPage,
    error,
    isLoading: Boolean(isLoading || isPlaceholderData),
    handleClear,
    handleSelect,
    handleBack,
  };
}
