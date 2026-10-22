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
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => permission,
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
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
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

describe("S07: 型別403失権は既存 editor を認可不能へ切替（draft 保持・5xx は対象外）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
  });

  it("list refetch 403 → editor 非表示・保存不能・PATCH なし → 成功 refetch で同一 draft 復帰", async () => {
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

    vi.mocked(axios.get).mockResolvedValue({
      data: { items: [pkg("201", "セットA")], order_version: "17" },
    });
    await act(async () => {
      await client.invalidateQueries();
    });
    const restored = (await screen.findByLabelText("名称")) as HTMLInputElement;
    expect(restored.value).toBe("未保存の入力"); // draft は破棄されず復帰。
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
    permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
  });

  const latestPkg = { ...pkg("201", "最新名"), version: "5" };
  const mockGets = (detail: unknown) =>
    vi
      .mocked(axios.get)
      .mockImplementation((url: string) =>
        url === DETAIL_URL
          ? detail
          : Promise.resolve({ data: { items: [pkg("201", "セットA")], order_version: "17" } }),
      );

  it("409 → 1回だけ detail GET → 両値を比較表示 → (a) 最新採用で draft 置換", async () => {
    const user = userEvent.setup();
    mockGets(Promise.resolve({ data: latestPkg }));
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
    mockGets(Promise.resolve({ data: latestPkg }));
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
    mockGets(Promise.reject(apiError(500)));
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
});
