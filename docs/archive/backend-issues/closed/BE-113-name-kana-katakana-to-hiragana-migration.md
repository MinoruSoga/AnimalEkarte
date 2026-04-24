# BE-113: name_kana 既存データのカタカナ→ひらがな一括変換マイグレーション

**Status**: Closed (2026-04-14)
**Priority**: Medium
**Affects**: `owners.name_kana`, `pets.name_kana`, シードデータ `003_seed_demo.sql`
**Date Created**: 2026-04-14
**Related**: BUG-375, BE-114, FE-251

## Summary

`owners.name_kana` と `pets.name_kana` に登録されているカタカナ文字 (U+30A1〜U+30F6) を
ひらがな (U+3041〜U+3096) へ一括変換する DB マイグレーションを追加し、シードもひらがな化する。

## 現状のコード

```sql
-- backend/migrations/001_init.sql:202-203 (owners)
name             text            NOT NULL,
name_kana        text            NOT NULL DEFAULT '',

-- backend/migrations/001_init.sql:574-576 (pets)
name              text            NOT NULL,
name_kana         text            NOT NULL DEFAULT '',
```

```sql
-- backend/migrations/003_seed_demo.sql:914 (例)
(1,  3, '林 文明', 'ハヤシ フミアキ', '1980-05-15', ...
```

## 必要な変更

### 1. 新規マイグレーション (`backend/migrations/004_convert_kana_to_hiragana.sql`)

`translate()` で 1 文字単位マッピング。86 文字すべてを 1 行で対応。

```sql
-- BUG-375 BE-113: owners.name_kana / pets.name_kana のカタカナをひらがなに変換
-- 文字マッピング: U+30A1 (ァ) 〜 U+30F6 (ヶ) → U+3041 (ぁ) 〜 U+3096 (ゖ)
-- translate() は文字単位置換のため冪等（再実行しても結果不変）

UPDATE owners
SET name_kana = translate(
    name_kana,
    'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ',
    'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ'
)
WHERE name_kana IS NOT NULL AND name_kana <> '';

UPDATE pets
SET name_kana = translate(
    name_kana,
    'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ',
    'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ'
)
WHERE name_kana IS NOT NULL AND name_kana <> '';
```

### 2. シードデータ更新 (`backend/migrations/003_seed_demo.sql`)

既存の `name_kana` 値（owners + pets）をひらがなに置換。
sed コマンド例:
```bash
# カタカナ→ひらがな対応表で sed 一括変換
# 実装時は手動置換 or スクリプト経由
```

主な対象（grep で全件特定）:
- `owners.name_kana` 値 (例: `ハヤシ フミアキ` → `はやし ふみあき`)
- `pets.name_kana` 値 (例: `モモ` → `もも`)

## API レスポンス形式

変更なし。`name_kana` フィールドの型は string のまま。

## フロントエンド影響

- API レスポンスがひらがなで返るようになる
- FE-251 で UI ラベル変更が必要

## 完了条件

- [ ] `backend/migrations/004_convert_kana_to_hiragana.sql` 新規追加
- [ ] `make migrate` または DB リセット運用で適用確認
- [ ] `SELECT name_kana FROM owners WHERE name_kana ~ '[ァ-ヶ]'` の結果が **0 件**
- [ ] `SELECT name_kana FROM pets WHERE name_kana ~ '[ァ-ヶ]'` の結果が **0 件**
- [ ] 003_seed_demo.sql の `name_kana` リテラル値もひらがな化
- [ ] 既存テスト全件パス (`go test ./...`)
- [ ] マイグレーション 2 回実行しても結果が変わらない（冪等性確認）

## 参照

- PostgreSQL `translate()` 関数: 文字単位置換、長さ不一致時は from 末尾の余剰文字を削除
- カタカナ Unicode 範囲: U+30A1〜U+30F6 (86 文字、ァ〜ヶ)
- ひらがな Unicode 範囲: U+3041〜U+3096 (86 文字、ぁ〜ゖ)

## リスク

| リスク | 影響 | 対策 |
|--------|------|------|
| マッピング表の文字数不一致でずれ発生 | 高 | from/to 共に 86 文字を厳密一致させる（本イシュー本文に記載済み） |
| 既存外部連携（CSV インポート等）がカタカナ前提 | 低 | 連携機能なし or BE-114 検索が両対応するため新規入力もカタカナで通る |
| 半角カナ（ﾊﾔｼ）残存 | 低 | スコープ外。本マイグレーションは全角カタカナのみ対象 |
