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
  const [searchParams, setSearchParams] =
    useState<PetSelectionSearchParams>(INITIAL_SEARCH_PARAMS);

  // 入力停止後にだけ問い合わせる。「検索」ボタンは持たない（自動検索）。
  const debouncedSearchParams = useDebouncedValue(
    searchParams,
    SEARCH_DEBOUNCE_MS,
  );

  // 入力からデバウンス確定までの間、表示中の一覧は「まだ古い検索条件の結果」である。
  const isSearchPending = searchParams !== debouncedSearchParams;

  // ページ番号は「それを選んだ時点で確定していた検索条件」と対で保持する。
  // デバウンス待ちの間にページ送りされても、新しい条件が確定した時点で先頭へ戻る。
  // 対にしないと、古い totalPages を根拠に選んだ page が新条件のクエリへ持ち越され、
  // 「5件中 21-5件」のような反転した範囲や、総件数はあるのに一覧が空という
  // 自己矛盾した表示になる（BUG-451 と同じ「嘘の件数」クラス）。
  const [pageBinding, setPageBinding] = useState<{
    page: number;
    params: PetSelectionSearchParams;
  }>({ page: 1, params: INITIAL_SEARCH_PARAMS });
  const page = pageBinding.params === debouncedSearchParams ? pageBinding.page : 1;

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
      // 直接記録入力の7 selectorでは、本人同定と死亡 sentinel の表示に死亡個体も必要。
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
      setPageBinding({
        page: Math.min(Math.max(1, nextPage), totalPages),
        params: debouncedSearchParams,
      });
    },
    [debouncedSearchParams, totalPages],
  );

  // ページは debouncedSearchParams との対応で導出するため、ここでの reset は不要。
  const updateSearchParams = useCallback((params: PetSelectionSearchParams) => {
    setSearchParams(params);
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
    setPageBinding({ page: 1, params: INITIAL_SEARCH_PARAMS });
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
    // デバウンス待ちの間は「入力済みの条件と表示中の一覧が一致していない」。
    // 確定済みとして扱うと、古い結果から患者を選ばせてしまう。
    isLoading: Boolean(isLoading || isPlaceholderData || isSearchPending),
    handleClear,
    handleSelect,
    handleBack,
  };
}
