import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { PetHealthPage } from "./PetHealthPage";
import type { HealthCardResponse } from "../api/liff-api";
import { DEFAULT_ERROR_PAGE_THEME, SHARED_ERROR_PAGE_TITLE } from "@/shared-liff/ErrorPage";

const BASE_PROPS = {
  idToken: "test-id-token",
  displayName: "テストユーザー",
  pictureUrl: null,
};

const healthCard: HealthCardResponse = {
  owner_name: "山田太郎",
  pets: [],
};

const settingsWithClinicBrand = {
  liff_id: "liff-1",
  header_text: "ノア動物病院 八王子",
  phone_number: "",
  status: "running",
  request_example: "",
  reservation_notice: "",
  cancel_notice: "",
  privacy_policy: "",
  show_no_staff_option: false,
  booking_window: 30,
};

function renderWithClinicId() {
  window.history.pushState({}, "", "/liff/health-card?clinic_id=1");
  return render(<PetHealthPage {...BASE_PROPS} />);
}

function renderWithoutClinicId() {
  window.history.pushState({}, "", "/liff/health-card");
  return render(<PetHealthPage {...BASE_PROPS} />);
}

describe("PetHealthPage（R-F22/R-F23: 共通フェッチ状態管理・ステータス別エラー）", () => {
  it("clinic_id 欠落時は共有エラー chrome と専用文言を表示し、再試行は出さない（BUG-014/027）", async () => {
    const { container } = renderWithoutClinicId();

    expect(await screen.findByRole("alert")).toHaveTextContent("クリニックIDが見つかりません");
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(SHARED_ERROR_PAGE_TITLE);
    expect(container.firstElementChild?.className).toContain(DEFAULT_ERROR_PAGE_THEME.bg);
    expect(screen.queryByText("しばらく経ってから再度お試しください")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "再試行" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "再読み込み" })).not.toBeInTheDocument();
  });

  it("取得後は飼い主名を表示する", async () => {
    server.use(
      http.get("/api/liff/:clinicId/health-card", () => HttpResponse.json(healthCard)),
      http.get("/api/liff/:clinicId/settings", () => HttpResponse.json(settingsWithClinicBrand)),
    );

    renderWithClinicId();

    expect(await screen.findByText("山田太郎")).toBeInTheDocument();
  });

  it("BUG-026: クリニックブランド名をヘッダーに表示し、飼い主名は副次表示する", async () => {
    server.use(
      http.get("/api/liff/:clinicId/health-card", () => HttpResponse.json(healthCard)),
      http.get("/api/liff/:clinicId/settings", () => HttpResponse.json(settingsWithClinicBrand)),
    );

    const { container } = renderWithClinicId();

    expect(await screen.findByRole("heading", { name: "ノア動物病院 八王子" })).toBeInTheDocument();
    expect(screen.getByText("山田太郎")).toBeInTheDocument();
    const header = screen.getByRole("banner");
    expect(header.className).toMatch(/bg-liff-brand(?!-)/);
    expect(container.firstElementChild?.className).toContain("bg-liff-brand-bg");
  });

  it("BUG-026: settings の header_text が空でもクラッシュせず飼い主名を表示する", async () => {
    server.use(
      http.get("/api/liff/:clinicId/health-card", () => HttpResponse.json(healthCard)),
      http.get("/api/liff/:clinicId/settings", () =>
        HttpResponse.json({ ...settingsWithClinicBrand, header_text: "" }),
      ),
    );

    renderWithClinicId();

    expect(await screen.findByText("山田太郎")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "ノア動物病院 八王子" })).not.toBeInTheDocument();
    expect(screen.getByRole("banner")).toBeInTheDocument();
  });

  it("BUG-026: settings 取得失敗でも健康カードは表示し、偽の医院名を出さない", async () => {
    server.use(
      http.get("/api/liff/:clinicId/health-card", () => HttpResponse.json(healthCard)),
      http.get("/api/liff/:clinicId/settings", () => HttpResponse.json(null, { status: 500 })),
    );

    renderWithClinicId();

    expect(await screen.findByText("山田太郎")).toBeInTheDocument();
    expect(screen.queryByText("ノア動物病院")).not.toBeInTheDocument();
    expect(screen.getByRole("banner")).toBeInTheDocument();
  });

  it("API失敗(5xx)時は共有エラー chrome・サーバーエラーメッセージ・再試行を表示する（BUG-027）", async () => {
    server.use(
      http.get("/api/liff/:clinicId/health-card", () => HttpResponse.json(null, { status: 500 })),
    );

    const { container } = renderWithClinicId();

    expect(await screen.findByRole("alert")).toHaveTextContent("サーバーエラーが発生しました");
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(SHARED_ERROR_PAGE_TITLE);
    expect(container.firstElementChild?.className).toContain(DEFAULT_ERROR_PAGE_THEME.bg);
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない", async () => {
    server.use(
      http.get("/api/liff/:clinicId/health-card", () => HttpResponse.json(null, { status: 401 })),
    );

    renderWithClinicId();

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "LINEアプリを再起動して開き直してください",
    );
    expect(screen.queryByRole("button", { name: "再試行" })).not.toBeInTheDocument();
  });

  it("再試行ボタンをクリックすると健康記録を再取得する", async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get("/api/liff/:clinicId/health-card", () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json(healthCard);
      }),
    );

    renderWithClinicId();

    await user.click(await screen.findByRole("button", { name: "再試行" }));

    expect(await screen.findByText("山田太郎")).toBeInTheDocument();
    expect(callCount).toBe(2);
  });
});
