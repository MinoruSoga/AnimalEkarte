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
  status: "生存",
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
    onPageChange: vi.fn(),
    ...overrides,
  };
}

// 回帰防止: 7画面が共有する本コンポーネントは pets.length === 0 なら常に
// 「該当するペットが見つかりません」+「0件」を描画していたため、
// GET /v1/pets の失敗（400 等）が「該当0件」として利用者に嘘をついていた。
describe("PetSelectionResultsTable empty/error/loading states", () => {
  it("取得失敗時はエラー文言を出し、0件表示を一切しない", () => {
    render(<PetSelectionResultsTable pets={createResults([])} onSelect={vi.fn()} isError />);

    expect(screen.getByText(ERROR_MESSAGE)).toBeInTheDocument();
    expect(screen.queryByText(EMPTY_MESSAGE)).not.toBeInTheDocument();
    expect(screen.queryByText("0件")).not.toBeInTheDocument();
  });

  it("読み込み中は0件でも空表示・件数を出さない", () => {
    render(<PetSelectionResultsTable pets={createResults([])} onSelect={vi.fn()} isLoading />);

    expect(screen.getByText(LOADING_MESSAGE)).toBeInTheDocument();
    expect(screen.queryByText(EMPTY_MESSAGE)).not.toBeInTheDocument();
    expect(screen.queryByText("0件")).not.toBeInTheDocument();
  });

  it("エラーは読み込み中より優先して表示する", () => {
    render(
      <PetSelectionResultsTable pets={createResults([])} onSelect={vi.fn()} isError isLoading />,
    );

    expect(screen.getByText(ERROR_MESSAGE)).toBeInTheDocument();
    expect(screen.queryByText(LOADING_MESSAGE)).not.toBeInTheDocument();
  });

  it("prop 未指定なら従来どおり0件の空表示を維持する（既定の後方互換）", () => {
    render(<PetSelectionResultsTable pets={createResults([])} onSelect={vi.fn()} />);

    expect(screen.getByText(EMPTY_MESSAGE)).toBeInTheDocument();
    expect(screen.getByText("0件")).toBeInTheDocument();
    expect(screen.queryByText(ERROR_MESSAGE)).not.toBeInTheDocument();
  });

  it("キャッシュ済み行が残る再取得失敗はエラーと前回範囲を示し、選択を拒否する", () => {
    const onSelect = vi.fn();
    render(
      <PetSelectionResultsTable
        pets={createResults([PET], {
          totalCount: 15_654,
          totalPages: 783,
          startIndex: 1,
          endIndex: 20,
        })}
        onSelect={onSelect}
        isError
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent(ERROR_MESSAGE);
    expect(screen.getByText("前回取得: 15,654件中 1-20件")).toBeInTheDocument();
    expect(screen.queryByText("1件")).not.toBeInTheDocument();
    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(screen.queryByText(EMPTY_MESSAGE)).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", {
        name: "取得失敗・選択不可: ポチ (ID pet-1)",
      }),
    ).toBeDisabled();
    expect(onSelect).not.toHaveBeenCalled();
  });

  it("取得成功時は件数と行を表示しエラー文言を出さない", () => {
    render(<PetSelectionResultsTable pets={createResults([PET])} onSelect={vi.fn()} />);

    expect(screen.getByText("1件中 1-1件")).toBeInTheDocument();
    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(screen.queryByText(ERROR_MESSAGE)).not.toBeInTheDocument();
    expect(screen.queryByText(EMPTY_MESSAGE)).not.toBeInTheDocument();
  });

  it("ページ切替中の前回行は範囲を明示して選択を拒否し、paginationを保つ", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([PET], {
          totalCount: 100,
          totalPages: 5,
          startIndex: 1,
          endIndex: 20,
        })}
        onSelect={vi.fn()}
        isLoading
      />,
    );

    expect(screen.getByText("前回取得: 100件中 1-20件")).toBeInTheDocument();
    expect(
      screen.getByRole("button", {
        name: "読み込み中・選択不可: ポチ (ID pet-1)",
      }),
    ).toBeDisabled();
    // ページ送りは読み込み中も操作可能なまま保つ。フォーカス中のボタンを
    // disabled にするとブラウザが強制 blur し、ページ送りのたびに
    // キーボード/スクリーンリーダの焦点が失われるため（aria-busy で伝える）。
    const nextButton = screen.getByRole("button", { name: "次のページ" });
    expect(nextButton).toBeEnabled();
    expect(nextButton.closest("fieldset")).toHaveAttribute("aria-busy", "true");
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

    // 多ページ時は Pagination が可視表示し、常設の status 領域が読み上げを担う。
    expect(screen.getAllByText("15,654件中 1-20件")).toHaveLength(2);
    expect(screen.getByRole("status")).toHaveTextContent("15,654件中 1-20件");
    await user.click(screen.getByRole("button", { name: "次のページ" }));
    expect(onPageChange).toHaveBeenCalledWith(2);
  });

  // 単一ページへ絞り込まれた検索（患者を狙って探す最も普通のケース）でも
  // 件数変化が読み上げられること。live region を出し入れすると無通知になる。
  it("単一ページの結果でも常設のstatus領域が件数を告知する", () => {
    const { rerender } = render(
      <PetSelectionResultsTable
        pets={createResults([PET], {
          totalCount: 15_654,
          totalPages: 783,
          startIndex: 1,
          endIndex: 20,
        })}
        onSelect={vi.fn()}
      />,
    );
    const liveRegion = screen.getByRole("status");

    rerender(
      <PetSelectionResultsTable
        pets={createResults([PET], { totalCount: 1, totalPages: 1 })}
        onSelect={vi.fn()}
      />,
    );

    // 同一ノードが残り、テキストだけが差し替わる（remount していない）。
    expect(screen.getByRole("status")).toBe(liveRegion);
    expect(liveRegion).toHaveTextContent("1件中 1-1件");
  });

  // 応答に total が無い等で総件数が不明なまま行がある場合、
  // 「0件」と言いながら行を描くのは BUG-451 と同じ「嘘の件数」である。
  it("総件数が不明なまま行があるときは件数を主張しない", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([PET], {
          totalCount: 0,
          totalPages: 1,
          startIndex: 0,
          endIndex: 0,
        })}
        onSelect={vi.fn()}
      />,
    );

    expect(screen.queryByText("0件")).not.toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("");
    expect(screen.getByText("ポチ")).toBeInTheDocument();
  });

  it("キーボードで次ページへ移動してもfocusを維持し、更新範囲を通知する", async () => {
    const user = userEvent.setup();
    const onPageChange = vi.fn();
    const { rerender } = render(
      <PetSelectionResultsTable
        pets={createResults([PET], {
          totalCount: 100,
          totalPages: 5,
          startIndex: 1,
          endIndex: 20,
          onPageChange,
        })}
        onSelect={vi.fn()}
      />,
    );
    const nextButton = screen.getByRole("button", { name: "次のページ" });
    nextButton.focus();

    await user.keyboard("{Enter}");
    expect(onPageChange).toHaveBeenCalledWith(2);

    rerender(
      <PetSelectionResultsTable
        pets={createResults([PET], {
          totalCount: 100,
          currentPage: 2,
          totalPages: 5,
          startIndex: 21,
          endIndex: 40,
          onPageChange,
        })}
        onSelect={vi.fn()}
      />,
    );

    expect(document.activeElement).toBe(nextButton);
    expect(screen.getByRole("status")).toHaveTextContent("100件中 21-40件");
  });

  // BUG-451 回帰防止: 総件数を根拠に「N件中 X-Y件」と提示する以上、
  // 描画行はその範囲の中身そのものでなければならない。
  // ページ内クライアント側フィルタが復活すると件数と行数が食い違い、
  // 利用者は「全件を検索した」と誤解する。
  it("総件数の提示と描画行が食い違わない（ページ内絞り込みの注記を持たない）", () => {
    const pageItems = [
      PET,
      { ...PET, id: "pet-2", name: "タマ" },
      { ...PET, id: "pet-3", name: "ハチ" },
    ];
    render(
      <PetSelectionResultsTable
        pets={createResults(pageItems, {
          totalCount: 15_654,
          currentPage: 2,
          totalPages: 783,
          startIndex: 21,
          endIndex: 40,
        })}
        onSelect={vi.fn()}
      />,
    );

    expect(screen.getByRole("status")).toHaveTextContent("15,654件中 21-40件");
    // backend が返した行は1件も間引かれない。
    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(screen.getByText("タマ")).toBeInTheDocument();
    expect(screen.getByText("ハチ")).toBeInTheDocument();
    expect(
      screen.queryByText("追加条件は現在のページ内だけを絞り込んでいます"),
    ).not.toBeInTheDocument();
  });
});

const EMPTY_SEARCH_PARAMS = { search: "", ownerId: "", species: "" };

describe("PetSelectionSearchForm backend filters", () => {
  it("全体検索欄を表示し、動物種は数値IDで更新する", async () => {
    const user = userEvent.setup();
    const setSearchParams = vi.fn();
    render(
      <PetSelectionSearchForm
        searchParams={EMPTY_SEARCH_PARAMS}
        setSearchParams={setSearchParams}
        onClear={vi.fn()}
      />,
    );

    expect(
      screen.getByRole("textbox", {
        name: "検索（ペット名・飼主名・よみ・電話）",
      }),
    ).toBeInTheDocument();

    await user.click(screen.getByRole("combobox", { name: "種別" }));
    await user.click(screen.getByRole("option", { name: "犬" }));
    expect(setSearchParams).toHaveBeenCalledWith(expect.objectContaining({ species: "3" }));
  });

  // BUG-451: backend が述語を持たない条件を欄として出すと、利用者は
  // 「その条件で全件を検索した」と誤解する。backend の述語と1対1の
  // 3コントロールだけを出す（住所は述語が無いため機能ごと廃止）。
  it("backendの述語に対応する3コントロールだけを表示する", () => {
    render(
      <PetSelectionSearchForm
        searchParams={EMPTY_SEARCH_PARAMS}
        setSearchParams={vi.fn()}
        onClear={vi.fn()}
      />,
    );

    expect(screen.getAllByRole("textbox")).toHaveLength(2);
    expect(screen.getByRole("textbox", { name: "飼主No" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "種別" })).toBeInTheDocument();

    for (const removed of [
      "飼主名",
      "飼主名よみ",
      "電話番号",
      "ペット名",
      "ペット名よみ",
      "住所",
    ]) {
      expect(screen.queryByRole("textbox", { name: removed })).not.toBeInTheDocument();
    }
  });

  // 決定③: デバウンス付き自動検索。押しても何も起きない「検索」ボタンを残さない。
  it("no-opな検索ボタンを持たず、自動検索であることを伝える", () => {
    render(
      <PetSelectionSearchForm
        searchParams={EMPTY_SEARCH_PARAMS}
        setSearchParams={vi.fn()}
        onClear={vi.fn()}
      />,
    );

    expect(screen.queryByRole("button", { name: "検索" })).not.toBeInTheDocument();
    expect(screen.getByText("入力すると自動で検索します")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "クリア" })).toBeInTheDocument();
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
    expect(screen.getByText("ミドル").parentElement).not.toHaveTextContent("⚠ 危険");
    expect(screen.getByText("未判定").parentElement).not.toHaveTextContent("⚠ 危険");
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
  ])("高危険度理由が%sならclickで理由未登録を表示する", async (_caseName, dangerReason) => {
    const user = userEvent.setup();
    const trigger = renderHighDangerPet(dangerReason);

    await user.click(trigger);

    expect(await screen.findByText("理由未登録")).toBeInTheDocument();
  });

  it.each([
    ["Enter", "{Enter}"],
    ["Space", " "],
  ])("%sで高危険度理由を開き、同じキーで閉じる", async (_keyName, key) => {
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
  });

  it("行は選択せず、固有名の44px native buttonだけが選択する", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(<PetSelectionResultsTable pets={createResults([PET])} onSelect={onSelect} />);

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
    expect(selectButton.closest("tr")).toHaveClass("opacity-60", "grayscale-[0.5]");
    await user.click(selectButton);
    expect(onSelect).not.toHaveBeenCalled();
  });

  it("生死不明の個体もfail-closedで選択不可になる", () => {
    render(
      <PetSelectionResultsTable
        pets={createResults([{ ...PET, status: undefined }])}
        onSelect={vi.fn()}
      />,
    );

    expect(
      screen.getByRole("button", {
        name: "状態不明・選択不可: ポチ (ID pet-1)",
      }),
    ).toBeDisabled();
  });
});
