import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { InterviewHistory } from "./InterviewHistory";
import type { InterviewHistoryItem } from "../types";

const ITEMS: InterviewHistoryItem[] = [
  { id: "1", date: "2026-01-01", type: "診察", title: "タイトルカタカナ", content: "本文内容", author: "田中" },
  { id: "2", date: "2026-01-02", type: "検査", title: "別のタイトル", content: "別の本文", author: "佐藤" },
];

describe("InterviewHistory — カナ混同検索", () => {
  it("過去カルテ検索inputに明示labelとid/nameを接続する", () => {
    render(<InterviewHistory historyItems={ITEMS} />);

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
    render(<InterviewHistory historyItems={ITEMS} />);
    expect(screen.getByText("タイトルカタカナ")).toBeInTheDocument();
    expect(screen.getByText("別のタイトル")).toBeInTheDocument();
  });

  it("ひらがなで検索するとカタカナのタイトルにヒットする", async () => {
    const user = userEvent.setup();
    render(<InterviewHistory historyItems={ITEMS} />);

    await user.type(screen.getByPlaceholderText("検索..."), "たいとるかたかな");

    await waitFor(() => {
      expect(screen.getByText("タイトルカタカナ")).toBeInTheDocument();
      expect(screen.queryByText("別のタイトル")).not.toBeInTheDocument();
    });
  });

  it("カタカナで検索するとひらがなのタイトルにヒットする", async () => {
    const hiraganaItems: InterviewHistoryItem[] = [
      { id: "1", date: "2026-01-01", type: "診察", title: "たいとるひらがな", content: "本文内容", author: "田中" },
      { id: "2", date: "2026-01-02", type: "検査", title: "別のタイトル", content: "別の本文", author: "佐藤" },
    ];
    const user = userEvent.setup();
    render(<InterviewHistory historyItems={hiraganaItems} />);

    await user.type(screen.getByPlaceholderText("検索..."), "タイトルヒラガナ");

    await waitFor(() => {
      expect(screen.getByText("たいとるひらがな")).toBeInTheDocument();
      expect(screen.queryByText("別のタイトル")).not.toBeInTheDocument();
    });
  });

  it("本文 (content) のカナ混同検索も効く", async () => {
    const katakanaContentItems: InterviewHistoryItem[] = [
      { id: "1", date: "2026-01-01", type: "診察", title: "タイトル1", content: "ホンブンカタカナ", author: "田中" },
      { id: "2", date: "2026-01-02", type: "検査", title: "タイトル2", content: "ベツノホンブン", author: "佐藤" },
    ];
    const user = userEvent.setup();
    render(<InterviewHistory historyItems={katakanaContentItems} />);

    await user.type(screen.getByPlaceholderText("検索..."), "ほんぶんかたかな");

    await waitFor(() => {
      expect(screen.getByText("タイトル1")).toBeInTheDocument();
      expect(screen.queryByText("タイトル2")).not.toBeInTheDocument();
    });
  });
});
