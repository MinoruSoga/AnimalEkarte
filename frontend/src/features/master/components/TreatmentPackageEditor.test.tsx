import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TreatmentPackageEditor } from "./TreatmentPackageEditor";
import { emptyPackageDraft } from "../lib/treatment-package-form";

const base = {
  draft: emptyPackageDraft(),
  conflict: false,
  createUnknown: false,
  canSave: true,
  canDelete: true,
  saving: false,
  deleting: false,
  onChange: vi.fn(),
  onComponentsChange: vi.fn(),
  onSave: vi.fn(),
  onDeleteRequest: vi.fn(),
  onClose: vi.fn(),
};

describe("TreatmentPackageEditor（競合・結果不明・入力保持）", () => {
  it("create モードは名称/説明/有効/構成品と保存を出す", () => {
    render(<TreatmentPackageEditor mode="create" {...base} />);
    expect(screen.getByLabelText("名称")).toBeTruthy();
    expect(screen.getByLabelText("説明")).toBeTruthy();
    expect(screen.getByText("保存")).toBeTruthy();
    expect(screen.getByText("構成品を追加")).toBeTruthy();
  });
  it("競合 alert は再読込・比較を促し入力保持を明示", () => {
    render(<TreatmentPackageEditor mode="edit" {...base} conflict={true} />);
    expect(screen.getByRole("alert").textContent).toContain("入力は保持");
  });
  it("create 応答不明は盲目再送しない旨を明示", () => {
    render(<TreatmentPackageEditor mode="create" {...base} createUnknown={true} />);
    expect(screen.getByRole("alert").textContent).toContain("自動で再送しません");
  });
  it("canSave=false では保存を disabled（送信中・権限ゲートのみ）", () => {
    render(<TreatmentPackageEditor mode="edit" {...base} canSave={false} />);
    expect((screen.getByText("保存") as HTMLButtonElement).disabled).toBe(true);
  });
});
