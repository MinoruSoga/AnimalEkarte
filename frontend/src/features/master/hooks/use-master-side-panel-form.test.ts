import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useMasterSidePanelForm } from "./use-master-side-panel-form";

interface TestForm {
  name: string;
}

describe("useMasterSidePanelForm", () => {
  it("編集すると isDirty が true になり onDirtyChange が呼ばれる", () => {
    const onDirtyChange = vi.fn();
    const { result } = renderHook(() =>
      useMasterSidePanelForm<TestForm>({
        initialFormData: { name: "" },
        onSave: vi.fn().mockResolvedValue(true),
        onDirtyChange,
      }),
    );

    act(() => {
      result.current.setFormData({ name: "新しい名前" });
    });

    expect(result.current.isDirty).toBe(true);
    expect(onDirtyChange).toHaveBeenCalledWith(true);
  });

  it("onSave が true を返すと isDirty を false に戻す", async () => {
    const onSave = vi.fn().mockResolvedValue(true);
    const onDirtyChange = vi.fn();
    const { result } = renderHook(() =>
      useMasterSidePanelForm<TestForm>({
        initialFormData: { name: "初期名" },
        onSave,
        onDirtyChange,
      }),
    );

    act(() => {
      result.current.setFormData({ name: "更新後" });
    });
    expect(result.current.isDirty).toBe(true);

    await act(async () => {
      await result.current.handleAction();
    });

    expect(onSave).toHaveBeenCalledWith({ name: "更新後" });
    expect(result.current.isDirty).toBe(false);
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);
  });

  it("onSave が false を返すと isDirty を落とさない（保存失敗を未保存のまま維持する）", async () => {
    const onSave = vi.fn().mockResolvedValue(false);
    const { result } = renderHook(() =>
      useMasterSidePanelForm<TestForm>({
        initialFormData: { name: "初期名" },
        onSave,
      }),
    );

    act(() => {
      result.current.setFormData({ name: "権限なしで拒否される変更" });
    });

    await act(async () => {
      await result.current.handleAction();
    });

    expect(onSave).toHaveBeenCalledTimes(1);
    expect(result.current.isDirty).toBe(true);
  });

  it("validate が false を返すと onSave を呼ばない", async () => {
    const onSave = vi.fn().mockResolvedValue(true);
    const validate = vi.fn().mockReturnValue(false);
    const { result } = renderHook(() =>
      useMasterSidePanelForm<TestForm>({
        initialFormData: { name: "" },
        onSave,
        validate,
      }),
    );

    await act(async () => {
      await result.current.handleAction();
    });

    expect(validate).toHaveBeenCalledWith({ name: "" });
    expect(onSave).not.toHaveBeenCalled();
  });

  it("onSave が同期 boolean を返す場合でも動作する", async () => {
    const onSave = vi.fn().mockReturnValue(true);
    const { result } = renderHook(() =>
      useMasterSidePanelForm<TestForm>({
        initialFormData: { name: "初期名" },
        onSave,
      }),
    );

    act(() => {
      result.current.setFormData({ name: "更新後" });
    });

    await act(async () => {
      await result.current.handleAction();
    });

    expect(result.current.isDirty).toBe(false);
  });
});
