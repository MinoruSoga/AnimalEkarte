import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ClinicBooleanProperty, ClinicTextareaProperty } from "./ClinicMasterSidePanelProperties";

/**
 * EMR-411 C-2: 医院マスタの boolean プロパティボタンは、どの項目か分かる
 * accessible name（`${label}を切り替え`）と現在値（aria-pressed）を公開する。
 * textarea は項目名を accessible name として持つ。
 */
describe("ClinicBooleanProperty アクセシブル名と pressed 状態 (EMR-411 C-2)", () => {
  // ClinicMasterSidePanel の実 callsite 7 件と同じラベルセット
  const BOOLEAN_LABELS = [
    "ロゴ表示",
    "登録番号警告",
    "項目カテゴリ",
    "医院情報ヘッダー",
    "飼主・ペット情報",
    "明細テーブル",
    "お会計サマリー",
  ] as const;

  function Stateful({ label, initial }: { label: string; initial: boolean }) {
    const [value, setValue] = useState(initial);
    return <ClinicBooleanProperty label={label} value={value} onChange={setValue} />;
  }

  it.each([...BOOLEAN_LABELS])(
    "「%s」が固有の accessible name と pressed 遷移を持つ",
    async (label) => {
      const user = userEvent.setup();
      render(<Stateful label={label} initial={false} />);

      const toggle = screen.getByRole("button", { name: `${label}を切り替え` });
      expect(toggle).toHaveAttribute("aria-pressed", "false");

      await user.click(toggle);
      expect(toggle).toHaveAttribute("aria-pressed", "true");

      await user.click(toggle);
      expect(toggle).toHaveAttribute("aria-pressed", "false");
    },
  );

  it("7 項目が同時に描画されても accessible name が一意に解決する", () => {
    render(
      <>
        {BOOLEAN_LABELS.map((label) => (
          <ClinicBooleanProperty key={label} label={label} value={false} onChange={vi.fn()} />
        ))}
      </>,
    );

    for (const label of BOOLEAN_LABELS) {
      expect(screen.getByRole("button", { name: `${label}を切り替え` })).toBeInTheDocument();
    }
  });
});

describe("ClinicTextareaProperty accessible name (EMR-411 C-2)", () => {
  it("textarea が項目名の accessible name を持ち、入力で onChange が呼ばれる", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <ClinicTextareaProperty
        label="フッター"
        value=""
        onChange={onChange}
        placeholder="例: ご来院ありがとうございました。"
      />,
    );

    const textarea = screen.getByRole("textbox", { name: "フッター" });
    await user.type(textarea, "あ");
    expect(onChange).toHaveBeenCalledWith("あ");
  });
});
