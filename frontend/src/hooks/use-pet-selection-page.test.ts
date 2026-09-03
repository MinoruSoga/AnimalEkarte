import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Pet } from "@/types";

import { SEARCH_DEBOUNCE_MS, usePetSelectionPage } from "./use-pet-selection-page";

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
  (_ownerId?: string, _options?: MockGetPetsOptions): MockPetsQueryResult => ({
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

const CONFIG = { selectPath: "/trimming/new", backPath: "/trimming" };

/** 検索条件を変更し、デバウンスを満了させて backend への反映まで進める。 */
function applySearch(
  result: { current: ReturnType<typeof usePetSelectionPage> },
  partial: Partial<ReturnType<typeof usePetSelectionPage>["searchParams"]>,
) {
  act(() => {
    result.current.setSearchParams({
      ...result.current.searchParams,
      ...partial,
    });
  });
  act(() => {
    vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS);
  });
}

describe("usePetSelectionPage", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    navigate.mockClear();
    mockUseGetPets.mockClear();
    receivedQueryOptions = undefined;
    mockUseGetPets.mockReturnValue({ data: [], total: 0, page: 1, limit: 20 });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  // 回帰防止: 旧実装は `const { data: pets = [] } = useGetPets(...)` で error/isLoading を
  // 破棄していたため、GET /v1/pets が 400 を返しても呼び出し側は「0件」としか判別できなかった。
  it("取得失敗を握り潰さず error として返す（0件と区別できる）", () => {
    const apiError = new Error("Request failed with status code 400");
    mockUseGetPets.mockReturnValue({
      data: undefined,
      error: apiError,
      isLoading: false,
    });

    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    expect(result.current.error).toBe(apiError);
    expect(result.current.petPage.items).toEqual([]);
  });

  it("読み込み中を握り潰さず isLoading として返す", () => {
    mockUseGetPets.mockReturnValue({
      data: undefined,
      error: undefined,
      isLoading: true,
    });

    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    expect(result.current.isLoading).toBe(true);
    expect(result.current.petPage.items).toEqual([]);
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

    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    expect(result.current.isLoading).toBe(true);
    expect(result.current.petPage.items).toEqual([katakanaOwnerPet]);
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

    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    expect(result.current.error).toBeUndefined();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.petPage.items).toEqual([katakanaOwnerPet]);
    expect(result.current.petPage.totalCount).toBe(15_654);
    expect(result.current.petPage.startIndex).toBe(1);
    expect(result.current.petPage.endIndex).toBe(20);
  });

  it("死亡個体・先頭ページ・20件を指定して共有一覧を取得する", () => {
    renderHook(() => usePetSelectionPage(CONFIG));

    expect(mockUseGetPets).toHaveBeenCalledWith(undefined, {
      includeDeceased: true,
      page: 1,
      limit: 20,
    });
    expect(receivedQueryOptions).toEqual({ preservePreviousData: true });
  });

  it("ページ移動でbackendへページ番号を渡す", () => {
    mockUseGetPets.mockReturnValue({ data: [], total: 100, page: 1, limit: 20 });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    act(() => {
      result.current.petPage.onPageChange(2);
    });

    expect(mockUseGetPets).toHaveBeenLastCalledWith(undefined, {
      includeDeceased: true,
      page: 2,
      limit: 20,
    });
  });

  it("検索語をbackendへ渡し、検索条件の変更時に先頭ページへ戻す", () => {
    mockUseGetPets.mockReturnValue({ data: [], total: 100, page: 1, limit: 20 });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    act(() => {
      result.current.petPage.onPageChange(3);
    });
    applySearch(result, { search: "もも" });

    expect(mockUseGetPets).toHaveBeenLastCalledWith(undefined, {
      includeDeceased: true,
      page: 1,
      limit: 20,
      search: "もも",
    });
  });

  it("動物種IDと飼主IDをbackendへ渡す", () => {
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    applySearch(result, { ownerId: "30042", species: "3" });

    expect(mockUseGetPets).toHaveBeenLastCalledWith("30042", {
      includeDeceased: true,
      page: 1,
      limit: 20,
      species: "3",
    });
  });

  // BUG-451 の核心。backend が返した1ページ分を FE で再度絞り込むと、
  // 画面は総件数(例 15,654)を根拠に「N件中 1-20件」と表示しながら、
  // その20件から黙って行を消す。利用者は「全件を検索した」と誤解し、
  // 実在患者を未登録と誤認して重複カルテを作りうる。
  // 実装がページ内クライアント側フィルタへ退行したらこのテストは必ず落ちる。
  it("backendが返した行を検索語でクライアント側再絞り込みしない", () => {
    const pets = [katakanaOwnerPet, hiraganaOwnerPet];
    mockUseGetPets.mockReturnValue({
      data: pets,
      total: 15_654,
      page: 1,
      limit: 20,
    });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    // どの行のどのフィールドにも一致しない語。旧実装なら 0 件へ潰れていた。
    applySearch(result, { search: "該当しない検索語" });

    expect(result.current.petPage.items).toEqual(pets);
    expect(result.current.petPage.totalCount).toBe(15_654);
  });

  it("21件目以降にしか存在しない患者も、backendが返せばそのまま提示する", () => {
    const pageTwoOnlyPet = {
      ...katakanaOwnerPet,
      id: "9999",
      name: "ページ2にしか居ないペット",
    } as Pet;
    mockUseGetPets.mockReturnValue({
      data: [pageTwoOnlyPet],
      total: 15_654,
      page: 2,
      limit: 20,
    });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    applySearch(result, { search: "ページ2" });

    expect(mockUseGetPets).toHaveBeenLastCalledWith(
      undefined,
      expect.objectContaining({ search: "ページ2" }),
    );
    expect(result.current.petPage.items).toEqual([pageTwoOnlyPet]);
    expect(result.current.petPage.startIndex).toBe(21);
    expect(result.current.petPage.endIndex).toBe(40);
  });

  // デバウンス待ちの間に選ばれたページ番号を新しい検索条件へ持ち越すと、
  // 「5件中 21-5件」のような反転した範囲や、総件数はあるのに一覧が空という
  // 自己矛盾した表示になる（BUG-451 と同じ「嘘の件数」クラス）。
  it("デバウンス待ち中のページ送りを新しい検索条件へ持ち越さない", () => {
    mockUseGetPets.mockReturnValue({ data: [], total: 100, page: 1, limit: 20 });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    // 旧条件で3ページ目へ移動して確定させる。
    act(() => {
      result.current.petPage.onPageChange(3);
    });
    expect(mockUseGetPets).toHaveBeenLastCalledWith(
      undefined,
      expect.objectContaining({ page: 3 }),
    );

    // 入力直後（デバウンス未確定）に、まだ旧条件由来の totalPages でページ送り。
    act(() => {
      result.current.setSearchParams({
        ...result.current.searchParams,
        search: "もも",
      });
    });
    act(() => {
      result.current.petPage.onPageChange(4);
    });

    // 新条件が確定した時点で先頭ページへ戻る（page:4 を持ち越さない）。
    act(() => {
      vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS);
    });

    expect(mockUseGetPets).toHaveBeenLastCalledWith(undefined, {
      includeDeceased: true,
      page: 1,
      limit: 20,
      search: "もも",
    });
  });

  it("デバウンス待ち中は確定済みでないことをisLoadingで伝える", () => {
    mockUseGetPets.mockReturnValue({
      data: [katakanaOwnerPet],
      total: 100,
      page: 1,
      limit: 20,
      isLoading: false,
    });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));
    expect(result.current.isLoading).toBe(false);

    act(() => {
      result.current.setSearchParams({
        ...result.current.searchParams,
        search: "もも",
      });
    });

    // 入力済みの条件と表示中の一覧が一致していない間は選択させない。
    expect(result.current.isLoading).toBe(true);

    act(() => {
      vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS);
    });
    expect(result.current.isLoading).toBe(false);
  });

  describe("デバウンス配線", () => {
    it("入力直後はbackendへ問い合わせず、停止後に1回だけ反映する", () => {
      const { result } = renderHook(() => usePetSelectionPage(CONFIG));
      mockUseGetPets.mockClear();

      act(() => {
        result.current.setSearchParams({
          ...result.current.searchParams,
          search: "も",
        });
      });

      // 入力直後: 検索語はまだ backend へ届いていない。
      expect(mockUseGetPets).not.toHaveBeenCalledWith(
        undefined,
        expect.objectContaining({ search: "も" }),
      );

      act(() => {
        vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS);
      });

      expect(mockUseGetPets).toHaveBeenLastCalledWith(undefined, {
        includeDeceased: true,
        page: 1,
        limit: 20,
        search: "も",
      });
    });

    it("連続入力では中間の検索語をbackendへ送らない", () => {
      const { result } = renderHook(() => usePetSelectionPage(CONFIG));

      act(() => {
        result.current.setSearchParams({
          ...result.current.searchParams,
          search: "も",
        });
      });
      act(() => {
        vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS - 100);
      });
      act(() => {
        result.current.setSearchParams({
          ...result.current.searchParams,
          search: "もも",
        });
      });
      act(() => {
        vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS);
      });

      const sentSearches = mockUseGetPets.mock.calls
        .map(([, options]) => options?.search)
        .filter((search): search is string => Boolean(search));
      expect(sentSearches).not.toContain("も");
      expect(sentSearches).toContain("もも");
    });
  });

  it("死亡個体は callback 境界で選択を拒否する", () => {
    const deceasedPet = { ...katakanaOwnerPet, status: "死亡" } as Pet;
    mockUseGetPets.mockReturnValue({ data: [deceasedPet] });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    act(() => {
      result.current.handleSelect(deceasedPet);
    });

    expect(navigate).not.toHaveBeenCalled();
  });

  it("ペット選択後も既存クエリと state を作成画面へ引き継ぐ", () => {
    mockUseGetPets.mockReturnValue({ data: [katakanaOwnerPet] });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    act(() => {
      result.current.handleSelect(katakanaOwnerPet);
    });

    expect(navigate).toHaveBeenCalledWith(
      "/trimming/new?appointmentId=88&visitDate=2026-05-29&petId=10",
      { state: locationState },
    );
  });

  it("生死不明の個体はfail-closedで選択を拒否する", () => {
    const unknownStatusPet = { ...katakanaOwnerPet, status: undefined } as Pet;
    mockUseGetPets.mockReturnValue({ data: [unknownStatusPet] });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    act(() => {
      result.current.handleSelect(unknownStatusPet);
    });

    expect(navigate).not.toHaveBeenCalled();
  });

  it("クリアで検索条件と現在ページを初期化する", () => {
    mockUseGetPets.mockReturnValue({ data: [], total: 100, page: 1, limit: 20 });
    const { result } = renderHook(() => usePetSelectionPage(CONFIG));

    applySearch(result, { search: "もも", species: "3" });
    act(() => {
      result.current.petPage.onPageChange(2);
    });
    act(() => {
      result.current.handleClear();
    });
    act(() => {
      vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS);
    });

    expect(result.current.searchParams).toEqual({
      search: "",
      ownerId: "",
      species: "",
    });
    expect(mockUseGetPets).toHaveBeenLastCalledWith(undefined, {
      includeDeceased: true,
      page: 1,
      limit: 20,
    });
  });
});
