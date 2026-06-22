import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { openOwnerReport } from "./owner-report-window";

describe("openOwnerReport", () => {
  let openSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
  });
  afterEach(() => {
    openSpy.mockRestore();
  });

  it("別ウィンドウで /owners/:id/report?petId= を _blank + noopener,noreferrer で開く", () => {
    openOwnerReport("42", "7");
    expect(openSpy).toHaveBeenCalledTimes(1);
    expect(openSpy).toHaveBeenCalledWith(
      "/owners/42/report?petId=7",
      "_blank",
      "noopener,noreferrer",
    );
  });

  it("petId を URL エンコードする", () => {
    openOwnerReport("42", "a b/7");
    expect(openSpy).toHaveBeenCalledWith(
      "/owners/42/report?petId=a%20b%2F7",
      "_blank",
      "noopener,noreferrer",
    );
  });
});
