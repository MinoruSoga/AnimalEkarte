import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router";
import { InterviewHistory } from "./InterviewHistory";
import type { InterviewHistoryItem } from "../types";

const ITEMS: InterviewHistoryItem[] = [
  {
    id: "1",
    date: "2026-01-01",
    type: "診察",
    title: "タイトルカタカナ",
    content: "本文内容",
    author: "田中",
  },
  {
    id: "2",
    date: "2026-01-02",
    type: "検査",
    title: "別のタイトル",
    content: "別の本文",
    author: "佐藤",
  },
];

function renderHistory(historyItems: InterviewHistoryItem[]) {
  return render(
    <MemoryRouter>
      <InterviewHistory historyItems={historyItems} />
    </MemoryRouter>,
  );
}

function LocationPathname() {
  const location = useLocation();
  return <div>{location.pathname}</div>;
}

describe("InterviewHistory — カナ混同検索", () => {
  it("過去カルテ検索inputに明示labelとid/nameを接続する", () => {
    renderHistory(ITEMS);

    expect(screen.getByRole("textbox", { name: "過去のカルテを検索" })).toHaveAttribute(
      "id",
      "medical-record-history-search",
    );
    expect(screen.getByRole("textbox", { name: "過去のカルテを検索" })).toHaveAttribute(
      "name",
      "medicalRecordHistorySearch",
    );
  });

  it("空の検索語は全件表示する", () => {
    renderHistory(ITEMS);
    expect(screen.getByText("タイトルカタカナ")).toBeInTheDocument();
    expect(screen.getByText("別のタイトル")).toBeInTheDocument();
  });

  it("ひらがなで検索するとカタカナのタイトルにヒットする", async () => {
    const user = userEvent.setup();
    renderHistory(ITEMS);

    await user.type(screen.getByPlaceholderText("検索..."), "たいとるかたかな");

    await waitFor(() => {
      expect(screen.getByText("タイトルカタカナ")).toBeInTheDocument();
      expect(screen.queryByText("別のタイトル")).not.toBeInTheDocument();
    });
  });

  it("カタカナで検索するとひらがなのタイトルにヒットする", async () => {
    const hiraganaItems: InterviewHistoryItem[] = [
      {
        id: "1",
        date: "2026-01-01",
        type: "診察",
        title: "たいとるひらがな",
        content: "本文内容",
        author: "田中",
      },
      {
        id: "2",
        date: "2026-01-02",
        type: "検査",
        title: "別のタイトル",
        content: "別の本文",
        author: "佐藤",
      },
    ];
    const user = userEvent.setup();
    renderHistory(hiraganaItems);

    await user.type(screen.getByPlaceholderText("検索..."), "タイトルヒラガナ");

    await waitFor(() => {
      expect(screen.getByText("たいとるひらがな")).toBeInTheDocument();
      expect(screen.queryByText("別のタイトル")).not.toBeInTheDocument();
    });
  });

  it("本文 (content) のカナ混同検索も効く", async () => {
    const katakanaContentItems: InterviewHistoryItem[] = [
      {
        id: "1",
        date: "2026-01-01",
        type: "診察",
        title: "タイトル1",
        content: "ホンブンカタカナ",
        author: "田中",
      },
      {
        id: "2",
        date: "2026-01-02",
        type: "検査",
        title: "タイトル2",
        content: "ベツノホンブン",
        author: "佐藤",
      },
    ];
    const user = userEvent.setup();
    renderHistory(katakanaContentItems);

    await user.type(screen.getByPlaceholderText("検索..."), "ほんぶんかたかな");

    await waitFor(() => {
      expect(screen.getByText("タイトル1")).toBeInTheDocument();
      expect(screen.queryByText("タイトル2")).not.toBeInTheDocument();
    });
  });
});

describe("InterviewHistory — 過去行の詳細遷移", () => {
  it("過去行は /medical-records/:id へリンクし、引用ボタンは無い", () => {
    renderHistory(ITEMS);

    expect(screen.getByRole("link", { name: /タイトルカタカナ/ })).toHaveAttribute(
      "href",
      "/medical-records/1",
    );
    expect(screen.getByRole("link", { name: /別のタイトル/ })).toHaveAttribute(
      "href",
      "/medical-records/2",
    );
    expect(screen.queryByRole("button", { name: "引用" })).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "問診抜粋" })).toBeInTheDocument();
    expect(screen.getByText("全文は詳細で確認できます")).toBeInTheDocument();
  });

  it("treatments が無い移行カルテでも詳細自体を開く", async () => {
    const user = userEvent.setup();
    const migratedWithoutTreatments: InterviewHistoryItem = {
      id: "rec-empty-treatments",
      date: "2026-03-01",
      type: "診察",
      title: "治療タブ空の移行カルテ",
      content: "問診のみ",
      author: "田中",
    };

    render(
      <MemoryRouter initialEntries={["/medical-records/current"]}>
        <Routes>
          <Route
            path="/medical-records/current"
            element={<InterviewHistory historyItems={[migratedWithoutTreatments]} />}
          />
          <Route path="/medical-records/:id" element={<LocationPathname />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.queryByRole("button", { name: "引用" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("link", { name: /治療タブ空の移行カルテ/ }));
    expect(await screen.findByText("/medical-records/rec-empty-treatments")).toBeInTheDocument();
  });

  it("呼び出し側の 50 件上限をさらに切らず、各行が詳細へ開く", () => {
    const items: InterviewHistoryItem[] = Array.from({ length: 50 }, (_, index) => ({
      id: String(index + 1),
      date: "2026-01-01",
      type: "診察",
      title: `履歴${index + 1}`,
      content: "抜粋",
      author: "田中",
    }));

    renderHistory(items);

    expect(screen.getAllByRole("link")).toHaveLength(50);
    expect(screen.getByRole("link", { name: /履歴50/ })).toHaveAttribute(
      "href",
      "/medical-records/50",
    );
    expect(screen.queryByRole("button", { name: "引用" })).not.toBeInTheDocument();
  });

  it("空状態は問診抜粋であり全文は詳細にあると示す", () => {
    renderHistory([]);

    expect(screen.getByRole("heading", { name: "問診抜粋" })).toBeInTheDocument();
    expect(screen.getByText(/問診抜粋/)).toBeInTheDocument();
    expect(screen.getByText(/全文は詳細/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "引用" })).not.toBeInTheDocument();
  });
});

const COPYABLE_ITEMS: InterviewHistoryItem[] = [
  {
    id: "10",
    date: "2026-02-01",
    type: "再診",
    title: "コピー元カルテ",
    content: "元気がない",
    author: "田中",
    copySource: {
      chiefComplaint: "前回の主訴詳細",
      treatmentPolicy: "前回の治療方針",
      chiefComplaintTypeId: 7,
    },
  },
  {
    id: "11",
    date: "2026-02-02",
    type: "初診",
    title: "コピー不可カルテ",
    content: "問診なし",
    author: "佐藤",
  },
];

function renderCopyable(onCopyItem?: (item: InterviewHistoryItem) => void) {
  return render(
    <MemoryRouter>
      <InterviewHistory historyItems={COPYABLE_ITEMS} onCopyItem={onCopyItem} />
    </MemoryRouter>,
  );
}

describe("InterviewHistory — 前回複写（コピー）", () => {
  it("copySource を持つ行にのみ コピー ボタンを表示する", () => {
    renderCopyable(vi.fn());

    expect(screen.getAllByRole("button", { name: "コピー" })).toHaveLength(1);
  });

  it("コピー はリンク内にネストしない button で、行リンクは維持される", () => {
    renderCopyable(vi.fn());

    const link = screen.getByRole("link", { name: /コピー元カルテ/ });
    expect(link).toHaveAttribute("href", "/medical-records/10");
    const copyButton = screen.getByRole("button", { name: "コピー" });
    expect(copyButton.tagName).toBe("BUTTON");
    expect(link.contains(copyButton)).toBe(false);
    expect(screen.getByRole("link", { name: /コピー不可カルテ/ })).toHaveAttribute(
      "href",
      "/medical-records/11",
    );
  });

  it("コピー クリックで onCopyItem に行アイテムを渡し、詳細へ遷移しない", async () => {
    const user = userEvent.setup();
    const onCopyItem = vi.fn();
    renderCopyable(onCopyItem);

    await user.click(screen.getByRole("button", { name: "コピー" }));

    expect(onCopyItem).toHaveBeenCalledTimes(1);
    expect(onCopyItem).toHaveBeenCalledWith(COPYABLE_ITEMS[0]);
    expect(screen.getByRole("heading", { name: "問診抜粋" })).toBeInTheDocument();
  });

  it("disabled fieldset 内では コピー が押せないが行リンクは有効", () => {
    render(
      <MemoryRouter>
        <fieldset disabled>
          <InterviewHistory historyItems={COPYABLE_ITEMS} onCopyItem={vi.fn()} />
        </fieldset>
      </MemoryRouter>,
    );

    expect(screen.getByRole("button", { name: "コピー" })).toBeDisabled();
    expect(screen.getByRole("link", { name: /コピー元カルテ/ })).toHaveAttribute(
      "href",
      "/medical-records/10",
    );
  });

  it("copySource の無い行には コピー を表示しない（既定fixture相当）", () => {
    renderHistory(ITEMS);

    expect(screen.queryByRole("button", { name: "コピー" })).not.toBeInTheDocument();
  });
});
