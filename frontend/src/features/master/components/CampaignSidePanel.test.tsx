import type { ReactNode } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { CampaignSidePanel } from "./CampaignSidePanel";
import { useGetAllMerchandiseItems } from "../api/merchandise-items";

async function enterDate(user: ReturnType<typeof userEvent.setup>, label: string, value: string) {
  const input = screen.getByRole("textbox", { name: label });
  await user.clear(input);
  await user.type(input, value.replaceAll("-", "/"));
  await user.tab();
}

vi.mock("@/components/shared/SidePeek", () => ({
  MasterSidePanel: ({
    children,
    onTitleChange,
    onSave,
    titleError,
  }: {
    children: ReactNode;
    onTitleChange?: (value: string) => void;
    onSave?: () => void;
    titleError?: string;
  }) => (
    <div>
      <input
        aria-label="名称"
        aria-invalid={titleError ? true : undefined}
        onChange={(e) => onTitleChange?.(e.target.value)}
      />
      {titleError ? <p role="alert">{titleError}</p> : null}
      <button type="button" onClick={() => onSave?.()}>
        保存
      </button>
      {children}
    </div>
  ),
  StatusToggleButton: () => null,
  PropertyRow: ({ label, children }: { label?: ReactNode; children: ReactNode }) => (
    <div>
      {label}
      {children}
    </div>
  ),
}));

vi.mock("../api/merchandise-items", () => ({
  useGetAllMerchandiseItems: vi.fn(() => ({ data: [] })),
}));

describe("CampaignSidePanel", () => {
  it("paginates 101 targets and keeps selections across pages and search", async () => {
    const user = userEvent.setup();
    const items = Array.from({ length: 101 }, (_, index) => ({
      id: String(index + 1),
      name: `合成商品${index + 1}`,
      category: "goods" as const,
      unitPrice: 1100,
      taxType: "included" as const,
      taxRate: 0.1,
      isActive: true,
      sortOrder: index,
    }));
    vi.mocked(useGetAllMerchandiseItems, { partial: true }).mockReturnValue({
      data: items,
    });
    try {
      render(<CampaignSidePanel item={null} onClose={vi.fn()} onSave={() => true} />);
      expect(screen.getAllByRole("checkbox", { name: /^合成商品/ })).toHaveLength(50);
      await user.click(screen.getByRole("checkbox", { name: "合成商品1" }));
      await user.click(screen.getByRole("button", { name: "2" }));
      await user.click(screen.getByRole("checkbox", { name: "合成商品51" }));
      await user.click(screen.getByRole("button", { name: "1" }));
      expect(screen.getByRole("checkbox", { name: "合成商品1" })).toBeChecked();
      const search = screen.getByPlaceholderText("商品名で検索...");
      await user.type(search, "合成商品51");
      expect(screen.getByRole("checkbox", { name: "合成商品51" })).toBeChecked();
      await user.clear(search);
      await user.click(screen.getByRole("button", { name: "2" }));
      expect(screen.getByRole("checkbox", { name: "合成商品51" })).toBeChecked();
    } finally {
      vi.mocked(useGetAllMerchandiseItems, { partial: true }).mockReturnValue({
        data: [],
      });
    }
  });
  // e93ac185b で responsive grid から PropertyRow 行構成へ移行済み。
  // 行入力の存在と、対象カテゴリの単一列 grid を現行仕様として固定する。
  it("PropertyRow 行構成で開始日・割引種別・対象カテゴリを描画する", () => {
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={() => true} />);

    expect(screen.getByLabelText("開始日")).toBeInTheDocument();
    expect(screen.getByLabelText("終了日")).toBeInTheDocument();
    expect(screen.getByLabelText("割引種別")).toBeInTheDocument();

    const categoryGrid = screen.getByText("フード").closest(".grid");
    expect(categoryGrid).toHaveClass("w-full", "grid-cols-1");
  });

  it("readOnly でも DatePicker の値とエラー関連付けを保ち、操作を無効化する", async () => {
    const user = userEvent.setup();
    const item = {
      id: "campaign-1",
      name: "既存キャンペーン",
      startDate: "2026-01-01",
      endDate: "2026-01-31",
      discountType: "rate" as const,
      discountValue: "10",
      isActive: true,
      sortOrder: 1,
      targetCategories: [],
      targetItemIds: [],
    };
    const onSave = vi.fn(() => true);
    const onDirtyChange = vi.fn();
    const { rerender } = render(
      <CampaignSidePanel
        item={item}
        onClose={() => {}}
        onSave={onSave}
        onDirtyChange={onDirtyChange}
      />,
    );

    await enterDate(user, "開始日", "2026-02-01");
    await user.click(screen.getByRole("button", { name: "保存" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("終了日は開始日以降にしてください");
    const dirtyCallsBeforeReadOnly = onDirtyChange.mock.calls.length;

    rerender(
      <CampaignSidePanel
        item={item}
        onClose={() => {}}
        onSave={onSave}
        onDirtyChange={onDirtyChange}
        readOnly
      />,
    );
    for (const [label, value] of [
      ["開始日", "2026年2月1日"],
      ["終了日", "2026年1月31日"],
    ]) {
      const input = screen.getByRole<HTMLInputElement>("textbox", { name: label });
      expect(input).toBeDisabled();
      expect(input.value).toContain(value);
      expect(input).toHaveAttribute("aria-invalid", "true");
      expect(input).toHaveAttribute("aria-describedby", "campaign-period-error");
      expect(screen.getByRole("button", { name: `${label}を選択` })).toBeDisabled();
      expect(screen.getAllByRole("button", { name: "日付をクリア" })).toHaveLength(2);
      expect(
        screen
          .getAllByRole("button", { name: "日付をクリア" })
          .every((button) => button.hasAttribute("disabled")),
      ).toBe(true);
    }
    expect(screen.queryByRole("button", { name: "Today" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "開始日を選択" }));
    expect(screen.queryByRole("button", { name: "Today" })).not.toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
    expect(onDirtyChange).toHaveBeenCalledTimes(dirtyCallsBeforeReadOnly);
  });

  // EMR-319: 開始日・終了日の必須エラーで両 date Input に aria-invalid+describedby。
  it("期間未入力で保存すると開始日・終了日が aria-invalid で共通エラー文を指す", async () => {
    const user = userEvent.setup();
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={() => true} />);

    await user.type(screen.getByLabelText("名称"), "テストキャンペーン");
    await user.click(screen.getByRole("button", { name: "保存" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("開始日・終了日を入力してください");
    expect(alert).toHaveAttribute("id", "campaign-period-error");

    for (const label of ["開始日", "終了日"]) {
      const input = screen.getByLabelText(label);
      expect(input).toHaveAttribute("aria-invalid", "true");
      expect(input).toHaveAttribute("aria-describedby", "campaign-period-error");
    }
  });

  // PR#614 指摘: 片側欠落では入力済み側を invalid にしない（per-field 契約）。
  it("開始日のみ未入力で保存すると開始日だけが invalid になる", async () => {
    const user = userEvent.setup();
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={() => true} />);

    await user.type(screen.getByLabelText("名称"), "テストキャンペーン");
    await enterDate(user, "終了日", "2026-01-31");
    await user.click(screen.getByRole("button", { name: "保存" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("開始日・終了日を入力してください");
    expect(screen.getByLabelText("開始日")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByLabelText("終了日")).not.toHaveAttribute("aria-invalid");
  });

  it("終了日のみ未入力で保存すると終了日だけが invalid になり、入力で解除される", async () => {
    const user = userEvent.setup();
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={() => true} />);

    await user.type(screen.getByLabelText("名称"), "テストキャンペーン");
    await enterDate(user, "開始日", "2026-01-01");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("開始日・終了日を入力してください");
    expect(screen.getByLabelText("終了日")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByLabelText("開始日")).not.toHaveAttribute("aria-invalid");

    await enterDate(user, "終了日", "2026-01-31");
    expect(screen.getByLabelText("終了日")).not.toHaveAttribute("aria-invalid");
  });

  it("終了日 < 開始日 のペア課題では両欄が invalid になる", async () => {
    const user = userEvent.setup();
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={() => true} />);

    await user.type(screen.getByLabelText("名称"), "テストキャンペーン");
    await enterDate(user, "開始日", "2026-01-31");
    await enterDate(user, "終了日", "2026-01-01");
    await user.click(screen.getByRole("button", { name: "保存" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("終了日は開始日以降にしてください");
    expect(screen.getByLabelText("開始日")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByLabelText("終了日")).toHaveAttribute("aria-invalid", "true");
  });

  // PR#614 R2-3: 名称・開始日・終了日が全て空の初回保存では、独立した必須群を
  // 同一パスで評価して名称欄と両日付欄の invalid+エラー文が同時に立つ。
  it("全必須項目が空の初回保存で名称・開始日・終了日が同時に invalid になり保存は拒否される", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => true);
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={onSave} />);

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("名称を入力してください")).toBeInTheDocument();
    expect(screen.getByLabelText("名称")).toHaveAttribute("aria-invalid", "true");

    const periodAlert = screen.getByText("開始日・終了日を入力してください");
    expect(periodAlert).toHaveAttribute("id", "campaign-period-error");
    for (const label of ["開始日", "終了日"]) {
      const input = screen.getByLabelText(label);
      expect(input).toHaveAttribute("aria-invalid", "true");
      expect(input).toHaveAttribute("aria-describedby", "campaign-period-error");
    }
    expect(onSave).not.toHaveBeenCalled();
  });

  // PR#614 R3 P2: 名称空 + 順序不正で保存失敗後、名称空のまま期間だけ訂正して
  // 再保存すると、期間は正常なので両日付の aria-invalid/赤枠/期間エラー文は
  // 解除され、名称の必須エラー表示だけが残る（期間クリアが名称失敗に連動しない）。
  it("名称空のまま期間を訂正して再保存すると期間エラーだけ解除され名称エラーが残る", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => true);
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={onSave} />);

    await user.type(screen.getByLabelText("開始日"), "2026-01-31");
    await user.type(screen.getByLabelText("終了日"), "2026-01-01");
    await user.click(screen.getByRole("button", { name: "保存" }));

    const orderAlert = await screen.findByText("終了日は開始日以降にしてください");
    expect(orderAlert).toHaveAttribute("id", "campaign-period-error");
    expect(screen.getByText("名称を入力してください")).toBeInTheDocument();
    for (const label of ["開始日", "終了日"]) {
      expect(screen.getByLabelText(label)).toHaveAttribute("aria-invalid", "true");
    }
    expect(onSave).not.toHaveBeenCalled();

    // 名称は空のまま、終了日だけを有効日に訂正して再保存する。
    await enterDate(user, "終了日", "2026-02-01");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(screen.queryByText("終了日は開始日以降にしてください")).not.toBeInTheDocument();
    });
    for (const label of ["開始日", "終了日"]) {
      const input = screen.getByLabelText(label);
      expect(input).not.toHaveAttribute("aria-invalid");
      expect(input).not.toHaveAttribute("aria-describedby");
    }
    expect(screen.getByLabelText("名称")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByText("名称を入力してください")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  // 保存成功 oracle: 全フィールド正常（名称 + 有効期間）なら onSave が呼ばれ、
  // 名称/期間ともエラー表示なし。常に false を返す退化実装を排除する。
  it("名称と有効期間を入力して保存するとエラー表示なしで onSave が1回呼ばれる", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => true);
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={onSave} />);

    await user.type(screen.getByLabelText("名称"), "テストキャンペーン");
    await enterDate(user, "開始日", "2026-01-01");
    await enterDate(user, "終了日", "2026-01-31");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
    expect(screen.getByLabelText("名称")).not.toHaveAttribute("aria-invalid");
    for (const label of ["開始日", "終了日"]) {
      expect(screen.getByLabelText(label)).not.toHaveAttribute("aria-invalid");
    }
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("開始日と終了日が同日（start=end）でも有効期間として保存に成功する", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => true);
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={onSave} />);

    await user.type(screen.getByLabelText("名称"), "同日キャンペーン");
    await enterDate(user, "開始日", "2026-03-10");
    await enterDate(user, "終了日", "2026-03-10");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
    expect(screen.getByLabelText("名称")).not.toHaveAttribute("aria-invalid");
    for (const label of ["開始日", "終了日"]) {
      expect(screen.getByLabelText(label)).not.toHaveAttribute("aria-invalid");
    }
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("新規キャンペーンで往診カテゴリを選ぶと house_call だけを保存する", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => true);
    render(<CampaignSidePanel item={null} onClose={vi.fn()} onSave={onSave} />);

    await user.type(screen.getByLabelText("名称"), "往診割引");
    await enterDate(user, "開始日", "2026-01-01");
    await enterDate(user, "終了日", "2026-01-31");
    await user.click(screen.getByRole("checkbox", { name: "往診" }));
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ targetCategories: ["house_call"] }),
    );
  });

  it("編集時に既存の house_call 対象を保持して保存する", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => true);
    const item = {
      id: "campaign-house-call",
      name: "既存往診割引",
      startDate: "2026-01-01",
      endDate: "2026-01-31",
      discountType: "rate" as const,
      discountValue: "10",
      isActive: true,
      sortOrder: 1,
      targetCategories: ["house_call"],
      targetItemIds: [],
    };
    render(<CampaignSidePanel item={item} onClose={vi.fn()} onSave={onSave} />);

    expect(screen.getByRole("checkbox", { name: "往診" })).toBeChecked();
    await user.type(screen.getByLabelText("名称"), "変更");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ targetCategories: ["house_call"] }),
    );
  });
});
