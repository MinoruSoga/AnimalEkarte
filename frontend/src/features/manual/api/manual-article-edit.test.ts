/**
 * manual-article-edit.test.ts — 既存記事編集 wire adapter の契約テスト（EMR-513）
 *
 * MSW v2 ハンドラで固定API契約（backend/docs/api.yaml 採用SHA 30538608）のwireどおりに
 * fixtureを返す「契約mock」。runtime への黙ったfallbackはしない
 * （未定義経路はMSW onUnhandledRequest:"error" で失敗し、adapterは偽成功を返さない）。
 *
 * 検証範囲: success / 400 / 401 / 403 / 404 / 409(4 code) / 500 / 通信断→operations unknown→succeeded /
 * receipt不一致拒否 / current=null・legacy_unknownで保存停止判定 / number変換しない文字列ID往復。
 */

import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it } from "vitest";

import { server } from "@/testing/mocks/node";

import {
  getManualArticleEditView,
  getManualEditOperationResult,
  resolveManualEditSource,
  saveManualArticleEdit,
  verifyManualEditReceipt,
  ManualArticleSaveError,
  ManualEditConflictError,
  ManualReceiptMismatchError,
  type ManualArticleEditView,
  type ManualEditContentVersion,
  type ManualEditReceipt,
  type UpsertManualArticleRequest,
} from "./manual-article-edit";

// int64最大値。JS Numberへ変換すると 9223372036854775807 → 9223372036854776000 に丸まる値
const ARTICLE_ID = "9223372036854775807";
// Number.MAX_SAFE_INTEGER を超える内容版ID
const CONTENT_VERSION_ID = "4100000000000000123";
const NEW_CONTENT_VERSION_ID = "4100000000000000124";
const PUBLISHED_VERSION_ID = "4100000000000000100";
const CAS_VERSION = 41;
const SHA_DIGEST = "3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca4c6a1c";
const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const EDIT_PATH = "/api/v1/manual/articles/screens/00-overview";

function currentFixture(
  overrides: Partial<ManualEditContentVersion> = {},
): ManualEditContentVersion {
  return {
    id: CONTENT_VERSION_ID,
    article_id: ARTICLE_ID,
    content_revision: 7,
    content_digest: SHA_DIGEST,
    source_kind: "edited",
    restored_from_id: null,
    content: {
      title: "概要",
      order_value: 12.34,
      section: "基本",
      body_markdown: "現在の本文",
    },
    edited_by_staff_id: 12,
    edited_at: "2026-10-19T09:00:00Z",
    ...overrides,
  };
}

function editViewFixture(overrides: Partial<ManualArticleEditView> = {}): ManualArticleEditView {
  return {
    id: ARTICLE_ID,
    category: "screens",
    slug: "00-overview",
    version: CAS_VERSION,
    current: currentFixture(),
    published_version_id: PUBLISHED_VERSION_ID,
    hidden: false,
    ...overrides,
  };
}

function receiptFixture(overrides: Partial<ManualEditReceipt> = {}): ManualEditReceipt {
  return {
    operation_key: "00000000-0000-4000-8000-000000000000",
    article_id: ARTICLE_ID,
    intent: "edit",
    previous_version: CAS_VERSION,
    resulting_version: CAS_VERSION + 1,
    content_version_id: NEW_CONTENT_VERSION_ID,
    content_digest: SHA_DIGEST,
    published_version_id: PUBLISHED_VERSION_ID,
    ...overrides,
  };
}

/** PUTの応答だけを差し替えるハンドラ（要求内容の検証は各テストで個別に実装する） */
function mockPutResponse(status: number, body: unknown): void {
  server.use(http.put(EDIT_PATH, () => HttpResponse.json(body, { status })));
}

const saveParams = {
  category: "screens",
  slug: "00-overview",
  articleId: ARTICLE_ID,
  expectedVersion: CAS_VERSION,
  expectedContentVersionId: CONTENT_VERSION_ID,
  expectedPublishedVersionId: PUBLISHED_VERSION_ID,
  content: {
    title: "概要",
    order_value: 12.34,
    section: "基本",
    body_markdown: "更新後の本文",
  },
} as const;

describe("manual-article-edit adapter（EMR-513 固定API契約）", () => {
  beforeEach(() => {
    server.resetHandlers();
  });

  describe("GET /manual/articles/{category}/{slug}/edit", () => {
    it("編集元を文字列IDのまま取得する（Number変換しない）", async () => {
      server.use(http.get(`${EDIT_PATH}/edit`, () => HttpResponse.json(editViewFixture())));

      const view = await getManualArticleEditView("screens", "00-overview");

      expect(typeof view.id).toBe("string");
      expect(view.id).toBe(ARTICLE_ID);
      expect(view.version).toBe(CAS_VERSION);
      expect(view.current?.id).toBe(CONTENT_VERSION_ID);
      expect(typeof view.current?.id).toBe("string");
      expect(view.current?.article_id).toBe(ARTICLE_ID);
      expect(view.current?.content_digest).toBe(SHA_DIGEST);
      expect(view.published_version_id).toBe(PUBLISHED_VERSION_ID);
      expect(view.hidden).toBe(false);
    });
  });

  describe("PUT /manual/articles/{category}/{slug}", () => {
    it("success: UUIDのoperation_keyと文字列IDの要求を送り、receipt照合を通過した保存結果を返す", async () => {
      const bodies: UpsertManualArticleRequest[] = [];
      const paths: string[] = [];
      server.use(
        http.put("/api/v1/manual/articles/:category/:slug", async ({ request, params }) => {
          const body = (await request.json()) as UpsertManualArticleRequest;
          bodies.push(body);
          paths.push(`${String(params.category)}/${String(params.slug)}`);
          return HttpResponse.json(receiptFixture({ operation_key: body.operation_key }), {
            status: 200,
          });
        }),
      );

      const outcome = await saveManualArticleEdit(saveParams);

      // operation_keyはUUIDで、要求とreceiptが同一keyで拘束される
      expect(outcome.operationKey).toMatch(UUID_RE);
      expect(outcome.receipt.operation_key).toBe(outcome.operationKey);
      expect(outcome.receipt.article_id).toBe(ARTICLE_ID);
      expect(outcome.receipt.previous_version).toBe(CAS_VERSION);
      expect(outcome.receipt.resulting_version).toBe(CAS_VERSION + 1);

      // wire要求: 全field必須・文字列ID・全置換content
      expect(paths).toEqual(["screens/00-overview"]);
      expect(bodies).toHaveLength(1);
      const sent = bodies[0];
      expect(Object.keys(sent).sort()).toEqual([
        "article_id",
        "content",
        "expected_content_version_id",
        "expected_version",
        "operation_key",
      ]);
      expect(sent.operation_key).toBe(outcome.operationKey);
      expect(typeof sent.article_id).toBe("string");
      expect(sent.article_id).toBe(ARTICLE_ID);
      expect(typeof sent.expected_content_version_id).toBe("string");
      expect(sent.expected_content_version_id).toBe(CONTENT_VERSION_ID);
      expect(sent.expected_version).toBe(CAS_VERSION);
      expect(sent.content).toEqual({
        title: "概要",
        order_value: 12.34,
        section: "基本",
        body_markdown: "更新後の本文",
      });
    });

    it("body_markdown空文字の全置換contentをそのまま送る（空文字拒否を追加しない）", async () => {
      const bodies: UpsertManualArticleRequest[] = [];
      server.use(
        http.put(EDIT_PATH, async ({ request }) => {
          const body = (await request.json()) as UpsertManualArticleRequest;
          bodies.push(body);
          return HttpResponse.json(receiptFixture({ operation_key: body.operation_key }), {
            status: 200,
          });
        }),
      );

      await saveManualArticleEdit({
        ...saveParams,
        content: { ...saveParams.content, body_markdown: "" },
      });

      expect(bodies[0].content.body_markdown).toBe("");
    });

    it.each([400, 401, 403, 404, 500] as const)(
      "%i は確定できない応答としてstatusを保持してrejectする（偽成功を作らない）",
      async (status) => {
        // 401のみ共有axios interceptorが自動リフレッシュを挟むため、
        // リフレッシュ失敗(401)も含めてstatusが保存errorへ伝搬することを確認する。
        if (status === 401) {
          server.use(
            http.post("*/v1/auth/refresh", () =>
              HttpResponse.json({ error: "session expired" }, { status: 401 }),
            ),
          );
        }
        mockPutResponse(status, { error: "error from server" });

        const err = await saveManualArticleEdit(saveParams).catch((e: unknown) => e);

        expect(err).toBeInstanceOf(ManualArticleSaveError);
        const saveError = err as ManualArticleSaveError;
        expect(saveError.status).toBe(status);
        expect(saveError.operationKey).toMatch(UUID_RE);
      },
    );

    it.each([
      "version_conflict",
      "operation_key_reused",
      "target_mismatch",
      "source_unresolved",
    ] as const)("409 %s はconflict codeを解析してrejectする", async (code) => {
      mockPutResponse(409, { error: "conflict", code });

      const err = await saveManualArticleEdit(saveParams).catch((e: unknown) => e);

      expect(err).toBeInstanceOf(ManualEditConflictError);
      const conflictError = err as ManualEditConflictError;
      expect(conflictError.conflict.code).toBe(code);
      expect(conflictError.conflict.error).toBe("conflict");
      expect(conflictError.operationKey).toMatch(UUID_RE);
    });

    it("通信断はstatus未定義のManualArticleSaveErrorになり、同一operation_keyで結果照会できる", async () => {
      let operationKeyParam = "";
      server.use(
        http.put(EDIT_PATH, () => HttpResponse.error()),
        http.get(`${EDIT_PATH}/operations/:operationKey`, ({ params }) => {
          operationKeyParam = String(params.operationKey);
          return HttpResponse.json({
            operation_key: operationKeyParam,
            state: "unknown",
            receipt: null,
          });
        }),
      );

      const err = await saveManualArticleEdit(saveParams).catch((e: unknown) => e);

      expect(err).toBeInstanceOf(ManualArticleSaveError);
      const saveError = err as ManualArticleSaveError;
      expect(saveError.status).toBeUndefined();

      // 同一operation_keyで照会 → unknown（入力保持・新key再保存禁止の判定材料）
      const result = await getManualEditOperationResult(
        "screens",
        "00-overview",
        saveError.operationKey,
      );
      expect(operationKeyParam).toBe(saveError.operationKey);
      expect(result.state).toBe("unknown");
      expect(result.receipt).toBeNull();
    });

    it("照会がsucceededを返した場合、receipt照合を通す", async () => {
      const key = "11111111-2222-4333-8444-555555555555";
      server.use(
        http.get(`${EDIT_PATH}/operations/:operationKey`, () =>
          HttpResponse.json({
            operation_key: key,
            state: "succeeded",
            receipt: receiptFixture({ operation_key: key }),
          }),
        ),
      );

      const result = await getManualEditOperationResult("screens", "00-overview", key);

      expect(result.state).toBe("succeeded");
      expect(result.receipt?.operation_key).toBe(key);
      expect(() =>
        verifyManualEditReceipt(result.receipt as ManualEditReceipt, {
          operationKey: key,
          articleId: ARTICLE_ID,
          expectedVersion: CAS_VERSION,
          expectedContentVersionId: CONTENT_VERSION_ID,
          expectedPublishedVersionId: PUBLISHED_VERSION_ID,
        }),
      ).not.toThrow();
    });

    it.each([
      ["article_id mismatch", receiptFixture({ article_id: "999" })],
      [
        "resulting_version is not previous_version + 1",
        receiptFixture({ resulting_version: CAS_VERSION + 2 }),
      ],
      [
        "content_version_id is not a new content version",
        receiptFixture({ content_version_id: CONTENT_VERSION_ID }),
      ],
      ["intent mismatch", receiptFixture({ intent: "publish" as unknown as "edit" })],
    ] as const)("receipt不一致（%s）は拒否する", async (_reason, receipt) => {
      server.use(
        http.put(EDIT_PATH, async ({ request }) => {
          const body = (await request.json()) as UpsertManualArticleRequest;
          return HttpResponse.json(
            { ...receipt, operation_key: body.operation_key },
            { status: 200 },
          );
        }),
      );

      const err = await saveManualArticleEdit(saveParams).catch((e: unknown) => e);

      expect(err).toBeInstanceOf(ManualReceiptMismatchError);
    });

    it("receiptのdigestが64hexでない場合も拒否する（server算出digestを透過保持する）", async () => {
      server.use(
        http.put(EDIT_PATH, async ({ request }) => {
          const body = (await request.json()) as UpsertManualArticleRequest;
          return HttpResponse.json(
            receiptFixture({ operation_key: body.operation_key, content_digest: "abc123" }),
            { status: 200 },
          );
        }),
      );

      const err = await saveManualArticleEdit(saveParams).catch((e: unknown) => e);

      expect(err).toBeInstanceOf(ManualReceiptMismatchError);
    });

    it("receiptのpublished_version_idが編集開始時と異なる場合は拒否する（editは公開版を変えない）", async () => {
      server.use(
        http.put(EDIT_PATH, async ({ request }) => {
          const body = (await request.json()) as UpsertManualArticleRequest;
          return HttpResponse.json(
            receiptFixture({
              operation_key: body.operation_key,
              published_version_id: "4100000000000000999",
            }),
            { status: 200 },
          );
        }),
      );

      const err = await saveManualArticleEdit(saveParams).catch((e: unknown) => e);

      expect(err).toBeInstanceOf(ManualReceiptMismatchError);
    });

    it("edit viewのcurrent.article_idが対象と不一致なら元確認へ戻す（保存要求を送らない）", async () => {
      const brokenView = editViewFixture();
      brokenView.current = { ...brokenView.current, article_id: "999" };
      server.use(
        http.get(`${EDIT_PATH}/edit`, () => HttpResponse.json(brokenView), { once: true }),
      );

      const view = await getManualArticleEditView("screens", "00-overview");
      const resolution = resolveManualEditSource(view);

      expect(resolution.ok).toBe(false);
      if (!resolution.ok) {
        expect(resolution.reason).toBe("invalid_target");
      }
    });

    it("モック未定義の経路では偽成功を作らずrejectする（黙ってfallbackしない）", async () => {
      // PUTハンドラを登録しない（onUnhandledRequest:"error"）
      await expect(saveManualArticleEdit(saveParams)).rejects.toBeInstanceOf(
        ManualArticleSaveError,
      );
    });
  });

  describe("resolveManualEditSource（current=NULL / legacy_unknown / digest・revision欠落の停止判定）", () => {
    it("解決できる編集元はCAS版と内容版を別fieldで返す", () => {
      const resolution = resolveManualEditSource(editViewFixture());

      expect(resolution.ok).toBe(true);
      if (!resolution.ok) return;
      expect(resolution.base.articleId).toBe(ARTICLE_ID);
      expect(resolution.base.casVersion).toBe(CAS_VERSION);
      expect(resolution.base.contentVersionId).toBe(CONTENT_VERSION_ID);
      expect(resolution.base.contentRevision).toBe(7);
      expect(resolution.base.contentDigest).toBe(SHA_DIGEST);
      expect(resolution.base.initialBody).toBe("現在の本文");
    });

    it.each([
      ["current_null", editViewFixture({ current: null })],
      [
        "legacy_unknown",
        editViewFixture({ current: currentFixture({ source_kind: "legacy_unknown" }) }),
      ],
      [
        "content_unidentified(digest欠落)",
        editViewFixture({ current: currentFixture({ content_digest: null }) }),
      ],
      [
        "content_unidentified(revision欠落)",
        editViewFixture({ current: currentFixture({ content_revision: null }) }),
      ],
      ["invalid_target", editViewFixture({ id: "0" })],
    ] as const)("%s は架空補完せず未解決扱いで停止させる", (_reason, view) => {
      const resolution = resolveManualEditSource(view);

      expect(resolution.ok).toBe(false);
    });
  });

  describe("verifyManualEditReceipt", () => {
    const expectation = {
      operationKey: "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee",
      articleId: ARTICLE_ID,
      expectedVersion: CAS_VERSION,
      expectedContentVersionId: CONTENT_VERSION_ID,
      expectedPublishedVersionId: PUBLISHED_VERSION_ID,
    };

    it("対象/版/内容版が一致するreceiptは通過する", () => {
      expect(() =>
        verifyManualEditReceipt(
          receiptFixture({ operation_key: expectation.operationKey }),
          expectation,
        ),
      ).not.toThrow();
    });

    it.each([
      [
        "operation_key mismatch",
        receiptFixture({ operation_key: "ffffffff-0000-4000-8000-000000000000" }),
      ],
      ["article_id mismatch", receiptFixture({ article_id: "1" })],
      ["previous_version mismatch", receiptFixture({ previous_version: CAS_VERSION - 1 })],
    ] as const)("%s は拒否する", (_reason, receipt) => {
      expect(() => verifyManualEditReceipt(receipt, expectation)).toThrow(
        ManualReceiptMismatchError,
      );
    });
  });
});
