# S05: 入院サイクル（ケア記録→退院会計）

> **目的**: 入院登録からデイリーケア記録、ケージ稼働の可視化、退院時の会計統合までの入院業務一巡が成立し、退院済み患者への二重退院が拒否されることを納品前に証明する。
> **所要目安**: 25分 / **深度**: 深い
> **仕様正本**: [specification.md §3.1](../../../spec/specification.md)・[screens/07-hospitalization-list.md](../../../spec/screens/07-hospitalization-list.md)・[screens/08-hospitalization-detail.md](../../../spec/screens/08-hospitalization-detail.md)・[screens/09-hospitalization-form.md](../../../spec/screens/09-hospitalization-form.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: vet ロール（hospitalization の create/edit 権限あり）。
- 対象ペット: ペット検索で「ステータス=生存」のペットを 1 頭選ぶ。
- ケージマスタに、現在入院中（active）の患者が割り当てられていないケージが 2 つ以上あること（ケージ移動の確認に使用）。
- 依存シナリオ: なし。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 入院管理から新規入院登録 → ペット選択 → 入院タイプ「入院」・予定期間・空きケージ・主訴を設定し、ケアプランに給餌指示と投薬指示を各 1 件登録して保存（[09 §1/§2](../../../spec/screens/09-hospitalization-form.md)） | 保存成功。入院一覧の「入院中」タブに表示される（[07 §2](../../../spec/screens/07-hospitalization-list.md)） |
| 2 | 一覧をボードビューに切り替える | 割り当てたケージの位置に患者カードが表示され、ケージ稼働状況が可視化される（[07 §1](../../../spec/screens/07-hospitalization-list.md)・[specification.md §3.1](../../../spec/specification.md)) |
| 2b | リストビューへ切り替える | 入院期間・主訴・担当医を含む表形式で当該入院が表示される（[07 §1](../../../spec/screens/07-hospitalization-list.md)） |
| 3 | 入院詳細を開き、デイリーログにバイタル 1 件（時刻・測定値）とケアログ 1 件を追加する | それぞれ独立に記録され、日付単位の時系列で「バイタル」「ケアログ」セクションに表示される（[08 §2/主要機能1](../../../spec/screens/08-hospitalization-detail.md)） |
| 4 | ボードビューで患者カードを別の空きケージへドラッグする | 収容ケージが直感的に変更され、移動先ケージにカードが表示される（[07 主要なユーザーアクション](../../../spec/screens/07-hospitalization-list.md)） |
| 5 | 入院詳細から退院処理を開き、ダイアログで **「退院後、そのまま会計画面へ進む」にチェックを入れて** 実行する（チェック既定は **OFF** — `DischargeAlertDialog`） | 確認ダイアログ（「退院処理を実行」）を経て `POST …/hospitalizations/:id/discharge-with-billing`（`create_accounting: true`）が走り、ステータスが退院（discharged）になる。`accounting_id` があれば会計詳細へ、無ければ `?petId=` 付き新規会計フォームへ遷移（[08 §2 退院プロセスと会計連携](../../../spec/screens/08-hospitalization-detail.md)）。会計明細にケアプラン由来の項目が含まれる |
| 6 | 入院一覧に戻り、タブとボードを確認する | 当該患者が「入院中」タブから消え「退院済」タブに表示される（[07 §2](../../../spec/screens/07-hospitalization-list.md)）。「入院中」タブのボードで当該ケージは空きに戻る。「退院済」タブのボードは退院患者を最終ケージ位置に表示し続ける（履歴表示 — 2026-07-17 実測で確認） |
| 7 | 2 件目の入院を作成し、退院処理ダイアログで **「退院後、そのまま会計画面へ進む」を入れないまま**（既定 OFF）退院を実行する | 会計なし経路: FE は `PATCH` で `status=discharged` のみ更新（`DischargeWithBilling` は呼ばない）。会計画面への遷移なし・会計レコードが作られないこと |

## 確認観点

- **会計あり退院**のみ `DischargeWithBilling`（`create_accounting: true`）で退院と会計生成が同一トランザクション（`backend/internal/medicalrecord/hospitalization_service.go`）。片方だけ成立しない。
- **会計なし退院**は FE が通常 Update（`status=discharged` + `end_date`）を使う（`use-hospitalization-detail.ts`）。コメント上も PATCH-driven discharge は `DischargeWithBilling` を bypass する。
- 退院前の最終確認は `DischargeAlertDialog`（タイトル「退院処理を行いますか？」・チェック「退院後、そのまま会計画面へ進む」既定 OFF）。
- 二重退院（`discharge-with-billing`）は行ロック（FOR UPDATE）で直列化し `hospitalization is already discharged`（InvalidInput → 4xx）で拒否。
- 入院登録・ケア記録・退院が `audit_logs` に記録されること（[specification.md §2.1](../../../spec/specification.md)）— DB 参照は USER 実施。会計ありは `hospitalization.discharge_with_billing`。
- 入院フォームのケージ割り当ては空き状況によるフィルタリングを行わない（[09 §1](../../../spec/screens/09-hospitalization-form.md)）— 占有中ケージが選択できても仕様どおりであり逸脱ではない。
- ペット選択クエリは `petId`（camelCase）。退院後の新規会計遷移も `?petId=`。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 退院済み入院に対して `POST …/discharge-with-billing` を再度送る（API 直接・任意） | 「hospitalization is already discharged」が返り、会計レコードは二重生成されない（2026-07-17 実測で確認） |
| A2 | ケアプランが紐付いたままの入院レコードを一覧から削除する | BE: 「ケアプランが紐付いているため削除できません…」（`hospitalization_service.go`）。**runtime 2026-08-01 BLOCKED**: `/hospitalization` 入院 0 件（ケージ全空き） |

---

## 実装突合

- 突合日: 2026-08-07
- HEAD: `844e43f69`
- 変更サマリ:
  - 会計あり/なしで API 経路が分岐（`discharge-with-billing` vs PATCH）することを手順 5/7 と確認観点に反映
  - チェック「会計画面へ進む」既定 OFF を明記
  - A2 削除拒否メッセージを実装文言に合わせて更新
