import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

const hasPermissionMock = vi.fn();

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    hasPermission: hasPermissionMock,
  }),
}));

vi.mock("../components/LstepSettingsForm", () => ({
  LstepSettingsForm: () => <div data-testid="lstep-settings-form" />,
}));

vi.mock("../components/TriggerPrioritySection", () => ({
  TriggerPrioritySection: () => <div data-testid="trigger-priority-section" />,
}));

vi.mock("../components/LstepTagCodeMappingsSection", () => ({
  LstepTagCodeMappingsSection: () => <div data-testid="lstep-tag-code-mappings-section" />,
}));

vi.mock("../components/LstepTagConfigSection", () => ({
  LstepTagConfigSection: () => <div data-testid="lstep-tag-config-section" />,
}));

import { LstepSettingsPage } from "./LstepSettingsPage";

describe("LstepSettingsPage", () => {
  beforeEach(() => {
    hasPermissionMock.mockReset();
  });

  it("canEdit=true で設定フォームと各セクションを表示する", () => {
    hasPermissionMock.mockImplementation((resource: string, action: string) => {
      return resource === "hospital-settings" && action === "edit";
    });

    render(<LstepSettingsPage />);

    expect(screen.getByText("Lステップ連携設定")).toBeInTheDocument();
    expect(screen.getByTestId("lstep-settings-form")).toBeInTheDocument();
    expect(screen.getByTestId("trigger-priority-section")).toBeInTheDocument();
    expect(screen.getByTestId("lstep-tag-code-mappings-section")).toBeInTheDocument();
  });

  it("canEdit=false で権限なしメッセージを表示する", () => {
    hasPermissionMock.mockReturnValue(false);

    render(<LstepSettingsPage />);

    expect(screen.getByText("この画面を表示する権限がありません。")).toBeInTheDocument();
    expect(screen.queryByTestId("lstep-settings-form")).not.toBeInTheDocument();
  });
});
