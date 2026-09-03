import { describe, it, expect } from "vitest";
import {
  CREATE_ALLOWED_STATUSES,
  CREATE_STATUS_OPTIONS,
  EDIT_STATUS_OPTIONS,
} from "./estimate-status-options";

describe("estimate-status-options 契約", () => {
  it("CREATE_ALLOWED_STATUSES は draft / sent のみ", () => {
    expect([...CREATE_ALLOWED_STATUSES]).toEqual(["draft", "sent"]);
  });

  it("Create 用選択肢は CREATE_ALLOWED_STATUSES と一致（draft / sent）", () => {
    expect(CREATE_STATUS_OPTIONS.map((o) => o.value)).toEqual([...CREATE_ALLOWED_STATUSES]);
  });

  it("Edit 用選択肢は draft / sent / approved / rejected の 4 値", () => {
    expect(EDIT_STATUS_OPTIONS.map((o) => o.value)).toEqual([
      "draft",
      "sent",
      "approved",
      "rejected",
    ]);
  });

  it("Create 用選択肢に approved / rejected は含まれない", () => {
    const values = CREATE_STATUS_OPTIONS.map((o) => o.value);
    expect(values).not.toContain("approved");
    expect(values).not.toContain("rejected");
  });
});
