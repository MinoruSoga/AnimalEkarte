import { useState, useMemo, useCallback } from "react";
import { useLocation, useNavigate } from "react-router";
import { useGetPets } from "@/hooks/use-pet";
import { normalizedIncludes } from "@/lib/normalize-kana";
import type { Pet } from "@/types";
import type { PetSelectionSearchParams } from "@/components/shared/PetSelection/PetSelectionSearchForm";

interface PetSelectionPageConfig {
  /** 選択後の遷移先パス (例: "/examinations/new") */
  selectPath: string;
  /** 戻るボタンの遷移先 (例: "/examinations") */
  backPath: string;
}

export interface PetSelectionResultPage {
  items: Pet[];
  totalCount: number;
  currentPage: number;
  totalPages: number;
  startIndex: number;
  endIndex: number;
  isPageLocalFiltered: boolean;
  onPageChange: (page: number) => void;
}

const PAGE_SIZE = 20;

const INITIAL_SEARCH_PARAMS: PetSelectionSearchParams = {
  search: "",
  ownerId: "",
  ownerName: "",
  ownerNameKana: "",
  phone: "",
  petName: "",
  petNameKana: "",
  species: "",
  address: "",
};

export function usePetSelectionPage(config: PetSelectionPageConfig) {
  const navigate = useNavigate();
  const location = useLocation();
  const [page, setPage] = useState(1);
  const [searchParams, setSearchParams] =
    useState<PetSelectionSearchParams>(INITIAL_SEARCH_PARAMS);

  // error / isLoading を破棄してはならない。破棄すると API 失敗が
  // 「該当0件」と区別できなくなり、利用者に嘘の検索結果を見せる。
  const {
    data: pets = [],
    total = 0,
    page: responsePage = page,
    limit: responseLimit = PAGE_SIZE,
    error,
    isLoading,
  } = useGetPets(searchParams.ownerId || undefined, {
    includeDeceased: true,
    page,
    limit: PAGE_SIZE,
    ...(searchParams.search ? { search: searchParams.search } : {}),
    ...(searchParams.species ? { species: searchParams.species } : {}),
  });

  const totalPages = Math.max(1, Math.ceil(total / responseLimit));

  const handlePageChange = useCallback(
    (nextPage: number) => {
      setPage(Math.min(Math.max(1, nextPage), totalPages));
    },
    [totalPages],
  );

  const updateSearchParams = useCallback(
    (params: PetSelectionSearchParams) => {
      setSearchParams(params);
      setPage(1);
    },
    [],
  );

  const filteredPets = useMemo<PetSelectionResultPage>(() => {
    const items = pets.filter((pet) => {
      if (
        searchParams.ownerName &&
        !normalizedIncludes(pet.ownerName, searchParams.ownerName)
      )
        return false;
      if (
        searchParams.ownerNameKana &&
        (!pet.ownerNameKana ||
          !normalizedIncludes(pet.ownerNameKana, searchParams.ownerNameKana))
      )
        return false;
      if (
        searchParams.phone &&
        (!pet.phone || !pet.phone.includes(searchParams.phone))
      )
        return false;
      if (
        searchParams.address &&
        (!pet.address || !pet.address.includes(searchParams.address))
      )
        return false;
      if (
        searchParams.petName &&
        !normalizedIncludes(pet.name, searchParams.petName)
      )
        return false;
      if (
        searchParams.petNameKana &&
        (!pet.petNameKana ||
          !normalizedIncludes(pet.petNameKana, searchParams.petNameKana))
      )
        return false;
      return true;
    });

    const isPageLocalFiltered = Boolean(
      searchParams.ownerName ||
        searchParams.ownerNameKana ||
        searchParams.phone ||
        searchParams.address ||
        searchParams.petName ||
        searchParams.petNameKana,
    );
    const startIndex =
      total === 0 ? 0 : (responsePage - 1) * responseLimit + 1;
    const endIndex =
      total === 0 ? 0 : Math.min(responsePage * responseLimit, total);

    return {
      items,
      totalCount: total,
      currentPage: responsePage,
      totalPages,
      startIndex,
      endIndex,
      isPageLocalFiltered,
      onPageChange: handlePageChange,
    };
  }, [
    handlePageChange,
    pets,
    responseLimit,
    responsePage,
    searchParams,
    total,
    totalPages,
  ]);

  // owner_id / search / species は backend の全件集合へ適用する。
  // 個別欄は backend の単一 OR search と同じ意味を表せないため、
  // 現在ページ内フィルタとして維持する。

  // フィルタはリアクティブ（useMemo）のため、ボタン押下時の追加処理は不要
  const handleSearch = useCallback(() => {
    setPage(1);
  }, []);

  const handleClear = useCallback(() => {
    setSearchParams(INITIAL_SEARCH_PARAMS);
    setPage(1);
  }, []);

  const handleSelect = useCallback((pet: Pet) => {
    if (pet.status === "死亡") return;

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
    filteredPets,
    error,
    isLoading,
    handleSearch,
    handleClear,
    handleSelect,
    handleBack,
  };
}
