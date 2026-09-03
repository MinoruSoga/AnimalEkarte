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
});
