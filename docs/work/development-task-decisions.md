# 開発タスクの着手判断

判断基準日: 2026-09-11。ソース基準: `main` / `f0e238f10830cbc8cf12a059bf518c207556f614` と、着手前から存在したWorker WIP。

依頼者: 曽我 稔。目的は、実装対象のない保守方針と前提未確定の仮説を開発キューから分け、次の担当者が対象・変更内容を判断し直さず着手できるようにすること。技術判断は今回の依頼に基づきCodexが行う。

本書は判断の根拠を保存する資料であり、実行タスクの正本は [todo.md](../../todo.md#development-tasks)。測定・テスト・受入は [todo-verification.md](../../todo-verification.md)、環境操作は [todo-operations.md](../../todo-operations.md) を参照する。Linearへの読取は `TASK-444` / `BE-RC` / `PERF-STG-LOGIN` をTeam・Project指定で試みたが、`USER_NOT_LOGGED_IN` で失敗した。以下はローカルの技術判断であり、LinearのDone/Cancelled更新ではない。

## 性能3件

| 旧ID | 判断 | 現在の根拠 | 再判定の管理先 |
|---|---|---|---|
| PERF-DEV-OBSERVATION | 現行WIP不採用。常時ログ追加を開発キューに置かない | [Worker](../../backend/worker/index.ts) は全proxy要求の `container.fetch` 前後だけを計測。ブラウザーで観測した最終GET接続前22.5秒やContainer起動イベントの内訳は得られない | `PERF-V-CLIENT-TRACE` / `PERF-V-CF-EVENTS` / `PERF-V-DECIDE-OBSERVATION` |
| PERF-DEV-MITIGATION | 原因未確定の設定・通信経路変更は採用しない | [性能記録](../../todo-performance.md) のE1と再現しなかったE2だけでは、edge OPTIONS・sleep延長・CORS変更のいずれも選べない | `PERF-V-MITIGATION`。因果が特定できた場合に対象を固定して開発キューへ戻す |
| PERF-DEV-BUNDLE | 根拠のないchunk再編は採用しない | [Vite設定](../../frontend/vite.config.ts) にcharts/LIFFのchunk分類があり、ローカルdistにもpreloadがある。ただしdist生成revisionはUNKNOWNで、現行SHAのtransfer/parse/execute寄与は未測定 | `PERF-V-BUNDLE`。測定で必要とされた依存だけを開発対象にする |

既存のpending表示は [router.tsx](../../frontend/src/app/router.tsx)、[app-routes.tsx](../../frontend/src/app/routes/app-routes.tsx)、[AuthProvider.tsx](../../frontend/src/features/auth/components/AuthProvider.tsx) にある。同じ実装を未完了として起票しない。STG配信や実測改善はこのソース確認からは判定できない。

### Worker runnerの訂正

[Makefile](../../Makefile) の `test-worker` は既にrootをDockerへmountし、指定したWorkerテストを実行する。`scripts/check-test-worker-makefile.test.sh` の8項目は今回PASSした。したがって「Docker runnerを新設しないと開始できない」は誤りで、runner新設タスクは不要。

この確認はWIPテストの実行PASSではない。既存runnerは型検査を含まず、[tsconfig.test.json](../../backend/worker/tsconfig.test.json) に新規 `index.test.ts` は明示されていない。将来の観測採用時の検証条件は統合検証TODOに保存する。現在のWorker WIPの削除・実装・commit・deployは行っていない。

<a id="task-444"></a>

## TASK-444

**採用する最小範囲は、既存生成型を使ったペット送信キーの固定。** [pet.ts](../../frontend/src/types/pet.ts) の `PetWritable` はモデルから除外列を引くだけなので、現在もモデルにある `version` / `deceased_at` / `deceased_reason` を入力型へ取り込む。[モデル型](../../frontend/src/types/generated/models.ts) と [Go request](../../backend/internal/pet/pet_request.go) は別契約であり、当該3キーは通常の作成・更新requestにはない。バックエンドで任意フィールドを書けるという指摘ではなく、FEの入力型が実際のAPIより広いという保守上の問題である。

ここでは現在の送信フィールドを `Pick` の許可リストへ固定し、モデル追加列の自動混入を止める。型名、enum、`name_kana`、danger_reasonのtri-stateと死亡専用経路を維持する。既存生成物を手編集せず完結でき、codegen待ちは不要である。

公開response型は [pet-responses.ts](../../frontend/src/types/generated/pet-responses.ts)、[medicalrecord-responses.ts](../../frontend/src/types/generated/medicalrecord-responses.ts) などへ既に分離されている。全 `generated/models` importの一括移行やResource定数の移動は採用しない。import件数を減らすこと自体を目的にしない。

補助調査で挙がったaddendumのresponse型移行は、[Go response DTO](../../backend/internal/medicalrecord/medical_record_addendum_response.go) が [tygo設定](../../backend/tygo.yaml) の対象に入っておらず、user-run codegenが必要になる。今回の即時着手単位には混ぜず、生成経路の変更を伴う別範囲として記録する。上記の送信キー固定でTASK-444全体の型移行が完了したとは扱わない。

## バックエンド6件

| 旧ID | 判断 | 現在の根拠と固定した範囲 |
|---|---|---|
| BE-RC-005 | 旧横断タスクを終了し、保守制約として維持 | `724cbf3fc` で返す5xxの二重ログ削除が入っている。[httpapi/response.go](../../backend/internal/httpapi/response.go) は5xxをGinへ登録する一方、[owner/service_core.go](../../backend/internal/owner/service_core.go) の通知失敗や [clinic/clinic_service.go](../../backend/internal/clinic/clinic_service.go) のbest-effort経路にはserviceログが必要。新しい具体的重複経路がないまま一括削除しない。全ログ経路の重複ゼロを証明したという意味ではない |
| BE-RC-009 | READYの具体的利用側へ限定 | [liff_service.go](../../backend/internal/reservation/liff_service.go) は不可時間・可能時刻の各4メソッドrepositoryを要求するが、[liff_service_availability.go](../../backend/internal/reservation/liff_service_availability.go) が使うのは `FindAll` だけ。2依存を利用側1メソッドinterfaceへ変更できる。他domainのinterface一括分割は不要 |
| BE-RC-014 | 現行依存に対する置換は不採用。upstream更新時の保守条件 | [apperrors/errors.go](../../backend/internal/apperrors/errors.go) はPgErrorを既に `errors.As` で処理し、client encode/rangeだけを既存3メッセージに限定。[go.mod](../../backend/go.mod) のpgx v5.10.0の `pgtype.newEncodeError` は `fmt.Errorf` を使う。存在しないtyped EncodeErrorへ置換しない。needle追加や依存upgradeをこの残件のために行わない |
| BE-RC-015 | 命名の維持制約へ変更 | 同じdomain内に複数能力があり、`HospitalizationService` などは対象能力を識別する名前。[Go規約](../../.claude/rules/go-gin-backend-guidelines.md) の新規命名基準を適用し、公開名を機械的一括renameする独立タスクは作らない |
| BE-RC-017 | READYの残存公開mapへ限定 | 多くのrepositoryは `724cbf3fc` で型付き化済みだが、[owner/repository.go](../../backend/internal/owner/repository.go) の `UpdateAndFind` と `OwnerUpdateApplier` は公開mapを残す。owner内のtyped command化を採用し、既存のrow lock・割引再判定・原子的reloadを維持する |
| BE-RC-019 | 現時点のpackage追加分割は不採用 | [ADR-006](../architecture/adr/006-backend-domain-package-boundaries.md) のmedicalrecord能力群に入院と検査が属し、[hospitalization_service.go](../../backend/internal/medicalrecord/hospitalization_service.go) は治療計画と同一transactionを持つ。新しい独立consumerやwrite ownerの要求がないままlab/hospitalizationを別packageへ移す根拠はない |

既存型やpackageを維持する判断は、必要な業務変更時の見直しを禁止するものではない。新しい具体的なconsumer・不具合・変更対象ができた時に、その最小範囲を定義して開発キューへ追加する。

## 調査の実施範囲

親はmainでソース・履歴・既存差分を照合した。補助調査は `f0e238f10` の隔離worktreeでread-only実施。TASK-444のaddendum案は親が生成物と設定を照合し、今回の着手単位には不採用とした。補助エージェントの後続調査は利用上限で終了したため、バックエンド6件と代替TASK-444の最終判断は親が実コードから行った。独立レビューPASSとは扱わない。

## 文書の引き継ぎ

調査と文書更新を完了し、main上の未コミット差分として引き継ぐ。実装・実DB検証・配信は今回の完了範囲ではない。文書リンク、3件のREADY行、旧10IDの判断先、検証先、既存WIPのhashを確認した。アプリコードの変更はなく、runtime検証は不要だった。

この調査だけで作成した `claim/LEDGER-TODO-READY`、`claim/TASK-444`、`claim/BE-RC-005`、`claim/BE-RC-009`、`claim/BE-RC-014`、`claim/BE-RC-015`、`claim/BE-RC-017`、`claim/BE-RC-019`、`claim/PERF-DEV-OBSERVATION`、`claim/PERF-DEV-MITIGATION`、`claim/PERF-DEV-BUNDLE` は、文書の保全と調査担当の終了後に解放する。READY3件の実装を開始する担当者は、それぞれのclaimを新規取得する。これは実装済みの宣言ではない。
