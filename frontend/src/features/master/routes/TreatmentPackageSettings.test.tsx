import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import type { ReactNode } from "react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import { TreatmentPackageSettings } from "./TreatmentPackageSettings";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn(), post: vi.fn(), patch: vi.fn(), delete: vi.fn() },
}));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

let permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
// Q3: 型別 view 権限（checkups / master-merchandise）の上書き。未指定は permission と同じ。
let kindPermission: Record<string, Partial<typeof permission>> = {};
vi.mock("@/hooks/use-permission", () => ({
  usePermission: (resource: string) => ({ ...permission, ...kindPermission[resource] }),
}));
// R2: 再確認は /me の権限キャッシュ更新（refreshPermissions）を先に待つ。既定は成功。
// useAuth は安定参照を返し、テストごとに beforeEach で既定へ戻す。
const refreshPermissionsMock = vi.fn<() => Promise<void>>();
const authValue = { refreshPermissions: refreshPermissionsMock };
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => authValue,
}));
const dirtyMock = {
  markDirty: vi.fn(),
  markClean: vi.fn(),
  runWithDiscardCheck: (cb: () => void) => cb(),
  discardDialog: null as ReactNode,
};
vi.mock("@/hooks/use-side-peek-dirty", () => ({
  useSidePeekDirty: () => dirtyMock,
}));

const render_ = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  );
  return { ...render(<TreatmentPackageSettings />, { wrapper }), client };
};

const apiError = (status: number) => {
  const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
  return new AxiosError(`status ${status}`, String(status), config, undefined, {
    status,
    statusText: "",
    headers: {},
    config,
    data: { error: "e", code: "internal-error" },
  });
};

const pkg = (id: string, name: string, active = true) => ({
  id,
  clinic_id: "7",
  version: "3",
  name,
  description: "",
  is_active: active,
  sort_order: 0,
  created_at: "t",
  updated_at: "t",
  components: [
    {
      component_key: "a0000000-0000-4000-8000-000000000011",
      item: { kind: "consultation", consultation_id: "101" },
      quantity: "1",
    },
  ],
});

describe("TreatmentPackageSettings（読込/0件/取得失敗/失権・権限差）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  it("0件は empty state", async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { items: [], order_version: "17" } });
    render_();
    expect(await screen.findByText(/登録されたパッケージはありません/)).toBeTruthy();
  });

  it("取得失敗は error state", async () => {
    vi.mocked(axios.get).mockRejectedValue(new Error("500"));
    render_();
    expect(await screen.findByText(/取得に失敗しました/)).toBeTruthy();
  });

  it("失権（viewなし）は一覧を出さない", async () => {
    permission = { canView: false, canCreate: false, canEdit: false, canDelete: false };
    render_();
    expect(await screen.findByText(/権限がありません/)).toBeTruthy();
    expect(axios.get).toHaveBeenCalled();
  });

  it("表示のみ権限では reorder/編集操作を出さない", async () => {
    permission = { canView: true, canCreate: false, canEdit: false, canDelete: false };
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA", false)], order_version: "17" },
    });
    render_();
    expect(await screen.findByText("セットA")).toBeTruthy();
    expect(screen.getByText("無効")).toBeTruthy(); // inactive も一覧へ含む
    expect(screen.queryByLabelText(/を上へ/)).toBeNull();
  });

  it("編集権限ありなら reorder ボタンを出す（全件・inactive込み）", async () => {
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "A", false), pkg("202", "B")], order_version: "17" },
    });
    render_();
    expect(await screen.findByText("A")).toBeTruthy();
    expect(screen.getAllByLabelText(/を上へ|を下へ/).length).toBeGreaterThan(0);
  });
});

const LIST_URL = "/v1/masters/treatment-packages";
const DETAIL_URL = `${LIST_URL}/201`;

describe("S07/Q3: 型別403失権は editor を認可不能へ切替。再表示は明示再確認のみ", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  const listOk = () =>
    Promise.resolve({ data: { items: [pkg("201", "セットA")], order_version: "17" } });

  it("list 403 → editor 非表示・保存不能・PATCH なし → 一覧成功だけでは復帰しない", async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA")], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "未保存の入力");
    expect(screen.getByText("保存")).toBeTruthy();

    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();
    expect(screen.queryByLabelText("名称")).toBeNull();
    expect(screen.queryByText("保存")).toBeNull();
    expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();

    // Q3: 一覧成功だけでは再表示しない（明示再確認が必要）。
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL ? Promise.resolve({ data: pkg("201", "セットA") }) : listOk(),
    );
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(screen.queryByLabelText("名称")).toBeNull();
    expect(screen.queryByText("保存")).toBeNull();
    expect(screen.getByText("権限を再確認")).toBeTruthy();
  });

  it("Q3: 再確認で対象 GET が 403 → 非表示維持", async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA")], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();

    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? Promise.reject(apiError(403))
        : Promise.resolve({ data: { items: [pkg("201", "セットA")], order_version: "17" } }),
    );
    await user.click(screen.getByText("権限を再確認"));
    expect(await screen.findByText(/対象パッケージへのアクセス権がありません/)).toBeTruthy();
    expect(screen.queryByLabelText("名称")).toBeNull();
  });

  it("Q3: 再確認で GET 成功だが draft 型（物販）の view なし → 非表示維持", async () => {
    const user = userEvent.setup();
    kindPermission = { "master-merchandise": { canView: false } };
    const withMerch = {
      ...pkg("201", "セットA"),
      components: [
        ...pkg("201", "セットA").components,
        {
          component_key: "a0000000-0000-4000-8000-000000000021",
          item: { kind: "merchandise_item", merchandise_item_id: "301" },
          quantity: "1",
        },
      ],
    };
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [withMerch], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();

    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? Promise.resolve({ data: withMerch })
        : Promise.resolve({ data: { items: [withMerch], order_version: "17" } }),
    );
    await user.click(screen.getByText("権限を再確認"));
    expect(await screen.findByText(/構成に含まれる型の閲覧権限がありません/)).toBeTruthy();
    expect(screen.queryByLabelText("名称")).toBeNull();
  });

  it("A2-02: 再確認のGET待機中に draft 型の view 権限が失われたら非表示維持", async () => {
    const user = userEvent.setup();
    // 保存済み構成には物販がない（fresh GET から消えている）が、draft には物販が残る。
    const savedWithoutMerch = pkg("201", "セットA");
    const listWithMerch = {
      ...savedWithoutMerch,
      components: [
        ...savedWithoutMerch.components,
        {
          component_key: "a0000000-0000-4000-8000-000000000021",
          item: { kind: "merchandise_item", merchandise_item_id: "301" },
          quantity: "1",
        },
      ],
    };
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [listWithMerch], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();

    let resolveDetail!: (v: unknown) => void;
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? new Promise((r) => {
            resolveDetail = r;
          })
        : Promise.resolve({ data: { items: [listWithMerch], order_version: "17" } }),
    );
    await user.click(screen.getByText("権限を再確認"));
    expect(screen.getByText(/権限を確認しています/)).toBeTruthy();

    // GET 待機中に物販 view が失われる（クリック開始時は true だった）。
    kindPermission = { "master-merchandise": { canView: false } };
    await act(async () => {
      resolveDetail({ data: savedWithoutMerch }); // fresh GET 成功（保存済み構成から物販は消えている）
    });

    // 解除判定は GET 応答後の「現在」権限で行われるため非表示維持。
    expect(await screen.findByText(/構成に含まれる型の閲覧権限がありません/)).toBeTruthy();
    expect(screen.queryByLabelText("名称")).toBeNull();
    expect(screen.queryByText("保存")).toBeNull();
  });

  it("Q3: 再確認 all ok → 同一 draft で editor 復帰・PATCH なし", async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA")], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "未保存の入力");
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();

    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL ? Promise.resolve({ data: pkg("201", "セットA") }) : listOk(),
    );
    await user.click(screen.getByText("権限を再確認"));
    const restored = (await screen.findByLabelText("名称")) as HTMLInputElement;
    expect(restored.value).toBe("未保存の入力");
    expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();
  });

  it("Q3: view のみ回復（canEdit false）→ editor 表示・保存 disabled", async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA")], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();

    permission = { canView: true, canCreate: false, canEdit: false, canDelete: false };
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL ? Promise.resolve({ data: pkg("201", "セットA") }) : listOk(),
    );
    await user.click(screen.getByText("権限を再確認"));
    expect(await screen.findByLabelText("名称")).toBeTruthy();
    const save = screen.getByText("保存").closest("button");
    expect((save as HTMLButtonElement | null)?.disabled).toBe(true);
  });

  it("Q3: 対象切替後の遅延再確認応答は破棄される", async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA"), pkg("202", "セットB")], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();

    let resolveDetail!: (v: unknown) => void;
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? new Promise((r) => {
            resolveDetail = r;
          })
        : Promise.resolve({
            data: { items: [pkg("201", "セットA"), pkg("202", "セットB")], order_version: "17" },
          }),
    );
    // 一覧を成功応答へ戻して行を出す（latch は維持される）。
    await act(async () => {
      await client.invalidateQueries();
    });
    await user.click(screen.getByText("権限を再確認"));
    expect(screen.getByText(/権限を確認しています/)).toBeTruthy();

    await user.click(await screen.findByText("セットB")); // 対象切替で token 無効化。
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();
    await act(async () => {
      resolveDetail({ data: pkg("201", "セットA") });
    });
    // 遅延応答は破棄: A の editor は表示されない（latch も維持）。
    expect(screen.queryByLabelText("名称")).toBeNull();
    expect(screen.queryByText("保存")).toBeNull();
  });

  it("B2-02: 保存済み構成から消えた型が draft に残り view なし → 非表示維持", async () => {
    const user = userEvent.setup();
    kindPermission = { "master-merchandise": { canView: false } };
    const savedWithoutMerch = pkg("201", "セットA");
    const listWithMerch = {
      ...savedWithoutMerch,
      components: [
        ...savedWithoutMerch.components,
        {
          component_key: "a0000000-0000-4000-8000-000000000021",
          item: { kind: "merchandise_item", merchandise_item_id: "301" },
          quantity: "1",
        },
      ],
    };
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [listWithMerch], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();

    // fresh GET は成功するが、保存済み構成からは物販が消えている（draft のみに残る）。
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? Promise.resolve({ data: savedWithoutMerch })
        : Promise.resolve({ data: { items: [listWithMerch], order_version: "17" } }),
    );
    await user.click(screen.getByText("権限を再確認"));
    expect(await screen.findByText(/構成に含まれる型の閲覧権限がありません/)).toBeTruthy();
    expect(screen.queryByLabelText("名称")).toBeNull();
  });

  it("B2-02: checkups 型のみの view 喪失でも非表示維持", async () => {
    const user = userEvent.setup();
    kindPermission = { checkups: { canView: false } };
    const withCheckup = {
      ...pkg("201", "セットA"),
      components: [
        ...pkg("201", "セットA").components,
        {
          component_key: "a0000000-0000-4000-8000-000000000022",
          item: { kind: "checkup_type", checkup_type_id: "401" },
          quantity: "1",
        },
      ],
    };
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [withCheckup], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();

    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? Promise.resolve({ data: withCheckup })
        : Promise.resolve({ data: { items: [withCheckup], order_version: "17" } }),
    );
    await user.click(screen.getByText("権限を再確認"));
    expect(await screen.findByText(/構成に含まれる型の閲覧権限がありません/)).toBeTruthy();
    expect(screen.queryByLabelText("名称")).toBeNull();
  });

  it("B2-02: 新規 draft の再確認は GET せず master-medical + draft 各型を確認する", async () => {
    const user = userEvent.setup();
    kindPermission = { "master-merchandise": { canView: false } };
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA")], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("新規登録"));
    // 新規 draft に物販を追加（draft 各型の確認対象にする）。
    await user.click(screen.getByText("構成品を追加"));
    const kindSelects = screen.getAllByLabelText("種別");
    await user.selectOptions(kindSelects[kindSelects.length - 1]!, "merchandise_item");
    const masterInputs = screen.getAllByLabelText(/マスタ ID/);
    await user.type(masterInputs[masterInputs.length - 1]!, "301");

    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();

    // 再確認 1 回目: 新規 draft なので対象 GET は発生しない。物販 view なしで denied。
    await user.click(screen.getByText("権限を再確認"));
    expect(await screen.findByText(/構成に含まれる型の閲覧権限がありません/)).toBeTruthy();
    const detailGets = vi.mocked(axios.get).mock.calls.filter(([u]) => u !== LIST_URL);
    expect(detailGets).toHaveLength(0); // 新規 draft では GET を送らない。
    expect(screen.queryByLabelText("名称")).toBeNull();

    // 再確認 2 回目: 物販 view が回復すれば新規 draft のまま再表示できる。
    kindPermission = {};
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA")], order_version: "17" },
    });
    await user.click(screen.getByText("権限を再確認"));
    expect(await screen.findByLabelText("名称")).toBeTruthy();
    expect(screen.getAllByLabelText("種別")).toHaveLength(2); // draft は保持（診察 + 物販）。
    expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();
  });

  it("refetch 500 は失権扱いせず editor を維持", async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA")], order_version: "17" },
    });
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    vi.mocked(axios.get).mockRejectedValue(apiError(500));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByLabelText("名称")).toBeTruthy();
    expect(screen.getByText("保存")).toBeTruthy();
  });
});

describe("S08: 409 → 最新取得 → 比較 → 明示選択（自動送信なし）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  const latestPkg = { ...pkg("201", "最新名"), version: "5" };
  // A1-03/A1-04: detail 応答は呼出時に生成する型付き factory。Promise.reject を
  // 呼出前に作らない（未処理 rejection 防止）。
  const mockGets = (detail: () => Promise<{ data: unknown }>) =>
    vi
      .mocked(axios.get)
      .mockImplementation((url: string) =>
        url === DETAIL_URL
          ? detail()
          : Promise.resolve({ data: { items: [pkg("201", "セットA")], order_version: "17" } }),
      );

  it("409 → 1回だけ detail GET → 両値を比較表示 → (a) 最新採用で draft 置換", async () => {
    const user = userEvent.setup();
    mockGets(() => Promise.resolve({ data: latestPkg }));
    vi.mocked(axios.patch).mockRejectedValue(apiError(409));
    render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));

    await vi.waitFor(() => {
      expect(vi.mocked(axios.get)).toHaveBeenCalledWith(DETAIL_URL);
    });
    expect(vi.mocked(axios.get).mock.calls.filter(([u]) => u === DETAIL_URL)).toHaveLength(1);
    expect(await screen.findByText("あなたの入力")).toBeTruthy();
    expect(screen.getAllByText("自分の入力").length).toBeGreaterThan(0);
    expect(screen.getAllByText("最新名").length).toBeGreaterThan(0);

    await user.click(screen.getByText("最新を採用（入力を破棄）"));
    expect(((await screen.findByLabelText("名称")) as HTMLInputElement).value).toBe("最新名");
    expect(dirtyMock.markClean).toHaveBeenCalled();
  });

  it("(b) 入力保持で CAS 基準版のみ最新へ → 次の保存は version=最新", async () => {
    const user = userEvent.setup();
    mockGets(() => Promise.resolve({ data: latestPkg }));
    vi.mocked(axios.patch).mockRejectedValue(apiError(409));
    render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));

    await user.click(await screen.findByText("入力を保持して最新版を基準に編集を続ける"));
    expect(((await screen.findByLabelText("名称")) as HTMLInputElement).value).toBe("自分の入力");
    expect(vi.mocked(axios.patch).mock.calls.filter(() => true)).toHaveLength(1); // 選択だけでは送信しない。

    vi.mocked(axios.patch).mockResolvedValue({ data: latestPkg });
    await user.click(screen.getByText("保存"));
    await vi.waitFor(() => {
      expect(vi.mocked(axios.patch)).toHaveBeenCalledWith(
        DETAIL_URL,
        expect.objectContaining({ version: "5" }),
      );
    });
  });

  it("最新取得失敗 → draft 保持 + エラー + 再試行（404 は削除済み表示）", async () => {
    const user = userEvent.setup();
    mockGets(() => Promise.reject(apiError(500)));
    vi.mocked(axios.patch).mockRejectedValue(apiError(409));
    render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));

    expect(await screen.findByText(/最新の取得に失敗しました/)).toBeTruthy();
    expect(((await screen.findByLabelText("名称")) as HTMLInputElement).value).toBe("自分の入力");
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? Promise.reject(apiError(404))
        : Promise.resolve({ data: { items: [pkg("201", "セットA")], order_version: "17" } }),
    );
    await user.click(screen.getByText("再試行"));
    expect(await screen.findByText(/このパッケージは削除されています/)).toBeTruthy();
    expect(((await screen.findByLabelText("名称")) as HTMLInputElement).value).toBe("自分の入力");
  });

  it("閉鎖後の遅延応答は破棄される", async () => {
    const user = userEvent.setup();
    let resolveDetail!: (v: unknown) => void;
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? new Promise((r) => {
            resolveDetail = r;
          })
        : Promise.resolve({ data: { items: [pkg("201", "セットA")], order_version: "17" } }),
    );
    vi.mocked(axios.patch).mockRejectedValue(apiError(409));
    render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));
    await screen.findByText(/最新を取得しています/);

    await user.click(screen.getByText("キャンセル")); // editor を閉じる。
    await act(async () => {
      resolveDetail({ data: latestPkg });
    });
    expect(screen.queryByText("あなたの入力")).toBeNull();
    expect(screen.queryByText("最新を採用（入力を破棄）")).toBeNull();
  });

  it("B2-02: 比較中の背景 list refetch は baseline/target を変えない", async () => {
    const user = userEvent.setup();
    mockGets(() => Promise.resolve({ data: latestPkg }));
    vi.mocked(axios.patch).mockRejectedValue(apiError(409));
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));
    expect(await screen.findByText("あなたの入力")).toBeTruthy();

    // 背景 refetch: 一覧だけが別応答（別版・別名）へ替わる。
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? Promise.resolve({ data: latestPkg })
        : Promise.resolve({
            data: {
              items: [pkg("201", "背景refetch後"), pkg("202", "セットB")],
              order_version: "99",
            },
          }),
    );
    await act(async () => {
      await client.invalidateQueries();
    });

    // 比較は維持（draft も target も不変・送信もしない）。
    expect(screen.getByText("あなたの入力")).toBeTruthy();
    expect(screen.getAllByText("最新名").length).toBeGreaterThan(0);
    expect((screen.getByLabelText("名称") as HTMLInputElement).value).toBe("自分の入力");
    expect(vi.mocked(axios.patch)).toHaveBeenCalledTimes(1);

    // (b) 入力保持 → CAS 基準版は明示選択の snapshot（version 5）。背景 refetch の "99" を使わない。
    await user.click(screen.getByText("入力を保持して最新版を基準に編集を続ける"));
    vi.mocked(axios.patch).mockResolvedValue({ data: latestPkg });
    await user.click(screen.getByText("保存"));
    await vi.waitFor(() => {
      expect(vi.mocked(axios.patch)).toHaveBeenCalledWith(
        DETAIL_URL,
        expect.objectContaining({ version: "5" }),
      );
    });
  });

  it("B2-02: (b) 選択後の再保存で再 409 → draft 保持のまま再比較", async () => {
    const user = userEvent.setup();
    mockGets(() => Promise.resolve({ data: latestPkg }));
    vi.mocked(axios.patch).mockRejectedValue(apiError(409));
    render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));

    await user.click(await screen.findByText("入力を保持して最新版を基準に編集を続ける"));
    expect(((await screen.findByLabelText("名称")) as HTMLInputElement).value).toBe("自分の入力");

    // 再保存 → 再 409 → draft を保持したまま再比較へ（自動送信・上書きなし）。
    await user.click(screen.getByText("保存"));
    expect(await screen.findByText("あなたの入力")).toBeTruthy();
    expect((screen.getByLabelText("名称") as HTMLInputElement).value).toBe("自分の入力");
    expect(vi.mocked(axios.patch)).toHaveBeenCalledTimes(2);
    expect(vi.mocked(axios.get).mock.calls.filter(([u]) => u === DETAIL_URL)).toHaveLength(2);
  });
});

describe("Q1: raw 10,000件は並替を事前抑止する", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  it("10,000件: メッセージ表示・ボタン無効・PATCH なし", async () => {
    const items = Array.from({ length: 10000 }, (_, i) => pkg(String(i + 1), `P${i}`));
    vi.mocked(axios.get).mockResolvedValue({ data: { items, order_version: "17" } });
    render_();
    expect(await screen.findByText(/全件取得を確認できないため並替できません/)).toBeTruthy();
    const ups = screen.getAllByLabelText(/を上へ/);
    expect(ups.length).toBe(10000);
    expect((ups[0] as HTMLButtonElement).disabled).toBe(true);
    expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();
  });
});

describe("A1-01: 保存中の対象切替後の遅延応答は新 session へ侵入しない", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  const twoItems = {
    data: { items: [pkg("201", "セットA"), pkg("202", "セットB")], order_version: "17" },
  };
  const latestPkg = { ...pkg("201", "最新名"), version: "5" };

  const setupPendingSave = async () => {
    const user = userEvent.setup();
    let resolvePatch!: (v: unknown) => void;
    let rejectPatch!: (e: unknown) => void;
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL ? Promise.resolve({ data: latestPkg }) : Promise.resolve(twoItems),
    );
    vi.mocked(axios.patch).mockImplementation(
      () =>
        new Promise((res, rej) => {
          resolvePatch = res;
          rejectPatch = rej;
        }),
    );
    render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));
    // 保存 pending 中にセットBへ切替（編集 session 変更）。
    await user.click(await screen.findByText("セットB"));
    const bName = (await screen.findByLabelText("名称")) as HTMLInputElement;
    expect(bName.value).toBe("セットB");
    return { user, resolvePatch, rejectPatch };
  };

  it("遅延 409 → 比較 UI も対象 jump も起きない", async () => {
    const { rejectPatch } = await setupPendingSave();
    await act(async () => {
      rejectPatch(apiError(409));
    });
    expect(screen.queryByText("あなたの入力")).toBeNull();
    expect(screen.queryByText(/最新を採用/)).toBeNull();
    expect((screen.getByLabelText("名称") as HTMLInputElement).value).toBe("セットB");
    // 新 session で旧対象 A の競合取得（detail GET）は起きない。
    expect(vi.mocked(axios.get).mock.calls.filter(([u]) => u === DETAIL_URL)).toHaveLength(0);
  });

  it("遅延成功 → 対象 jump も draft 置換も起きない", async () => {
    const { resolvePatch } = await setupPendingSave();
    await act(async () => {
      resolvePatch({ data: latestPkg });
    });
    expect((screen.getByLabelText("名称") as HTMLInputElement).value).toBe("セットB");
    expect(screen.queryByText("最新名")).toBeNull();
    expect(dirtyMock.markClean).not.toHaveBeenCalled();
  });
});

describe("B2-02 Q1 fixtures: 9,999件・pending/error/版なしで PATCH しない", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  it("9,999件はボタン有効・メッセージなし", async () => {
    const items = Array.from({ length: 9999 }, (_, i) => pkg(String(i + 1), `P${i}`));
    vi.mocked(axios.get).mockResolvedValue({ data: { items, order_version: "17" } });
    render_();
    expect(await screen.findByText("P0")).toBeTruthy();
    expect(screen.queryByText(/全件取得を確認できないため並替できません/)).toBeNull();
    expect((screen.getAllByLabelText(/を上へ/)[0] as HTMLButtonElement).disabled).toBe(false);
  });

  it("order_version なしは schema 違反で error state・PATCH なし", async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { items: [pkg("201", "セットA")] } });
    render_();
    expect(await screen.findByText(/取得に失敗しました/)).toBeTruthy();
    expect(screen.queryByLabelText(/を上へ/)).toBeNull();
    expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();
  });

  it("pending の間は一覧も操作も出ず PATCH しない", async () => {
    vi.mocked(axios.get).mockReturnValue(new Promise(() => {}));
    render_();
    expect(screen.queryByLabelText(/を上へ/)).toBeNull();
    expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();
  });

  it("error の間は一覧も操作も出ず PATCH しない", async () => {
    vi.mocked(axios.get).mockRejectedValue(new Error("500"));
    render_();
    expect(await screen.findByText(/取得に失敗しました/)).toBeTruthy();
    expect(screen.queryByLabelText(/を上へ/)).toBeNull();
    expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();
  });
});

describe("B3-01: 比較取得待機中の医院切替（clear + reload 相当）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  it("遅延 detail GET は新ツリーへ引き継がれず・自動送信もしない", async () => {
    const user = userEvent.setup();
    let resolveDetail!: (v: unknown) => void;
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? new Promise((r) => {
            resolveDetail = r;
          })
        : Promise.resolve({ data: { items: [pkg("201", "セットA")], order_version: "17" } }),
    );
    vi.mocked(axios.patch).mockRejectedValue(apiError(409));
    const first = render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));
    await screen.findByText(/最新を取得しています/);

    // 医院切替の契約（clinic-selection-recovery.ts）: queryClient.clear() + フル reload。
    // テストでは unmount + 新 QueryClient で代替する（jsdom で location.reload は不可）。
    await act(async () => {
      first.client.clear();
      first.unmount();
    });

    // 新医院の一覧（別 package）。旧 pending GET はまだ未解決。
    vi.mocked(axios.get).mockImplementation(() =>
      Promise.resolve({ data: { items: [pkg("301", "別医院セット")], order_version: "41" } }),
    );
    render_();
    expect(await screen.findByText("別医院セット")).toBeTruthy();
    expect(screen.queryByText("セットA")).toBeNull();

    // 旧医院の遅延 detail GET を後から解決しても新ツリーへ適用されない。
    await act(async () => {
      resolveDetail({ data: { ...pkg("201", "最新名"), version: "5" } });
    });
    await act(async () => {});
    expect(screen.queryByText("あなたの入力")).toBeNull();
    expect(screen.queryByText(/最新を採用/)).toBeNull();
    expect(screen.queryByDisplayValue("自分の入力")).toBeNull();
    // 切替後に PATCH/POST は送られない（旧 draft の自動再送・自動保存なし）。
    expect(vi.mocked(axios.patch)).toHaveBeenCalledTimes(1); // 切替前の 409 PATCH のみ
    expect(vi.mocked(axios.post)).not.toHaveBeenCalled();
  });
});

describe("EMR-532 R1/R2: 再確認の順序・conflict terminal・refreshPermissions fail-closed", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  const latestPkg = { ...pkg("201", "最新名"), version: "5" };
  const listBody = { data: { items: [pkg("201", "セットA")], order_version: "17" } };

  /** 失権（一覧 403）を作る。draft は「未保存の入力」を保持したまま editor は非表示になる。 */
  const openAndLoseAuthorization = async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockResolvedValue(listBody);
    const { client } = render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "未保存の入力");
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    await act(async () => {
      await client.invalidateQueries();
    });
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL
        ? Promise.resolve({ data: pkg("201", "セットA") })
        : Promise.resolve(listBody),
    );
    return { user };
  };

  it("F1: 409 後の detail GET が 403 → 再確認 all ok で editor 復帰し「最新を取得しています…」は残らない", async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL ? Promise.reject(apiError(403)) : Promise.resolve(listBody),
    );
    vi.mocked(axios.patch).mockRejectedValue(apiError(409));
    render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));

    // 競合取得の 403 → 失権表示（editor 非表示）。
    expect(await screen.findByText(/この構成を表示する権限がありません/)).toBeTruthy();
    expect(screen.queryByLabelText("名称")).toBeNull();

    // 全て許可へ戻して明示再確認。
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL ? Promise.resolve({ data: latestPkg }) : Promise.resolve(listBody),
    );
    await user.click(screen.getByText("権限を再確認"));

    const restored = (await screen.findByLabelText("名称")) as HTMLInputElement;
    expect(restored.value).toBe("自分の入力");
    // conflictFetch は failed（terminal）。fetching のまま固着しない。
    expect(screen.queryByText(/最新を取得しています/)).toBeNull();
    expect(screen.getByText(/最新の取得に失敗しました/)).toBeTruthy();
    expect(vi.mocked(axios.patch)).toHaveBeenCalledTimes(1); // 再確認は再送しない。
  });

  it("F2a: refreshPermissions の完了まで editor は復帰しない（latch 解除より先に await）", async () => {
    const { user } = await openAndLoseAuthorization();
    let resolveRefresh!: () => void;
    refreshPermissionsMock.mockImplementation(
      () =>
        new Promise<void>((r) => {
          resolveRefresh = r;
        }),
    );
    await user.click(screen.getByText("権限を再確認"));
    await vi.waitFor(() => expect(refreshPermissionsMock).toHaveBeenCalledTimes(1));

    // 対象 GET は成功済みでも、refreshPermissions 待機中は非表示のまま。
    expect(screen.getByText(/権限を確認しています/)).toBeTruthy();
    expect(screen.queryByLabelText("名称")).toBeNull();

    await act(async () => {
      resolveRefresh();
    });
    expect(((await screen.findByLabelText("名称")) as HTMLInputElement).value).toBe("未保存の入力");
  });

  it("F2a: 解除判定は refreshPermissions 後の現在権限で行う（更新で view 喪失 → 非表示維持）", async () => {
    const { user } = await openAndLoseAuthorization();
    // 更新（/me 再取得）が完了した時点で master-medical の view が失われている。
    refreshPermissionsMock.mockImplementation(async () => {
      permission = { canView: false, canCreate: false, canEdit: false, canDelete: false };
    });
    await user.click(screen.getByText("権限を再確認"));
    expect(await screen.findByText(/構成に含まれる型の閲覧権限がありません/)).toBeTruthy();
    expect(refreshPermissionsMock).toHaveBeenCalledTimes(1);
    expect(screen.queryByLabelText("名称")).toBeNull();
  });

  it("F2b: refreshPermissions が reject → 非表示維持・失敗表示・PATCH なし（古い権限で解除しない）", async () => {
    const { user } = await openAndLoseAuthorization();
    refreshPermissionsMock.mockRejectedValue(new Error("failed to refresh permissions"));
    await user.click(screen.getByText("権限を再確認"));

    expect(
      await screen.findByText("再確認中に取得エラーが発生しました。権限状態は不明です。"),
    ).toBeTruthy();
    expect(screen.queryByLabelText("名称")).toBeNull();
    expect(screen.queryByText("保存")).toBeNull();
    expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();
  });

  it.each([
    ["500", () => apiError(500)],
    ["通信断（response なし）", () => new AxiosError("Network Error")],
  ])(
    "B7-02: 再確認で対象 GET が %s → 失敗表示・editor 非表示維持・PATCH なし",
    async (_label, makeError) => {
      const { user } = await openAndLoseAuthorization();
      vi.mocked(axios.get).mockImplementation((url: string) =>
        url === DETAIL_URL ? Promise.reject(makeError()) : Promise.resolve(listBody),
      );
      await user.click(screen.getByText("権限を再確認"));

      expect(
        await screen.findByText("再確認中に取得エラーが発生しました。権限状態は不明です。"),
      ).toBeTruthy();
      expect(screen.queryByLabelText("名称")).toBeNull();
      expect(screen.queryByText("保存")).toBeNull();
      expect(refreshPermissionsMock).not.toHaveBeenCalled(); // 対象 GET 失敗時は先へ進まない。
      expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();
    },
  );
});

describe("EMR-532 F3: 409 比較後の明示選択は一覧を再取得し、背景データで baseline/draft を置換しない", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  const latestPkg = { ...pkg("201", "最新名"), version: "5" };
  const listGetCount = () => vi.mocked(axios.get).mock.calls.filter(([u]) => u === LIST_URL).length;
  const BACKGROUND_LIST = {
    data: { items: [pkg("201", "背景refetch後"), pkg("202", "セットB")], order_version: "99" },
  };

  /** 比較表示まで進める。以後の一覧 GET は背景データ（別名・別 order_version）を返す。 */
  const reachCompare = async () => {
    const user = userEvent.setup();
    let listResponse: unknown = {
      data: { items: [pkg("201", "セットA")], order_version: "17" },
    };
    vi.mocked(axios.get).mockImplementation((url: string) =>
      url === DETAIL_URL ? Promise.resolve({ data: latestPkg }) : Promise.resolve(listResponse),
    );
    vi.mocked(axios.patch).mockRejectedValue(apiError(409));
    render_();
    await user.click(await screen.findByText("セットA"));
    const nameInput = screen.getByLabelText("名称") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "自分の入力");
    await user.click(screen.getByText("保存"));
    expect(await screen.findByText("あなたの入力")).toBeTruthy();
    listResponse = BACKGROUND_LIST;
    return { user };
  };

  it("(a) 最新を採用 → 一覧を再取得。editor は最新 snapshot（背景 list の名称ではない）", async () => {
    const { user } = await reachCompare();
    const before = listGetCount();
    await user.click(screen.getByText("最新を採用（入力を破棄）"));
    await vi.waitFor(() => expect(listGetCount()).toBeGreaterThan(before));
    await screen.findByText("背景refetch後"); // 一覧は再取得結果へ更新される。
    expect(((await screen.findByLabelText("名称")) as HTMLInputElement).value).toBe("最新名");
    expect(vi.mocked(axios.patch)).toHaveBeenCalledTimes(1); // 自動保存しない。
  });

  it("(b) 入力を保持 → 一覧を再取得。draft は入力のまま・CAS 基準は比較 snapshot の version 5", async () => {
    const { user } = await reachCompare();
    const before = listGetCount();
    await user.click(screen.getByText("入力を保持して最新版を基準に編集を続ける"));
    await vi.waitFor(() => expect(listGetCount()).toBeGreaterThan(before));
    await screen.findByText("背景refetch後");
    expect((screen.getByLabelText("名称") as HTMLInputElement).value).toBe("自分の入力");
    expect(vi.mocked(axios.patch)).toHaveBeenCalledTimes(1); // 選択だけでは送信しない。

    vi.mocked(axios.patch).mockResolvedValue({ data: latestPkg });
    await user.click(screen.getByText("保存"));
    await vi.waitFor(() => {
      expect(vi.mocked(axios.patch)).toHaveBeenLastCalledWith(
        DETAIL_URL,
        expect.objectContaining({ version: "5" }), // 背景 list の order_version "99" や版を使わない。
      );
    });
  });
});

describe("EMR-532 Codex 4354539658: editor なし（list-only）の失権にも再確認の復帰経路がある", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    refreshPermissionsMock.mockReset().mockResolvedValue(undefined);
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    kindPermission = {};
  });

  const twoItems = {
    data: { items: [pkg("201", "セットA"), pkg("202", "セットB")], order_version: "17" },
  };

  // A7-02: target なしの失権は「構成」ではなくマスタ操作の権限確認として案内する。
  const LIST_ONLY_COPY = /パッケージマスタを操作する権限を確認できません/;

  /** editor を開かず並替を試み、PATCH reorder が 403 を返して失権 latch が掛かった状態にする。 */
  const loseAuthorizationByReorder = async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockResolvedValue(twoItems);
    vi.mocked(axios.patch).mockRejectedValue(apiError(403));
    render_();
    await screen.findByText("セットA");
    await user.click(screen.getByLabelText("セットAを下へ"));
    expect(await screen.findByText(LIST_ONLY_COPY)).toBeTruthy();
    return { user };
  };

  const reorderDisabled = () =>
    (screen.getByLabelText("セットAを下へ") as HTMLButtonElement).disabled;

  it("target なしの reorder 403 → 説明と「権限を再確認」を表示し、並替は無効", async () => {
    await loseAuthorizationByReorder();
    expect(screen.getByText("権限を再確認")).toBeTruthy();
    expect(screen.queryByText(/この構成を表示する権限がありません/)).toBeNull();
    expect(reorderDisabled()).toBe(true);
    expect(screen.queryByLabelText("名称")).toBeNull();
  });

  it("再確認 all ok → latch 解除・並替ボタンが再び有効（対象 GET はしない）", async () => {
    const { user } = await loseAuthorizationByReorder();
    vi.mocked(axios.get).mockClear();
    await user.click(screen.getByText("権限を再確認"));

    await vi.waitFor(() => expect(screen.queryByText(LIST_ONLY_COPY)).toBeNull());
    expect(refreshPermissionsMock).toHaveBeenCalledTimes(1);
    expect(reorderDisabled()).toBe(false);
    // 対象がないので detail GET は送らず、一覧のみ再取得する。
    const urls = vi.mocked(axios.get).mock.calls.map(([u]) => u);
    expect(urls).toContain(LIST_URL);
    expect(urls.every((u) => u === LIST_URL)).toBe(true);
    expect(vi.mocked(axios.patch)).toHaveBeenCalledTimes(1); // 自動再送しない。
  });

  it("refreshPermissions が reject → latch 維持・失敗表示・並替は無効のまま", async () => {
    const { user } = await loseAuthorizationByReorder();
    refreshPermissionsMock.mockRejectedValue(new Error("failed to refresh permissions"));
    await user.click(screen.getByText("権限を再確認"));

    expect(
      await screen.findByText("再確認中に取得エラーが発生しました。権限状態は不明です。"),
    ).toBeTruthy();
    expect(screen.getByText(LIST_ONLY_COPY)).toBeTruthy();
    expect(reorderDisabled()).toBe(true);
    expect(vi.mocked(axios.patch)).toHaveBeenCalledTimes(1);
  });

  it("再確認で master-medical の view が回復しない → denied-view で latch 維持", async () => {
    const { user } = await loseAuthorizationByReorder();
    kindPermission = { "master-medical": { canView: false } };
    await user.click(screen.getByText("権限を再確認"));

    expect(await screen.findByText(/パッケージマスタの閲覧権限がありません/)).toBeTruthy();
    expect(screen.getByText("権限を再確認")).toBeTruthy();
  });

  it("B7-02: 初回ロードの一覧 403（editor なし）→ 再確認ボタン表示 → 一覧成功+権限 OK で解除", async () => {
    const user = userEvent.setup();
    vi.mocked(axios.get).mockRejectedValue(apiError(403));
    render_();
    expect(await screen.findByText(LIST_ONLY_COPY)).toBeTruthy();
    expect(screen.queryByText(/この構成を表示する権限がありません/)).toBeNull();
    expect(screen.getByText("権限を再確認")).toBeTruthy();

    vi.mocked(axios.get).mockResolvedValue(twoItems);
    await user.click(screen.getByText("権限を再確認"));

    await vi.waitFor(() => expect(screen.queryByText(LIST_ONLY_COPY)).toBeNull());
    expect(refreshPermissionsMock).toHaveBeenCalledTimes(1);
    await vi.waitFor(() => expect(reorderDisabled()).toBe(false));
    expect(vi.mocked(axios.patch)).not.toHaveBeenCalled();
  });
});
