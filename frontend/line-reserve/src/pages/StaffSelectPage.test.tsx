import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { StaffSelectPage } from "./StaffSelectPage";
import type { Staff } from "../types/models";
import { AUTO_ADVANCE_HELPER_TEXT } from "../lib/advance-policy";

const BASE_PROPS = {
  clinicId: "1",
  idToken: "test-id-token",
  courseId: 10,
  showNoStaffOption: false,
  isTrimming: false,
};

const staff: Staff = {
  id: 1,
  name: "山田先生",
  reservation_comment: "",
  reservation_image_url: "",
  sort_order: 1,
};

describe("StaffSelectPage（R-F22/R-F23: 共通フェッチ状態管理・ステータス別エラー）", () => {
  it("取得後はスタッフ一覧を表示する", async () => {
    server.use(http.get("/api/liff/:clinicId/staffs", () => HttpResponse.json([staff])));

    render(<StaffSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByText("山田先生")).toBeInTheDocument();
  });

  it("API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する", async () => {
    server.use(
      http.get("/api/liff/:clinicId/staffs", () => HttpResponse.json(null, { status: 503 })),
    );

    render(<StaffSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole("alert")).toHaveTextContent("サーバーエラーが発生しました");
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない", async () => {
    server.use(
      http.get("/api/liff/:clinicId/staffs", () => HttpResponse.json(null, { status: 401 })),
    );

    render(<StaffSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "LINEアプリを再起動して開き直してください",
    );
    expect(screen.queryByRole("button", { name: "再試行" })).not.toBeInTheDocument();
  });

  it("再試行ボタンをクリックするとスタッフ一覧を再取得する", async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get("/api/liff/:clinicId/staffs", () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json([staff]);
      }),
    );

    render(<StaffSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "再試行" }));

    expect(await screen.findByText("山田先生")).toBeInTheDocument();
    expect(callCount).toBe(2);
  });

  it("BUG-030: auto-on-select のヘルパー文言を表示し、一覧タップで onSelect する", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    server.use(http.get("/api/liff/:clinicId/staffs", () => HttpResponse.json([staff])));

    render(<StaffSelectPage {...BASE_PROPS} onSelect={onSelect} onBack={vi.fn()} />);

    expect(await screen.findByTestId("auto-advance-hint")).toHaveTextContent(
      AUTO_ADVANCE_HELPER_TEXT,
    );
    expect(screen.queryByRole("button", { name: "次へ" })).not.toBeInTheDocument();

    await user.click(await screen.findByText("山田先生"));
    expect(onSelect).toHaveBeenCalledWith(1, "山田先生");
  });
});
