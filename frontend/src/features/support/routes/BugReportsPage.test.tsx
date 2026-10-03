import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { BugReport } from "../types";

const createTicketMutateMock = vi.fn();
const deleteMutateMock = vi.fn();
let reportsMock: BugReport[] = [];

vi.mock("../api/get-bug-reports", () => ({
  useGetBugReports: () => ({ data: reportsMock, isLoading: false, isError: false }),
}));

vi.mock("../api/update-bug-report-status", () => ({
  useUpdateBugReportStatus: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock("../api/create-plane-ticket", () => ({
  useCreatePlaneTicket: () => ({ mutate: createTicketMutateMock, isPending: false }),
}));

vi.mock("../api/delete-bug-report", () => ({
  useDeleteBugReport: () => ({ mutate: deleteMutateMock, isPending: false }),
}));

import { BugReportsPage } from "./BugReportsPage";

function makeReport(overrides: Partial<BugReport> = {}): BugReport {
  return {
    id: 1,
    title: "受付でエラー",
    detail: "詳細",
    page_url: "https://stg.noah-karte.com/reception",
    route_path: "/reception",
    user_agent: "ua",
    viewport: "1200x800",
    app_version: "1.0.0",
    status: "open",
    reporter_staff_id: 5,
    reporter_name: "田中",
    clinic_name: "さくら動物病院",
    created_at: "2026-10-01T00:00:00Z",
    updated_at: "2026-10-01T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  createTicketMutateMock.mockReset();
  deleteMutateMock.mockReset();
  reportsMock = [makeReport()];
});

describe("BugReportsPage 一覧", () => {
  it("全医院共有ボードとして医院列を表示し、権限チェックなしで描画する", () => {
    // 権限ゲートは意図的に無い — useAuth/usePermission をモックせず全操作が描画される
    render(<BugReportsPage />);

    expect(screen.getByRole("columnheader", { name: "医院" })).toBeInTheDocument();
    expect(screen.getByText("さくら動物病院")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "詳細: 受付でエラー" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "削除: 受付でエラー" })).toBeInTheDocument();
  });
});

describe("BugReportsPage Plane 列", () => {
  it("plane_issue_url があるとチケットへの外部リンクを表示する", () => {
    reportsMock = [
      makeReport({ plane_issue_url: "https://app.plane.so/baritechllc/browse/EMR-42/" }),
    ];
    render(<BugReportsPage />);

    const link = screen.getByRole("link", { name: "チケット" });
    expect(link).toHaveAttribute("href", "https://app.plane.so/baritechllc/browse/EMR-42/");
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noopener noreferrer");
  });

  it("plane_sync_error があると「起票失敗」と再送ボタンを表示し、再送で mutation を呼ぶ", async () => {
    const user = userEvent.setup();
    reportsMock = [makeReport({ id: 7, plane_sync_error: "plane api error (status 500)" })];
    render(<BugReportsPage />);

    expect(screen.getByText("起票失敗")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "再送" }));
    expect(createTicketMutateMock).toHaveBeenCalledWith(7);
  });

  it("Plane 情報がない報告は ― を表示する", () => {
    render(<BugReportsPage />);

    expect(screen.queryByRole("link", { name: "チケット" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "再送" })).not.toBeInTheDocument();
  });
});

describe("BugReportsPage 削除", () => {
  it("削除ボタンを表示し、確認ダイアログ経由で削除する", async () => {
    const user = userEvent.setup();
    reportsMock = [makeReport({ id: 7 })];
    render(<BugReportsPage />);

    await user.click(screen.getByRole("button", { name: "削除: 受付でエラー" }));

    const dialog = await screen.findByRole("alertdialog");
    expect(dialog).toHaveTextContent("「受付でエラー」を削除しますか？");

    await user.click(screen.getByRole("button", { name: "削除する" }));
    expect(deleteMutateMock).toHaveBeenCalledWith(7, expect.anything());
  });
});
