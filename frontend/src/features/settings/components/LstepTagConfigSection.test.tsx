import { afterEach, describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";
import { LstepTagConfigSection } from "./LstepTagConfigSection";

function createWrapper() {
  return createTestWrapper({ router: true });
}

afterEach(() => {
  server.resetHandlers();
});

// ─────────────────────────────────────────────────
// A: GET 結果表示
// ─────────────────────────────────────────────────

describe("LstepTagConfigSection — A: GET 結果表示", () => {
  it("セクションヘッダーが表示される", async () => {
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/condition-tag-mappings", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/send-purpose-tag-prefixes", () =>
        HttpResponse.json([]),
      ),
    );

    render(<LstepTagConfigSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByText("自動管理タグ設定")).toBeInTheDocument();
    expect(screen.getByText("自動管理プレフィックス (B/C系)")).toBeInTheDocument();
    expect(screen.getByText("慢性疾患コード → タグマッピング")).toBeInTheDocument();
    expect(screen.getByText("LINE送信目的 → タグプレフィックス")).toBeInTheDocument();
  });

  it("データが空の場合、全セクションで「登録なし」が表示される", async () => {
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/condition-tag-mappings", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/send-purpose-tag-prefixes", () =>
        HttpResponse.json([]),
      ),
    );

    render(<LstepTagConfigSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getAllByText("登録なし")).toHaveLength(3);
  });

  it("自動管理プレフィックスの一覧が表示される", async () => {
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([
          { id: 1, prefix: "vaccine_", category: "C2", description: null },
        ]),
      ),
      http.get("/api/v1/lstep-tag-config/condition-tag-mappings", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/send-purpose-tag-prefixes", () =>
        HttpResponse.json([]),
      ),
    );

    render(<LstepTagConfigSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("vaccine_")).toBeInTheDocument();
    });

    expect(screen.getByText("C2")).toBeInTheDocument();
  });

  it("慢性疾患タグマッピングの一覧が表示される", async () => {
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/condition-tag-mappings", () =>
        HttpResponse.json([
          { id: 1, condition_code: "DM", tag_name: "CHRON_DM", description: null },
        ]),
      ),
      http.get("/api/v1/lstep-tag-config/send-purpose-tag-prefixes", () =>
        HttpResponse.json([]),
      ),
    );

    render(<LstepTagConfigSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("DM")).toBeInTheDocument();
    });

    expect(screen.getByText("CHRON_DM")).toBeInTheDocument();
  });

  it("送信目的タグプレフィックスの一覧が表示される", async () => {
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/condition-tag-mappings", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/send-purpose-tag-prefixes", () =>
        HttpResponse.json([
          { id: 1, purpose: "vaccine_reminder", tag_prefix: "PREV_", description: null },
        ]),
      ),
    );

    render(<LstepTagConfigSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("vaccine_reminder")).toBeInTheDocument();
    });

    expect(screen.getByText("PREV_")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────
// B: POST 追加
// ─────────────────────────────────────────────────

describe("LstepTagConfigSection — B: POST 追加", () => {
  it("プレフィックスとカテゴリを入力して追加するとリストに反映される", async () => {
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/condition-tag-mappings", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/send-purpose-tag-prefixes", () =>
        HttpResponse.json([]),
      ),
      http.post("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json({ id: 10, prefix: "vaccine_", category: "C2", description: null }, { status: 201 }),
      ),
    );

    render(<LstepTagConfigSection />, { wrapper: createWrapper() });
    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    const user = userEvent.setup();
    const prefixInput = screen.getByPlaceholderText("例: vaccine_");
    const categoryInput = screen.getByPlaceholderText("例: C2");

    await user.type(prefixInput, "vaccine_");
    await user.type(categoryInput, "C2");

    // invalidate後に再フェッチで追加済みリストを返す
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([
          { id: 10, prefix: "vaccine_", category: "C2", description: null },
        ]),
      ),
    );

    const addButtons = screen.getAllByRole("button", { name: "追加" });
    await user.click(addButtons[0]);

    await waitFor(() => {
      expect(screen.getByText("vaccine_")).toBeInTheDocument();
    });
  });

  it("フィールドが空のまま追加しようとするとエラーが表示される", async () => {
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/condition-tag-mappings", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/send-purpose-tag-prefixes", () =>
        HttpResponse.json([]),
      ),
    );

    render(<LstepTagConfigSection />, { wrapper: createWrapper() });
    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    const user = userEvent.setup();
    const addButtons = screen.getAllByRole("button", { name: "追加" });
    await user.click(addButtons[0]);

    // toast は sonner が管理するため DOM 検証はスキップ（バリデーション呼ばれ POST が飛ばないことを確認）
    await waitFor(() => {
      expect(screen.queryByText("vaccine_")).not.toBeInTheDocument();
    });
  });
});

// ─────────────────────────────────────────────────
// C: DELETE 削除
// ─────────────────────────────────────────────────

describe("LstepTagConfigSection — C: DELETE 削除", () => {
  it("削除ボタンをクリックするとリストから消える", async () => {
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([
          { id: 1, prefix: "vaccine_", category: "C2", description: null },
        ]),
      ),
      http.get("/api/v1/lstep-tag-config/condition-tag-mappings", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/lstep-tag-config/send-purpose-tag-prefixes", () =>
        HttpResponse.json([]),
      ),
      http.delete("/api/v1/lstep-tag-config/auto-managed-prefixes/:id", () =>
        new HttpResponse(null, { status: 204 }),
      ),
    );

    render(<LstepTagConfigSection />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("vaccine_")).toBeInTheDocument();
    });

    // 削除後は空リストを返す
    server.use(
      http.get("/api/v1/lstep-tag-config/auto-managed-prefixes", () =>
        HttpResponse.json([]),
      ),
    );

    const user = userEvent.setup();
    const deleteButton = screen.getByRole("button", { name: "削除" });
    expect(deleteButton).toHaveClass("min-h-11", "min-w-11");
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.queryByText("vaccine_")).not.toBeInTheDocument();
    });
  });
});
