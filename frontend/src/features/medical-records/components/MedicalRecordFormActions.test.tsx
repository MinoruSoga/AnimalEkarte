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

async function expectFinalizePhysicallyBlocked(
  extraProps: {
    billingConfirmationStatus?: "pending" | "confirmed" | "returned";
    isBillingConfirmationLoading?: boolean;
    isBillingConfirmationError?: boolean;
  },
) {
  const onFinalizeClick = vi.fn();
  render(
    <MedicalRecordFloatingActions
      {...baseProps}
      {...extraProps}
      onFinalizeClick={onFinalizeClick}
    />,
  );
  const button = screen.getByRole("button", { name: "確定する" });
  expect(button).toBeDisabled();
  expect(button).toHaveAttribute("title", "会計確認が未完了です");
  // Never(0): disabled:pointer-events-none でも click 試行を到達させ、handler 非発火を検証する
  const user = userEvent.setup({ pointerEventsCheck: 0 });
  await user.click(button);
  expect(onFinalizeClick).not.toHaveBeenCalled();
}

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

  // BUG-503: auto-draft 前の floating 保存は create を装わない
  it("新規作成中（未保存）のカルテでは保存ボタンを表示しない", () => {
    render(<MedicalRecordFloatingActions {...baseProps} isNewRecord />);
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
  });

  it("編集権限がない場合は確定ボタンを表示しない", () => {
    render(<MedicalRecordFloatingActions {...baseProps} canEdit={false} />);
    expect(screen.queryByRole("button", { name: "確定する" })).not.toBeInTheDocument();
  });

  it("確定ボタンクリックで onFinalizeClick が呼ばれる", async () => {
    const user = userEvent.setup();
    const onFinalizeClick = vi.fn();
    render(
      <MedicalRecordFloatingActions
        {...baseProps}
        billingConfirmationStatus="confirmed"
        onFinalizeClick={onFinalizeClick}
      />,
    );
    await user.click(screen.getByRole("button", { name: "確定する" }));
    expect(onFinalizeClick).toHaveBeenCalledOnce();
  });

  it("会計確認が pending のときは確定するを物理ブロックする", async () => {
    await expectFinalizePhysicallyBlocked({ billingConfirmationStatus: "pending" });
  });

  it("会計確認が returned のときは確定するを物理ブロックする", async () => {
    await expectFinalizePhysicallyBlocked({ billingConfirmationStatus: "returned" });
  });

  it("会計確認の読み込み中は確定するを物理ブロックする", async () => {
    await expectFinalizePhysicallyBlocked({ isBillingConfirmationLoading: true });
  });

  it("会計確認の取得エラー時は確定するを物理ブロックする", async () => {
    await expectFinalizePhysicallyBlocked({ isBillingConfirmationError: true });
  });

  it("会計確認props省略時は確定するを物理ブロックする（fail-closed）", async () => {
    await expectFinalizePhysicallyBlocked({});
  });

  it("会計確認が confirmed なら確定するをクリックできる", async () => {
    const user = userEvent.setup();
    const onFinalizeClick = vi.fn();
    render(
      <MedicalRecordFloatingActions
        {...baseProps}
        billingConfirmationStatus="confirmed"
        onFinalizeClick={onFinalizeClick}
      />,
    );
    const button = screen.getByRole("button", { name: "確定する" });
    expect(button).toBeEnabled();
    expect(button).not.toHaveAttribute("title");
    await user.click(button);
    expect(onFinalizeClick).toHaveBeenCalledOnce();
  });

  it("会計(医師確認)タブではフローティングバーごと確定するを出さない", () => {
    render(
      <MedicalRecordFloatingActions
        {...baseProps}
        activeTab="会計(医師確認)"
        billingConfirmationStatus="confirmed"
      />,
    );
    expect(screen.queryByRole("button", { name: "確定する" })).not.toBeInTheDocument();
  });

  it("予防接種タブでは保存を出さず確定する・印刷は残す", () => {
    render(<MedicalRecordFloatingActions {...baseProps} activeTab="予防接種" />);

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "確定する" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "印刷" })).toBeInTheDocument();
  });

  it("問診タブでは保存を表示する", () => {
    render(<MedicalRecordFloatingActions {...baseProps} activeTab="問診" />);

    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
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
