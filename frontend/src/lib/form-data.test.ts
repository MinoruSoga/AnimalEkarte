import { describe, expect, it } from "vitest";
import { getFormEnum, getFormOptionalString, getFormString } from "./form-data";

describe("getFormString", () => {
  it("returns string values", () => {
    const fd = new FormData();
    fd.set("name", "  alice  ");
    expect(getFormString(fd, "name")).toBe("  alice  ");
  });

  it("returns empty string for missing keys", () => {
    expect(getFormString(new FormData(), "missing")).toBe("");
  });

  it("returns empty string for File values", () => {
    const fd = new FormData();
    fd.set("file", new File(["x"], "x.txt"));
    expect(getFormString(fd, "file")).toBe("");
  });
});

describe("getFormOptionalString", () => {
  it("returns null when key is absent", () => {
    expect(getFormOptionalString(new FormData(), "x")).toBeNull();
  });

  it("returns string when present", () => {
    const fd = new FormData();
    fd.set("x", "y");
    expect(getFormOptionalString(fd, "x")).toBe("y");
  });
});

describe("getFormEnum", () => {
  const isPeriod = (v: string): v is "am" | "pm" => v === "am" || v === "pm";

  it("returns typed enum when valid", () => {
    const fd = new FormData();
    fd.set("period", "am");
    expect(getFormEnum(fd, "period", isPeriod)).toBe("am");
  });

  it("returns null when invalid", () => {
    const fd = new FormData();
    fd.set("period", "xx");
    expect(getFormEnum(fd, "period", isPeriod)).toBeNull();
  });
});
