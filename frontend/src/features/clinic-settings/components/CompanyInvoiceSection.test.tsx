import { describe, it, expect, afterEach, vi } from "vitest";
import { act, render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";
import { server } from "@/testing/mocks/node";
import { CompanyInvoiceSection } from "./CompanyInvoiceSection";

const mockCompany = {
  id: 1,
  name: "ノア動物病院",
  postal_code: "",
  address: "",
  phone_number: "",
  fax_number: "",
  email: "",
  website: "",
  director_name: "",
  registration_number: "",
  invoice_registration_number: "T1234567890123",
  logo_url: "",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

afterEach(() => {
  cleanup();
});

describe("CompanyInvoiceSection", () => {
  it("既存のインボイス登録番号を入力欄に表示する", async () => {
    server.use(http.get("/api/v1/company", () => HttpResponse.json(mockCompany)));

    render(<CompanyInvoiceSection canEdit={true} />, { wrapper: createWrapper() });

    const input = await screen.findByLabelText("インボイス登録番号");
    expect(input).toHaveValue("T1234567890123");
  });

  it("編集して保存すると PATCH /v1/company が新しい値で呼ばれる", async () => {
    server.use(http.get("/api/v1/company", () => HttpResponse.json(mockCompany)));

    let receivedBody: unknown = null;
    server.use(
      http.patch("/api/v1/company", async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json({
          ...mockCompany,
          invoice_registration_number: "T9999999999999",
          updated_at: "2026-01-02T00:00:00Z",
        });
      }),
    );

    const user = userEvent.setup();
    render(<CompanyInvoiceSection canEdit={true} />, { wrapper: createWrapper() });

    const input = await screen.findByLabelText("インボイス登録番号");
    await user.clear(input);
    await user.type(input, "T9999999999999");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(receivedBody).toEqual({ invoice_registration_number: "T9999999999999" });
    });
  });

  it("canEdit=false の場合、入力欄は disabled で保存ボタンは表示されない", async () => {
    server.use(http.get("/api/v1/company", () => HttpResponse.json(mockCompany)));

    render(<CompanyInvoiceSection canEdit={false} />, { wrapper: createWrapper() });

    const input = await screen.findByLabelText("インボイス登録番号");
    expect(input).toBeDisabled();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
  });

  it("canEdit が false に変わった後は formAction が PATCH しない", async () => {
    server.use(http.get("/api/v1/company", () => HttpResponse.json(mockCompany)));

    let patchCount = 0;
    server.use(
      http.patch("/api/v1/company", () => {
        patchCount += 1;
        return HttpResponse.json(mockCompany);
      }),
    );

    const toastError = vi.spyOn(toast, "error").mockImplementation(() => "toast");

    try {
      const { rerender } = render(<CompanyInvoiceSection canEdit={true} />, {
        wrapper: createWrapper(),
      });

      const input = await screen.findByLabelText("インボイス登録番号");
      const form = input.closest("form");
      expect(form).not.toBeNull();

      rerender(<CompanyInvoiceSection canEdit={false} />);

      await waitFor(() => {
        expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
      });

      await act(async () => {
        form?.requestSubmit();
      });

      await waitFor(() => {
        expect(toastError).toHaveBeenCalledWith("この操作を行う権限がありません");
      });
      expect(patchCount).toBe(0);
    } finally {
      toastError.mockRestore();
    }
  });
});
