# PR #49 Post-Merge Smoke Checklist

> **対象**: PR #49 — 混在会計 (payment_splits)、予約フォーム改善、健診 doctor フィールド、seed 拡充
> **作成**: 2026-05-24
> **使用タイミング**: PR #49 が `main` → `staging` へ反映されたデプロイ完了直後

---

## A. インフラ基本確認

| # | チェック項目 | 期待値 |
|---|------------|--------|
| A-1 | STG deploy 成功 | GitHub Actions `backend-deploy.yml` → success |
| A-2 | `GET https://api.stg.noah-karte.com/health` | `200 OK`, `"status":"ok"` |
| A-3 | ECS サービス状態 | `runningCount == desiredCount`（タスク再起動なし） |
| A-4 | CloudWatch ログ | `ERROR` / `panic` が 0 件（deploy 完了から 5 分間） |
| A-5 | Vercel Preview | STG フロントエンドが正常ロードされる |

```bash
# A-3 確認コマンド
export AWS_PROFILE=AnimalEkarte
aws ecs describe-services \
  --cluster animalekarte-stg-cluster \
  --services animalekarte-stg-service \
  --region us-east-1 \
  --query 'services[0].{status:status,runningCount:runningCount,desiredCount:desiredCount}'

# A-4 確認コマンド
aws logs tail /ecs/animalekarte-stg --since 5m --region us-east-1 | grep -i 'error\|panic'
```

---

## B. 予約 (Reservation)

> PR #49 変更: 週次カレンダービュー 5日/7日 toggle、クリック日時の初期値セット、新規飼主インライン作成

| # | チェック項目 | 期待値 |
|---|------------|--------|
| B-1 | 5日表示 | 従来通り横スクロール表示、クリニック業務時間帯が表示される |
| B-2 | 7日表示 | `/calendar?view=week7` で正常表示（overflow なし） |
| B-3 | 5日/7日 toggle | ボタンで切り替わり、URL パラメータが反映される |
| B-4 | スロットクリック → フォーム初期値 | クリックした日時がフォームの「日付」「開始時刻」に反映される |
| B-5 | 既存飼主で予約作成 | 患者検索 → 選択 → 予約確定が正常完了する |
| B-6 | 新規飼主インライン作成 | 「新規飼主」ボタン押下で4フィールドフォームが表示される |
| B-7 | 新規飼主バリデーション | 電話番号が `0` 始まりでない場合にフォームエラーが表示される |
| B-8 | 新規飼主からの予約完結 | owner → pet → reservation の順で作成が完了する |
| B-9 | 電話番号 BE 整合 | `090-1234-5678` / `09012345678` 形式が通る、`1234-5678` は弾かれる |

---

## C. 会計 / 混在会計 (Mixed Payment)

> 詳細テスト手順は **[MIXED-PAYMENT-SMOKE-TEST.md](./MIXED-PAYMENT-SMOKE-TEST.md)** を参照。
> 以下はキーポイントのみ列挙。

| # | チェック項目 | 期待値 |
|---|------------|--------|
| C-1 | 単一支払い: 現金 | `payment_splits` 1行 (method=cash), status=completed |
| C-2 | 単一支払い: クレジットカード | `payment_splits` 1行 (method=credit_card), 正常完了 |
| C-3 | 単一支払い: 電子マネー | `payment_splits` 1行 (method=electronic_money), 正常完了 |
| C-4 | 2種混在: 現金 + クレジット | `payment_splits` 2行、合計 = billing_amount |
| C-5 | 3種混在: 現金 + クレジット + 電子マネー | `payment_splits` 3行、合計 = billing_amount |
| C-6 | 現金 received_amount / change_amount | 預り金 - 現金金額 = お釣り が正しく保存 |
| C-7 | 支払合計不足 | 保存ボタンが無効のまま |
| C-8 | 支払合計超過 | 保存ボタンが無効のまま |
| C-9 | 保存後再表示 | `payment_splits` の行数・method・amount が復元される |
| C-10 | 既存会計（splits なし） | 古い会計が単一支払いとして表示される（後方互換） |
| C-11 | 本日会計タブ: 支払方法表示 | 混在会計が「現金 / カード」など `/` 区切りで表示される |
| C-12 | 本日会計タブ: 支払方法別合計 | `payment_splits.amount` ベースの集計値が正しい |
| C-13 | レジ締め: 理論現金 | 現金 split の amount のみが加算される（billing_amount 非依存） |
| C-14 | レジ締め: 支払方法別集計 | 混在会計の各 split が正しい支払方法欄に計上される |

---

## D. 返金 (Refund)

> **仕様制約**: 返金は会計 (billing) 単位のみ。支払方法別の返金は未実装。
> **関連 Issue**: [GitHub Issue #60](https://github.com/MinoruSoga/AnimalEkarte/issues/60) — 支払方法別返金の将来仕様

| # | チェック項目 | 期待値 |
|---|------------|--------|
| D-1 | 混在会計に対して返金操作が完了する | `billing_refunds` に 1 行追加 |
| D-2 | `total_refunded_amount` が更新される | 返金額が加算された値になる |
| D-3 | レジ締め・日次集計の返金反映 | `billing_refunds` 単位で集計（split 単位ではない） |
| D-4 | 現金理論値の返金計算 | **制約**: 返金全額が現金扱いで計算される（cash split 按分は対応外） |

---

## E. CRUD

> 基本導線は **[CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md)** を参照。
> 以下は CRUD-SMOKE-TEST.md 未カバー項目のみ。

| # | チェック項目 | 期待値 |
|---|------------|--------|
| E-1 | 医院: 編集・保存後に値が維持される | 再読み込み後も変更値が表示される |
| E-2 | 権限グループ: `system_admin` グループの削除が拒否される | 403 または 409 エラー |
| E-3 | スタッフ: カルテが紐付くスタッフの削除が拒否される | 409 Conflict エラーと適切なメッセージ |
| E-4 | スタッフ: ログイン中の自分自身が削除できない | 削除ボタンが無効化されるか、実行時エラーが表示される |
| E-5 | スタッフ: 削除後に権限グループ参照が消える | 削除されたスタッフの FK が孤立しない |

---

## F. 入院・ホテル (Hospitalization)

| # | チェック項目 | 期待値 |
|---|------------|--------|
| F-1 | 入院詳細ページが表示される | `/hospitalization/:id` が正常ロード |
| F-2 | id なし / 存在しない id | Blank page にならず、エラー UI または 404 表示 |
| F-3 | サーバーエラー時 | スタックトレースが露出しない、ユーザー向けメッセージが表示 |
| F-4 | `admitted` 状態の表示 | 入院中バッジ・日付が正しく表示される |
| F-5 | `discharged` 状態の表示 | 退院済みバッジ・退院日が表示される |
| F-6 | `reserved` 状態の表示 | 予定バッジが表示される |

---

## G. 健診 / カルテ (Checkup / Medical Record)

> PR #49 変更: CheckupsTab に `doctor` フィールド追加、Path B カルテ `visit_type` NULL 修正

| # | チェック項目 | 期待値 |
|---|------------|--------|
| G-1 | CheckupsTab に担当医フィールドが表示される | 健診タブを開いたとき `担当医` セレクトが存在する |
| G-2 | 担当医を設定して保存できる | 保存後に再表示で設定値が維持される |
| G-3 | 担当医を「なし」に戻せる | 空白 / `-` 選択後に保存できる |
| G-4 | Path B カルテの `visit_type` が NULL でない | 健診から自動生成したカルテの `visit_type` に値が入っている |
| G-5 | finalized カルテでは健診タブ編集不可 | 確定済みカルテの健診タブが readonly 表示になる |

---

## H. マスタ (Masters)

| # | チェック項目 | 期待値 |
|---|------------|--------|
| H-1 | 診療項目: 保険対象外フラグの作成・保存 | `is_insurance_applicable=false` で保存、再表示で値が維持される |
| H-2 | 薬剤: 保険対象外フラグの作成・保存 | 同上 |
| H-3 | マスタ: 更新後の再表示 | 一覧・詳細いずれも最新値が表示される（キャッシュ問題なし） |
| H-4 | 予約区分: 新規作成・色設定・保存 | 色ピッカーで設定した色が一覧に反映される |

---

## I. Seed / Demo データ整合性

> DB_RESET=true でデプロイした場合のみ確認。通常デプロイでは不要。

| # | チェック項目 | 期待値 |
|---|------------|--------|
| I-1 | 2026-05-22 前後の予約が表示される | 週次カレンダーで城東・八王子両クリニックの予約が表示される |
| I-2 | 本日会計データが 2026-05-22 に存在する | 本日会計タブで対象日の会計が表示される |
| I-3 | 城東 `medical_records` id=21-28 が壊れていない | カルテ詳細が正常表示、関連 exam_results が存在する |
| I-4 | 八王子当日 `medical_records` id=61-65 が存在する | カルテ一覧・詳細が表示される |
| I-5 | `exam_results` id=46-50 が城東 exam に正しく紐付く | 検査結果タブに値が表示される（別クリニックのデータと混在しない） |
| I-6 | `payment_splits` に seed データが存在する | 本日会計の混在支払い会計で splits が表示される |

---

## J. NG 時の切り分け

| 現象 | 初動確認箇所 |
|------|------------|
| API 500 エラー | CloudWatch `/ecs/animalekarte-stg` → `ERROR` / `panic` 行を確認 |
| 画面のみ崩れる（API は正常） | Vercel build ログ / ブラウザ console の JS エラーを確認 |
| 会計集計値が不一致 | `payment_splits` / `payments` / `billings` テーブルを SQL で確認、JOIN 条件をレビュー |
| 混在会計の splits が保存されない | `SavePaymentSplits` のトランザクション / DELETE 後 INSERT の実行ログを確認 |
| 再表示で splits が消える | API レスポンスの `payment_splits` フィールド / FE `transformToAccounting` の fallback 条件を確認 |
| 予約作成失敗（新規飼主） | owner → pet → reservation の作成順序ログを確認、FK 制約エラーがないか |
| 予約フォームに日時が反映されない | フロントの `initialData` → `formData` への初期値セット処理を確認 |
| 健診 doctor フィールドが空欄 | `checkup_results.doctor_id` / GET レスポンスの `doctor` フィールドを確認 |
| 入院ページが blank | `useHospitalization` クエリのエラー状態 / `id` パラメータの型変換を確認 |
| seed 不整合（DB_RESET 後） | migration `001_init.sql` 〜 `004_seed_demo.sql` のログを確認、checksum mismatch がないか |

---

## 関連ドキュメント

- [MIXED-PAYMENT-SMOKE-TEST.md](./MIXED-PAYMENT-SMOKE-TEST.md) — 混在会計の詳細テスト手順 (Section C の全項目)
- [CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) — 医院/権限グループ/スタッフ CRUD の詳細手順 (Section E の基本)
- [STG_PRE_DEPLOY_READINESS_CHECK.md](./runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md) — デプロイ前後のインフラ確認ランブック
- [GitHub Issue #60](https://github.com/MinoruSoga/AnimalEkarte/issues/60) — 支払方法別返金の将来仕様

---

## チェックリスト合格基準

- A〜E: **全項目 PASS** が merge / STG 昇格の最低条件
- F〜H: **重大な Blank page / 500 がないこと** を確認
- I: DB_RESET デプロイ時のみ必須
- NG が 1 件でも発見された場合は即座にロールバックを判断し、[リリース準備ランブック](./runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md) の §5 ロールバック基準に従う
