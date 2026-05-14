import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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
    code_type: "checkup_type",
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
      HttpResponse.json(data),
    ),
  );
}

function setupPutHandler(
  onRequest?: (body: unknown) => void,
  status = 200,
) {
  server.use(
    http.put(
      `/api/v1/clinics/${CLINIC_ID}/lstep-tag-code-mappings/:tagName`,
      async ({ request }) => {
        const body = await request.json();
        onRequest?.(body);
        return status === 200
          ? HttpResponse.json([])
          : new HttpResponse(null, { status });
      },
    ),
  );
}

async function renderAndWait(data: TagCodeMappingItem[] = []) {
  setupGetHandler(data);
  render(<LstepTagCodeMappingsSection />, { wrapper: createWrapper() });
  await waitFor(() => {
    expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
  });
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", CLINIC_ID);
});

afterEach(() => {
  localStorage.removeItem("auth_current_clinic:v1");
  server.resetHandlers();
});

// ─────────────────────────────────────────────────
// A: GET 結果表示（スケルトン既存テスト）
// ─────────────────────────────────────────────────

describe("LstepTagCodeMappingsSection — A: GET 結果表示", () => {
  it("読み込み後に骨組みのタグ一覧と未確定バナーが表示される", async () => {
    await renderAndWait([]);

    expect(screen.getByText("タグコードマッピング")).toBeInTheDocument();
    expect(screen.getByText(/SPEC-002 Q5 確定待ち/)).toBeInTheDocument();
    expect(screen.getByText("HLTH_健診あり")).toBeInTheDocument();
    expect(screen.getAllByText("未投入").length).toBeGreaterThan(0);
  });

  it("取得した mapping を tag 単位で表示する", async () => {
    await renderAndWait(mappings);

    expect(screen.getByText("設定済")).toBeInTheDocument();
    expect(screen.getByText("checkup_type")).toBeInTheDocument();
    expect(screen.getByText("CHK_A, CHK_B")).toBeInTheDocument();
  });

  it("API エラー時は失敗メッセージを出す", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep-tag-code-mappings`, () =>
        new HttpResponse(null, { status: 500 }),
      ),
    );

    render(<LstepTagCodeMappingsSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByText("読み込みに失敗しました")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────
// B: 編集フォーム開閉
// ─────────────────────────────────────────────────

describe("LstepTagCodeMappingsSection — B: 編集フォーム開閉", () => {
  it("編集ボタンをクリックすると保存・キャンセルボタンが表示される", async () => {
    await renderAndWait([]);

    const user = userEvent.setup();
    const editButtons = screen.getAllByRole("button", { name: "編集" });
    await user.click(editButtons[0]);

    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "キャンセル" }),
    ).toBeInTheDocument();
  });

  it("キャンセルボタンで編集フォームが閉じる", async () => {
    await renderAndWait([]);

    const user = userEvent.setup();
    await user.click(screen.getAllByRole("button", { name: "編集" })[0]);

    expect(
      screen.getByRole("button", { name: "キャンセル" }),
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "キャンセル" }));

    expect(
      screen.queryByRole("button", { name: "キャンセル" }),
    ).not.toBeInTheDocument();
  });

  it("既存マッピングが編集フォームの初期値として反映される", async () => {
    await renderAndWait(mappings);

    const user = userEvent.setup();
    await user.click(screen.getAllByRole("button", { name: "編集" })[0]);

    const codeTypeInputs = screen.getAllByPlaceholderText("例: checkup_type");
    expect((codeTypeInputs[0] as HTMLInputElement).value).toBe("checkup_type");

    const codesInputs = screen.getAllByPlaceholderText("例: CHK_A, CHK_B");
    expect((codesInputs[0] as HTMLInputElement).value).toBe("CHK_A, CHK_B");
  });
});

// ─────────────────────────────────────────────────
// C: PUT 送信
// ─────────────────────────────────────────────────

describe("LstepTagCodeMappingsSection — C: PUT 送信", () => {
  it("保存ボタンで PUT が正しい body で呼ばれる", async () => {
    let capturedBody: unknown = null;
    setupPutHandler((body) => {
      capturedBody = body;
    });
    await renderAndWait(mappings);

    const user = userEvent.setup();
    await user.click(screen.getAllByRole("button", { name: "編集" })[0]);
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });

    const body = capturedBody as {
      entries: { code_type: string; codes: string[] }[];
    };
    expect(Array.isArray(body.entries)).toBe(true);
    const first = body.entries.find((e) => e.code_type === "checkup_type");
    expect(first?.codes).toEqual(["CHK_A", "CHK_B"]);
  });

  it("コードを編集してから保存すると変更内容が PUT body に反映される", async () => {
    let capturedBody: unknown = null;
    setupPutHandler((body) => {
      capturedBody = body;
    });
    await renderAndWait([]);

    const user = userEvent.setup();
    await user.click(screen.getAllByRole("button", { name: "編集" })[0]);

    const codeTypeInput = screen.getAllByPlaceholderText("例: checkup_type")[0];
    const codesInput = screen.getAllByPlaceholderText("例: CHK_A, CHK_B")[0];

    fireEvent.change(codeTypeInput, { target: { value: "my_type" } });
    fireEvent.change(codesInput, { target: { value: "CODE_1, CODE_2" } });

    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });

    const body = capturedBody as {
      entries: { code_type: string; codes: string[] }[];
    };
    expect(body.entries[0].code_type).toBe("my_type");
    expect(body.entries[0].codes).toEqual(["CODE_1", "CODE_2"]);
  });

  it("保存成功後に編集モードが閉じる", async () => {
    setupPutHandler();
    await renderAndWait([]);

    const user = userEvent.setup();
    await user.click(screen.getAllByRole("button", { name: "編集" })[0]);
    expect(
      screen.getByRole("button", { name: "キャンセル" }),
    ).toBeInTheDocument();

    const codeTypeInput = screen.getAllByPlaceholderText("例: checkup_type")[0];
    fireEvent.change(codeTypeInput, { target: { value: "t" } });

    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(
        screen.queryByRole("button", { name: "キャンセル" }),
      ).not.toBeInTheDocument();
    });
  });
});
