import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { LstepDeliveryMonitorLogsTable } from "./LstepDeliveryMonitorLogsTable";
import type { DeliveryTriggerLogsPageResponse } from "../api/get-lstep-delivery-trigger-logs";

const mockLogsPage: DeliveryTriggerLogsPageResponse = {
  items: [
    {
      id: "log-1",
      owner_id: "owner-1",
      owner_name: "非常に長い飼主名を持つ田中太郎さん",
      trigger_type: "birthday_message",
      scheduled_at: "2026-05-02T10:00:00Z",
      status: "excluded",
      excluded_reason:
        "LINE連携が解除されているため送信対象外となりました。再連携のうえ配信設定をご確認ください。",
    },
  ],
  total: 1,
  page: 1,
  per_page: 20,
};

function renderTable(logsPage: DeliveryTriggerLogsPageResponse = mockLogsPage) {
  return render(
    <LstepDeliveryMonitorLogsTable
      logsPage={logsPage}
      logsLoading={false}
      page={1}
      totalPages={1}
      onPreviousPage={() => {}}
      onNextPage={() => {}}
    />,
  );
}

describe("LstepDeliveryMonitorLogsTable", () => {
  // EMR-389: 768px / 390px で 予定日時・ステータス・除外理由 などの固定ヘッダが
  // 1文字ずつ断片化して折り返される不具合の回帰防止。テーブルは既存の横スクロール内で伸長させる。
  it("keeps every fixed column header on a single line", () => {
    renderTable();

    for (const label of ["種別", "飼主", "予定日時", "ステータス", "送信日時", "除外理由"]) {
      const header = screen.getByRole("columnheader", { name: label });
      expect(header.className).toContain("whitespace-nowrap");
    }
  });

  // EMR-389: 日時・コンパクトなステータスバッジは途中で折り返さず一貫性を保つ
  it("keeps datetime and status badge cells on a single line", () => {
    renderTable();

    const scheduledCell = screen.getByText(/2026-05-02/).closest("td");
    expect(scheduledCell?.className).toContain("whitespace-nowrap");

    const statusCell = screen.getByText("除外", { selector: "span" }).closest("td");
    expect(statusCell?.className).toContain("whitespace-nowrap");
  });

  // EMR-389: 長い飼主名・除外理由の本文は折り返しを維持する（ヘッダ nowrap を波及させない）
  it("does not force nowrap on long body text cells", () => {
    renderTable();

    const ownerCell = screen.getByText("非常に長い飼主名を持つ田中太郎さん").closest("td");
    expect(ownerCell?.className).not.toContain("whitespace-nowrap");

    const reasonCell = screen.getByText(/LINE連携が解除されているため/).closest("td");
    expect(reasonCell?.className).not.toContain("whitespace-nowrap");
  });

  it("shows the loading state while logs are being fetched", () => {
    render(
      <LstepDeliveryMonitorLogsTable
        logsPage={undefined}
        logsLoading={true}
        page={1}
        totalPages={1}
        onPreviousPage={() => {}}
        onNextPage={() => {}}
      />,
    );

    expect(screen.getByTestId("logs-loading")).toBeInTheDocument();
  });
});
