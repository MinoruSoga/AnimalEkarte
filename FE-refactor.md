# FE-refactor 第9期-B（FE9-3）— ページ毎 DESIGN.md 視覚遵守スイープ

> 起票: 2026-07-21 再作成（要件責任者: 曽我。「per-page・checklist で全画面を目視して DESIGN.md 準拠を確認せよ」の指示）
> **本ファイルは対応後削除する使い捨てトラッカー**。恒久規約は `docs/spec/design-system.md`、機械ガードは audit C1〜C11（既に緑）。
> **遵守の正本 = `docs/spec/design-system.md`**（DESIGN.md の製品翻訳。字義遵守でない — primary=#038B94・semantic 色維持・pill/display はマーケ専用）。
> 実行 = ログイン済みブラウザ（port 9222）で1画面ずつ screenshot → P1〜P7 判定 → 逸脱を最小差分修正 → batch 単位でコミット → 本表に記録。

## 方法

各画面で以下を判定（機械 audit C1〜C11 でカバー済みの色/角丸/影/font-size/maxWidth は除く。ここは目視でしか分からない項目のみ）:

```
P1 図地   : ページ canvas=bgPage・カード/フィールド=白 surface。全面純白でない
P2 境界   : カード境界=hairline。通常カードに重い影/二重枠がない
P3 階層   : 見出しが §3.4 ロール(title>section>body)どおり。font-bold が本文/数値セルに漏れてない
P4 余白   : セクション間隔が spacing スケール(8/12/16/24)で階段状。詰まり/過剰空きがない。罫線でなく whitespace でグルーピング
P5 アクセント: brand teal が CTA/リンク/active/focus のみ。装飾に構造色/sticker が漏れてない(semantic は臨床用途のみ)
P6 テーブル: 一貫 house 様式(plain muted ヘッダ + hairline 行区切り)を維持(§7.5 裁定)
P7 状態   : hover/focus が知覚可能。disabled が RBAC 非活性表現(C6a)を退行させてない
```

判定記号: ✅ 準拠 / 🔧 修正した / ⚠️ 好み分岐(裁定要) / — 該当なし

## 進捗表（ルート正本 = docs/spec/ui-design-compliance.md §2）

| Batch | ルート | P1 | P2 | P3 | P4 | P5 | P6 | P7 | 修正/所見 | 状態 |
|-------|--------|----|----|----|----|----|----|----|-----------|------|
| 0 | /login | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | ✅済 |
| 0 | /（受付カンバン） | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | — | ✅済 |
| 0 | /owners（一覧） | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | P6=house様式一貫 | ✅済 |
| 0 | /medical-records（一覧） | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | — | owners と同様式 | ✅済 |
| B1 | /reservations（週カレンダー） | ✅ | ✅ | ✅ | ✅ | ✅ | — | ✅ | カテゴリ凡例=sanctioned dots・CTA/active は teal のみ | ✅済 |

## Batch 計画（feature 単位・1 batch=1 コミット）

- **B1 予約系**: /reservations・/reservations 詳細モーダル・予約登録モーダル
- **B2 入院系**: /hospitalization・入院詳細・デイリー記録タブ
- **B3 カルテ編集**: /medical-records/:id（問診タブ・診察/治療プランタブ ← FE8 修正の視覚確認込み）
- **B4 会計系**: /accounting・/accounting/reports・/estimates
- **B5 マスタ/設定**: /settings・/settings/clinic・master 各種
- **B6 その他一覧**: /inventory・/trimming・/vaccinations・/checkups・/examinations
- **B7 分析/連携**: /aggregation・/lstep・/shifts・/manual

## 前提の再確認（やらないこと）

- primary 色変更・semantic 色削除・pill/display/sticker 装飾の導入（マーケ専用・§1/§3.4/§7.2）
- liff/line-reserve のページ監査（本体84ルートに限定）
- 機械 audit で緑の項目の再チェック（色/角丸/影/font-size は C1〜C11 が保証）

## 検証規約

- 各 batch 修正後: 影響 feature の `npx vitest run <path>` + `node scripts/design-system-audit.mjs`（緑維持）。コミットは Co-Authored-By 禁止。
- フル lint/type-check/build は USER 手動。
