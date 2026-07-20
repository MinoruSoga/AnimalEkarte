# FE-refactor 第9期（FE9）— design-system.md（DESIGN.md 製品翻訳）への完全遵守

> 起票: 2026-07-21（要件責任者: 曽我。「DESIGN.md に完全遵守させたい」指示を受け、翻訳→機械化→ページ掃討の3フェーズで起票）
> **本ファイルは対応後削除する使い捨て計画書**。恒久規約は `docs/spec/design-system.md` / audit スクリプトへ同梱してから削除する。
> **遵守の正本 = `docs/spec/design-system.md`**（DESIGN.md の製品翻訳。SSOT 優先順位: design-tokens.ts ＞ design-system.md ＞ DESIGN.md）。DESIGN.md への字義遵守ではない — primary は #038B94・semantic 色は臨床安全のため維持・pill CTA/display 型はマーケ専用（同書 §3.4 / §7.2 裁定・2026-07-21 確定済み）。
> 読者 = 着手するエージェント（Sonnet 想定）。実行順は FE9-1 → 9-2 → 9-3（逆行禁止 — 機械化前にページ監査すると偽違反を量産する）。

---

## 0. 要約

| ID | 内容 | 規模 | 状態 |
|----|------|------|------|
| FE9-0 | design-system.md 未翻訳章の追補（§3.4 タイポ採用範囲・§5.1 shadow 実装・§7.2 pill 裁定） | — | ✅ 起票と同時完了（2026-07-21） |
| FE9-1 | shadow Level トークン実装（--shadow-level1/2） | 小 | 未着手 |
| FE9-2 | 機械置換＋audit C10/C11 新設（shadow 70件・font-size 任意値 66件） | 中 | 未着手 |
| FE9-3 | ページ毎スイープ（84 リーフルート・機械判定不能項目のみ） | 大 | 未着手・**前提条件あり** |

## 1. 実測スナップショット（2026-07-21・src+liff/src+line-reserve/src 非test）

```bash
# 再計測（着手時に必ず実行）
cd frontend
for p in shadow-sm shadow-md shadow-lg "shadow-\["; do echo -n "$p: "; grep -rE "[\"' ]${p}" src liff/src line-reserve/src --include="*.tsx" --include="*.ts" | grep -v test | wc -l; done
grep -rEn "text-\[[0-9.]+(px|rem)\]" src liff/src line-reserve/src --include="*.tsx" --include="*.ts" | grep -v test | wc -l
```

- shadow: `shadow-sm`×55 / `shadow-md`×4 / `shadow-lg`×5 / `shadow-[...]`任意値×6 — **Tailwind 既定影が無統制**。DESIGN.md Level 0/1/2 体系（design-system.md §5）は仕様のみで実装トークンなし。`drop-shadow`・CSS `box-shadow` は 0
- font-size 任意値 `text-[Npx|rem]`: **66件** — §3.4 のロール表（title/section/body-md/body-sm/caption/eyebrow）へ正規化対象
- text クラス分布: xs×177 / sm×485 / base×150 / lg×10 / xl×6 / 2xl×2 — 高密度業務 UI の実態は body-sm 中心。display/h1 系は不在（=§3.4 裁定と整合）
- font-weight: bold×110 / semibold×93 / medium×349 — §3.4 の「700 は見出し専用」に対し bold 110件は本文・数値セル混入の疑いあり（FE9-3 で判定）
- rounded・maxWidth・色系は FE8-5 で機械化済み（C1/C3/C5/C6b/C7/C8/C9・make ci 配線済み）

## 2. タスク

### FE9-1: shadow Level トークン実装【小・即着手可】

- **作業**: `frontend/src/styles/globals.css` の `@theme inline` に追加:
  - `--shadow-level1: 0 0.175px 1.041px rgba(0,0,0,0.01), 0 0.8px 2.925px rgba(0,0,0,0.02), 0 2.025px 7.847px rgba(0,0,0,0.027), 0 4px 18px rgba(0,0,0,0.04);`
  - `--shadow-level2`: design-system.md §5 の Level-2 deep stack（末尾 `0 23px 52px rgba(0,0,0,0.05)`。中間段は Level-1 の等比を踏襲して構成し、採用値を §5.1 へ追記）
  - Tailwind v4 が `shadow-level1` / `shadow-level2` クラスを自動生成することを確認
- **検証**: 代表1箇所（Dialog）へ適用し Chrome 目視 + `npx vitest run src/components/ui`
- **完了条件**: 両トークンが globals.css に存在し、§5.1 の値記載と一致

### FE9-2: 機械置換＋audit C10/C11 新設

- **置換1（shadow・70件）**: 使用箇所ごとに文脈判定して置換 — 機械一括ではない:
  - 通常カード・セクション枠の `shadow-sm` → **削除**（Level 0 = hairline のみ。border が無い場合は `border ${C.borderLight}` を付与）
  - ドロップダウン・ポップオーバー・浮動パネル・フォーカス強調の `shadow-sm/md` → `shadow-level1`
  - モーダル（Dialog）・トーストの `shadow-lg` / 任意値 → `shadow-level2`
  - 判定に迷う箇所は「Level 0（削除）」に倒す（②削除原則・barely-there 哲学）。全置換箇所を Completion Report に列挙
- **置換2（font-size 任意値・66件）**: 値分布実測 = `[11px]`×25・`[10px]`×23・`[12px]`×14・`[24px]`×4。10/11px を caption(13px) へ丸めると視覚激変のため **`--text-2xs: 11px` を micro ロールとして公式追加**（radius 3px と同型裁定・design-system.md §3.4 反映済み）。対応: `text-[10px]`→`text-2xs`（+1px）／`text-[11px]`→`text-2xs`（等値）／`text-[12px]`→`text-xs`（+1px）／`text-[24px]`→`text-2xl`（等値）。他の値が現れたら最近傍ロールへ・±2px 以上は列挙報告
- **audit 追補**（`frontend/scripts/design-system-audit.mjs` + `design-system-audit.test.mjs`・既存 C7〜C9 と同型）:
  - **C10**: `shadow-(sm|md|lg|xl|2xl)\b` および `shadow-\[` 禁止（`shadow-level1/2`・`shadow-none` のみ許可。ui/ 配下の shadcn 基底は置換済み前提で allowlist なし）
  - **C11**: `text-\[[0-9.]+(px|rem)\]` 禁止
  - `docs/spec/ui-design-compliance.md` §1 表へ C10/C11 を同一コミットで追記
- **スコープ**: FE7 教訓どおり **src+liff/src+line-reserve/src の3アプリ全域**。置換後 grep 残数 0 確認
- **検証**: `node --test scripts/design-system-audit.test.mjs` + `node scripts/design-system-audit.mjs`（違反0）+ 影響 feature の scoped vitest + Chrome 目視（影の消失/変化は視覚確認必須）

### FE9-3: ページ毎スイープ（機械判定不能項目のみ）【前提条件あり】

**前提条件（先に解消 — 未解消なら着手しない）**: エージェントがログイン済みブラウザで全画面を実測できること。FE8/FEAT-CHECKIN で2度「DevTools ログイン不可→USER 目視依頼」になった。解消手段はユーザーと合意の上で用意する（例: 開発環境の DevTools ポート 127.0.0.1:9222 にログイン済みタブを開いておく運用）。**84画面の USER 手動目視への丸投げは不可**。

- **ページ一覧の正本**: `docs/spec/ui-design-compliance.md` §2 の全リーフルート表（84 routes・83 準拠/1 対象外）。**本書へ転記しない**（台帳への番号列挙は必ずドリフトする — 着手時に §2 を読み batch を編成する）
- **batch 編成**: feature 単位で 8〜12 ページ/batch（目安: ①owners/pets ②reception/reservations ③medical-records 系 ④accounting/cash-register/estimates ⑤hospitalization/trimming/vaccinations/checkups ⑥master/settings 系 ⑦inventory/aggregation/lstep/manual/auth ほか）。1 batch = 1 コミット
- **ページ毎チェックリスト（v1 — 機械 audit C1〜C11 でカバー済みの項目は含めない）**:

```
□ P1 図地: ページ canvas = bgPage(#F6F5F4)、カード/フィールド = 白 surface。全面純白ページになっていない
□ P2 境界: カード境界は hairline（C.borderLight）。Level 0 に影がない。重い枠・二重枠がない
□ P3 階層: 見出しが §3.4 ロール表どおり（title>section>body）。font-bold が本文・数値セルに漏れていない
□ P4 余白リズム: セクション間隔が spacing スケール（8/12/16/24px）で階段状。詰まり・過剰な空きがない。
     グルーピングは罫線でなく whitespace が主
□ P5 単一アクセント: brand teal が Primary CTA・リンク・active/focus のみ。装飾に構造色・
     sticker 色が漏れていない（semantic 色 = 臨床安全用途のみ）
□ P6 テーブル chrome: ex-data-table-cell 準拠（header = canvas-soft + eyebrow / body = body-sm / hairline 行区切り）
□ P7 状態: hover/focus が知覚可能。disabled が RBAC 非活性表現（C6a）を退行させていない
```

- **手順/batch**: 各ページをブラウザで開き screenshot → P1〜P7 判定 → 違反を最小差分で修正 → scoped vitest + design-audit → screenshot 再取得。判定と証跡（違反項目・修正 file:line）を Completion Report に列挙
- **エスカレーション**: チェックリストで判定できない構造問題（レイアウト再設計級）は修正せず「発見事項」として報告 — 本計画のスコープ外（別起票）

## 3. 裁定記録（2026-07-21 確定・design-system.md へ反映済み）

| 論点 | 裁定 |
|------|------|
| pill CTA | 不採用（アプリ本体は utility 系角丸。§7.2） |
| typography 採用範囲 | title 以下のみ。display/h1 系はマーケ専用（§3.4） |
| shadow Level 1/2 | 採用。Tailwind 既定影は禁止化（§5.1） |
| caption/eyebrow サイズ | 13px に統一（DESIGN.md 14px/12px の製品上書き。§3.4） |

## 4. やらないこと（決定済み）

- **primary 色の変更** — #038B94 維持。DESIGN.md の #0075DE は C1 で禁止済みのテンプレート値
- **semantic 色（danger 等）の削除** — 臨床安全 UI（C6a）は本計画に優先する
- **pill CTA・display 型・sticker 装飾の導入** — マーケ面専用（§3.4/§7.2 裁定）
- **hover 状態の網羅仕様化** — DESIGN.md 未文書化。各プリミティブ実装に委譲（§7 注記どおり）
- **liff/line-reserve のページ毎スイープ** — FE9-2 の機械置換のみ対象。画面監査は 84 routes（本体）に限定

## 5. 検証規約（frontend/CLAUDE.md 準拠）

- フル `pnpm lint`/`test:run`/`type-check`/`build` は自動実行禁止 → scoped は `npx vitest run <path>` / `npx eslint <paths>`。audit は `node scripts/design-system-audit.mjs`（直接実行可・make ci 配線済み）
- FE9-2/9-3 は視覚変更のため Chrome 実測必須。コミットはタスク/batch 単位・conventional commit・Co-Authored-By 禁止（Cursor は自動付与するため commit 直後に `git log -1 --format=%B` を確認し混入時は HEAD のうちに amend 除去）
