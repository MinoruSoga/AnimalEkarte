import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { MedicalRecordAutoCreateFailure } from "./MedicalRecordAutoCreateFailure";

describe("MedicalRecordAutoCreateFailure", () => {
  it("カルテ作成失敗を通知し、明示的な再試行 action を提供する", async () => {
    const user = userEvent.setup();
    const onRetry = vi.fn();

    render(
      <MedicalRecordAutoCreateFailure
        failurePhase="medical-record"
        isRetrying={false}
        onRetry={onRetry}
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent(
      "カルテの作成に失敗しました。作成済みの予約は保持されています。",
    );
    const retryButton = screen.getByRole("button", {
      name: "カルテ作成を再試行する",
    });
    expect(retryButton).toHaveAttribute("type", "button");
    expect(retryButton).toBeEnabled();

    await user.click(retryButton);

    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it("appointment 作成失敗を通知し、再試行中は action を無効化する", () => {
    render(
      <MedicalRecordAutoCreateFailure failurePhase="appointment" isRetrying onRetry={vi.fn()} />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent("予約の作成に失敗しました。");
    expect(screen.getByRole("button", { name: "カルテ作成を再試行する" })).toBeDisabled();
  });

  // BUG-MR-DRAFT-AUTOPOST-FAILED: master 欠落は再試行では解決しないため、
  // 復旧導線（マスタ追加 + 再読み込み）を表示し再試行 action は出さない。
  it("予約区分マスタ欠落では復旧導線を表示し、再試行 action を出さない", () => {
    render(
      <MedicalRecordAutoCreateFailure
        failurePhase="appointment-master-missing"
        isRetrying={false}
        onRetry={vi.fn()}
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent(
      "予約区分マスタに診察系の予約区分が登録されていません。",
    );
    expect(
      screen.queryByRole("button", { name: "カルテ作成を再試行する" }),
    ).not.toBeInTheDocument();
  });
});
