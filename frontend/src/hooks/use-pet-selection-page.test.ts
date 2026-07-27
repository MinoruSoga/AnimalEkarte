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
  error?: unknown;
  isLoading?: boolean;
}

const mockUseGetPets = vi.fn(
  (
    _ownerId?: string,
    _options?: { includeDeceased?: boolean },
  ): MockPetsQueryResult => ({
    data: [] as Pet[],
  }),
);

vi.mock("@/hooks/use-pet", () => ({
  useGetPets: (
    ownerId?: string,
    options?: { includeDeceased?: boolean },
  ) => mockUseGetPets(ownerId, options),
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
    mockUseGetPets.mockReturnValue({ data: [] });
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
    expect(result.current.filteredPets).toEqual([]);
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
    expect(result.current.filteredPets).toEqual([]);
  });

  it("取得成功時は error が無く isLoading も false になる", () => {
    mockUseGetPets.mockReturnValue({
      data: [katakanaOwnerPet],
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
    expect(result.current.filteredPets).toEqual([katakanaOwnerPet]);
  });

  it("死亡個体を含める option で共有一覧を取得する", () => {
    renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    expect(mockUseGetPets).toHaveBeenCalledWith(undefined, {
      includeDeceased: true,
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

  it("生死不明の個体は従来どおり選択できる", () => {
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

    expect(navigate).toHaveBeenCalledOnce();
  });

  describe("かな正規化フィルタ", () => {
    const config = { selectPath: "/trimming/new", backPath: "/trimming" };
    const pets = [katakanaOwnerPet, hiraganaOwnerPet];

    function setup() {
      mockUseGetPets.mockReturnValue({ data: pets });
      return renderHook(() => usePetSelectionPage(config));
    }

    it("ひらがな入力でカタカナ ownerName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, ownerName: "やまだ" });
      });
      expect(result.current.filteredPets).toEqual([katakanaOwnerPet]);
    });

    it("カタカナ入力でひらがな ownerName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, ownerName: "サトウ" });
      });
      expect(result.current.filteredPets).toEqual([hiraganaOwnerPet]);
    });

    it("ひらがな入力でカタカナ petName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, petName: "ぽち" });
      });
      expect(result.current.filteredPets).toEqual([katakanaOwnerPet]);
    });

    it("カタカナ入力でひらがな petName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, petName: "タロウ" });
      });
      expect(result.current.filteredPets).toEqual([hiraganaOwnerPet]);
    });
  });
});
