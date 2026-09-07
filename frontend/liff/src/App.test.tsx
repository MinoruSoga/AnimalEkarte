import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/shared-liff/use-liff", () => ({
  useLiff: () => ({
    idToken: "test-id-token",
    isReady: true,
    initError: false,
    displayName: "テストユーザー",
    pictureUrl: null,
  }),
}));

vi.mock("./pages/PetHealthPage", () => ({
  PetHealthPage: () => <div>health-card-route</div>,
}));

function setSearch(search: string) {
  window.history.pushState({}, "", `/liff/health-card${search}`);
}

describe("LIFF App token routing (S12 / V05-5)", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    setSearch("");
  });

  it("token なし clinic_id のみは健康手帳導線（仕様38 / S12手順5）", async () => {
    setSearch("?clinic_id=1");
    const { App } = await import("./App");
    render(<App />);
    expect(screen.getByText("health-card-route")).toBeInTheDocument();
  });

  it("空 token キーは連携導線の無効 URL になる（V05-5）", async () => {
    setSearch("?token=&clinic_id=1");
    const { App } = await import("./App");
    render(<App />);
    expect(
      await screen.findByText("無効なURLです。QRコードを再度読み取ってください"),
    ).toBeInTheDocument();
  });
});
