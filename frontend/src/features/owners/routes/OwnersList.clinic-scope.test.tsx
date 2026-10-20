import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act } from "react";
import { ME_QUERY_KEY } from "@/lib/query-keys";
import { AuthContext } from "@/hooks/auth-context";
import { OwnersList } from "./OwnersList";
import type { OwnersLoaderData } from "../loaders";
import type { AuthUser, ResourceAction } from "@/types/auth";
import type { Pet } from "@/types";

// EMR-463: このファイルは活性時（OWNERS_CLINIC_SCOPE_ACTIVE=true）の経路を検証する。
vi.mock("../lib/owners-clinic-scope-activation", () => ({ OWNERS_CLINIC_SCOPE_ACTIVE: true }));

// EMR-463: 一覧の医院フィルタは loader が確定した現在許可医院だけを表示し、
// 明示選択を URL ?clinics= と保存先 owners-list.clinics.v1.<user.id> の両方へ反映する。
// 値はすべて合成（実患者・実飼主の値なし）。

vi.mock("@/features/owners/components/PetEditModal", () => ({
  PetEditModal: () => null,
}));

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: vi.fn(() => ({
    activeSpecies: [{ id: 1, name: "犬" }],
    allSpecies: [{ id: 1, name: "犬" }],
  })),
}));

const USER_ID = "10";
const STORAGE_KEY = `owners-list.clinics.v1.${USER_ID}`;

function makeAuthCtx(userId = USER_ID) {
  const user: AuthUser = {
    id: userId,
    email: "staff@example.com",
    displayName: "合成スタッフ",
    isSystemAdmin: false,
    mainClinicId: "1",
    clinic: null,
    // AuthProvider の所属一覧は表示医院の根拠にしない（3 は許可外の所属）
    clinics: [
      { clinicId: "1", clinicName: "合成本院", isMain: true },
      { clinicId: "2", clinicName: "合成分院", isMain: false },
      { clinicId: "3", clinicName: "合成許可外院", isMain: false },
    ],
    permissions: {},
  };
  return {
    user,
    currentClinicId: "1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: (_resource: string, _action: ResourceAction) => true,
    refreshPermissions: async () => {},
  };
}

function makePet(): Pet {
  return {
    id: "1",
    clinicId: "1",
    ownerId: "1",
    ownerName: "合成飼主",
    ownerNumber: 1,
    name: "合成ペット",
    species: "犬",
  } as unknown as Pet;
}

const TWO_CLINICS = [
  { clinicId: "1", clinicName: "合成本院", isMain: true },
  { clinicId: "2", clinicName: "合成分院", isMain: false },
];

function readyData(selectedClinicIds: string[]): OwnersLoaderData {
  return {
    pets: [makePet()],
    page: 1,
    limit: 20,
    total: 1,
    listState: { kind: "ready", scopeRechecked: false },
    clinicScope: { userId: USER_ID, clinics: TWO_CLINICS, selectedClinicIds },
  };
}

function renderList(
  loader: (request: Request) => OwnersLoaderData | Promise<OwnersLoaderData>,
  {
    initialPath = "/owners",
    initialEntries = [initialPath],
    authUserId = USER_ID,
    queryClient,
  }: {
    initialPath?: string;
    initialEntries?: string[];
    authUserId?: string;
    queryClient?: QueryClient;
  } = {},
) {
  const router = createMemoryRouter(
    [
      { path: "/elsewhere", element: <p>合成の別画面</p> },
      { path: "/owners", element: <OwnersList />, loader: ({ request }) => loader(request) },
    ],
    { initialEntries, initialIndex: initialEntries.length - 1 },
  );
  const tree = (
    <AuthContext.Provider value={makeAuthCtx(authUserId)}>
      <RouterProvider router={router} />
    </AuthContext.Provider>
  );
  render(
    queryClient ? <QueryClientProvider client={queryClient}>{tree}</QueryClientProvider> : tree,
  );
  return router;
}

const THREE_CLINICS = [
  ...TWO_CLINICS,
  { clinicId: "3", clinicName: "合成第三院", isMain: false },
];

function readyDataFromUrl(request: Request): OwnersLoaderData {
  const clinics = new URL(request.url).searchParams.get("clinics");
  return {
    ...readyData([]),
    clinicScope: {
      userId: USER_ID,
      clinics: THREE_CLINICS,
      selectedClinicIds: clinics ? clinics.split(",") : ["1", "2", "3"],
    },
  };
}

beforeEach(() => {
  window.localStorage.clear();
});

describe("OwnersList — EMR-463 現在許可医院のフィルタ", () => {
  it("許可医院だけを表示し、loader の有効選択（初回は全許可）を選択状態にする", async () => {
    renderList(() => readyData(["1", "2"]));

    expect(await screen.findByRole("button", { name: "合成本院" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    expect(screen.getByRole("button", { name: "合成分院" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    expect(screen.queryByRole("button", { name: "合成許可外院" })).not.toBeInTheDocument();
  });

  it("明示選択は URL ?clinics= と保存先の両方へ反映し、page を戻す", async () => {
    const requests: string[] = [];
    const router = renderList(
      (request) => {
        requests.push(new URL(request.url).search);
        const clinics = new URL(request.url).searchParams.get("clinics");
        return readyData(clinics ? clinics.split(",") : ["1", "2"]);
      },
      { initialPath: "/owners?page=2&search=%E5%90%88%E6%88%90" },
    );

    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "合成分院" }));

    await waitFor(() => {
      expect(router.state.location.search).toContain("clinics=1");
    });
    expect(router.state.location.search).not.toContain("page=2");
    // 既存の検索条件は保持する
    expect(router.state.location.search).toContain("search=");
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('["1"]');
    expect(requests.at(-1)).toContain("clinics=1");
  });

  it("最後の 1 院は解除できない（空選択を URL・保存へ書かない）", async () => {
    const router = renderList(() => readyData(["1"]));

    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "合成本院" }));

    expect(router.state.location.search).not.toContain("clinics=");
    expect(window.localStorage.getItem(STORAGE_KEY)).toBeNull();
  });

  it("ログイン中の利用者と scope の利用者が異なる旧画面では保存しない", async () => {
    renderList(() => readyData(["1", "2"]), { authUserId: "11" });

    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "合成分院" }));

    expect(window.localStorage.getItem(STORAGE_KEY)).toBeNull();
    expect(window.localStorage.getItem("owners-list.clinics.v1.11")).toBeNull();
  });

  it("許可医院が 1 院ならフィルタを表示しない", async () => {
    renderList(() => ({
      ...readyData(["1"]),
      clinicScope: { userId: USER_ID, clinics: [TWO_CLINICS[0]], selectedClinicIds: ["1"] },
    }));

    await screen.findByText("合成飼主");
    expect(screen.queryByTestId("clinic-scope-filter")).not.toBeInTheDocument();
  });
});

describe("OwnersList — EMR-463 一覧を取得しなかった状態の表示", () => {
  it("scope 取得失敗は一覧 0 件と区別したエラーを出し、フィルタを出さない", async () => {
    renderList(() => ({
      pets: [],
      page: 1,
      limit: 20,
      total: 0,
      listState: { kind: "scope-error" },
      clinicScope: null,
    }));

    expect(await screen.findByRole("alert")).toHaveTextContent("表示できる医院を確認できなかった");
    expect(screen.getByRole("button", { name: "再読み込み" })).toBeInTheDocument();
    expect(screen.queryByTestId("clinic-scope-filter")).not.toBeInTheDocument();
  });

  it("許可ゼロは「何が無いか／条件／次の操作」を示す", async () => {
    renderList(() => ({
      pets: [],
      page: 1,
      limit: 20,
      total: 0,
      listState: { kind: "no-permitted-clinics" },
      clinicScope: { userId: USER_ID, clinics: [], selectedClinicIds: [] },
    }));

    expect(
      await screen.findByText("飼主・ペット一覧を表示できる医院がありません。"),
    ).toBeInTheDocument();
    expect(screen.getByText(/閲覧権限がありません/)).toBeInTheDocument();
    expect(screen.getByText(/管理者に権限の付与を依頼/)).toBeInTheDocument();
  });

  it("選択超過はフィルタを残して選択を減らす操作を案内する", async () => {
    const clinics = Array.from({ length: 51 }, (_, i) => ({
      clinicId: String(i + 1),
      clinicName: `合成医院${i + 1}`,
      isMain: i === 0,
    }));
    renderList(() => ({
      pets: [],
      page: 1,
      limit: 20,
      total: 0,
      listState: { kind: "selection-over-limit", selectedCount: 51, limit: 50 },
      clinicScope: {
        userId: USER_ID,
        clinics,
        selectedClinicIds: clinics.map((c) => c.clinicId),
      },
    }));

    expect(await screen.findByRole("alert")).toHaveTextContent("51院選択されている");
    expect(screen.getByTestId("clinic-scope-filter")).toBeInTheDocument();
  });

  it("403 後の再照合で取得し直したときは通知する", async () => {
    renderList(() => ({
      ...readyData(["1"]),
      listState: { kind: "ready", scopeRechecked: true },
    }));

    expect(await screen.findByText(/表示できる医院が変わったため/)).toBeInTheDocument();
    expect(screen.getByText("合成飼主")).toBeInTheDocument();
  });
});

describe("OwnersList — EMR-463 トグルの合成・履歴・/me 更新", () => {
  it("loader 進行中の連続トグルは直前のトグル結果から合成する（古い選択で上書きしない）", async () => {
    let callCount = 0;
    const router = renderList((request) => {
      callCount += 1;
      if (callCount === 1) return readyDataFromUrl(request);
      // 2 回目以降の loader は遅延させ、進行中に次のトグルを押せるようにする
      return new Promise((resolve) => setTimeout(() => resolve(readyDataFromUrl(request)), 50));
    });

    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "合成分院" }));
    await user.click(screen.getByRole("button", { name: "合成第三院" }));

    await waitFor(() => {
      expect(new URLSearchParams(router.state.location.search).get("clinics")).toBe("1");
    });
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('["1"]');
  });

  it("暗黙選択からの最初のトグルは replace、以後の明示選択からのトグルは push", async () => {
    const router = renderList(readyDataFromUrl, { initialEntries: ["/elsewhere", "/owners"] });

    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "合成分院" }));
    await waitFor(() => {
      expect(router.state.location.search).toContain("clinics=1%2C3");
    });
    expect(router.state.historyAction).toBe("REPLACE");

    await user.click(screen.getByRole("button", { name: "合成第三院" }));
    await waitFor(() => {
      expect(new URLSearchParams(router.state.location.search).get("clinics")).toBe("1");
    });
    expect(router.state.historyAction).toBe("PUSH");

    // 戻る 1 回で明示選択 1,3 へ、もう 1 回で一覧の外へ（暗黙状態の履歴は残らない）
    await act(async () => {
      await router.navigate(-1);
    });
    await waitFor(() => {
      expect(new URLSearchParams(router.state.location.search).get("clinics")).toBe("1,3");
    });
    await act(async () => {
      await router.navigate(-1);
    });
    await waitFor(() => {
      expect(router.state.location.pathname).toBe("/elsewhere");
    });
  });

  it("ME_QUERY_KEY のデータ更新で owners loader を再検証する（profile は根拠にしない）", async () => {
    const queryClient = new QueryClient();
    const loader = vi.fn(readyDataFromUrl);
    renderList(loader, { queryClient });
    await screen.findByText("合成飼主");
    expect(loader).toHaveBeenCalledTimes(1);

    act(() => {
      queryClient.setQueryData(ME_QUERY_KEY, { id: USER_ID });
    });

    await waitFor(() => {
      expect(loader).toHaveBeenCalledTimes(2);
    });
  });

  it("ME_QUERY_KEY 以外の query 更新では再検証しない", async () => {
    const queryClient = new QueryClient();
    const loader = vi.fn(readyDataFromUrl);
    renderList(loader, { queryClient });
    await screen.findByText("合成飼主");

    act(() => {
      queryClient.setQueryData(["unrelated"], { value: 1 });
    });

    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(loader).toHaveBeenCalledTimes(1);
  });
});
