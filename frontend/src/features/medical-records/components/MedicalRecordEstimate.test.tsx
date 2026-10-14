import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createElement, startTransition, Suspense, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import { MedicalRecordEstimate } from "./MedicalRecordEstimate";

const mockCreateMutateAsync = vi.fn();
const mockUpdateMutateAsync = vi.fn();
let mockExisting: {
  id: number;
  title: string;
  comment?: string;
  notes?: string;
  discount_amount?: number;
  items?: unknown[];
} | null = null;

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true, canCreate: true, canDelete: true }),
}));

// FE-RC-048: useClinicTaxRates は独立に useAuth() を呼ぶため最小モックが必要。
// mockClinic は医院マスタの税率設定をテストごとに切り替えるための可変参照。
const { mockClinic } = vi.hoisted(() => ({
  mockClinic: { current: {} as { standardTaxRate?: number; reducedTaxRate?: number } },
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { clinic: mockClinic.current } }),
}));

vi.mock("../api/save-estimate", () => ({
  useGetEstimateByRecord: () => ({ data: mockExisting }),
  useCreateEstimateRecord: () => ({ mutateAsync: mockCreateMutateAsync }),
  useUpdateEstimateRecord: () => ({ mutateAsync: mockUpdateMutateAsync }),
}));

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return createElement(QueryClientProvider, { client: qc }, children);
}

describe("MedicalRecordEstimate BUG-016 resave", () => {
  beforeEach(() => {
    mockExisting = null;
    mockCreateMutateAsync.mockReset();
    mockUpdateMutateAsync.mockReset();
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
    mockClinic.current = {};
  });

  it("初回保存は POST (create)、2回目以降は PATCH (update) を呼び成功トーストは実成功時のみ", async () => {
    const user = userEvent.setup();
    const registered: Array<() => Promise<void>> = [];

    mockCreateMutateAsync.mockResolvedValue({
      id: 42,
      title: "初回件名",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [],
    });
    mockUpdateMutateAsync.mockResolvedValue({
      id: 42,
      title: "更新件名",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [],
    });

    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "99",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    const subject = await screen.findByRole("textbox", { name: /見積書件名/i }).catch(async () => {
      // Label が htmlFor 未接続の場合は placeholder 無しの textbox を拾う
      const inputs = screen.getAllByRole("textbox");
      return inputs[0];
    });

    await user.clear(subject);
    await user.type(subject, "初回件名");

    // 最新の登録ハンドラを使用
    const firstSave = registered[registered.length - 1];
    expect(firstSave).toEqual(expect.any(Function));

    await act(async () => {
      await firstSave();
    });

    expect(mockCreateMutateAsync).toHaveBeenCalledOnce();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(toast.success).toHaveBeenCalledWith("見積書を保存しました");

    // 件名変更 → 2回目は PATCH
    await user.clear(subject);
    await user.type(subject, "更新件名");

    const secondSave = registered[registered.length - 1];
    await act(async () => {
      await secondSave();
    });

    await waitFor(() => expect(mockUpdateMutateAsync).toHaveBeenCalledOnce());
    expect(mockCreateMutateAsync).toHaveBeenCalledOnce(); // 増えない
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: 42,
      payload: expect.objectContaining({
        title: "更新件名",
        medical_record_id: 99,
      }),
    });
    expect(toast.success).toHaveBeenCalledTimes(2);
  });

  it("existing がある状態からの保存は最初から PATCH", async () => {
    mockExisting = {
      id: 7,
      title: "既存",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [],
    };
    mockUpdateMutateAsync.mockResolvedValue({ ...mockExisting, title: "既存更新" });

    const registered: Array<() => Promise<void>> = [];
    const user = userEvent.setup();

    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "5",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    const inputs = await screen.findAllByRole("textbox");
    const subject = inputs[0];
    await user.clear(subject);
    await user.type(subject, "既存更新");

    await act(async () => {
      await registered[registered.length - 1]();
    });

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: 7,
      payload: expect.objectContaining({ title: "既存更新" }),
    });
  });

  it("FE-RC-048: 医院マスタの税率設定 (8%) を消費税計算に反映する", () => {
    mockClinic.current = { standardTaxRate: 0.08 };
    mockExisting = {
      id: 8,
      title: "税率テスト",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [
        {
          id: 1,
          name: "検査A",
          unit_price: 1000,
          quantity: 1,
          discount_rate: 0,
          discount_amount: 0,
          is_insurance_applicable: false,
        },
      ],
    };

    render(createElement(MedicalRecordEstimate, { medicalRecordId: "8" }), { wrapper });

    // 1000円 * 8% = 80円（ハードコード 10% なら 100円になるはず）
    expect(screen.getByText("￥80")).toBeInTheDocument();
    expect(screen.queryByText("￥100")).not.toBeInTheDocument();
  });

  it("EMR-401: 既存明細の tax_type/tax_rate が保存 payload に保持される", async () => {
    mockExisting = {
      id: 9,
      title: "内税見積",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [
        {
          id: 1,
          name: "内税項目",
          unit_price: 1100,
          quantity: 1,
          tax_type: "included",
          tax_rate: 0.1,
          discount_rate: 0,
          discount_amount: 0,
          is_insurance_applicable: false,
        },
      ],
    };
    mockUpdateMutateAsync.mockResolvedValue(mockExisting);

    const registered: Array<() => Promise<void>> = [];
    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "9",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    await act(async () => {
      await registered[registered.length - 1]!();
    });

    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: 9,
      payload: expect.objectContaining({
        items: [
          expect.objectContaining({
            name: "内税項目",
            unit_price: 1100,
            quantity: 1,
            tax_type: "included",
            tax_rate: 0.1,
          }),
        ],
      }),
    });
  });

  it("EMR-401: 内税1100円の明細は消費税100円・合計1100円（外税二重計算しない）", () => {
    mockExisting = {
      id: 10,
      title: "内税",
      items: [
        {
          id: 1,
          name: "内税項目",
          unit_price: 1100,
          quantity: 1,
          tax_type: "included",
          tax_rate: 0.1,
          discount_rate: 0,
          discount_amount: 0,
          is_insurance_applicable: false,
        },
      ],
    };

    render(createElement(MedicalRecordEstimate, { medicalRecordId: "10" }), { wrapper });

    expect(screen.getByText("￥100")).toBeInTheDocument();
    expect(screen.getAllByText("￥1,100").length).toBeGreaterThanOrEqual(2);
    expect(screen.queryByText("￥110")).not.toBeInTheDocument();
    expect(screen.queryByText("￥1,210")).not.toBeInTheDocument();
  });

  it("EMR-401: 外税1000円は消費税100円・非課税1000円は消費税0円", () => {
    mockExisting = {
      id: 11,
      title: "混在",
      items: [
        {
          id: 1,
          name: "外税項目",
          unit_price: 1000,
          quantity: 1,
          tax_type: "excluded",
          tax_rate: 0.1,
          discount_rate: 0,
          discount_amount: 0,
          is_insurance_applicable: false,
        },
        {
          id: 2,
          name: "非課税項目",
          unit_price: 1000,
          quantity: 1,
          tax_type: "exempt",
          tax_rate: 0.1,
          discount_rate: 0,
          discount_amount: 0,
          is_insurance_applicable: false,
        },
      ],
    };

    render(createElement(MedicalRecordEstimate, { medicalRecordId: "11" }), { wrapper });

    expect(screen.getByText("￥100")).toBeInTheDocument();
    expect(screen.getAllByText("￥2,100").length).toBeGreaterThanOrEqual(1);
    expect(screen.queryByText("￥200")).not.toBeInTheDocument();
  });

  it("EMR-401: 新規追加行は内税・医院標準税率で payload に載る", async () => {
    mockClinic.current = { standardTaxRate: 0.1 };
    mockExisting = null;
    mockCreateMutateAsync.mockResolvedValue({
      id: 50,
      title: "",
      items: [],
    });

    const registered: Array<() => Promise<void>> = [];
    const user = userEvent.setup();
    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "50",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    const subject = (await screen.findAllByRole("textbox"))[0]!;
    await user.type(subject, "新規見積");

    await user.click(screen.getByRole("button", { name: "行を追加" }));
    const contentInputs = screen.getAllByPlaceholderText("治療内容を入力");
    await user.type(contentInputs[0]!, "新規項目");

    await act(async () => {
      await registered[registered.length - 1]!();
    });

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        items: [
          expect.objectContaining({
            name: "新規項目",
            tax_type: "included",
            tax_rate: 0.1,
          }),
        ],
      }),
    );
  });

  it("PR665: 取得行の改名保存は existing_item_id を運び、新規クライアント行は運ばない", async () => {
    mockClinic.current = { standardTaxRate: 0.1 };
    mockExisting = {
      id: 9,
      title: "既存見積",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [
        {
          id: 1,
          name: "値引き済み項目",
          unit_price: 1000,
          quantity: 1,
          tax_type: "excluded",
          tax_rate: 0.1,
          discount_rate: 0,
          discount_amount: 500,
          is_insurance_applicable: false,
        },
      ],
    };
    mockUpdateMutateAsync.mockResolvedValue(mockExisting);

    const registered: Array<() => Promise<void>> = [];
    const user = userEvent.setup();
    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "9",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    const contentInput = await screen.findByDisplayValue("値引き済み項目");
    await user.clear(contentInput);
    await user.type(contentInput, "改名後項目");

    await user.click(screen.getByRole("button", { name: "行を追加" }));
    const newContentInput = screen.getAllByPlaceholderText("治療内容を入力").at(-1)!;
    await user.type(newContentInput, "新規クライアント行");

    await act(async () => {
      await registered[registered.length - 1]!();
    });

    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: 9,
      payload: expect.objectContaining({
        items: [
          expect.objectContaining({ name: "改名後項目", existing_item_id: 1 }),
          expect.objectContaining({ name: "新規クライアント行" }),
        ],
      }),
    });
    const call = mockUpdateMutateAsync.mock.calls[0]![0] as {
      payload: { items: Array<Record<string, unknown>> };
    };
    expect("existing_item_id" in call.payload.items[1]!).toBe(false);
  });

  it.each(["新規項目", " 新規項目 "])(
    "PR665: 名前 %s の作成・更新応答から次回の明細IDを維持する",
    async (inputName) => {
      mockClinic.current = { standardTaxRate: 0.1 };
      mockExisting = null;
      mockCreateMutateAsync.mockResolvedValue({
        id: 50,
        title: "",
        discount_amount: 0,
        items: [
          {
            id: 77,
            name: "新規項目",
            unit_price: 1100,
            quantity: 1,
            tax_type: "included",
            tax_rate: 0.1,
            discount_rate: 0,
            discount_amount: 0,
            is_insurance_applicable: false,
            sort_order: 0,
          },
        ],
      });
      mockUpdateMutateAsync.mockResolvedValue({
        id: 50,
        title: "",
        discount_amount: 0,
        items: [{ id: 88, name: "改名後項目", sort_order: 0 }],
      });

      const registered: Array<() => Promise<void>> = [];
      const user = userEvent.setup();
      render(
        createElement(MedicalRecordEstimate, {
          medicalRecordId: "50",
          onRegisterSave: (fn) => {
            registered.push(fn);
          },
        }),
        { wrapper },
      );

      const subject = (await screen.findAllByRole("textbox"))[0]!;
      await user.type(subject, "初回保存");
      await user.click(screen.getByRole("button", { name: "行を追加" }));
      const contentInput = screen.getAllByPlaceholderText("治療内容を入力")[0]!;
      await user.type(contentInput, inputName);

      await act(async () => {
        await registered[registered.length - 1]!();
      });
      expect(mockCreateMutateAsync).toHaveBeenCalledOnce();

      await user.clear(contentInput);
      await user.type(contentInput, " 改名後項目 ");
      await act(async () => {
        await registered[registered.length - 1]!();
      });

      expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
        id: 50,
        payload: expect.objectContaining({
          items: [expect.objectContaining({ name: "改名後項目", existing_item_id: 77 })],
        }),
      });
      await user.clear(contentInput);
      await user.type(contentInput, "更新後項目");
      await act(async () => {
        await registered[registered.length - 1]!();
      });
      expect(mockUpdateMutateAsync).toHaveBeenLastCalledWith({
        id: 50,
        payload: expect.objectContaining({
          items: [expect.objectContaining({ name: "更新後項目", existing_item_id: 88 })],
        }),
      });
    },
  );

  it("破棄された別見積の描画後も、表示中の見積の明細IDで保存できる", async () => {
    const item = {
      id: 71,
      name: "元項目",
      unit_price: 1000,
      quantity: 1,
      tax_type: "excluded",
      tax_rate: 0.1,
      discount_rate: 0,
      discount_amount: 0,
      is_insurance_applicable: false,
      sort_order: 0,
    };
    const estimateA = { id: 7, title: "見積A", discount_amount: 0, items: [item] };
    mockExisting = estimateA;
    mockUpdateMutateAsync.mockResolvedValue(estimateA);
    const registered: Array<() => Promise<void>> = [];
    const neverResolves = new Promise<void>(() => {});
    let suspended = false;
    const probe = vi.fn(() => {
      if (suspended) throw neverResolves;
      return null;
    });
    const onRegisterSave = (fn: () => Promise<void>) => registered.push(fn);
    const ui = () =>
      createElement(
        Suspense,
        { fallback: "読み込み中" },
        createElement(MedicalRecordEstimate, {
          medicalRecordId: "5",
          ownerDiscountRate: suspended ? 1 : 0,
          onRegisterSave,
        }),
        createElement(probe),
      );
    const view = render(ui(), { wrapper });
    await act(async () => {
      mockExisting = { id: 8, title: "見積B", items: [{ ...item, id: 81 }] };
      suspended = true;
      startTransition(() => view.rerender(ui()));
    });
    expect(probe.mock.calls.length).toBeGreaterThan(1);
    await act(async () => {
      mockExisting = estimateA;
      suspended = false;
      view.rerender(ui());
    });
    expect(screen.getByPlaceholderText("治療内容を入力")).toHaveValue("元項目");
    await act(async () => {
      await registered[registered.length - 1]!();
    });
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: 7,
      payload: expect.objectContaining({
        items: [expect.objectContaining({ name: "元項目", existing_item_id: 71 })],
      }),
    });
  });

  it("PR665: 表示と payload は同じ有効行から計算され、請求額は割引後の正味合計", async () => {
    mockClinic.current = { standardTaxRate: 0.1 };
    mockExisting = {
      id: 9,
      title: "割引見積",
      comment: "",
      notes: "",
      discount_amount: 100,
      items: [
        {
          id: 1,
          name: "外税項目",
          unit_price: 1000,
          quantity: 1,
          tax_type: "excluded",
          tax_rate: 0.1,
          discount_rate: 0,
          discount_amount: 0,
          is_insurance_applicable: false,
        },
      ],
    };
    mockUpdateMutateAsync.mockResolvedValue(mockExisting);

    const registered: Array<() => Promise<void>> = [];
    const user = userEvent.setup();
    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "9",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    // 空名行を追加（単価のみ入力、内容は空）→ 合計・payload に寄与しない
    await user.click(screen.getByRole("button", { name: "行を追加" }));
    const lastContentInput = screen.getAllByPlaceholderText("治療内容を入力").at(-1)!;
    let container: HTMLElement | null = lastContentInput.closest("div");
    let unitPriceInput: HTMLInputElement | null = null;
    while (container) {
      unitPriceInput = container.querySelector<HTMLInputElement>('input[type="number"]');
      if (unitPriceInput) break;
      container = container.parentElement;
    }
    if (!unitPriceInput) throw new Error("unit price input not found for the added row");
    await user.type(unitPriceInput, "9999");
    // 内容は空のまま（保存対象外）であることを確認
    expect(
      (screen.getAllByPlaceholderText("治療内容を入力").at(-1) as HTMLInputElement).value,
    ).toBe("");

    // 請求額（正味）= 1100 - 100 = 1000
    expect(screen.getAllByText("￥1,000").length).toBeGreaterThan(0);

    await act(async () => {
      await registered[registered.length - 1]!();
    });
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: 9,
      payload: expect.objectContaining({
        subtotal: 1000,
        tax_total: 100,
        total_amount: 1000,
        discount_amount: 100,
        items: [expect.objectContaining({ name: "外税項目", existing_item_id: 1 })],
      }),
    });
    const call = mockUpdateMutateAsync.mock.calls[0]![0] as { payload: { items: unknown[] } };
    expect(call.payload.items).toHaveLength(1);
  });

  it("PR665: 過大なヘッダー割引は mutation 前に拒否される", async () => {
    mockClinic.current = { standardTaxRate: 0.1 };
    mockExisting = {
      id: 9,
      title: "割引見積",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [
        {
          id: 1,
          name: "外税項目",
          unit_price: 1000,
          quantity: 1,
          tax_type: "excluded",
          tax_rate: 0.1,
          discount_rate: 0,
          discount_amount: 0,
          is_insurance_applicable: false,
        },
      ],
    };
    mockUpdateMutateAsync.mockResolvedValue(mockExisting);

    const registered: Array<() => Promise<void>> = [];
    const user = userEvent.setup();
    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "9",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    // 値引額 NumberInput はラベル関連付けを持たないため、セル構造から解決する
    const discountCell = screen.getByText("値引額").closest("div")!.parentElement!;
    const discountInput = discountCell.querySelector("input")!;
    await user.type(discountInput, "99999");

    await expect(
      act(async () => {
        await registered[registered.length - 1]!();
      }),
    ).rejects.toThrow();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalled();
  });

  it.each([-1, 0.5, Number.NaN, Number.POSITIVE_INFINITY, Number.MAX_SAFE_INTEGER + 1])(
    "不正なヘッダー割引 %s は明細が空でも送信しない",
    async (discountAmount) => {
      mockExisting = { id: 9, title: "見積", discount_amount: discountAmount, items: [] };
      const registered: Array<() => Promise<void>> = [];
      render(
        createElement(MedicalRecordEstimate, {
          medicalRecordId: "9",
          onRegisterSave: (fn) => registered.push(fn),
        }),
        { wrapper },
      );
      await expect(
        act(async () => {
          await registered[registered.length - 1]!();
        }),
      ).rejects.toThrow("invalid discount amount");
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    },
  );

  it("空明細の保存は正味合計0と既存の有効な割引メタデータを維持する", async () => {
    mockExisting = { id: 9, title: "見積", discount_amount: 200, items: [] };
    mockUpdateMutateAsync.mockResolvedValue(mockExisting);
    const registered: Array<() => Promise<void>> = [];
    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "9",
        onRegisterSave: (fn) => registered.push(fn),
      }),
      { wrapper },
    );
    await act(async () => {
      await registered[registered.length - 1]!();
    });
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: 9,
      payload: expect.objectContaining({ total_amount: 0, discount_amount: 200, items: [] }),
    });
  });

  it("API 失敗時は成功トーストを出さない", async () => {
    mockCreateMutateAsync.mockRejectedValue(new Error("network"));
    const registered: Array<() => Promise<void>> = [];
    const user = userEvent.setup();

    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "1",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    const subject = (await screen.findAllByRole("textbox"))[0];
    await user.type(subject, "x");

    await expect(
      act(async () => {
        await registered[registered.length - 1]();
      }),
    ).rejects.toThrow();

    expect(toast.success).not.toHaveBeenCalled();
  });
});
