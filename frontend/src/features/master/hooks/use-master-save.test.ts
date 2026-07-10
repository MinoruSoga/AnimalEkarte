import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { toast } from "sonner";
import type { UseMutationResult } from "@tanstack/react-query";
import { useMasterSave } from "./use-master-save";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

interface TestEntity {
  id: string;
  name: string;
}
interface TestForm {
  name: string;
}
interface TestCreateReq {
  name: string;
}
interface TestUpdateReq {
  name: string;
}

type MutationOpts<TData> = {
  onSuccess?: (data: TData) => void | Promise<void>;
  onError?: (error: Error) => void;
};

function buildCrud(editTarget: TestEntity | "new" | null) {
  const setEditTarget = vi.fn();
  const startSaveTransition = vi.fn((cb: () => void) => cb());
  return { crud: { editTarget, setEditTarget, startSaveTransition }, setEditTarget, startSaveTransition };
}

function buildMutation<TVars>(
  impl?: (vars: TVars, opts?: MutationOpts<TestEntity>) => void,
): { mutation: UseMutationResult<TestEntity, Error, TVars>; mutate: ReturnType<typeof vi.fn> } {
  const mutate = vi.fn(impl);
  return { mutation: { mutate } as unknown as UseMutationResult<TestEntity, Error, TVars>, mutate };
}

const savedEntity: TestEntity = { id: "1", name: "保存済み" };

describe("useMasterSave", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("(a) validate失敗時はmutateを呼ばずvalidationErrorを設定する", () => {
    const { crud } = buildCrud(null);
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation } = buildMutation<{ id: string; req: TestUpdateReq }>();

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => result.current.handleSave({ name: "" }));

    expect(result.current.validationError).toBe("名称は必須です");
    expect(createMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("名称は必須です");
    expect(toast.error).toHaveBeenCalledTimes(1);
  });

  it("validate成功時はvalidationErrorをクリアする", () => {
    const { crud } = buildCrud(null);
    const { mutation: createMutation } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation } = buildMutation<{ id: string; req: TestUpdateReq }>();

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => result.current.handleSave({ name: "" }));
    expect(result.current.validationError).toBe("名称は必須です");
    expect(toast.error).toHaveBeenCalledTimes(1);

    act(() => result.current.handleSave({ name: "有効な名称" }));
    expect(result.current.validationError).toBeNull();
    // validate成功時は追加のtoast.errorは呼ばれない(失敗時の1回のみ)
    expect(toast.error).toHaveBeenCalledTimes(1);
  });

  it("(b) editTargetId===nullの場合はcreateMutation経路を通る", () => {
    const { crud, setEditTarget } = buildCrud(null);
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>(
      (_vars, opts) => opts?.onSuccess?.(savedEntity),
    );
    const { mutation: updateMutation, mutate: updateMutate } = buildMutation<{ id: string; req: TestUpdateReq }>();

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => result.current.handleSave({ name: "新規ケージ" }));

    expect(createMutate).toHaveBeenCalledWith({ name: "新規ケージ" }, expect.objectContaining({ onSuccess: expect.any(Function), onError: expect.any(Function) }));
    expect(updateMutate).not.toHaveBeenCalled();
    expect(toast.success).toHaveBeenCalledWith("登録しました");
    expect(setEditTarget).toHaveBeenCalledWith(null);
  });

  it("(c) editTargetIdが存在する場合はupdateMutation経路を通り、idとリクエストを渡す", () => {
    const editTarget: TestEntity = { id: "42", name: "既存ケージ" };
    const { crud, setEditTarget } = buildCrud(editTarget);
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation, mutate: updateMutate } = buildMutation<{ id: string; req: TestUpdateReq }>(
      (_vars, opts) => opts?.onSuccess?.(savedEntity),
    );

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => result.current.handleSave({ name: "更新後の名称" }));

    expect(updateMutate).toHaveBeenCalledWith(
      { id: "42", req: { name: "更新後の名称" } },
      expect.objectContaining({ onSuccess: expect.any(Function), onError: expect.any(Function) }),
    );
    expect(createMutate).not.toHaveBeenCalled();
    expect(toast.success).toHaveBeenCalledWith("更新しました");
    expect(setEditTarget).toHaveBeenCalledWith(null);
  });

  it("editTarget==='new'の場合はcreateMutation経路を通る(editTargetIdはnull扱い)", () => {
    const { crud } = buildCrud("new");
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>(
      (_vars, opts) => opts?.onSuccess?.(savedEntity),
    );
    const { mutation: updateMutation, mutate: updateMutate } = buildMutation<{ id: string; req: TestUpdateReq }>();

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => result.current.handleSave({ name: "新規" }));

    expect(createMutate).toHaveBeenCalled();
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("(d) onSuccessコールバックがrejectした場合はhandleApiErrorが呼ばれcrudSetEditTarget(null)されない", async () => {
    const { handleApiError } = await import("@/lib/handle-api-error");
    const { crud, setEditTarget } = buildCrud(null);
    const { mutation: createMutation } = buildMutation<TestCreateReq>((_vars, opts) => {
      void opts?.onSuccess?.(savedEntity);
    });
    const { mutation: updateMutation } = buildMutation<{ id: string; req: TestUpdateReq }>();

    const onSuccess = vi.fn().mockRejectedValue(new Error("post-save failed"));

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
        onSuccess,
      }),
    );

    await act(async () => {
      result.current.handleSave({ name: "新規" });
      // onSuccess は async 関数内の await を挟むため、マイクロタスクをフラッシュする
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(onSuccess).toHaveBeenCalledWith(savedEntity, { name: "新規" });
    expect(handleApiError).toHaveBeenCalledWith(expect.any(Error), "保存");
    expect(setEditTarget).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("mutate自体がonErrorを呼んだ場合はhandleApiErrorが'更新'/'登録'ラベルで呼ばれる", async () => {
    const { handleApiError } = await import("@/lib/handle-api-error");
    const editTarget: TestEntity = { id: "42", name: "既存ケージ" };
    const { crud, setEditTarget } = buildCrud(editTarget);
    const { mutation: createMutation } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation } = buildMutation<{ id: string; req: TestUpdateReq }>(
      (_vars, opts) => opts?.onError?.(new Error("network error")),
    );

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => result.current.handleSave({ name: "更新" }));

    expect(handleApiError).toHaveBeenCalledWith(expect.any(Error), "更新");
    expect(setEditTarget).not.toHaveBeenCalled();
  });
});
