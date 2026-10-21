import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import type { CertificateContext } from "../api/certificate-schemas";
import { createMockCertificateApi } from "../mock/certificate-mock-api";
import { CertificateHistory } from "./CertificateHistory";

/** 履歴だけを描画して確認する。契約モックを明示注入しており、実 API・実 BE の検証ではない。 */
function wrapperFor(queryClient: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

function newQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function renderHistory(api: CertificateApi, context: CertificateContext, queryClient = newQueryClient()) {
  const element = (ctx: CertificateContext) => (
    <CertificateHistory api={api} vaccinationId={101} context={ctx} onRefetch={() => {}} getCanEdit={() => true} />
  );
  const view = render(element(context), { wrapper: wrapperFor(queryClient) });
  return { ...view, rerenderWith: (ctx: CertificateContext) => view.rerender(element(ctx)), queryClient };
}

beforeEach(() => {
  URL.createObjectURL = vi.fn(() => "blob:mock");
  URL.revokeObjectURL = vi.fn();
});

describe("履歴の様式選択", () => {
  it("狂犬病の対象でなくなり選べなくなったら、通常様式の履歴へ戻る（切替ボタンが消えても閉じ込めない）", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: true });
    const eligible = await api.getContext(101);
    const view = renderHistory(api, eligible);
    await user.click(await screen.findByRole("button", { name: "狂犬病予防注射済証" }));
    await screen.findByText(/狂犬病予防注射済証の発行履歴はありません/);

    view.rerenderWith({
      ...eligible,
      eligibility: { ...eligible.eligibility, state: "ineligible", allowed_kinds: ["normal"] },
    });
    expect(await screen.findByText(/通常の予防接種証明書の発行履歴はありません/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "狂犬病予防注射済証" })).toBeNull();
    const kinds = controls.calls.filter((c) => c.operation === "listEditions").map((c) => (c.detail as { kind: string }).kind);
    expect(kinds.at(-1)).toBe("normal");
  });
});

describe("履歴の取得失敗", () => {
  it("一時的な失敗では再読み込みでき、成功すると履歴が表示される", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const context = await api.getContext(101);
    controls.failNext("listEditions", { error: new CertificateApiError("network") });
    renderHistory(api, context);
    expect(await screen.findByText("履歴を取得できませんでした。")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "履歴を再読み込み" }));
    expect(await screen.findByText(/発行履歴はありません/)).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("履歴を取得できませんでした。")).toBeNull());
  });

  it("権限なし・接種記録なしでは再読み込みを出さない", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const context = await api.getContext(101);
    controls.failNext("listEditions", { error: new CertificateApiError("forbidden", { httpStatus: 403, code: "not-authorized" }) });
    renderHistory(api, context);
    expect(await screen.findByText("履歴を閲覧する権限がありません。")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "履歴を再読み込み" })).toBeNull();
  });
});
