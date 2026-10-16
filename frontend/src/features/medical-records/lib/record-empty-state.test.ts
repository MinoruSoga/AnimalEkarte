import { describe, expect, it } from "vitest";

import { getLockedEmptyStateMessage } from "./record-empty-state";

describe("getLockedEmptyStateMessage (EMR-412)", () => {
  it.each([
    { isFinalized: true, isLocked: true },
    { isFinalized: true, isLocked: false },
  ])("確定済みは権限によるロックより優先する: %o", (input) => {
    expect(getLockedEmptyStateMessage(input)).toBe("確定済のため変更できません");
  });

  it("未確定で編集権限がない場合は権限を理由として伝える", () => {
    expect(getLockedEmptyStateMessage({ isFinalized: false, isLocked: true })).toBe(
      "編集権限がありません",
    );
  });

  it("編集できる場合は通常の空状態を呼び出し側に任せる", () => {
    expect(getLockedEmptyStateMessage({ isFinalized: false, isLocked: false })).toBeNull();
  });
});
