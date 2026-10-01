import { test, expect } from "@playwright/test";
import type { BrowserContext, Page } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";

// S16 再現スクリプト: 問診抜粋（治療履歴パネル）→ 過去カルテ詳細への履歴導線
// 正本: docs/ops/testing/scenarios/S16-interview-history-navigation.md
// 実行: scripts/run-e2e.sh e2e/s16-interview-history-navigation.spec.ts（E2E_LOGIN_* はホスト env から供給）
// 証跡: stdout の "S16-EVIDENCE {...}" 行を reports/uat-YYYY-MM-DD/ へ保存する。
// fixture は run 内で作成し、afterAll で medical-records → pet → owner の順に削除する。
// 確定済みカルテは API で削除不可（status=draft 限定・設計どおり）のため、fixture は
// 全て draft で作る（S16 の検証対象は導線/スコープ/空状態で、status に依存しない。
// finalized への遷移成功は 2026-10-01 実行で実証済み）。
//
// コード観測メモ:
// - パネル見出しは「治療履歴」（NO32 でカルテ+トリミング統合）。シナリオ確認観点の
//   「問診抜粋」は旧称であり、見出し文言の一致は別途 doc 突合対象として記録する。
// - MedicalRecordInterview の旧実装は historyItems が空のとき DEFAULT_HISTORY_ITEMS
//   （デモ3行、id=1..3 → /medical-records/{1,2,3} へリンク）を表示していた。
//   実測で製品 FAIL を確認し修正済み（BUG-INTERVIEW-HISTORY-DEMO-ROWS）。
//   履歴0件ペットのテストはその回帰ガードとして残す。

let context: BrowserContext;
let clinicId = "";
const createdOwnerIds: number[] = [];
const createdPetIds: number[] = [];
const createdRecordIds: number[] = [];
const cleanupFailures: string[] = [];

const CC = {
  aCurrent: "S16-A-現在の主訴",
  aPast: "S16-A-過去の主訴",
  bCurrent: "S16-B-現在の主訴",
  bPast: "S16-B-過去の主訴",
  cCurrent: "S16-C-現在の主訴",
} as const;

const IDS = {
  petA: 0,
  petB: 0,
  petC: 0,
  recACurrent: 0,
  recAPast: 0,
  recAEmpty: 0,
  recBCurrent: 0,
  recBPast: 0,
  recCCurrent: 0,
};

function clinicHeaders(): Record<string, string> {
  return clinicId ? { "X-Clinic-ID": clinicId } : {};
}

function evidence(caseName: string, expected: string, actual: string): void {
  console.log(`S16-EVIDENCE ${JSON.stringify({ case: caseName, expected, actual })}`);
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
    const text = await res.text().catch(() => "");
    throw new Error(
      `fixture API failed: ${method} ${path} -> ${res.status()} ${text.slice(0, 300)}`,
    );
  }
  if (method === "DELETE") return null;
  return res.json();
}

async function createOwnerPet(page: Page, ownerName: string, speciesId: number, petName: string) {
  const owner = (await jsonOrThrow(page, "POST", "/api/v1/owners", {
    owner_name: ownerName,
    owner_name_kana: ownerName,
  })) as { id: number };
  createdOwnerIds.push(owner.id);
  const pet = (await jsonOrThrow(page, "POST", "/api/v1/pets", {
    owner_id: owner.id,
    animal_species_id: speciesId,
    name: petName,
  })) as { id: number };
  createdPetIds.push(pet.id);
  return { ownerId: owner.id, petId: pet.id };
}

async function createRecord(
  page: Page,
  ownerId: number,
  petId: number,
  visitDate: string,
  chiefComplaint?: string,
): Promise<number> {
  const record = (await jsonOrThrow(page, "POST", "/api/v1/medical-records", {
    owner_id: String(ownerId),
    pet_id: String(petId),
    visit_date: visitDate,
    status: "draft",
    ...(chiefComplaint !== undefined ? { chief_complaint: chiefComplaint } : {}),
  })) as { id: number };
  createdRecordIds.push(record.id);
  return record.id;
}

/** カルテ編集画面を開き、問診タブの治療履歴パネルが描画されるまで待つ。 */
async function openRecord(page: Page, recordId: number): Promise<void> {
  await page.goto(`/medical-records/${recordId}`);
  await expect(page.getByRole("heading", { name: "カルテ編集" })).toBeVisible({
    timeout: 20000,
  });
  await expect(page.getByRole("heading", { name: "治療履歴" })).toBeVisible();
}

/** 治療履歴パネル内のカルテ詳細リンク href 集合を返す（リンクは a[href^="/medical-records/"]）。 */
async function historyLinks(page: Page): Promise<string[]> {
  // パネルは見出し「治療履歴」を含むカード。リンクは行全体を覆う Link。
  // h3 → [1]タイトル行内div → [2]ヘッダdiv → [3]パネルルート の3段上がパネル全体。
  const panel = page.getByRole("heading", { name: "治療履歴" }).locator("xpath=ancestor::div[3]");
  const links = panel.locator('a[href^="/medical-records/"]');
  const count = await links.count();
  const hrefs: string[] = [];
  for (let i = 0; i < count; i++) {
    const href = await links.nth(i).getAttribute("href");
    if (href) hrefs.push(href.split("?")[0]);
  }
  return hrefs.sort();
}

test.beforeAll(async ({ browser }) => {
  context = await createAuthedContext(browser);
  const page = await context.newPage();
  await page.goto("/owners");
  await expect(page.getByRole("heading", { name: "飼主・ペット一覧" })).toBeVisible({
    timeout: 30000,
  });
  const me = (await jsonOrThrow(page, "GET", "/api/v1/me")) as { main_clinic_id?: string };
  clinicId = (me.main_clinic_id ?? "").trim();
  if (!clinicId) throw new Error("main clinic not found in /v1/me");

  const list = (await jsonOrThrow(page, "GET", "/api/v1/masters/animal-species")) as Array<{
    id: number;
    name?: string;
  }>;
  const dog = list.find((s) => s.name === "犬") || list[0];
  if (!dog) throw new Error("animal species master is empty");

  const a = await createOwnerPet(page, "検証一六 A", dog.id, "S16ポチA");
  IDS.petA = a.petId;
  const b = await createOwnerPet(page, "検証一六 B", dog.id, "S16タロウB");
  IDS.petB = b.petId;
  const c = await createOwnerPet(page, "検証一六 C", dog.id, "S16ミケC");
  IDS.petC = c.petId;

  // pet A: 現在編集中のカルテ + 過去カルテ（主訴あり）+ 明細未移行相当の古い記録（主訴なし・治療なし）
  IDS.recAPast = await createRecord(page, a.ownerId, a.petId, "2026-09-20", CC.aPast);
  IDS.recAEmpty = await createRecord(page, a.ownerId, a.petId, "2026-09-25");
  IDS.recACurrent = await createRecord(page, a.ownerId, a.petId, "2026-10-01", CC.aCurrent);
  // pet B: 別ペットの履歴スコープ検証用
  IDS.recBPast = await createRecord(page, b.ownerId, b.petId, "2026-09-21", CC.bPast);
  IDS.recBCurrent = await createRecord(page, b.ownerId, b.petId, "2026-10-01", CC.bCurrent);
  // pet C: 履歴0件（現在の1件のみ）— デモ行フォールバックの回帰ガード
  IDS.recCCurrent = await createRecord(page, c.ownerId, c.petId, "2026-10-01", CC.cCurrent);

  evidence(
    "fixtures",
    "3 owners+pets, 6 records (A:3, B:2, C:1)",
    `records=${createdRecordIds.join(",")}`,
  );
  await page.close();
});

test.afterAll(async () => {
  if (!context) return;
  const page = await context.newPage();
  for (const id of createdRecordIds) {
    try {
      await jsonOrThrow(page, "DELETE", `/api/v1/medical-records/${id}`);
    } catch (error) {
      cleanupFailures.push(`record:${id}`);
      console.log(`S16-CLEANUP-FAIL record:${id} ${String(error)}`);
    }
  }
  for (const petId of createdPetIds) {
    try {
      await jsonOrThrow(page, "DELETE", `/api/v1/pets/${petId}`);
    } catch (error) {
      cleanupFailures.push(`pet:${petId}`);
      console.log(`S16-CLEANUP-FAIL pet:${petId} ${String(error)}`);
    }
  }
  for (const ownerId of createdOwnerIds) {
    try {
      await jsonOrThrow(page, "DELETE", `/api/v1/owners/${ownerId}`);
    } catch (error) {
      cleanupFailures.push(`owner:${ownerId}`);
      console.log(`S16-CLEANUP-FAIL owner:${ownerId} ${String(error)}`);
    }
  }
  await page.close();
  await context.close();
  evidence(
    "cleanup",
    "6 records + 3 pets + 3 owners deleted",
    cleanupFailures.length === 0 ? "all deleted" : `failed=${cleanupFailures.join(",")}`,
  );
});

test.describe("S16: 問診抜粋（治療履歴）→ カルテ詳細の履歴導線", () => {
  test("手順1: 編集中カルテの履歴パネルに同一ペットの過去記録のみ列挙", async () => {
    const page = await context.newPage();
    await openRecord(page, IDS.recACurrent);
    // 実データ行が読み込まれるまで待つ（ロード中はデモ行が出得るため先に待機）
    await expect(page.locator(`a[href="/medical-records/${IDS.recAPast}"]`)).toBeVisible({
      timeout: 15000,
    });
    const hrefs = await historyLinks(page);
    evidence(
      "history-links",
      `links={/${IDS.recAPast},/${IDS.recAEmpty}} (current ${IDS.recACurrent} excluded)`,
      `links=${hrefs.join(",")}`,
    );
    expect(hrefs).toContain(`/medical-records/${IDS.recAPast}`);
    expect(hrefs).toContain(`/medical-records/${IDS.recAEmpty}`);
    expect(hrefs).not.toContain(`/medical-records/${IDS.recACurrent}`);
    await page.close();
  });

  test("手順2+3: 履歴行クリックで同一ペット過去カルテへ遷移・戻るで文脈維持・dirty 離脱は確認", async () => {
    const page = await context.newPage();
    await openRecord(page, IDS.recACurrent);
    const pastLink = page.locator(`a[href="/medical-records/${IDS.recAPast}"]`);
    await expect(pastLink).toBeVisible({ timeout: 15000 });

    // 手順3 後半: 未保存入力があるとき履歴遷移は NavigationBlocker の確認を経由する
    await page.locator("#medical-record-chief-complaint").fill("S16 dirty unsaved input");
    await pastLink.click();
    const blockerDialog = page.getByRole("alertdialog");
    await expect(blockerDialog).toBeVisible({ timeout: 5000 });
    await expect(blockerDialog).toContainText("変更が保存されていません");
    await blockerDialog.getByRole("button", { name: "このページに留まる" }).click();
    await expect(page).toHaveURL(new RegExp(`/medical-records/${IDS.recACurrent}$`));
    evidence(
      "navigation-blocker",
      "dirty state blocks history navigation with confirm dialog",
      "dialog shown, stayed on current record",
    );

    // クリーンな状態へ戻してから手順2の遷移を実行
    await page.reload();
    await openRecord(page, IDS.recACurrent);
    const pastLinkClean = page.locator(`a[href="/medical-records/${IDS.recAPast}"]`);
    await expect(pastLinkClean).toBeVisible({ timeout: 15000 });
    await pastLinkClean.click();
    await expect(page).toHaveURL(new RegExp(`/medical-records/${IDS.recAPast}$`), {
      timeout: 15000,
    });
    await expect(page.getByRole("heading", { name: "カルテ編集" })).toBeVisible();
    await expect(page.getByText("S16ポチA").first()).toBeVisible();
    evidence(
      "history-navigate",
      `row click → /medical-records/${IDS.recAPast} (same pet record)`,
      `url=${page.url()}`,
    );

    // 手順3 前半: 戻る操作で元の編集画面・患者ヘッダーが維持される
    await page.goBack();
    await expect(page).toHaveURL(new RegExp(`/medical-records/${IDS.recACurrent}$`), {
      timeout: 15000,
    });
    await expect(page.getByRole("heading", { name: "カルテ編集" })).toBeVisible();
    await expect(page.getByText("S16ポチA").first()).toBeVisible();
    await expect(page.getByRole("tab", { name: "問診" })).toHaveAttribute("aria-selected", "true");
    evidence(
      "history-back",
      `back → /medical-records/${IDS.recACurrent} with 問診 tab + pet header`,
      `url=${page.url()}`,
    );
    await page.close();
  });

  test("手順4: 治療明細が空の過去記録でも詳細が開ける", async () => {
    const page = await context.newPage();
    await openRecord(page, IDS.recACurrent);
    const emptyLink = page.locator(`a[href="/medical-records/${IDS.recAEmpty}"]`);
    await expect(emptyLink).toBeVisible({ timeout: 15000 });
    await emptyLink.click();
    await expect(page).toHaveURL(new RegExp(`/medical-records/${IDS.recAEmpty}$`), {
      timeout: 15000,
    });
    await expect(page.getByRole("heading", { name: "カルテ編集" })).toBeVisible();
    // 治療タブを開き、空明細が空状態として表示されること（真っ白/エラーにならない）
    await page.getByRole("tab", { name: "治療", exact: true }).click();
    await expect(page.getByText("治療明細がありません")).toBeVisible({ timeout: 10000 });
    evidence(
      "empty-record-detail",
      `record ${IDS.recAEmpty} (no treatments) opens; 治療 tab shows empty state`,
      "opened, empty-state text visible",
    );
    await page.close();
  });

  test("手順5: 別ペットの履歴はそのペット自身の記録のみ", async () => {
    const page = await context.newPage();
    await openRecord(page, IDS.recBCurrent);
    await expect(page.locator(`a[href="/medical-records/${IDS.recBPast}"]`)).toBeVisible({
      timeout: 15000,
    });
    const hrefs = await historyLinks(page);
    evidence(
      "pet-b-history-scope",
      `links={/${IDS.recBPast}} only (pet A records must not appear)`,
      `links=${hrefs.join(",")}`,
    );
    expect(hrefs).toContain(`/medical-records/${IDS.recBPast}`);
    expect(hrefs).not.toContain(`/medical-records/${IDS.recACurrent}`);
    expect(hrefs).not.toContain(`/medical-records/${IDS.recAPast}`);
    expect(hrefs).not.toContain(`/medical-records/${IDS.recAEmpty}`);
    await page.close();
  });

  test("履歴0件ペット: デモ行（DEFAULT_HISTORY_ITEMS）が履歴として表示されない", async () => {
    const page = await context.newPage();
    await openRecord(page, IDS.recCCurrent);
    // 履歴クエリ完了を待つため、ロード中表示の解消をポーリングで待機
    await page.waitForTimeout(2500);
    const hrefs = await historyLinks(page);
    const demoVisible = await page.getByText("消化器症状").count();
    evidence(
      "empty-history-behavior",
      "履歴0件 → 空状態のみ（デモ行なし）",
      `links=${hrefs.join(",")} demoText=${demoVisible}`,
    );
    // デモ行（id=1..3 → /medical-records/{1,2,3} へのリンク）は他記録へ飛び得るため
    // 出れば製品 FAIL。期待は空状態のみ。
    expect(hrefs).toEqual([]);
    expect(demoVisible).toBe(0);
    await page.close();
  });
});
