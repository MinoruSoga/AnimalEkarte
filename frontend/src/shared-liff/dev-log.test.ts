import { afterEach, describe, expect, it, vi } from "vitest";
import { devError } from "./dev-log";

describe("devError（FE-RC-077: DEV限定のconsole.errorラッパー）", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("DEV環境ではconsole.errorに引数をそのまま渡す", () => {
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.stubEnv("DEV", true);

    devError("[test] BE body:", { status: 500, body: "secret detail" });

    expect(spy).toHaveBeenCalledWith("[test] BE body:", {
      status: 500,
      body: "secret detail",
    });

    vi.unstubAllEnvs();
  });

  it("本番相当（DEV=false）ではconsole.errorを呼ばない", () => {
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.stubEnv("DEV", false);

    devError("[test] BE body:", { status: 500, body: "secret detail" });

    expect(spy).not.toHaveBeenCalled();

    vi.unstubAllEnvs();
  });
});
