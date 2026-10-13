import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ClinicScopeFilter } from "./ClinicScopeFilter";
import type { ClinicMembership } from "@/types/auth";

const TWO_CLINICS: ClinicMembership[] = [
  { clinicId: "1", clinicName: "本院", isMain: true },
  { clinicId: "2", clinicName: "分院", isMain: false },
];

describe("ClinicScopeFilter — #86 医院横断フィルタ", () => {
  it("単一所属では何も描画しない", () => {
    render(
      <ClinicScopeFilter clinics={[TWO_CLINICS[0]]} selectedIds={["1"]} onToggle={() => {}} />,
    );
    expect(screen.queryByTestId("clinic-scope-filter")).not.toBeInTheDocument();
  });

  it("複数所属で全医院のトグルを表示し、選択状態を aria-pressed で示す", () => {
    render(<ClinicScopeFilter clinics={TWO_CLINICS} selectedIds={["1"]} onToggle={() => {}} />);
    expect(screen.getByRole("button", { name: "本院" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "分院" })).toHaveAttribute("aria-pressed", "false");
  });

  it("トグルクリックで onToggle が clinicId 付きで呼ばれる", async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();
    render(<ClinicScopeFilter clinics={TWO_CLINICS} selectedIds={["1"]} onToggle={onToggle} />);
    await user.click(screen.getByRole("button", { name: "分院" }));
    expect(onToggle).toHaveBeenCalledWith("2");
  });

  it("全角スペースを含む正式名を正規化せず1ボタンで表示し、トグルも維持する (EMR-388)", async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();
    const OFFICIAL_NAME = "テスト　動物病院"; // 全角スペース(U+3000)を含む正式名
    const clinics: ClinicMembership[] = [
      { clinicId: "1", clinicName: OFFICIAL_NAME, isMain: true },
      { clinicId: "2", clinicName: "分院", isMain: false },
    ];
    render(<ClinicScopeFilter clinics={clinics} selectedIds={["1"]} onToggle={onToggle} />);

    const buttons = screen.getAllByRole("button");
    expect(buttons).toHaveLength(2);
    const official = buttons.find((b) => b.textContent === OFFICIAL_NAME);
    expect(official).toBeDefined();
    expect(official?.textContent).toContain("　");

    await user.click(official!);
    expect(onToggle).toHaveBeenCalledWith("1");
  });
});
