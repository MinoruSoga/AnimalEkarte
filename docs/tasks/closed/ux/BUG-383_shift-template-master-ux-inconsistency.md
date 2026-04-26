# BUG-383: シフトテンプレートマスタで他マスタと UX が不整合

**作成日**: 2026-04-15
**Status**: CLOSED
**Priority**: LOW (UX 一貫性 / 軽微な機能欠落)
**Affects**: `features/settings/shift-template`, `/settings/shift-template`

---

## 概要

`/settings/shift-template`（シフトテンプレートマスタ）が他のマスタ設定画面と UI・UX パターンが揃っていない。以下 3 点が不整合。

## 不整合点（ブラウザ検証 2026-04-15）

### 1. 件数表示の欠落

- 他マスタ（主訴 / 薬剤 / 保険 / 動物種類 / 問診テンプレート / 物販 / 処置 / 診断病名 / ケージ / 入院 / トリミング / 予約区分 / 動物種類 / スタッフ）はすべて「**XX 件**」をヘッダー直下に表示
- シフトテンプレートマスタ **のみ**表示なし
- ユーザーが一覧に何件あるか視認できない／登録・削除後の変化を確認できない

### 2. 削除確認ダイアログに対象名が無い

他マスタ（主訴/薬剤/問診/動物種/物販/処置/保険/ケージ/トリミング）:
```
「ブラウザテスト主訴_更新」を削除します。この操作は取り消せません。
```

シフトテンプレート:
```
このテンプレートを削除しますか？
この操作は取り消せません。
```

- 対象名が無いため **誤削除リスク**（複数行を連続処理する運用で致命的）
- ダイアログ文言が疑問形（他マスタは「削除します」の断定形）

### 3. 開始/終了時刻の必須バリデーション不足

- 「全日」種別でテンプレート名のみ入力・時刻 `00:00` のまま保存すると **POST 201 で成功**
- 一覧の「時間」列は **「-」** と表示されるだけで、完全にゴミデータ
- 正しくは種別が「全日・午前・午後」の場合は時刻必須にすべき（「休日・有休」のみ時刻不要）

### 4. URL の命名不整合

- 他マスタ: 複数形 kebab-case（`/settings/reservation-types`, `/settings/insurance` 等）
- シフトテンプレート: **単数形**（`/settings/shift-template`）
- CLAUDE.md の「API パス: kebab-case / テーブル名と 1:1」ルールに照らすと改名候補

## 再現手順

1. `admin@noavet.jp` で `/settings/shift-template` へ遷移
2. 初期 5 件（通常勤務/午前勤務/午後勤務/休日/有給休暇）表示。件数「5 件」表示なしを確認
3. 「+ 新規登録」 → テンプレート名のみ入力（時刻未設定） → 「保存」ボタン活性化 → クリック
4. **POST 201** 成功、一覧に「ブラウザテストシフト 全日 - 有効」として追加（時刻列「-」）
5. 追加行 → 操作メニュー → 編集パネル → 削除アイコン
6. 確認ダイアログ「このテンプレートを削除しますか？」表示（対象名なし）
7. 「削除」→ DELETE 成功

## 修正方針

### 1. 件数表示の追加
他マスタ同様 `<MasterPageHeader count={items.length} />` 等の共通コンポーネントに揃える。

### 2. 削除確認ダイアログに対象名埋め込み
shadcn/ui `AlertDialog` の description に `{template.name}` を差し込む。他マスタ実装を参考。

### 3. 時刻バリデーション
`features/settings/shift-template/hooks/use-shift-template-form.ts` に:
- 種別が `fullday` / `morning` / `afternoon` のときは `start_time` / `end_time` 必須
- 種別が `holiday` / `paid_leave` のときは時刻フィールド自体を非表示

### 4. URL 命名
- router を `/settings/shift-templates` に変更（複数形）
- 旧 URL は 301 redirect（外部ブックマーク互換）

## 関連

- CLAUDE.md `.claude/rules/naming-conventions.md` (API パス kebab-case / 複数形)
- BUG-382（他 dead route / 命名不整合）と同系統の UX 一貫性問題

## 確認事項

- [ ] バックエンド `shift_templates` テーブルの `start_time` / `end_time` が nullable か要確認
- [ ] 休日・有休でも時刻必須にしている API 仕様があるか
- [ ] 既存シードの休日/有休データの時刻値（`"-"` になっているのでおそらく NULL）
