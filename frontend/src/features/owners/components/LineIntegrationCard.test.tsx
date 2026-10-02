import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, waitFor, within, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";
import { server } from "@/testing/mocks/node";
import { AuthContext } from "@/hooks/auth-context";
import { LineIntegrationCard } from "./LineIntegrationCard";
import type { Owner } from "@/types/owner";
import type { Pet } from "@/types";
import { LSTEP_EXCL_DELIVERY_STOP } from "@/constants/lstep-tag-names";

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}));

const CLINIC_ID = "clinic-test-1";
const OWNER_ID = "owner-test-1";
const LINE_USER_ID = "U1234567890abcdef1234567890abcdef1";

const mockAuthContext = {
  user: null,
  currentClinicId: CLINIC_ID,
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: () => true as boolean,
  refreshPermissions: async () => {},
};

const baseOwner: Owner = {
  id: OWNER_ID,
  ownerName: "テスト飼い主",
  company: "",
  postalCode: "",
  address1: "",
  address2: "",
  homePostalCode: "",
  homeAddress1: "",
  homeAddress2: "",
  phone: "",
  companyPhone: "",
  email: "",
  remarks: "",
  isDangerous: false,
  discountRate: 0,
  membershipType: "一般",
  deliveryExcluded: false,
  deliveryCaution: false,
  isTransferred: false,
  lstepOptOut: false,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

// mutation ハンドラーが transformOwner を通せる最低限 OwnerResponse（detail wire）。
// line_user_id / lstep_opt_out は detail DTO に無い — LINE 専用 API fixture 側に置く。
const minimalOwnerApiResponse = {
  id: 1,
  clinic_id: 1,
  owner_name: "テスト飼い主",
  owner_name_kana: "",
  company: "",
  postal_code: "",
  address1: "",
  address2: "",
  home_postal_code: "",
  home_address1: "",
  home_address2: "",
  phone: "",
  company_phone: "",
  email: "",
  remarks: "",
  is_dangerous: false,
  discount_rate: 0,
  membership_type: "member",
  line_id_confirmed_at: "2026-05-01T00:00:00Z",
  delivery_excluded: false,
  delivery_caution: false,
  is_transferred: false,
  pets: [] as const,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-05-01T00:00:00Z",
};

const linkedLineTags = {
  line_user_id: LINE_USER_ID,
  is_linked: true,
  lstep_opt_out: false,
  tags: [] as string[],
  fetched_at: "2026-05-01T00:00:00Z",
};

function setupLineTagsHandler(data: typeof linkedLineTags = linkedLineTags) {
  server.use(
    http.get(`/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/lstep/tags`, () =>
      HttpResponse.json(data),
    ),
  );
}

function createWrapper(authValue: typeof mockAuthContext = mockAuthContext) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <AuthContext.Provider value={authValue}>
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>
  );
}

async function renderAndWait(
  owner: Owner = baseOwner,
  lineTags: typeof linkedLineTags = linkedLineTags,
) {
  setupLineTagsHandler(lineTags);
  render(<LineIntegrationCard ownerId={OWNER_ID} ownerName="テスト飼い主" owner={owner} />, {
    wrapper: createWrapper(),
  });
  await waitFor(() => {
    expect(screen.getByText("LINE / Lステップ連携")).toBeInTheDocument();
  });
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", CLINIC_ID);
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.removeItem("auth_current_clinic:v1");
  server.resetHandlers();
});

// ─────────────────────────────────────────────────────────────
// A: LINE ID 確認セクション（test 7 → A1, A2）
// ─────────────────────────────────────────────────────────────

describe("LineIntegrationCard — A: LINE ID 確認セクション", () => {
  it("lineIdConfirmedAt なし → LINE ID 未確認 + 確認するボタン表示", async () => {
    await renderAndWait({ ...baseOwner, lineUserId: LINE_USER_ID });
    expect(screen.getByText("LINE ID 未確認")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "確認する" })).toBeInTheDocument();
  });

  it("lineIdConfirmedAt あり → LINE ID 確認済み、確認するボタンなし", async () => {
    await renderAndWait({
      ...baseOwner,
      lineUserId: LINE_USER_ID,
      lineIdConfirmedAt: "2026-01-01T00:00:00Z",
    });
    expect(screen.getByText("LINE ID 確認済み")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "確認する" })).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// B: 確認する → line-id-confirm エンドポイント（test 1）
// ─────────────────────────────────────────────────────────────

describe("LineIntegrationCard — B: LINE ID confirm エンドポイント呼び出し", () => {
  it("確認するクリック → PATCH line-id-confirm が body なしで呼ばれる", async () => {
    let called = false;
    server.use(
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/line-id-confirm`,
        async ({ request }) => {
          called = true;
          const bodyText = await request.text();
          expect(bodyText).toBe("");
          return HttpResponse.json(minimalOwnerApiResponse);
        },
      ),
    );
    await renderAndWait({ ...baseOwner, lineUserId: LINE_USER_ID });
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "確認する" }));
    await waitFor(() => expect(called).toBe(true));
  });
});

// ─────────────────────────────────────────────────────────────
// C: 配信停止バナー（test 5）
// ─────────────────────────────────────────────────────────────

describe("LineIntegrationCard — C: 配信停止バナー表示条件", () => {
  it("owner.deliveryExcluded=true → 配信停止中バナー表示", async () => {
    await renderAndWait({ ...baseOwner, deliveryExcluded: true });
    expect(screen.getByText("配信停止中")).toBeInTheDocument();
  });

  it("EXCL_配信停止 タグ付き → 配信停止中バナー表示", async () => {
    await renderAndWait(baseOwner, {
      ...linkedLineTags,
      tags: [LSTEP_EXCL_DELIVERY_STOP],
    });
    expect(screen.getByText("配信停止中")).toBeInTheDocument();
  });

  it("owner.isTransferred=true → 配信停止中バナー表示", async () => {
    await renderAndWait({ ...baseOwner, isTransferred: true });
    expect(screen.getByText("配信停止中")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// D: 配信除外スイッチ（test 3, 6）
// ─────────────────────────────────────────────────────────────

describe("LineIntegrationCard — D: 配信除外スイッチ", () => {
  it("配信制御スイッチは用途を区別できる accessible name を持つ", async () => {
    await renderAndWait();

    expect(screen.getByRole("switch", { name: "配信除外" })).toBeInTheDocument();
    expect(screen.getByRole("switch", { name: "配信注意" })).toBeInTheDocument();
    expect(screen.getByRole("switch", { name: "転院済み" })).toBeInTheDocument();
  });

  it("除外理由入力フィールドは maxLength=100 属性を持つ", async () => {
    await renderAndWait();
    const input = screen.getByPlaceholderText("除外理由（任意・100文字以内）");
    expect(input).toHaveAttribute("maxlength", "100");
  });

  it("配信除外スイッチ OFF → PATCH delivery-exclusion に excluded:false / reason:null が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/delivery-exclusion`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json(minimalOwnerApiResponse);
        },
      ),
    );
    // deliveryExcluded=true でスイッチが checked 状態
    await renderAndWait({ ...baseOwner, deliveryExcluded: true });
    const user = userEvent.setup();
    const row = screen.getByText("配信除外").parentElement;
    expect(row).not.toBeNull();
    const switchEl = within(row!).getByRole("switch");
    await user.click(switchEl);
    await waitFor(() => {
      expect(capturedBody).toEqual(expect.objectContaining({ excluded: false, reason: null }));
    });
  });
});

// ─────────────────────────────────────────────────────────────
// E: 転院確認ダイアログ（test 4, 8）
// ─────────────────────────────────────────────────────────────

describe("LineIntegrationCard — E: 転院確認ダイアログ", () => {
  it("転院済みスイッチ ON → 確認ダイアログに停止コピーが表示される", async () => {
    await renderAndWait();
    const user = userEvent.setup();
    const row = screen.getByText("転院済み").parentElement;
    expect(row).not.toBeNull();
    const switchEl = within(row!).getByRole("switch");
    await user.click(switchEl);
    await waitFor(() => {
      expect(
        screen.getByText(
          "転院フラグを設定すると Lステップへの配信が停止されます。よろしいですか？",
        ),
      ).toBeInTheDocument();
    });
  });

  it("確認ダイアログ確定 → PATCH transfer-status に is_transferred:true が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/transfer-status`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json(minimalOwnerApiResponse);
        },
      ),
    );
    await renderAndWait();
    const user = userEvent.setup();
    const row = screen.getByText("転院済み").parentElement;
    expect(row).not.toBeNull();
    const switchEl = within(row!).getByRole("switch");
    await user.click(switchEl);
    await waitFor(() =>
      screen.getByText("転院フラグを設定すると Lステップへの配信が停止されます。よろしいですか？"),
    );
    await user.click(screen.getByRole("button", { name: "転院済みに設定" }));
    await waitFor(() => {
      expect(capturedBody).toEqual(expect.objectContaining({ is_transferred: true }));
    });
  });
});

// ─────────────────────────────────────────────────────────────
// F: 配信注意バナー + スイッチ（FEAT-382-2 Commit 1）
// ─────────────────────────────────────────────────────────────

describe("LineIntegrationCard — F: 配信注意バナー + スイッチ", () => {
  it("caution=false → 配信注意バナー非表示、配信注意スイッチ OFF", async () => {
    await renderAndWait({ ...baseOwner, deliveryCaution: false });
    expect(screen.queryByTestId("delivery-caution-banner")).not.toBeInTheDocument();
    const row = screen.getByText("配信注意").parentElement;
    expect(row).not.toBeNull();
    const switchEl = within(row!).getByRole("switch");
    expect(switchEl).toHaveAttribute("aria-checked", "false");
  });

  it("caution=true → 黄バナー表示（配信注意テキスト確認）", async () => {
    await renderAndWait({ ...baseOwner, deliveryCaution: true });
    const banner = screen.getByTestId("delivery-caution-banner");
    expect(banner).toBeInTheDocument();
    expect(within(banner).getByText("配信注意")).toBeInTheDocument();
  });

  it("caution=true かつ excluded=true → 赤バナー表示、黄バナー非表示（優先順位）", async () => {
    await renderAndWait({ ...baseOwner, deliveryCaution: true, deliveryExcluded: true });
    expect(screen.queryByTestId("delivery-caution-banner")).not.toBeInTheDocument();
    expect(screen.getByText("配信停止中")).toBeInTheDocument();
  });

  it("配信注意スイッチ ON → PATCH delivery-caution に caution:true が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/delivery-caution`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({ ...minimalOwnerApiResponse, delivery_caution: true });
        },
      ),
    );
    await renderAndWait({ ...baseOwner, deliveryCaution: false });
    const user = userEvent.setup();
    const row = screen.getByText("配信注意").parentElement;
    expect(row).not.toBeNull();
    const switchEl = within(row!).getByRole("switch");
    await user.click(switchEl);
    await waitFor(() => {
      expect(capturedBody).toEqual(expect.objectContaining({ caution: true }));
    });
  });

  it("caution=true 状態で理由入力後「理由を保存」クリック → PATCH delivery-caution に reason が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/delivery-caution`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({
            ...minimalOwnerApiResponse,
            delivery_caution: true,
            delivery_caution_reason: "アレルギー注意",
          });
        },
      ),
    );
    await renderAndWait({ ...baseOwner, deliveryCaution: true });
    const user = userEvent.setup();
    const reasonInput = screen.getByPlaceholderText("注意理由（任意・100文字以内）");
    await user.clear(reasonInput);
    await user.type(reasonInput, "アレルギー注意");
    await user.click(screen.getByTestId("delivery-caution-save-btn"));
    await waitFor(() => {
      expect(capturedBody).toEqual(
        expect.objectContaining({ caution: true, reason: "アレルギー注意" }),
      );
    });
  });
});

// ─────────────────────────────────────────────────────────────
// G: 連携用URL発行（SD-14）
// ─────────────────────────────────────────────────────────────

const unlinkedLineTags = {
  line_user_id: undefined as unknown as string,
  is_linked: false,
  lstep_opt_out: false,
  tags: [] as string[],
  fetched_at: "2026-05-01T00:00:00Z",
};

describe("LineIntegrationCard — G: 連携用URL発行", () => {
  it("未連携時は「連携用URLを発行」ボタンが表示される", async () => {
    await renderAndWait(baseOwner, unlinkedLineTags);
    expect(screen.getByRole("button", { name: "連携用URLを発行" })).toBeInTheDocument();
  });

  it("ボタンクリック → POST link-token が呼ばれ、返却された liff_url が読み取り専用入力欄に表示される", async () => {
    const issuedUrl = "https://liff.line.me/1234567-abcdefgh?token=abc123&clinic_id=1";
    server.use(
      http.post(`/api/v1/owners/${OWNER_ID}/line/link-token`, () =>
        HttpResponse.json(
          { token: "abc123", expires_at: "2026-07-17T00:00:00+09:00", liff_url: issuedUrl },
          { status: 201 },
        ),
      ),
    );
    await renderAndWait(baseOwner, unlinkedLineTags);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "連携用URLを発行" }));

    const urlInput = await screen.findByLabelText("LINE連携用URL");
    expect(urlInput).toHaveValue(issuedUrl);
    expect(screen.getByRole("button", { name: "コピー" })).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// H: LSTEP 配信停止/再開（EMR-251）
// 停止・再開は既存 PATCH /owners/{id}/delivery-exclusion 経路を使い、
// BE 側で delivery_excluded と lstep_opt_out を同時更新する。
// 表示の正本は LINE tags API の lstep_opt_out と owner.deliveryExcluded であり、
// transform で固定 false になる owner.lstepOptOut は正本にしない。
// ─────────────────────────────────────────────────────────────

const optedOutLineTags = {
  ...linkedLineTags,
  lstep_opt_out: true,
};

describe("LineIntegrationCard — H: LSTEP 配信停止/再開（EMR-251）", () => {
  it("永続値 lstep_opt_out=true → 配信停止中バナー + 「配信を再開」ボタン（実値を正本に読む）", async () => {
    await renderAndWait(baseOwner, optedOutLineTags);
    expect(screen.getByText("配信停止中")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "配信を再開" })).toBeInTheDocument();
  });

  it("owner.lstepOptOut=true だけでは停止表示しない（固定フィールドを正本にしない）", async () => {
    await renderAndWait({ ...baseOwner, lstepOptOut: true }, linkedLineTags);
    expect(screen.queryByText("配信停止中")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "配信を停止する" })).toBeInTheDocument();
  });

  it("「配信を再開」→ PATCH delivery-exclusion に excluded:false / reason:null・外部タグ API 呼出 0 件", async () => {
    let capturedBody: unknown = null;
    let capturedUrl = "";
    let tagApiCalls = 0;
    server.use(
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/delivery-exclusion`,
        async ({ request }) => {
          capturedUrl = request.url;
          capturedBody = await request.json();
          return HttpResponse.json(minimalOwnerApiResponse);
        },
      ),
      http.post(`/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/lstep/tags`, () => {
        tagApiCalls += 1;
        return HttpResponse.json({});
      }),
      http.delete(`/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/lstep/tags/:tagName`, () => {
        tagApiCalls += 1;
        return HttpResponse.json({});
      }),
    );
    await renderAndWait(baseOwner, optedOutLineTags);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "配信を再開" }));
    await waitFor(() => {
      expect(capturedBody).toEqual({ excluded: false, reason: null });
    });
    // 医院スコープの endpoint を通る（clinic isolation）
    expect(capturedUrl).toContain(`/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/`);
    // 外部 LSTEP タグ API（付与/解除）は一切呼ばない — 同期は BE 側の責務
    expect(tagApiCalls).toBe(0);
  });

  it("「配信を停止する」→ 確認確定 → PATCH delivery-exclusion に excluded:true（理由空=省略許容）", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/delivery-exclusion`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json(minimalOwnerApiResponse);
        },
      ),
    );
    await renderAndWait();
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "配信を停止する" }));
    await user.click(await screen.findByRole("button", { name: "停止する" }));
    await waitFor(() => {
      expect(capturedBody).toEqual({ excluded: true });
    });
  });

  it("除外理由 100 文字 → PATCH reason がそのまま送られる（上限受理）", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/delivery-exclusion`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({
            ...minimalOwnerApiResponse,
            delivery_excluded: true,
          });
        },
      ),
    );
    await renderAndWait();
    const reason = "あ".repeat(100);
    const reasonInput = screen.getByPlaceholderText("除外理由（任意・100文字以内）");
    fireEvent.change(reasonInput, { target: { value: reason } });
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "配信を停止する" }));
    await user.click(await screen.findByRole("button", { name: "停止する" }));
    await waitFor(() => {
      expect(capturedBody).toEqual({ excluded: true, reason });
    });
  });

  it("PATCH が 400（BE の 100 文字超拒否など）→ エラートースト表示・成功トーストなし", async () => {
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/delivery-exclusion`, () =>
        HttpResponse.json({ error: "理由は100文字以内で入力してください" }, { status: 400 }),
      ),
    );
    await renderAndWait();
    const user = userEvent.setup();
    const row = screen.getByText("配信除外").parentElement;
    expect(row).not.toBeNull();
    const switchEl = within(row!).getByRole("switch");
    await user.click(switchEl);
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled();
    });
    // 保存失敗を成功として表示しない
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("再開後も独立停止（転院済み）が残る → refetch 後も「配信停止中」を維持する", async () => {
    let capturedBody: unknown = null;
    await renderAndWait({ ...baseOwner, isTransferred: true }, optedOutLineTags);
    // PATCH 成功後の refetch では lstep_opt_out=false が返るが is_transferred は残る
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/lstep/tags`, () =>
        HttpResponse.json({ ...linkedLineTags, lstep_opt_out: false }),
      ),
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/owners/${OWNER_ID}/delivery-exclusion`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({
            ...minimalOwnerApiResponse,
            is_transferred: true,
          });
        },
      ),
    );
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "配信を再開" }));
    await waitFor(() => {
      expect(capturedBody).toEqual({ excluded: false, reason: null });
    });
    // refetch 反映（opt_out=false）で再開ボタンは消えるが、転院の独立停止は残るため
    // 「配信停止中」のまま — 配信可能と誤表示しない
    await waitFor(() => {
      expect(screen.queryByRole("button", { name: "配信を再開" })).not.toBeInTheDocument();
    });
    expect(screen.getByText("配信停止中")).toBeInTheDocument();
    expect(screen.getByText("— 転院済み")).toBeInTheDocument();
  });

  it("全ペット死亡の飼主は独立停止として「配信停止中」を表示する", async () => {
    const deadPets = [
      { id: "p1", status: "死亡" } as unknown as Pet,
      { id: "p2", status: "死亡", deceasedAt: "2026-01-01T00:00:00Z" } as unknown as Pet,
    ];
    await renderAndWait({ ...baseOwner, pets: deadPets });
    expect(screen.getByText("配信停止中")).toBeInTheDocument();
    expect(screen.getByText("— 全ペット死亡")).toBeInTheDocument();
  });

  it("membershipType=退亡者 の飼主は独立停止として「配信停止中」を表示する", async () => {
    await renderAndWait({ ...baseOwner, membershipType: "退亡者" });
    expect(screen.getByText("配信停止中")).toBeInTheDocument();
  });

  it("canEdit=false → 「配信を再開」ボタンも配信除外スイッチも表示しない", async () => {
    const readOnlyAuth = { ...mockAuthContext, hasPermission: () => false };
    setupLineTagsHandler(optedOutLineTags);
    render(<LineIntegrationCard ownerId={OWNER_ID} ownerName="テスト飼い主" owner={baseOwner} />, {
      wrapper: createWrapper(readOnlyAuth),
    });
    await waitFor(() => {
      expect(screen.getByText("LINE / Lステップ連携")).toBeInTheDocument();
    });
    expect(screen.getByText("配信停止中")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "配信を再開" })).not.toBeInTheDocument();
    expect(screen.queryByRole("switch", { name: "配信除外" })).not.toBeInTheDocument();
  });
});
