import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {
  MedicalRecordFloatingActions,
  MedicalRecordFinalizeDialog,
} from "./MedicalRecordFormActions";

// SPEC-GAP: カルテ確定(Lock)のUI導線。BEにはfinalize APIが存在するがFEから
// 確定できない状態だったため、確定ボタン・確認ダイアログの表示条件を検証する。

const baseProps = {
  activeTab: "問診",
  canDelete: true,
  canEdit: true,
  canSubmit: true,
  isNewRecord: false,
  isCreating: false,
  isSaving: false,
  isFinalized: false,
  onDeleteClick: vi.fn(),
  onVitalsClick: vi.fn(),
  onPrintClick: vi.fn(),
  onFinalizeClick: vi.fn(),
};

describe("MedicalRecordFloatingActions", () => {
  it("編集権限があり未確定の既存カルテでは確定ボタンを表示する", () => {
    render(<MedicalRecordFloatingActions {...baseProps} />);
    expect(screen.getByRole("button", { name: "確定する" })).toBeInTheDocument();
  });

  it("確定済みカルテでは確定ボタンを表示しない", () => {
    render(<MedicalRecordFloatingActions {...baseProps} isFinalized />);
    expect(screen.queryByRole("button", { name: "確定する" })).not.toBeInTheDocument();
  });

  it("確定済みカルテでは保存ボタンを表示しない", () => {
    render(<MedicalRecordFloatingActions {...baseProps} isFinalized />);

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
  });

  it("保存処理中は重複submitを防ぐため保存ボタンを操作不可にする", () => {
    render(<MedicalRecordFloatingActions {...baseProps} isSaving />);

    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
  });

  it("保存権限がない場合は保存ボタンを表示しない", () => {
    render(<MedicalRecordFloatingActions {...baseProps} canEdit={false} canSubmit={false} />);

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
  });

  it("新規作成中（未保存）のカルテでは確定ボタンを表示しない", () => {
    render(<MedicalRecordFloatingActions {...baseProps} isNewRecord />);
    expect(screen.queryByRole("button", { name: "確定する" })).not.toBeInTheDocument();
  });

  it("編集権限がない場合は確定ボタンを表示しない", () => {
    render(<MedicalRecordFloatingActions {...baseProps} canEdit={false} />);
    expect(screen.queryByRole("button", { name: "確定する" })).not.toBeInTheDocument();
  });

  it("確定ボタンクリックで onFinalizeClick が呼ばれる", async () => {
    const user = userEvent.setup();
    const onFinalizeClick = vi.fn();
    render(<MedicalRecordFloatingActions {...baseProps} onFinalizeClick={onFinalizeClick} />);
    await user.click(screen.getByRole("button", { name: "確定する" }));
    expect(onFinalizeClick).toHaveBeenCalledOnce();
  });
});

describe("MedicalRecordFinalizeDialog", () => {
  it("open時に不可逆であることを説明する確認ダイアログを表示する", () => {
    render(
      <MedicalRecordFinalizeDialog
        open
        isFinalizing={false}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    expect(screen.getByText("カルテを確定しますか？")).toBeInTheDocument();
    expect(screen.getByText(/元に戻せません/)).toBeInTheDocument();
  });

  it("確定するボタンクリックで onConfirm が呼ばれる", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    render(
      <MedicalRecordFinalizeDialog
        open
        isFinalizing={false}
        onClose={vi.fn()}
        onConfirm={onConfirm}
      />,
    );
    await user.click(screen.getByRole("button", { name: "確定する" }));
    expect(onConfirm).toHaveBeenCalledOnce();
  });

  it("isFinalizing 中はボタンラベルが変わり操作不可になる", () => {
    render(
      <MedicalRecordFinalizeDialog
        open
        isFinalizing
        onClose={vi.fn()}
        onConfirm={vi.fn()}
      />,
    );
    const button = screen.getByRole("button", { name: "確定中..." });
    expect(button).toBeDisabled();
  });
});
