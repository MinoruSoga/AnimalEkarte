import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TreatmentPackageComponentList } from "./TreatmentPackageComponentList";
import { newComponentDraft, type ComponentDraft } from "../lib/treatment-package-form";

const row = (kind: ComponentDraft["kind"], masterId = ""): ComponentDraft => ({
  ...newComponentDraft(kind),
  masterId,
});

describe("TreatmentPackageComponentList", () => {
  it("種別選択・数量入力・追加/削除/上下移動を onChange へ反映", () => {
    const onChange = vi.fn();
    const components = [row("consultation", "1"), row("medicine", "2")];
    render(<TreatmentPackageComponentList components={components} onChange={onChange} />);
    // kind 変更で masterId をリセット（他型FK混入しない）
    fireEvent.change(screen.getAllByLabelText("種別")[0]!, {
      target: { value: "manual_other" },
    });
    expect(onChange).toHaveBeenLastCalledWith([
      expect.objectContaining({ kind: "manual_other", masterId: "" }),
      components[1],
    ]);
    fireEvent.click(screen.getByText("構成品を追加"));
    expect(onChange).toHaveBeenLastCalledWith(
      expect.arrayContaining([expect.objectContaining({ kind: "consultation" })]),
    );
    expect(onChange.mock.calls.at(-1)![0]).toHaveLength(3);
  });

  it("manual_other は手入力欄、master 種は typed ID 欄を出す", () => {
    const onChange = vi.fn();
    const { rerender } = render(
      <TreatmentPackageComponentList components={[row("manual_other")]} onChange={onChange} />,
    );
    expect(screen.getByLabelText("手入力名称")).toBeTruthy();
    expect(screen.getByLabelText("税率")).toBeTruthy();
    rerender(
      <TreatmentPackageComponentList components={[row("checkup_type")]} onChange={onChange} />,
    );
    // 健診は1品目として扱い、nested package の編集欄は出さない
    expect(screen.getByLabelText(/マスタ ID/)).toBeTruthy();
    expect(screen.queryByLabelText("手入力名称")).toBeNull();
  });

  it("行別エラーを componentKey 単位で表示", () => {
    const components = [row("consultation")];
    const errors = new Map([[components[0]!.componentKey, ["マスタを選択してください"]]]);
    render(
      <TreatmentPackageComponentList
        components={components}
        onChange={vi.fn()}
        componentErrors={errors}
      />,
    );
    expect(screen.getByRole("alert").textContent).toContain("マスタを選択してください");
  });
});
