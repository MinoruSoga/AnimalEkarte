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
  const mutateAsync = vi.fn(
    (vars: TVars) =>
      new Promise<TestEntity>((resolve, reject) => {
        mutate(vars, {
          onSuccess: (saved) => resolve(saved),
          onError: (error) => reject(error),
        });
      }),
  );
  return {
    mutation: { mutate, mutateAsync } as unknown as UseMutationResult<TestEntity, Error, TVars>,
    mutate,
  };
}

const savedEntity: TestEntity = { id: "1", name: "保存済み" };
const allowSavePermissions = { canCreate: true, canEdit: true };

describe("useMasterSave", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("(a) validate失敗時はmutateを呼ばずfalseを返してvalidationErrorを設定する", async () => {
    const { crud } = buildCrud(null);
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation } = buildMutation<{ id: string; req: TestUpdateReq }>();

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        permissions: allowSavePermissions,
        validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    let saveResult: boolean | undefined;
    await act(async () => {
      saveResult = await result.current.handleSave({ name: "" });
    });

    expect(saveResult).toBe(false);
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
        permissions: allowSavePermissions,
        validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => {
      void result.current.handleSave({ name: "" });
    });
    expect(result.current.validationError).toBe("名称は必須です");
    expect(toast.error).toHaveBeenCalledTimes(1);

    act(() => {
      void result.current.handleSave({ name: "有効な名称" });
    });
    expect(result.current.validationError).toBeNull();
    // validate成功時は追加のtoast.errorは呼ばれない(失敗時の1回のみ)
    expect(toast.error).toHaveBeenCalledTimes(1);
  });

  it("create許可時は成功後にtrueを返す", async () => {
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
        permissions: allowSavePermissions,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    let saveResult: boolean | undefined;
    await act(async () => {
      saveResult = await result.current.handleSave({ name: "新規ケージ" });
    });

    expect(saveResult).toBe(true);
    expect(createMutate).toHaveBeenCalledWith({ name: "新規ケージ" }, expect.objectContaining({ onSuccess: expect.any(Function), onError: expect.any(Function) }));
    expect(updateMutate).not.toHaveBeenCalled();
    expect(toast.success).toHaveBeenCalledWith("登録しました");
    expect(setEditTarget).toHaveBeenCalledWith(null);
  });

  it("update許可時は成功後にtrueを返し、idとリクエストを渡す", async () => {
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
        permissions: allowSavePermissions,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    let saveResult: boolean | undefined;
    await act(async () => {
      saveResult = await result.current.handleSave({ name: "更新後の名称" });
    });

    expect(saveResult).toBe(true);
    expect(updateMutate).toHaveBeenCalledWith(
      { id: "42", req: { name: "更新後の名称" } },
      expect.objectContaining({ onSuccess: expect.any(Function), onError: expect.any(Function) }),
    );
    expect(createMutate).not.toHaveBeenCalled();
    expect(toast.success).toHaveBeenCalledWith("更新しました");
    expect(setEditTarget).toHaveBeenCalledWith(null);
  });

  it("canCreateがtrueでない場合はfalseを返してmutationを発行しない", async () => {
    const { crud, startSaveTransition } = buildCrud("new");
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation, mutate: updateMutate } = buildMutation<{ id: string; req: TestUpdateReq }>();
    const onSuccess = vi.fn();

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        permissions: { canCreate: false, canEdit: true },
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
        onSuccess,
      }),
    );

    let saveResult: boolean | undefined;
    await act(async () => {
      saveResult = await result.current.handleSave({ name: "拒否対象" });
    });

    expect(saveResult).toBe(false);
    expect(startSaveTransition).not.toHaveBeenCalled();
    expect(createMutate).not.toHaveBeenCalled();
    expect(updateMutate).not.toHaveBeenCalled();
    expect(onSuccess).not.toHaveBeenCalled();
  });

  it("canEditがtrueでない場合はfalseを返してmutationを発行しない", async () => {
    const editTarget: TestEntity = { id: "42", name: "既存ケージ" };
    const { crud, startSaveTransition } = buildCrud(editTarget);
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation, mutate: updateMutate } = buildMutation<{ id: string; req: TestUpdateReq }>();
    const onSuccess = vi.fn();

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        permissions: { canCreate: true, canEdit: false },
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
        onSuccess,
      }),
    );

    let saveResult: boolean | undefined;
    await act(async () => {
      saveResult = await result.current.handleSave({ name: "拒否対象" });
    });

    expect(saveResult).toBe(false);
    expect(startSaveTransition).not.toHaveBeenCalled();
    expect(createMutate).not.toHaveBeenCalled();
    expect(updateMutate).not.toHaveBeenCalled();
    expect(onSuccess).not.toHaveBeenCalled();
  });

  it("canCreateがtrueならcreate payloadを維持する", () => {
    const { crud } = buildCrud("new");
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation, mutate: updateMutate } = buildMutation<{ id: string; req: TestUpdateReq }>();

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        permissions: { canCreate: true, canEdit: false },
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => {
      void result.current.handleSave({ name: "新規グループ" });
    });

    expect(createMutate).toHaveBeenCalledWith(
      { name: "新規グループ" },
      expect.objectContaining({ onSuccess: expect.any(Function), onError: expect.any(Function) }),
    );
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("canEditがtrueならupdate payloadを維持する", () => {
    const editTarget: TestEntity = { id: "42", name: "既存ケージ" };
    const { crud } = buildCrud(editTarget);
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation, mutate: updateMutate } = buildMutation<{ id: string; req: TestUpdateReq }>();

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        permissions: { canCreate: false, canEdit: true },
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => {
      void result.current.handleSave({ name: "更新後のグループ" });
    });

    expect(updateMutate).toHaveBeenCalledWith(
      { id: "42", req: { name: "更新後のグループ" } },
      expect.objectContaining({ onSuccess: expect.any(Function), onError: expect.any(Function) }),
    );
    expect(createMutate).not.toHaveBeenCalled();
  });

  it("権限剥奪後はcaptured済みhandleSaveでも最新のdenyを使う", () => {
    const { crud } = buildCrud("new");
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation } = buildMutation<{ id: string; req: TestUpdateReq }>();
    const options = {
      crud,
      createMutation,
      updateMutation,
      validate: () => null,
      toCreateRequest: (d: TestForm) => ({ name: d.name }),
      toUpdateRequest: (d: TestForm) => ({ name: d.name }),
    };

    const { result, rerender } = renderHook(
      ({ canCreate }: { canCreate: boolean }) =>
        useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
          ...options,
          permissions: { canCreate, canEdit: false },
        }),
      { initialProps: { canCreate: true } },
    );
    const capturedHandleSave = result.current.handleSave;

    rerender({ canCreate: false });
    act(() => {
      void capturedHandleSave({ name: "拒否対象" });
    });

    expect(createMutate).not.toHaveBeenCalled();
  });

  it("deny時も既存validationを先に実行する", () => {
    const { crud, startSaveTransition } = buildCrud("new");
    const { mutation: createMutation, mutate: createMutate } = buildMutation<TestCreateReq>();
    const { mutation: updateMutation } = buildMutation<{ id: string; req: TestUpdateReq }>();
    const validate = vi.fn(() => "名称は必須です");

    const { result } = renderHook(() =>
      useMasterSave<TestEntity, TestForm, TestCreateReq, TestUpdateReq>({
        crud,
        createMutation,
        updateMutation,
        permissions: { canCreate: false, canEdit: false },
        validate,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => {
      void result.current.handleSave({ name: "" });
    });

    expect(validate).toHaveBeenCalledWith({ name: "" });
    expect(result.current.validationError).toBe("名称は必須です");
    expect(toast.error).toHaveBeenCalledWith("名称は必須です");
    expect(startSaveTransition).not.toHaveBeenCalled();
    expect(createMutate).not.toHaveBeenCalled();
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
        permissions: allowSavePermissions,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    act(() => {
      void result.current.handleSave({ name: "新規" });
    });

    expect(createMutate).toHaveBeenCalled();
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("(d) onSuccessコールバックがrejectした場合はfalseを返してpanelを閉じない", async () => {
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
        permissions: allowSavePermissions,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
        onSuccess,
      }),
    );

    let saveResult: boolean | undefined;
    await act(async () => {
      saveResult = await result.current.handleSave({ name: "新規" });
    });

    expect(saveResult).toBe(false);
    expect(onSuccess).toHaveBeenCalledWith(savedEntity, { name: "新規" });
    expect(handleApiError).toHaveBeenCalledWith(expect.any(Error), "保存");
    expect(setEditTarget).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("mutate自体がonErrorを呼んだ場合はfalseを返す(handleApiErrorはmutation自身のonErrorに委譲し二重通知しない)", async () => {
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
        permissions: allowSavePermissions,
        validate: () => null,
        toCreateRequest: (d) => ({ name: d.name }),
        toUpdateRequest: (d) => ({ name: d.name }),
      }),
    );

    let saveResult: boolean | undefined;
    await act(async () => {
      saveResult = await result.current.handleSave({ name: "更新" });
    });

    expect(saveResult).toBe(false);
    // create/updateMutation の onError (master/api/*.ts) が既に handleApiError 済みのため、
    // ここで再度呼ぶと二重 toast になる。呼ばれないことを確認する。
    expect(handleApiError).not.toHaveBeenCalled();
    expect(setEditTarget).not.toHaveBeenCalled();
  });
});
