import { describe, it, expect } from "vitest";
import { readFileSync } from "fs";
import { resolve } from "path";

/**
 * /accounting 配下 route の権限ガード構造保証テスト
 *
 * 権限分離の設計:
 * - 親 `/accounting`     → ResourceAccounting (会計一覧/詳細/新規のみ)
 * - `/accounting/close`  → ResourceCashRegisterClose (レジ締め・締め履歴)
 * - `/accounting/reports`→ ResourceAccountingReports (月次集計レポート)
 *
 * close/history/reports は /accounting 親ガードから切り出した独立 route。
 * ResourceAccounting しか持たないユーザーが /accounting/close を直接叩いても
 * AccessDenied になること、逆に ResourceCashRegisterClose しか持たないユーザーが
 * /accounting を叩いても AccessDenied になることが保証される。
 *
 * router の構造変更で誤って view ガードが外れると、未認可ユーザーが
 * 直接 URL を叩いて履歴・領収書情報にアクセスできてしまう。
 * ソースコード解析テストでこの構造を固定する。
 */
describe("/accounting route — accounting:view guard", () => {
  const routerSource = readFileSync(
    resolve(__dirname, "../../../../app/router.tsx"),
    "utf-8",
  );

  // /accounting ブロックだけを切り出してパターン検査する
  // （他リソースの RequirePermission 行を誤って拾わないため）
  const accountingBlock = (() => {
    const start = routerSource.indexOf('path: "/accounting"');
    expect(start).toBeGreaterThan(-1);
    // 次のトップレベル route コメント（"// ──"）または末尾までを範囲とする
    const rest = routerSource.slice(start);
    const nextSection = rest.search(/\n\s*\/\/ ── (?!Accounting)/);
    return nextSection > 0 ? rest.slice(0, nextSection) : rest;
  })();

  it("親 /accounting に RequirePermission resource={ResourceAccounting} が掛かっている", () => {
    // path: "/accounting" の直後の element に RequirePermission resource={ResourceAccounting}
    // が記述されていることを正規表現で確認。action 指定なし = デフォルト "view"。
    expect(accountingBlock).toMatch(
      /path:\s*"\/accounting"[\s\S]*?element:\s*\([\s\S]*?<RequirePermission\s+resource=\{ResourceAccounting\}\s*>/,
    );
  });

  it("親ガードに action 上書きがない（= デフォルト view を維持）", () => {
    // 親 element 内で action="create" 等の上書きがないことを確認。
    // 範囲を「path: \"/accounting\"」の直後の element 〜 children の手前に限定。
    const elementMatch = accountingBlock.match(
      /path:\s*"\/accounting"[\s\S]*?element:\s*\(([\s\S]*?)\)\s*,\s*errorElement/,
    );
    expect(elementMatch).not.toBeNull();
    expect(elementMatch![1]).not.toMatch(/action\s*=\s*"(create|edit|delete)"/);
  });

  it("子 select-pet には action=\"create\" の個別ガードが掛かっている", () => {
    expect(accountingBlock).toMatch(
      /path:\s*"select-pet"[\s\S]*?<RequirePermission\s+resource=\{ResourceAccounting\}\s+action="create"/,
    );
  });

  it("子 new には action=\"create\" の個別ガードが掛かっている", () => {
    expect(accountingBlock).toMatch(
      /path:\s*"new"[\s\S]*?<RequirePermission\s+resource=\{ResourceAccounting\}\s+action="create"/,
    );
  });

  it("詳細 :id には個別ガードなし（親 view ガードを継承）", () => {
    // ":id" route 自体の宣言ブロックには RequirePermission がない。
    // ガードは親 /accounting の view が効く。
    const idRouteMatch = accountingBlock.match(
      /path:\s*":id"[\s\S]*?lazy:\s*async/,
    );
    expect(idRouteMatch).not.toBeNull();
    expect(idRouteMatch![0]).not.toMatch(/<RequirePermission/);
  });

  it("一覧 (index: true) には個別ガードなし（親 view ガードを継承）", () => {
    // index 配下の lazy 宣言ブロックに RequirePermission が無いこと。
    // 「親 element の RequirePermission」に該当する行は除外したいので、
    // index 直後の lazy ブロックだけを抽出して見る。
    const indexMatch = accountingBlock.match(
      /index:\s*true,\s*lazy:\s*async\s*\(\)[\s\S]*?\}\s*,/,
    );
    expect(indexMatch).not.toBeNull();
    expect(indexMatch![0]).not.toMatch(/<RequirePermission/);
  });

  it("/accounting 親ブロックに close / close/history / reports が含まれない（権限分離の確認）", () => {
    // close・history・reports は /accounting 親ガードから独立した route のため、
    // accountingBlock の範囲内に現れてはならない。
    expect(accountingBlock).not.toMatch(/path:\s*"close"/);
    expect(accountingBlock).not.toMatch(/path:\s*"reports"/);
    expect(accountingBlock).not.toMatch(/CashRegisterClosePage/);
    expect(accountingBlock).not.toMatch(/CashRegisterHistoryPage/);
    expect(accountingBlock).not.toMatch(/AccountingReportsPage/);
  });
});

describe("/accounting/close route — cash-register-close:view guard", () => {
  const routerSource = readFileSync(
    resolve(__dirname, "../../../../app/router.tsx"),
    "utf-8",
  );

  // /accounting/close ブロックを切り出す（/accounting/close/history は別ブロック）
  const closeBlock = (() => {
    const start = routerSource.indexOf('path: "/accounting/close"');
    expect(start).toBeGreaterThan(-1);
    const rest = routerSource.slice(start);
    // 次の top-level route comment か末尾までを範囲とする
    const nextSection = rest.search(/\n\s*\{[\s\S]*?path:\s*"\/accounting\/close\/history"/);
    return nextSection > 0 ? rest.slice(0, nextSection) : rest.slice(0, 400);
  })();

  it("/accounting/close に RequirePermission resource={ResourceCashRegisterClose} が掛かっている", () => {
    expect(closeBlock).toMatch(
      /path:\s*"\/accounting\/close"[\s\S]*?element:\s*\([\s\S]*?<RequirePermission\s+resource=\{ResourceCashRegisterClose\}\s*>/,
    );
  });

  it("/accounting/close の親ガードに action 上書きがない（= デフォルト view）", () => {
    const elementMatch = closeBlock.match(
      /path:\s*"\/accounting\/close"[\s\S]*?element:\s*\(([\s\S]*?)\)\s*,\s*errorElement/,
    );
    expect(elementMatch).not.toBeNull();
    expect(elementMatch![1]).not.toMatch(/action\s*=\s*"(create|edit|delete)"/);
  });

  it("/accounting/close/history に RequirePermission resource={ResourceCashRegisterClose} が掛かっている", () => {
    // /accounting/close/history のブロックを検索
    const historyStart = routerSource.indexOf('path: "/accounting/close/history"');
    expect(historyStart).toBeGreaterThan(-1);
    const historySnippet = routerSource.slice(historyStart, historyStart + 400);
    expect(historySnippet).toMatch(
      /path:\s*"\/accounting\/close\/history"[\s\S]*?<RequirePermission\s+resource=\{ResourceCashRegisterClose\}/,
    );
  });

  it("/accounting/close/history は ResourceAccounting ガードを持たない（/accounting 親から独立）", () => {
    const historyStart = routerSource.indexOf('path: "/accounting/close/history"');
    const historySnippet = routerSource.slice(historyStart, historyStart + 400);
    expect(historySnippet).not.toMatch(/ResourceAccounting/);
  });
});

describe("/accounting/reports route — accounting-reports:view guard", () => {
  const routerSource = readFileSync(
    resolve(__dirname, "../../../../app/router.tsx"),
    "utf-8",
  );

  const reportsBlock = (() => {
    const start = routerSource.indexOf('path: "/accounting/reports"');
    expect(start).toBeGreaterThan(-1);
    return routerSource.slice(start, start + 400);
  })();

  it("/accounting/reports に RequirePermission resource={ResourceAccountingReports} が掛かっている", () => {
    expect(reportsBlock).toMatch(
      /path:\s*"\/accounting\/reports"[\s\S]*?element:\s*\([\s\S]*?<RequirePermission\s+resource=\{ResourceAccountingReports\}\s*>/,
    );
  });

  it("/accounting/reports の親ガードに action 上書きがない（= デフォルト view）", () => {
    const elementMatch = reportsBlock.match(
      /path:\s*"\/accounting\/reports"[\s\S]*?element:\s*\(([\s\S]*?)\)\s*,\s*errorElement/,
    );
    expect(elementMatch).not.toBeNull();
    expect(elementMatch![1]).not.toMatch(/action\s*=\s*"(create|edit|delete)"/);
  });

  it("/accounting/reports は ResourceAccounting ガードを持たない（/accounting 親から独立）", () => {
    expect(reportsBlock).not.toMatch(/ResourceAccounting[^R]/);
  });
});
