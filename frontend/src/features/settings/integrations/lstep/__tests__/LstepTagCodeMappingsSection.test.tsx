import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { LstepTagCodeMappingsSection } from "../LstepTagCodeMappingsSection";
import type { TagCodeMappingItem } from "../hooks/useLstepTagCodeMappings";

const CLINIC_ID = "clinic-test-1";

const mappings: TagCodeMappingItem[] = [
  {
    id: 1,
    clinic_id: 1,
    tag_name: "HLTH_健診あり",
    code_type: "checkup_type",
    codes: ["CHK_A", "CHK_B"],
  },
  {
    id: 2,
    clinic_id: 1,
    tag_name: "HLTH_健診あり",
    code_type: "specialty_dental",
    codes: [],
  },
];

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  );
}

function setupGetHandler(data: TagCodeMappingItem[]) {
  server.use(
    http.get(`/api/v1/clinics/${CLINIC_ID}/lstep-tag-code-mappings`, () =>
      HttpResponse.json(data)
    )
  );
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", CLINIC_ID);
});

afterEach(() => {
  localStorage.removeItem("auth_current_clinic:v1");
  server.resetHandlers();
});

describe("LstepTagCodeMappingsSection", () => {
  it("読み込み後に骨組みのタグ一覧と未確定バナーが表示される", async () => {
    setupGetHandler([]);
    render(<LstepTagCodeMappingsSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByText("タグコードマッピング")).toBeInTheDocument();
    expect(screen.getByText(/SPEC-002 Q5 確定待ち/)).toBeInTheDocument();
    expect(screen.getByText("HLTH_健診あり")).toBeInTheDocument();
    expect(screen.getByText("LTV_サプリ購入あり")).toBeInTheDocument();
    expect(screen.getAllByText("未投入").length).toBeGreaterThan(0);
  });

  it("取得した mapping を tag 単位で表示する", async () => {
    setupGetHandler(mappings);
    render(<LstepTagCodeMappingsSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByText("設定済")).toBeInTheDocument();
    expect(screen.getByText("checkup_type")).toBeInTheDocument();
    expect(screen.getByText("CHK_A, CHK_B")).toBeInTheDocument();
  });

  it("API エラー時は失敗メッセージを出す", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep-tag-code-mappings`, () =>
        new HttpResponse(null, { status: 500 })
      )
    );

    render(<LstepTagCodeMappingsSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByText("読み込みに失敗しました")).toBeInTheDocument();
  });
});
