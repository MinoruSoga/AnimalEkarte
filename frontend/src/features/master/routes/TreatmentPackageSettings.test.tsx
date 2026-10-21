import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
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
vi.mock("@/hooks/use-side-peek-dirty", () => ({
  useSidePeekDirty: () => ({
    markDirty: vi.fn(),
    markClean: vi.fn(),
    runWithDiscardCheck: (cb: () => void) => cb(),
    discardDialog: null,
  }),
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
  return render(<TreatmentPackageSettings />, { wrapper });
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
