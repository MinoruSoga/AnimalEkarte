import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ClinicScopeFilter } from "./ClinicScopeFilter";
import type { ClinicMembership } from "@/types/auth";

const TWO_CLINICS: ClinicMembership[] = [
  { clinicId: "1", clinicName: "本院", isMain: true },
  { clinicId: "2", clinicName: "分院", isMain: false },
];

describe("ClinicScopeFilter — #86 拠点横断フィルタ", () => {
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
});
