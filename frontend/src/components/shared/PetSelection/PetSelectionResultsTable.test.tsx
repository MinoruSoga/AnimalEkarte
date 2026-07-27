import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";
import { PetSelectionSearchForm } from "./PetSelectionSearchForm";
import { PetSelectionResultsTable } from "./PetSelectionResultsTable";

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: () => ({
    activeSpecies: [{ id: 3, name: "犬", label: "犬" }],
    isLoading: false,
    isError: false,
  }),
}));

const PET = {
  id: "pet-1",
  ownerId: "owner-1",
  ownerNumber: 1,
  ownerName: "山田 太郎",
  phone: "",
  name: "ポチ",
  species: "犬",
  gender: "雄",
} satisfies Pet;

const DANGER_REASON = "診察時に咬傷歴あり";

function renderHighDangerPet(dangerReason: string | undefined) {
  render(
    <PetSelectionResultsTable
      pets={createResults([{ ...PET, dangerLevel: "高", dangerReason }])}
      onSelect={vi.fn()}
    />,
  );

  return screen.getByRole("button", {
    name: "ポチの危険理由を表示",
  });
}

const EMPTY_MESSAGE = "該当するペットが見つかりません";
const ERROR_MESSAGE = "ペット一覧の取得に失敗しました";
const LOADING_MESSAGE = "ペット一覧を読み込み中です";

function createResults(
  items: Pet[],
  overrides: Partial<{
    totalCount: number;
    currentPage: number;
    totalPages: number;
    startIndex: number;
    endIndex: number;
    isPageLocalFiltered: boolean;
    onPageChange: (page: number) => void;
  }> = {},
) {
  return {
    items,
    totalCount: items.length,
    currentPage: 1,
    totalPages: 1,
    startIndex: items.length === 0 ? 0 : 1,
    endIndex: items.length,
    isPageLocalFiltered: false,
    onPageChange: vi.fn(),
    ...overrides,
  };
}

// 回帰防止: 7画面が共有する本コンポーネントは pets.length === 0 なら常に
// 「該当するペットが見つかりません」+「0件」を描画していたため、
// GET /v1/pets の失敗（400 等）が「該当0件」として利用者に嘘をついていた。
describe("PetSelectionResultsTable empty/error/loading states", () => {
  it("取得失敗時はエラー文言を出し、0件表示を一切しない", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([])}
        onSelect={vi.fn()}
        isError
      />,
    );

    expect(screen.getByText(ERROR_MESSAGE)).toBeInTheDocument();
    expect(screen.queryByText(EMPTY_MESSAGE)).not.toBeInTheDocument();
    expect(screen.queryByText("0件")).not.toBeInTheDocument();
  });

  it("読み込み中は0件でも空表示・件数を出さない", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([])}
        onSelect={vi.fn()}
        isLoading
      />,
    );

    expect(screen.getByText(LOADING_MESSAGE)).toBeInTheDocument();
    expect(screen.queryByText(EMPTY_MESSAGE)).not.toBeInTheDocument();
    expect(screen.queryByText("0件")).not.toBeInTheDocument();
  });

  it("エラーは読み込み中より優先して表示する", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([])}
        onSelect={vi.fn()}
        isError
        isLoading
      />,
    );

    expect(screen.getByText(ERROR_MESSAGE)).toBeInTheDocument();
    expect(screen.queryByText(LOADING_MESSAGE)).not.toBeInTheDocument();
  });

  it("prop 未指定なら従来どおり0件の空表示を維持する（既定の後方互換）", () => {
    render(
      <PetSelectionResultsTable pets={createResults([])} onSelect={vi.fn()} />,
    );

    expect(screen.getByText(EMPTY_MESSAGE)).toBeInTheDocument();
    expect(screen.getByText("0件")).toBeInTheDocument();
    expect(screen.queryByText(ERROR_MESSAGE)).not.toBeInTheDocument();
  });

  // 共有キャッシュ（staleTime: STATIC）により、前回取得分の行が残ったまま
  // 再取得だけが失敗する状態が起こりうる。行が出ているのに件数を隠すのは不親切。
  it("キャッシュ済みの行が残ったまま再取得が失敗した場合は件数を表示する", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([PET])}
        onSelect={vi.fn()}
        isError
      />,
    );

    expect(screen.getByText("1件")).toBeInTheDocument();
    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(screen.queryByText(EMPTY_MESSAGE)).not.toBeInTheDocument();
  });

  it("取得成功時は件数と行を表示しエラー文言を出さない", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([PET])}
        onSelect={vi.fn()}
      />,
    );

    expect(screen.getByText("1件中 1-1件")).toBeInTheDocument();
    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(screen.queryByText(ERROR_MESSAGE)).not.toBeInTheDocument();
    expect(screen.queryByText(EMPTY_MESSAGE)).not.toBeInTheDocument();
  });
});

describe("PetSelectionResultsTable server pagination", () => {
  it("backendの全体件数と現在範囲を表示し、次ページへ移動できる", async () => {
    const user = userEvent.setup();
    const onPageChange = vi.fn();
    render(
      <PetSelectionResultsTable
        pets={createResults([PET], {
          totalCount: 15_654,
          totalPages: 783,
          startIndex: 1,
          endIndex: 20,
          onPageChange,
        })}
        onSelect={vi.fn()}
      />,
    );

    expect(screen.getByText("15,654件中 1-20件")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "次のページ" }));
    expect(onPageChange).toHaveBeenCalledWith(2);
  });

  it("ページ内条件がある場合は全体検索ではないことを明示する", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([PET], {
          totalCount: 100,
          currentPage: 2,
          totalPages: 5,
          startIndex: 21,
          endIndex: 40,
          isPageLocalFiltered: true,
        })}
        onSelect={vi.fn()}
      />,
    );

    expect(
      screen.getByText("追加条件は現在のページ内だけを絞り込んでいます"),
    ).toBeInTheDocument();
    expect(screen.getByText("100件中 21-40件")).toBeInTheDocument();
  });
});

describe("PetSelectionSearchForm backend filters", () => {
  it("全体検索欄を表示し、動物種は数値IDで更新する", async () => {
    const user = userEvent.setup();
    const setSearchParams = vi.fn();
    render(
      <PetSelectionSearchForm
        searchParams={{
          search: "",
          ownerId: "",
          ownerName: "",
          ownerNameKana: "",
          phone: "",
          petName: "",
          petNameKana: "",
          species: "",
          address: "",
        }}
        setSearchParams={setSearchParams}
        onSearch={vi.fn()}
        onClear={vi.fn()}
      />,
    );

    expect(
      screen.getByRole("textbox", {
        name: "全体検索（ペット名・飼主名・よみ・電話）",
      }),
    ).toBeInTheDocument();

    await user.click(screen.getByRole("combobox", { name: "種別" }));
    await user.click(screen.getByRole("option", { name: "犬" }));
    expect(setSearchParams).toHaveBeenCalledWith(
      expect.objectContaining({ species: "3" }),
    );
  });
});

describe("PetSelectionResultsTable row actions", () => {
  it("危険度が高の個体だけ非色警告を表示する", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([
          { ...PET, id: "pet-high", name: "ハイ", dangerLevel: "高" },
          { ...PET, id: "pet-medium", name: "ミドル", dangerLevel: "中" },
          { ...PET, id: "pet-unknown", name: "未判定" },
        ])}
        onSelect={vi.fn()}
      />,
    );

    expect(screen.getAllByText("⚠ 危険")).toHaveLength(1);
    expect(screen.getByRole("button", { name: "ハイの危険理由を表示" })).toHaveClass(
      C.bgDanger10,
      C.danger,
      C.borderDanger20,
    );
    expect(screen.getByText("ハイ").parentElement).toHaveTextContent("⚠ 危険");
    expect(screen.getByText("ミドル").parentElement).not.toHaveTextContent(
      "⚠ 危険",
    );
    expect(screen.getByText("未判定").parentElement).not.toHaveTextContent(
      "⚠ 危険",
    );
  });

  it("保存済みの高危険度理由をclickで開き、同じtriggerの再clickで閉じる", async () => {
    const user = userEvent.setup();
    const trigger = renderHighDangerPet(DANGER_REASON);

    expect(trigger.tagName).toBe("BUTTON");
    expect(trigger).toHaveAttribute("type", "button");
    expect(trigger).toHaveAttribute("aria-expanded", "false");

    await user.click(trigger);
    expect(await screen.findByText(DANGER_REASON)).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(trigger).toHaveAttribute("aria-controls");

    await user.click(trigger);
    await waitFor(() => {
      expect(screen.queryByText(DANGER_REASON)).not.toBeInTheDocument();
    });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it.each([
    ["undefined", undefined],
    ["空文字", ""],
    ["空白のみ", "   "],
  ])(
    "高危険度理由が%sならclickで理由未登録を表示する",
    async (_caseName, dangerReason) => {
      const user = userEvent.setup();
      const trigger = renderHighDangerPet(dangerReason);

      await user.click(trigger);

      expect(await screen.findByText("理由未登録")).toBeInTheDocument();
    },
  );

  it.each([
    ["Enter", "{Enter}"],
    ["Space", " "],
  ])(
    "%sで高危険度理由を開き、同じキーで閉じる",
    async (_keyName, key) => {
      const user = userEvent.setup();
      const trigger = renderHighDangerPet(DANGER_REASON);

      trigger.focus();
      await user.keyboard(key);
      expect(await screen.findByText(DANGER_REASON)).toBeInTheDocument();

      trigger.focus();
      await user.keyboard(key);
      await waitFor(() => {
        expect(screen.queryByText(DANGER_REASON)).not.toBeInTheDocument();
      });
    },
  );

  it("行は選択せず、固有名の44px native buttonだけが選択する", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(
      <PetSelectionResultsTable
        pets={createResults([PET])}
        onSelect={onSelect}
      />,
    );

    await user.click(screen.getByText("山田 太郎"));
    expect(onSelect).not.toHaveBeenCalled();

    const selectButton = screen.getByRole("button", { name: "選択: ポチ (ID pet-1)" });
    expect(selectButton.tagName).toBe("BUTTON");
    expect(selectButton).toHaveClass("min-h-11", "min-w-11");
    await user.click(selectButton);
    expect(onSelect).toHaveBeenCalledWith(PET);
  });

  it("死亡個体は非色依存の名前で選択不可になる", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(
      <PetSelectionResultsTable
        pets={createResults([{ ...PET, status: "死亡" }])}
        onSelect={onSelect}
      />,
    );

    const selectButton = screen.getByRole("button", {
      name: "死亡・選択不可: ポチ (ID pet-1)",
    });
    expect(selectButton).toBeDisabled();
    expect(selectButton).toHaveTextContent("選択不可");
    expect(screen.getByText("死亡")).toBeInTheDocument();
    expect(selectButton.closest("tr")).toHaveClass(
      "opacity-60",
      "grayscale-[0.5]",
    );
    await user.click(selectButton);
    expect(onSelect).not.toHaveBeenCalled();
  });
});
