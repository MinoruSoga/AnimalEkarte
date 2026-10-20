# EMR 予約関連票の実装・検証記録（2026-10-08）

> **2026-10-19の状態補足**：[EMR-270](https://app.plane.so/baritechllc/browse/EMR-270/)は本人が解消済みと認識しているため、下記の当時のSTG再現待ちを現在の仕様・設計blockerとして残さない。以下は10/8の調査記録であり、今回の再現・修正・実機検証を意味しない。

## 作業範囲と所有

Plane workspace `baritechllc` / project `EMR` の本文・最新コメント・受入条件を読み、EMR-270 / 274 / 275 / 318 / 321 を claim して Running に変更した。`origin/main` の `1b1b1c5ed` から専用 worktree `emr-reservation` / branch `fix/emr-reservation` を作成し、子担当も別 worktree で実装・レビューした。共有 main と他セッションの変更は保持した。claim は PR の受入まで保全する。

受付画面本体は変更しない。EMR-275 / 278 の表示名称「予約内容」を Plane コメントで共有した。ホテルのケージ任意化は EMR-316、連結は EMR-306 の範囲とする。

`agent-fast-campaign --all-recommended --boost` の正式 preflight は子エージェント容量不足で停止したため、Skill の fallback に従い通常の isolated execution として進めた。正式 campaign の COMPLETE 判定は行わない。

## 各票の結果と受入境界

| 票 | 実装・調査 | 残る確認 |
| --- | --- | --- |
| EMR-270 | 編集中に開始時刻を繰り返し変更し、終了時刻が更新される回帰テスト。担当者あり／なし、満員以外の枠を実 Radix Select で選ぶ。現行 API は満員を含む全枠を返すため、過去の「自己予約で候補が消える」仮説から backend を変更しない。 | STG の available-times A/B 応答と対象データで再現できていない。修正済みとは判定しない。EMR-285 のデータ準備後に人間が再現確認。 |
| EMR-274 | 作成時は新規飼主を初診、既存ペットを再診に固定。表示と送信値を一致。編集・次回予定確定では変更可能。 | 実診療の初診履歴を判定する仕組みではない。未受診の登録済みペットも再診となる。通知への伝播は画面仕様に記載。 |
| EMR-275 | 作成／編集フォーム、予約詳細の表示名称と placeholder を「予約内容」に統一。 | API / DB は `notes` のまま。受付側 EMR-278 の実装は別担当。 |
| EMR-318 表示 | 患者結果の高さ制約とスクロール領域を整理し、フォーカス可能な結果領域を設ける。 | Windows 実ブラウザで長い検索結果をスクロールする受入が必要。DOM / class 検証だけでは実際のスクロールを証明しない。 |
| EMR-318 受付追加 | 飼主詳細の保存済み生存ペットに受付ボタン。予約作成権限、最新ペット状態、当日予約候補を確認し既存予約を再利用。確認済み以降は再刻印せず、複数候補は受付画面へ案内。 | PR #609レビュー対応で患者単位のtransaction lock後に当日候補を再検索する新規作成ガードを追加。並行POSTの後続は409として再取得を求める。実DB同時作成・ボード反映・一連の画面遷移は未実測。 |
| EMR-321 | 承認済みホテル区分を手動投入 SQL に追加。既存 live 行の属性を保持し、inactive を含む事後条件違反は全体 rollback。全医院・再実行・属性保持の契約テストと手動 runbook。 | データ投入、実 DB 契約テスト、医院別ケージ件数、STG UI は未実行。ケージ候補空の原因は通信／権限と未登録を network 応答で区別する。 |

EMR-318 の既存予約更新は、読取後に別スタッフが進めた状態を巻き戻さないよう、サーバー側の版照合と transaction 内の患者状態確認を追加した。未指定の従来更新経路は維持する。

## PR #609 レビュー対応

GitHubのレビュー正本は [REQUEST CHANGESコメント](https://github.com/MinoruSoga/AnimalEkarte/pull/609#issuecomment-6060355935)（対象旧HEAD `c0feb3a8a`）。規約軸と要件軸のHIGHは同じ「2タブ／2端末からの新規受付の競合」1件で、旧HEADでも問題が残っていた。権限迂回の指摘はレビュアーが訂正済みで、既存の受付mutationと同じ reservations create/editを使用することを再確認した。別途のreception viewをmutation権限に混同しない。

既存予約のCASに加え、新規受付POST／batchだけ `prevent_duplicate_reception=true` を指定する。サーバーはreception・checked_in・当日・医院と患者／飼主・生存を検証し、既存の患者intent lockの取得後に当日の未完了予約を検索する。候補が0件の場合だけ新規作成／次回予定の昇格へ進み、候補がある場合は409。後続要求が既存予約を黙って変更せず、再取得後に既存GET→CASまたは既受付の再利用を行う。409では入力を保持する。

完了／未来院／キャンセルは候補から除外し、同日中の正当な別来院は許可する。全日一律の患者×日unique制約を追加しない。複数ペットは同じtransactionで処理し、全バッチの患者ロック順を固定して応答順を保持する。通常予約の重複ルールと既存の手動作成契約は保持する。

ローカルに残っていた古いレビュー初稿は保全し、GitHubの訂正済みコメントへ同期した。変更後の承認・再判定・スレッド解決はレビュアーまたは人間に委ねる。本文の過去のPASSは旧HEAD時点の記録であり、今回の追加検証と区別する。

## 検証環境と証跡

対象 worktree を read-only mount した Docker の一時 offline runner を使用。共有 main を mount した既存 Compose コンテナでは候補差分の検証を行わない。dependency volume も read-only、ネットワークなし、明示 entrypoint。アプリ起動・migration 適用・codegen・seed 書換えは行わない。

```sh
python3 -B scripts/checks/verify-agent-task.py --base origin/main \
  --frontend-image animalekarte-frontend \
  --frontend-dependency-volume ekarte-frontend-node-modules \
  --backend-image animalekarte-backend \
  --backend-dependency-volume ekarte-go-mod-cache \
  --evidence /tmp/emr-reservation-evidence/final-verification.json
```

初回PR（`c0feb3a8a`）の統合結果: フロントエンド関連360件、backend reservation1688件・API契約33件・cmd/migrate51件・live insert3件、全16checks PASS。 scoped ESLint・Prettier・gofmt も PASS。backend の scoped golangci-lint は0 issues、build / vet も PASS（担当 worktree 証跡）。実DB CAS / seed contract は SKIP。

ローカル証跡は `/tmp/emr-reservation-evidence/`、再現可能な命令・結果は PR と Plane コメントに記録する。限定 coverage はフォーム4 component の statements 88.34 / branches 83.67 / functions 80.00 / lines 90.04、受付関連 statements 81.85 / branches 81.71 / functions 79.62 / lines 85.05（追加安全修正前）。全プロジェクト 80% 達成の証拠ではない。

依存 audit は Docker `pnpm audit --json` で脆弱性0件。`python3 scripts/old-db/verify_seed.py` は PASS。生成 CSV / manifest は変更していない。backend の DB 接続が必要なテストは SKIP として別記し、static テスト通過を DB 適用成功と扱わない。

## レビュー修正の統合検証

修正コミット: frontend `180d6db80`、backend `7640b051c`。旧HEADからの修正差分を専用worktreeで検証し、11 checks PASS。Frontend関連210件、reservation1695件、API契約33件、scoped ESLint／Prettier／gofmtがPASS。担当worktreeのbackend scoped golangci-lintは0 issues、build／vetはPASS。React／TypeScript／backend・DBの独立差分レビューは指摘0件。これは実DBでの並行作成を証明する結果ではない。

```sh
python3 -B scripts/checks/verify-agent-task.py --base c0feb3a8a \
  --frontend-image animalekarte-frontend \
  --frontend-dependency-volume ekarte-frontend-node-modules \
  --backend-image animalekarte-backend \
  --backend-dependency-volume ekarte-go-mod-cache \
  --evidence /tmp/emr-reservation-evidence/pr609-review-update-final.json
```

新規受付のDB回帰テスト4本は `-short` でSKIP。fixtureがschema／indexを準備するため、ユーザーがデータ消失影響を確認して専用使い捨てtest DBを準備し、候補worktreeをmountしたbackend環境とtest DB接続設定を確認してから次を手動実行する。既存／共有DBへ向けない。エージェントは隔離DB作成・破棄やmigrationを実行していない。

```sh
docker compose exec backend go test -count=1 ./internal/reservation \
  -run '^TestReceptionCreationGuard(ConcurrentTransactions|RepositoryBoundaries|IgnoresOtherDayClinicAndDeleted|BatchConflictRollsBackEarlierInsert)$'
```

独立2接続の並行作成、進行中候補と完了後の別来院、医院／患者／死亡の境界、tentative昇格、他日／他医院／削除行の除外、バッチの途中競合rollbackを確認する。最終coverage／全体型検査、Windows／STG／受付ボードE2Eは未確認のまま人間の受入へ引き継ぐ。

## 独立レビュー

React のフック・依存・a11y と TypeScript/API の最終 frontend 差分は指摘0件。初回レビューで見つかった既存予約の無条件更新と API の不完全応答受入は、版照合と unknown 境界検証で対応した。開始時刻テストは実 Radix に対する user-event と選択後の閉じ状態確認へ修正し、7件・lint・format が PASS。

SQL は既存値保持・医院分離・transaction と事後条件を静的レビュー。受付の backend はロック順、医院・患者・所有者、生存の positive match、CAS、0件409、同 transaction の再取得を静的レビューし、最終指摘0件。実DB測定とは区別する。型 hook は UNKNOWN、生成・全体型検査の代わりにテスト通過を使わない。

## 手動作業

1. EMR-270: STG で同一予約の担当者指定あり／なしを比較し、空き枠応答・開始／終了時刻を記録する。認証情報・患者情報は貼らない。
2. EMR-318: Windows で検索結果をスクロール。別途、飼主詳細→受付追加→既存／新規→受付ボード、権限不足・死亡・複数候補・競合409を確認する。
3. EMR-321: [ホテルとケージの runbook](../../ops/deploy/runbooks/EMR-321_HOTEL_RESERVATION.md) に従い、対象環境・医院別件数・同名 inactive／削除済み行を事前確認してからユーザーが SQL を実行する。fresh DB 用 CSV は使い捨て DB の正規 export 経路で手動再生成する。
4. API 定義の追加があるため、ユーザーが `make codegen` を実行し、生成物と必要な契約チェックを確認する。エージェントは生成処理を行わない。
5. migration を含む更新を取り込んだ開発者は、プロジェクト規則に従いアプリ使用前に `make migrate` を手動実行する。本 PR の live insert SQL は自動 migration 対象ではない。
6. マージと Done 遷移は人間の受入後に行う。
