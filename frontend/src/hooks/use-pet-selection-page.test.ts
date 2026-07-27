import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Pet } from "@/types";

import { usePetSelectionPage } from "./use-pet-selection-page";

const navigate = vi.fn();
const locationState = {
  from: "/reservations",
  appointmentId: "88",
  visitDate: "2026-05-29",
};

vi.mock("react-router", () => ({
  useNavigate: () => navigate,
  useLocation: () => ({
    search: "?appointmentId=88&visitDate=2026-05-29",
    state: locationState,
  }),
}));

interface MockPetsQueryResult {
  data?: Pet[];
  total?: number;
  page?: number;
  limit?: number;
  error?: unknown;
  isLoading?: boolean;
  isPlaceholderData?: boolean;
}

interface MockGetPetsOptions {
  includeDeceased?: boolean;
  page?: number;
  limit?: number;
  search?: string;
  species?: string;
}

interface MockGetPetsQueryOptions {
  preservePreviousData?: boolean;
}

let receivedQueryOptions: MockGetPetsQueryOptions | undefined;

const mockUseGetPets = vi.fn(
  (
    _ownerId?: string,
    _options?: MockGetPetsOptions,
  ): MockPetsQueryResult => ({
    data: [] as Pet[],
    total: 0,
    page: 1,
    limit: 20,
  }),
);

vi.mock("@/hooks/use-pet", () => ({
  useGetPets: (
    ownerId?: string,
    options?: MockGetPetsOptions,
    queryOptions?: MockGetPetsQueryOptions,
  ) => {
    receivedQueryOptions = queryOptions;
    return mockUseGetPets(ownerId, options);
  },
}));

const katakanaOwnerPet = {
  id: "10",
  ownerId: "20",
  name: "ポチ",
  ownerName: "ヤマダハナコ",
  ownerNameKana: "ヤマダハナコ",
  petNameKana: "ポチ",
  species: "犬",
  status: "生存",
} as unknown as Pet;

const hiraganaOwnerPet = {
  id: "11",
  ownerId: "21",
  name: "たろう",
  ownerName: "さとうけんじ",
  ownerNameKana: "さとうけんじ",
  petNameKana: "たろう",
  species: "猫",
} as unknown as Pet;

describe("usePetSelectionPage", () => {
  beforeEach(() => {
    navigate.mockClear();
    mockUseGetPets.mockClear();
    receivedQueryOptions = undefined;
    mockUseGetPets.mockReturnValue({ data: [], total: 0, page: 1, limit: 20 });
  });

  // 回帰防止: 旧実装は `const { data: pets = [] } = useGetPets(...)` で error/isLoading を
  // 破棄していたため、GET /v1/pets が 400 を返しても呼び出し側は「0件」としか判別できなかった。
  it("取得失敗を握り潰さず error として返す（0件と区別できる）", () => {
    const apiError = new Error("Request failed with status code 400");
    mockUseGetPets.mockReturnValue({ data: undefined, error: apiError, isLoading: false });

    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    expect(result.current.error).toBe(apiError);
    expect(result.current.filteredPets.items).toEqual([]);
  });

  it("読み込み中を握り潰さず isLoading として返す", () => {
    mockUseGetPets.mockReturnValue({ data: undefined, error: undefined, isLoading: true });

    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    expect(result.current.isLoading).toBe(true);
    expect(result.current.filteredPets.items).toEqual([]);
  });

  it("ページ切替中の前回データも読み込み中として選択側へ伝える", () => {
    mockUseGetPets.mockReturnValue({
      data: [katakanaOwnerPet],
      total: 100,
      page: 1,
      limit: 20,
      isLoading: false,
      isPlaceholderData: true,
    });

    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    expect(result.current.isLoading).toBe(true);
    expect(result.current.filteredPets.items).toEqual([katakanaOwnerPet]);
  });

  it("取得成功時は error が無く isLoading も false になる", () => {
    mockUseGetPets.mockReturnValue({
      data: [katakanaOwnerPet],
      total: 15_654,
      page: 1,
      limit: 20,
      error: undefined,
      isLoading: false,
    });

    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    expect(result.current.error).toBeUndefined();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.filteredPets.items).toEqual([katakanaOwnerPet]);
    expect(result.current.filteredPets.totalCount).toBe(15_654);
    expect(result.current.filteredPets.startIndex).toBe(1);
    expect(result.current.filteredPets.endIndex).toBe(20);
  });

  it("死亡個体・先頭ページ・20件を指定して共有一覧を取得する", () => {
    renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    expect(mockUseGetPets).toHaveBeenCalledWith(undefined, {
      includeDeceased: true,
      page: 1,
      limit: 20,
    });
    expect(receivedQueryOptions).toEqual({ preservePreviousData: true });
  });

  it("ページ移動でbackendへページ番号を渡す", () => {
    mockUseGetPets.mockReturnValue({
      data: [],
      total: 100,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    act(() => {
      result.current.filteredPets.onPageChange(2);
    });

    expect(mockUseGetPets).toHaveBeenLastCalledWith(undefined, {
      includeDeceased: true,
      page: 2,
      limit: 20,
    });
  });

  it("検索語をbackendへ渡し、検索条件の変更時に先頭ページへ戻す", () => {
    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    act(() => {
      result.current.filteredPets.onPageChange(3);
    });
    act(() => {
      result.current.setSearchParams({
        ...result.current.searchParams,
        search: "もも",
      });
    });

    expect(mockUseGetPets).toHaveBeenLastCalledWith(undefined, {
      includeDeceased: true,
      page: 1,
      limit: 20,
      search: "もも",
    });
  });

  it("動物種IDと飼主IDをbackendへ渡す", () => {
    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    act(() => {
      result.current.setSearchParams({
        ...result.current.searchParams,
        ownerId: "30042",
        species: "3",
      });
    });

    expect(mockUseGetPets).toHaveBeenLastCalledWith("30042", {
      includeDeceased: true,
      page: 1,
      limit: 20,
      species: "3",
    });
  });

  it("死亡個体は callback 境界で選択を拒否する", () => {
    const deceasedPet = { ...katakanaOwnerPet, status: "死亡" } as Pet;
    mockUseGetPets.mockReturnValue({ data: [deceasedPet] });
    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    act(() => {
      result.current.handleSelect(deceasedPet);
    });

    expect(navigate).not.toHaveBeenCalled();
  });

  it("ペット選択後も既存クエリと state を作成画面へ引き継ぐ", () => {
    mockUseGetPets.mockReturnValue({ data: [katakanaOwnerPet] });
    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    act(() => {
      result.current.handleSelect(katakanaOwnerPet);
    });

    expect(navigate).toHaveBeenCalledWith(
      "/trimming/new?appointmentId=88&visitDate=2026-05-29&petId=10",
      { state: locationState },
    );
  });

  it("生死不明の個体はfail-closedで選択を拒否する", () => {
    const unknownStatusPet = {
      ...katakanaOwnerPet,
      status: undefined,
    } as Pet;
    mockUseGetPets.mockReturnValue({ data: [unknownStatusPet] });
    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    act(() => {
      result.current.handleSelect(unknownStatusPet);
    });

    expect(navigate).not.toHaveBeenCalled();
  });

  describe("かな正規化フィルタ", () => {
    const config = { selectPath: "/trimming/new", backPath: "/trimming" };
    const pets = [katakanaOwnerPet, hiraganaOwnerPet];

    function setup() {
      mockUseGetPets.mockReturnValue({
        data: pets,
        total: pets.length,
        page: 1,
        limit: 20,
      });
      return renderHook(() => usePetSelectionPage(config));
    }

    it("ひらがな入力でカタカナ ownerName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, ownerName: "やまだ" });
      });
      expect(result.current.filteredPets.items).toEqual([katakanaOwnerPet]);
    });

    it("カタカナ入力でひらがな ownerName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, ownerName: "サトウ" });
      });
      expect(result.current.filteredPets.items).toEqual([hiraganaOwnerPet]);
    });

    it("ひらがな入力でカタカナ petName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({
          ...result.current.searchParams,
          petName: "ぽち",
        });
      });
      expect(result.current.filteredPets.items).toEqual([katakanaOwnerPet]);
      expect(result.current.filteredPets.isPageLocalFiltered).toBe(true);
    });

    it("カタカナ入力でひらがな petName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({
          ...result.current.searchParams,
          petName: "タロウ",
        });
      });
      expect(result.current.filteredPets.items).toEqual([hiraganaOwnerPet]);
    });
  });
});
