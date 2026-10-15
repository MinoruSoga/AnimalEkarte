import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { todayJSTISO } from "@/lib/jst-date";
import { addYearsIso } from "./next-visit-intent-date";

import { NextVisitResetDialog } from "./NextVisitResetDialog";

/**
 * EMR-308: 「次回予定を再設定」共有ダイアログ。
 * キャンセル後オプション/期限超過パネルの再設定から開く。自動作成はせず、
 * スタッフが日付を選んで確定した時だけ onSubmit（PUT intent）を呼ぶ。
 */

// 意図日の上限は「今日から2年」（validateNextVisitIntentDate）— 上限内の有効日付を使う。
const VALID_DATE = addYearsIso(todayJSTISO(), 1);

function renderDialog(overrides: Partial<Parameters<typeof NextVisitResetDialog>[0]> = {}) {
  const onSubmit = overrides.onSubmit ?? vi.fn().mockResolvedValue(undefined);
  const onClose = overrides.onClose ?? vi.fn();
  render(
    <NextVisitResetDialog
      open={true}
      petName="ポチ"
      previousIntendedDate="2026-09-01"
      onSubmit={onSubmit}
      onClose={onClose}
      {...overrides}
    />,
  );
  return { onSubmit, onClose };
}

describe("NextVisitResetDialog (EMR-308)", () => {
  it("ペット名・前回意図日を説明文に表示する", () => {
    renderDialog();
    expect(screen.getByText(/ポチちゃんの次回予定（前回: 2026-09-01）/)).toBeInTheDocument();
  });

  it("日付未選択では onSubmit を呼ばずバリデーションエラーを出す", async () => {
    const user = userEvent.setup();
    const { onSubmit } = renderDialog();

    await user.click(screen.getByRole("button", { name: "設定する" }));

    expect(onSubmit).not.toHaveBeenCalled();
    expect(screen.getByText("日付を選択してください")).toBeInTheDocument();
  });

  it("有効日付で onSubmit を呼び、成功したら onClose で閉じる", async () => {
    const user = userEvent.setup();
    const { onSubmit, onClose } = renderDialog();

    const dateInput = screen.getByRole("textbox", { name: "予定日" });
    await user.clear(dateInput);
    await user.type(dateInput, VALID_DATE.replaceAll("-", "/"));
    await user.tab();
    await user.click(screen.getByRole("button", { name: "設定する" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledWith(VALID_DATE));
    await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
  });

  it("onSubmit 失敗時はダイアログを閉じずインラインエラーを出す", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockRejectedValue(new Error("api failed"));
    const { onClose } = renderDialog({ onSubmit });

    const dateInput = screen.getByRole("textbox", { name: "予定日" });
    await user.clear(dateInput);
    await user.type(dateInput, VALID_DATE.replaceAll("-", "/"));
    await user.tab();
    await user.click(screen.getByRole("button", { name: "設定する" }));

    await waitFor(() =>
      expect(
        screen.getByText("次回予定の再設定に失敗しました。日付を確認して再度お試しください。"),
      ).toBeInTheDocument(),
    );
    // キャンセル自体は完了済み — 再設定失敗で親を閉じない
    expect(onClose).not.toHaveBeenCalled();
  });

  it("「設定しない」は onSubmit を呼ばずに閉じる", async () => {
    const user = userEvent.setup();
    const { onSubmit, onClose } = renderDialog();

    await user.click(screen.getByRole("button", { name: "設定しない" }));

    expect(onSubmit).not.toHaveBeenCalled();
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("isPending 中は入力・両ボタンを無効化しラベルを「設定中...」にする（mutation.isPending 由来）", () => {
    renderDialog({ isPending: true });

    expect(screen.getByLabelText("予定日")).toBeDisabled();
    expect(screen.getByRole("button", { name: "設定しない" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "設定中..." })).toBeDisabled();
  });
});
