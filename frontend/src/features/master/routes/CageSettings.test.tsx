import { describe, it, expect, afterEach, vi } from "vitest";
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryRouter, RouterProvider } from "react-router";
import { useState } from "react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import type { Cage as ModelCage } from "@/types/generated/models";
import { CageSettings } from "./CageSettings";

const permissionMock = vi.hoisted(() => ({
  current: {
    view: true,
    create: true,
    edit: true,
    delete: true,
  },
}));

// R-F15: use-master-crud / use-master-save の共有状態機械を、実ページ経由で
// end-to-end に検証するリファレンス実装（横展開の型）。
// PageLayout の resource prop 経由で PermissionBadges → usePermission → useAuth
// が呼ばれるため、CashRegisterHistoryPage.test.tsx と同パターンで最小限モックする。
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }] },
    currentClinicId: "1",
    hasPermission: (_resource: unknown, action: keyof typeof permissionMock.current) =>
      permissionMock.current[action],
  }),
}));

function makeCage(
  id: number,
  name: string,
  // EMR-306: merged_into_cage_id は generated ModelCage に含まれる（codegen 済み）。
  overrides: Partial<ModelCage> = {},
): ModelCage {
  return {
    id,
    clinic_id: 1,
    name,
    cage_type: "general",
    cage_size: "medium",
    price: 1000,
    description: "",
    is_active: true,
    sort_order: id,
    created_at: "2026-06-01T00:00:00Z",
    updated_at: "2026-06-01T00:00:00Z",
    ...overrides,
  };
}

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  let refreshPermissions = () => undefined;
  function PermissionHarness() {
    const [, setPermissionVersion] = useState(0);
    refreshPermissions = () => setPermissionVersion((version) => version + 1);
    return <CageSettings />;
  }
  // NavigationBlocker が useBlocker(react-router のデータルーターAPI)を使うため、
  // MemoryRouter ではなく createMemoryRouter + RouterProvider を使う。
  const router = createMemoryRouter([{ path: "/", element: <PermissionHarness /> }], {
    initialEntries: ["/"],
  });
  const result = render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return {
    ...result,
    refreshPermissions,
  };
}

const stubCages = (cages: ModelCage[]) => {
  server.use(http.get("*/v1/masters/cages", () => HttpResponse.json(cages)));
};

afterEach(() => {
  permissionMock.current = {
    view: true,
    create: true,
    edit: true,
    delete: true,
  };
  server.resetHandlers();
});

describe("CageSettings (useMasterCRUD / useMasterSave 統合テスト)", () => {
  it("一覧を取得し表示する", async () => {
    stubCages([makeCage(1, "1番ケージ")]);
    renderPage();

    expect(await screen.findByText("1番ケージ")).toBeInTheDocument();
  });

  it("新規作成: サイドパネルで名称を入力して保存するとcreateエンドポイントを呼びリストに反映する", async () => {
    stubCages([makeCage(1, "1番ケージ")]);

    let createdBody: unknown;
    server.use(
      http.post("*/v1/masters/cages", async ({ request }) => {
        createdBody = await request.json();
        stubCages([makeCage(1, "1番ケージ"), makeCage(2, "2番ケージ")]);
        return HttpResponse.json(makeCage(2, "2番ケージ"));
      }),
    );

    const user = userEvent.setup();
    renderPage();

    expect(await screen.findByText("1番ケージ")).toBeInTheDocument();

    await user.click(await screen.findByRole("button", { name: "新規登録" }));

    const titleInput = await screen.findByPlaceholderText("無題");
    await user.type(titleInput, "2番ケージ");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(createdBody).toMatchObject({ name: "2番ケージ" });
    });
    expect(await screen.findByText("2番ケージ")).toBeInTheDocument();
  });

  it("編集: 固有名の行操作buttonでサイドパネルに現在値が表示され、保存でupdateエンドポイントを呼ぶ", async () => {
    stubCages([makeCage(1, "1番ケージ", { price: 500 })]);

    let updatedId: string | undefined;
    let updatedBody: unknown;
    server.use(
      http.patch("*/v1/masters/cages/:id", async ({ request, params }) => {
        updatedId = String(params.id);
        updatedBody = await request.json();
        return HttpResponse.json(makeCage(1, "1番ケージ(改)"));
      }),
    );

    const user = userEvent.setup();
    renderPage();

    await user.click(
      await screen.findByRole("button", {
        name: "詳細: ケージ 1番ケージ (ID 1)",
      }),
    );

    const titleInput = await screen.findByDisplayValue("1番ケージ");
    await user.clear(titleInput);
    await user.type(titleInput, "1番ケージ(改)");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(updatedId).toBe("1");
    });
    expect(updatedBody).toMatchObject({ name: "1番ケージ(改)" });
  });

  it("削除: 確認ダイアログでの確定によりdeleteエンドポイントを呼び、サイドパネルが閉じる", async () => {
    stubCages([makeCage(1, "1番ケージ")]);

    let deletedId: string | undefined;
    server.use(
      http.delete("*/v1/masters/cages/:id", ({ params }) => {
        deletedId = String(params.id);
        stubCages([]);
        return new HttpResponse(null, { status: 204 });
      }),
    );

    const user = userEvent.setup();
    renderPage();

    await user.click(
      await screen.findByRole("button", {
        name: "詳細: ケージ 1番ケージ (ID 1)",
      }),
    );
    await user.click(await screen.findByRole("button", { name: "削除" }));

    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "削除" }));

    await waitFor(() => {
      expect(deletedId).toBe("1");
    });
    await waitFor(() => {
      expect(screen.queryByPlaceholderText("無題")).not.toBeInTheDocument();
    });
  });

  it("削除確認後にdelete権限を剥奪された場合はdeleteエンドポイントを呼ばない", async () => {
    stubCages([makeCage(1, "1番ケージ")]);
    const deleteSpy = vi.fn();
    server.use(
      http.delete("*/v1/masters/cages/:id", () => {
        deleteSpy();
        return new HttpResponse(null, { status: 204 });
      }),
    );

    const user = userEvent.setup();
    const { refreshPermissions } = renderPage();

    await user.click(
      await screen.findByRole("button", {
        name: "詳細: ケージ 1番ケージ (ID 1)",
      }),
    );
    await user.click(await screen.findByRole("button", { name: "削除" }));
    expect(await screen.findByRole("alertdialog")).toBeInTheDocument();

    permissionMock.current = {
      ...permissionMock.current,
      delete: false,
    };
    act(() => {
      refreshPermissions();
    });

    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "削除" }));

    expect(deleteSpy).not.toHaveBeenCalled();
  });

  it("名称未入力での保存はクライアント側バリデーションで止まりcreateエンドポイントを呼ばない", async () => {
    stubCages([]);
    const createSpy = vi.fn();
    server.use(
      http.post("*/v1/masters/cages", async ({ request }) => {
        createSpy(await request.json());
        return HttpResponse.json(makeCage(1, "無題"));
      }),
    );

    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "新規登録" }));
    await screen.findByPlaceholderText("無題");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("名称を入力してください")).toBeInTheDocument();
    expect(createSpy).not.toHaveBeenCalled();
  });
});

describe("CageSettings ケージ連結 (EMR-306)", () => {
  it("2件選択し代表を選んで確定すると merge エンドポイントを呼ぶ", async () => {
    stubCages([makeCage(1, "1番ケージ"), makeCage(2, "2番ケージ"), makeCage(3, "3番ケージ")]);

    let mergedId: string | undefined;
    let mergeBody: unknown;
    server.use(
      http.post("*/v1/masters/cages/:id/merge", async ({ request, params }) => {
        mergedId = String(params.id);
        mergeBody = await request.json();
        stubCages([
          makeCage(1, "1番ケージ"),
          makeCage(2, "2番ケージ", { merged_into_cage_id: 3 }),
          makeCage(3, "3番ケージ"),
        ]);
        return HttpResponse.json(makeCage(3, "3番ケージ"));
      }),
    );

    const user = userEvent.setup();
    renderPage();
    await screen.findByText("1番ケージ");

    await user.click(screen.getByRole("checkbox", { name: "連結選択: ケージ 1番ケージ (ID 1)" }));
    await user.click(screen.getByRole("checkbox", { name: "連結選択: ケージ 3番ケージ (ID 3)" }));

    await user.click(screen.getByRole("button", { name: "選択したケージを連結" }));

    const dialog = await screen.findByRole("alertdialog");
    // 代表は選択肢から明示的に選べる（既定は先頭選択だが変更可能）
    await user.click(within(dialog).getByRole("radio", { name: "3番ケージ" }));
    await user.click(within(dialog).getByRole("button", { name: "連結する" }));

    await waitFor(() => {
      expect(mergedId).toBe("3");
    });
    expect(mergeBody).toEqual({ member_cage_ids: [1] });
  });

  it("連結メンバー行には代表ケージ名入りの連結中バッジが出て、連結済み行の選択チェックボックスは無効", async () => {
    stubCages([
      makeCage(1, "1番ケージ"),
      makeCage(2, "2番ケージ", { merged_into_cage_id: 1 }),
      makeCage(3, "3番ケージ"),
    ]);

    renderPage();
    await screen.findByText("1番ケージ");

    // メンバー行には代表名入りバッジ、代表行にも連結中表示が出る（EMR-306 レビュー対応）
    const badge = await screen.findByText("連結中: 1番ケージ");
    expect(badge).toBeInTheDocument();
    expect(screen.getByText("連結中（代表）")).toBeInTheDocument();
    // 連結済み（メンバー・代表）行は merge 対象に選べない
    expect(
      screen.getByRole("checkbox", { name: "連結選択: ケージ 1番ケージ (ID 1)" }),
    ).toBeDisabled();
    expect(
      screen.getByRole("checkbox", { name: "連結選択: ケージ 2番ケージ (ID 2)" }),
    ).toBeDisabled();
    // 未連結行は選択可能
    expect(
      screen.getByRole("checkbox", { name: "連結選択: ケージ 3番ケージ (ID 3)" }),
    ).toBeEnabled();
  });

  it("連結解除: メンバー行の解除アクションを確認ダイアログで確定すると代表IDの unmerge を呼ぶ", async () => {
    stubCages([makeCage(1, "1番ケージ"), makeCage(2, "2番ケージ", { merged_into_cage_id: 1 })]);

    let unmergedId: string | undefined;
    server.use(
      http.post("*/v1/masters/cages/:id/unmerge", ({ params }) => {
        unmergedId = String(params.id);
        stubCages([makeCage(1, "1番ケージ"), makeCage(2, "2番ケージ")]);
        return new HttpResponse(null, { status: 204 });
      }),
    );

    const user = userEvent.setup();
    renderPage();
    await screen.findByText("1番ケージ");

    await user.click(screen.getByRole("button", { name: "連結解除: ケージ 2番ケージ (ID 2)" }));
    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "解除" }));

    await waitFor(() => {
      expect(unmergedId).toBe("1");
    });
  });

  it("1件しか選択していないときは連結ボタンが無効", async () => {
    stubCages([makeCage(1, "1番ケージ"), makeCage(2, "2番ケージ")]);

    const user = userEvent.setup();
    renderPage();
    await screen.findByText("1番ケージ");

    expect(screen.getByRole("button", { name: "選択したケージを連結" })).toBeDisabled();

    await user.click(screen.getByRole("checkbox", { name: "連結選択: ケージ 1番ケージ (ID 1)" }));

    expect(screen.getByRole("button", { name: "選択したケージを連結" })).toBeDisabled();

    await user.click(screen.getByRole("checkbox", { name: "連結選択: ケージ 2番ケージ (ID 2)" }));

    expect(screen.getByRole("button", { name: "選択したケージを連結" })).toBeEnabled();
  });

  it("連結メンバーのサイドパネルには代表ケージ名入りの連結状態行が表示される", async () => {
    stubCages([makeCage(1, "1番ケージ"), makeCage(2, "2番ケージ", { merged_into_cage_id: 1 })]);

    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "詳細: ケージ 2番ケージ (ID 2)" }));

    // サイドパネルに「連結状態」PropertyRow が出る
    expect(await screen.findByText("連結状態")).toBeInTheDocument();
    // 「連結中: <代表名>」は行バッジとパネルバッジの2箇所に出る
    expect(await screen.findAllByText("連結中: 1番ケージ")).toHaveLength(2);
  });

  it("edit 権限がない場合は選択チェックボックスが無効で連結ボタンは出ない", async () => {
    permissionMock.current = { ...permissionMock.current, edit: false };
    stubCages([makeCage(1, "1番ケージ")]);

    renderPage();
    await screen.findByText("1番ケージ");

    expect(
      screen.getByRole("checkbox", { name: "連結選択: ケージ 1番ケージ (ID 1)" }),
    ).toBeDisabled();
    expect(screen.queryByRole("button", { name: "選択したケージを連結" })).not.toBeInTheDocument();
  });
});
