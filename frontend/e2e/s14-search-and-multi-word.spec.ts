import { test, expect } from "@playwright/test";
import type { BrowserContext, Page } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";

// S14 再現スクリプト: 飼主・ペット検索 — 複数語 AND / 語順非依存 / 1語 OR / 0件 / 空白 fail-closed
// 正本: docs/ops/testing/scenarios/S14-search-and-multi-word.md
// 実行: scripts/run-e2e.sh e2e/s14-search-and-multi-word.spec.ts（E2E_LOGIN_* はホスト env から供給）
// 証跡: stdout の "S14-EVIDENCE {...}" 行を reports/uat-YYYY-MM-DD/ へ保存する。
// fixture は run 内で作成し、afterAll で pet → owner の順に削除する。

let context: BrowserContext;
let clinicId = "";
let speciesId: number | null = null;
const createdOwnerIds: number[] = [];
const createdPetIds: number[] = [];
const cleanupFailures: string[] = [];

// seed データに存在しない合成姓を使う（山田等は seed に多数存在し、1語検索で
// ページネーション上限に埋もれて fixture 行を検証できなくなるため）。
const FIXTURES = [
  { owner: "検証一四 太郎", kana: "ケンショウイチヨン タロウ", pet: "ポチ" },
  { owner: "検証一四 次郎", kana: "ケンショウイチヨン ジロウ", pet: "タロウ" },
  { owner: "検証二四 三郎", kana: "ケンショウニヨン サブロウ", pet: "ポチ" },
  { owner: "無関係 花子", kana: "ムカンケイ ハナコ", pet: "ミケ" },
];

function clinicHeaders(): Record<string, string> {
  return clinicId ? { "X-Clinic-ID": clinicId } : {};
}

function evidence(caseName: string, expected: string, actual: string): void {
  // 1 行 JSON で stdout へ。実名検索語は記録しない（S14 確認観点）。
  console.log(`S14-EVIDENCE ${JSON.stringify({ case: caseName, expected, actual })}`);
}

async function jsonOrThrow(page: Page, method: string, path: string, body?: unknown) {
  const res = await page.request.fetch(path, {
    method,
    headers: {
      "Content-Type": "application/json",
      "X-Requested-With": "XMLHttpRequest",
      ...clinicHeaders(),
    },
    data: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!res.ok()) {
    throw new Error(`fixture API failed: ${method} ${path} -> ${res.status()}`);
  }
  if (method === "DELETE") return null;
  return res.json();
}

async function searchBy(page: Page, term: string): Promise<string[]> {
  // 一覧は GET /v1/pets（ペット行粒度）でロードされる。
  const listResponse = (hasSearch: boolean) =>
    page.waitForResponse(
      (res) =>
        res.url().includes("/v1/pets") &&
        (hasSearch ? res.url().includes("search=") : !res.url().includes("search=")),
      { timeout: 15000 },
    );
  if (term === "") {
    // 空文字は FE が search パラメータ自体を消すため、search= を含まない一覧リフェッチを待つ。
    const cleared = listResponse(false);
    await page.goto("/owners");
    await cleared;
  } else {
    await page.goto("/owners");
    const input = page.getByPlaceholder("飼主名、ペット名、電話番号、飼主No、ペット番号...");
    await expect(input).toBeVisible();
    const responsePromise = listResponse(true);
    await input.fill(String(term));
    // EMR-247: Enter / 検索ボタンの確定操作で初めて search リクエストが発行される
    await input.press("Enter");
    await responsePromise;
  }
  await page.waitForTimeout(300); // table re-render settle
  const rows = page.getByRole("row");
  const texts: string[] = [];
  const count = await rows.count();
  for (let i = 0; i < count; i++) {
    const row = rows.nth(i);
    if ((await row.getByRole("columnheader").count()) > 0) continue; // header row
    const text = (await row.innerText()).replace(/\s+/g, " ").trim();
    if (text.includes("データが見つかりません")) continue; // empty-state row は data row に数えない
    texts.push(text);
  }
  return texts;
}

test.beforeAll(async ({ browser }) => {
  context = await createAuthedContext(browser);
  const page = await context.newPage();
  await page.goto("/owners");
  await expect(page.getByRole("heading", { name: "飼主・ペット一覧" })).toBeVisible();
  // UI のアクティブ医院は localStorage(auth_current_clinic) 優先だが、cookie セッション
  // 復元のみの context では未書込のままになり得るため /v1/me の main_clinic_id に揃える。
  const me = (await jsonOrThrow(page, "GET", "/api/v1/me")) as { main_clinic_id?: string };
  clinicId = (me.main_clinic_id ?? "").trim();
  if (!clinicId) throw new Error("main clinic not found in /v1/me");

  const list = (await jsonOrThrow(page, "GET", "/api/v1/masters/animal-species")) as Array<{
    id: number;
    name?: string;
  }>;
  const dog = list.find((s) => s.name === "犬") || list[0];
  if (!dog) throw new Error("animal species master is empty");
  speciesId = dog.id;

  for (const f of FIXTURES) {
    const owner = (await jsonOrThrow(page, "POST", "/api/v1/owners", {
      owner_name: f.owner,
      owner_name_kana: f.kana,
    })) as { id: number };
    createdOwnerIds.push(owner.id);
    const pet = (await jsonOrThrow(page, "POST", "/api/v1/pets", {
      owner_id: owner.id,
      animal_species_id: speciesId,
      name: f.pet,
    })) as { id: number };
    createdPetIds.push(pet.id);
  }
  evidence(
    "fixtures",
    `4 owner+pet created (owner_ids=${createdOwnerIds.join(",")})`,
    `created=${createdOwnerIds.length}`,
  );
  await page.close();
});

test.afterAll(async () => {
  if (!context) return;
  const page = await context.newPage();
  for (const petId of createdPetIds) {
    try {
      await jsonOrThrow(page, "DELETE", `/api/v1/pets/${petId}`);
    } catch (error) {
      cleanupFailures.push(`pet:${petId}`);
      console.log(`S14-CLEANUP-FAIL pet:${petId} ${String(error)}`);
    }
  }
  for (const ownerId of createdOwnerIds) {
    try {
      await jsonOrThrow(page, "DELETE", `/api/v1/owners/${ownerId}`);
    } catch (error) {
      cleanupFailures.push(`owner:${ownerId}`);
      console.log(`S14-CLEANUP-FAIL owner:${ownerId} ${String(error)}`);
    }
  }
  await page.close();
  await context.close();
  evidence(
    "cleanup",
    "4 pets + 4 owners deleted",
    cleanupFailures.length === 0 ? "all deleted" : `failed=${cleanupFailures.join(",")}`,
  );
});

test.describe("S14: 複数語 AND 検索", () => {
  test("手順2: 複数語（姓+ペット名）は AND 絞り込み", async () => {
    const page = await context.newPage();
    const rows = await searchBy(page, "検証一四 ポチ");
    evidence("AND:rowcount", "rows all contain both words", `count=${rows.length}`);
    expect(rows.length).toBeGreaterThanOrEqual(1);
    for (const text of rows) {
      expect(text).toContain("検証一四");
      expect(text).toContain("ポチ");
    }
    await expect(page.getByRole("row", { name: /検証一四 太郎/ })).toBeVisible();
    await expect(page.getByRole("row", { name: /検証一四 次郎/ })).toHaveCount(0);
    await expect(page.getByRole("row", { name: /検証二四 三郎/ })).toHaveCount(0);
    await page.close();
  });

  test("手順3: 語順を逆にしても同じ結果", async () => {
    const page = await context.newPage();
    const rows = await searchBy(page, "ポチ 検証一四");
    evidence("AND:reversed", "same AND semantics regardless of order", `count=${rows.length}`);
    expect(rows.length).toBeGreaterThanOrEqual(1);
    for (const text of rows) {
      expect(text).toContain("検証一四");
      expect(text).toContain("ポチ");
    }
    await expect(page.getByRole("row", { name: /検証一四 次郎/ })).toHaveCount(0);
    await page.close();
  });

  test("手順4: 1語は部分一致 OR で緩い条件", async () => {
    const page = await context.newPage();
    const both = await searchBy(page, "検証一四 ポチ");
    const surname = await searchBy(page, "検証一四");
    evidence(
      "single-word",
      "count(surname) >= count(AND) and both surname fixture rows visible",
      `surname=${surname.length} and=${both.length}`,
    );
    expect(surname.length).toBeGreaterThanOrEqual(2);
    expect(surname.length).toBeGreaterThanOrEqual(both.length);
    await expect(page.getByRole("row", { name: /検証一四 太郎/ })).toBeVisible();
    await expect(page.getByRole("row", { name: /検証一四 次郎/ })).toBeVisible();
    await page.close();
  });

  test("手順5: 存在しない語は空状態・クリアで復帰", async () => {
    const page = await context.newPage();
    const rows = await searchBy(page, "存在しない検索語XYZQ");
    evidence("no-hit", "0 rows with empty state", `count=${rows.length}`);
    expect(rows.length).toBe(0);
    await expect(page.getByText("データが見つかりません")).toBeVisible();
    const cleared = await searchBy(page, "");
    evidence("cleared", "list restored", `count=${cleared.length}`);
    expect(cleared.length).toBeGreaterThan(1);
    await page.close();
  });

  test("手順6: 空白のみは fail-closed で 0 件", async () => {
    const page = await context.newPage();
    const half = await searchBy(page, "   ");
    const full = await searchBy(page, "　　");
    evidence(
      "whitespace-only",
      "0 rows for half/full-width spaces (no all-rows fallback)",
      `half=${half.length} full=${full.length}`,
    );
    expect(half.length).toBe(0);
    expect(full.length).toBe(0);
    await expect(page.getByText("データが見つかりません")).toBeVisible();
    await page.close();
  });

  test("手順1+7: 検索語なし一覧と医院スコープ", async () => {
    const page = await context.newPage();
    // 手順1: 検索語なしで一覧が表示される。一覧は owners.name_kana ASC ソートのため
    // fixture 行が1ページ目に来る保証はない（fixture 存在は手順2-4 で検証済み）。
    const rows = await searchBy(page, "");
    evidence("no-search", "unfiltered list renders data rows", `count=${rows.length}`);
    expect(rows.length).toBeGreaterThan(0);
    // 手順7（他医院候補の非表示）は複数医院 fixture が前提（S23）。本 run では未整備のため
    // BLOCKED 記録とし、PASS に数えない。
    evidence("clinic-scope", "step7 multi-clinic fixture", "BLOCKED (S23 fixture 未整備)");
    await page.close();
  });
});
