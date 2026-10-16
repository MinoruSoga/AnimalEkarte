# C04 認証・権限・医院scope

> 2026-10-15の仕様整理候補。PR #670の基準版から既存合意・既存仕様・未承認案を分離した。業務合意と実装／実機検証を別管理する。D22/D23は2026-10-15の一括承認を反映。技術設計案は業務承認と分ける。

## 目的・範囲

P1：全業務が本人・現在の権限・対象医院を分離して扱うための共通契約。実権限付与・credential発行・実環境設定は対象外。

## 正本と優先順

- [docs/spec/screens/21-login.md](../screens/21-login.md)
- [docs/spec/screens/settings/master-permission-group.md](../screens/settings/master-permission-group.md)
- [docs/architecture/auth.md](../../architecture/auth.md)
- [docs/architecture/adr/002-multitenancy-clinic-id-isolation.md](../../architecture/adr/002-multitenancy-clinic-id-isolation.md)
- [SECURITY.md](../../../SECURITY.md)

最新の明示合意が適用範囲内の旧記述に優先する。コードは実装事実の根拠で、業務承認の代替ではない。既存FR／ACの本文は元正本に維持し、本書の接続要件で改番・複製しない。

## 業務要件

<a id="requirement-fr-access-control-identity-session"></a>
### FR-access-control-identity-session — 認証と回復

1人1アカウントを基本とし、本人のログイン・ログアウト・パスワード再設定を扱う。未登録メールへの再設定申請も存在を明かさない同じ応答とし、期限切れ・使用済みtokenを再利用しない。認証済みであることから対象業務の権限を推定しない。

決定状態・根拠：既存auth・ログイン仕様継承。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-access-control-independent-authority"></a>
### FR-access-control-independent-authority — 独立した認可判定

resource/actionの許可、現在の医院所属、対象データの所有医院を独立に確認する。複数groupはOR合算、許可なしは拒否。職種名で権限を与えない。system_adminも有効な医院scopeの制約を免れない。明示された横断READ・共有master・support例外は個別境界を守り、全業務へ一般化しない。

決定状態・根拠：既存auth・SECURITY継承。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-access-control-revocation-recovery"></a>
### FR-access-control-revocation-recovery — 権限失効・医院切替

サーバーは現在のauthorityを再解決し取得障害はfail closed。UIの古いメニュー表示は最終権限ではない。選択医院失効時はwriteを停止して利用可能な医院へ復旧するが、失敗したwriteを別医院へ自動再送しない。有効医院なし・未認証・通常拒否・通信障害を区別する。

決定状態・根拠：既存auth §4.2,4.8,4.9継承。runtimeキャッシュの有効性は別検証。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-access-control-admin-self-lockout"></a>
### FR-access-control-admin-self-lockout — 管理権限変更の保護

非system_adminの自分への変更後、選択医院でmaster-permission:view/editを維持できない変更は拒否する。別groupのORが残れば許可。他の管理者がいることだけで例外にしない。資格情報・必須監査の失敗時は成功にせず変更を巻き戻す。

決定状態・根拠：既存auth §2.1,4.3継承。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

## 受入条件

全条件は文書上の期待結果。実機未実施。技術方式の提案は実装設計として区別する。

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-access-control-01 | 登録済み／未登録メール各1件 | 再設定を申請 | 画面応答で登録有無を露出しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-identity-session](#requirement-fr-access-control-identity-session)。実機未実施 |
| AC-access-control-02 | Agroup拒否・Bgroup許可 | 同じ操作を試す | 他の医院scope等を満たすならOR許可 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-independent-authority](#requirement-fr-access-control-independent-authority)。実機未実施 |
| AC-access-control-03 | 許可groupなし | 直接APIで操作 | 拒否されデータが不変 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-independent-authority](#requirement-fr-access-control-independent-authority)。実機未実施 |
| AC-access-control-04 | 選択A院の所属が解除された | A院でwriteし医院復旧する | A院writeを拒否しB院へ自動再送しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-revocation-recovery](#requirement-fr-access-control-revocation-recovery)。実機未実施 |
| AC-access-control-05 | 権限lookup障害 | 保護対象を取得・更新 | 古いtokenの医院一覧を根拠に許可しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-revocation-recovery](#requirement-fr-access-control-revocation-recovery)。実機未実施 |
| AC-access-control-06 | 非system_adminの自身に最後の管理view/editがある | その権限を削除 | 変更全体を拒否し既存管理権限を保持 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-admin-self-lockout](#requirement-fr-access-control-admin-self-lockout)。実機未実施 |
| AC-access-control-07 | 必須監査保存が失敗する | 資格情報を変更 | 資格情報と成功監査が片方だけ確定しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-admin-self-lockout](#requirement-fr-access-control-admin-self-lockout)。実機未実施 |
| AC-access-control-08 | 有効な所属医院がなくなった | ヘッダーなしmeで復旧 | 有効医院なしと示し管理者確認・ログアウトへ案内。通常403や通信障害と混同しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-revocation-recovery](#requirement-fr-access-control-revocation-recovery)。実機未実施 |
| AC-access-control-09 | 医院復旧中にログアウト | 遅延した復旧応答が届く | 医院保存・reloadでログアウトを取り消さない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-revocation-recovery](#requirement-fr-access-control-revocation-recovery)。実機未実施 |
| AC-access-control-10 | 期限切れまたは使用済み再設定リンク | 新パスワードを確定 | 拒否し成功したと表示しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-identity-session](#requirement-fr-access-control-identity-session)。実機未実施 |
| AC-access-control-11 | 自己groupを削除しても別groupにview/editが残る | 自己groupを変更 | OR後の権限を評価し自分のgroupという理由だけでは拒否しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-access-control-admin-self-lockout](#requirement-fr-access-control-admin-self-lockout)。実機未実施 |

## 未決・責務境界

業務上の新規grant・権限配布の承認は行わない。各D固有の取消／訂正／確定条件はそのowner正本へ委ねる。全routeの実装適合・権限実設定・cache有効設定・並行性・実機検証は未確認。実装事実のtoken寿命やDB構造はarchitecture正本を参照し複製しない。今回合意したD22公開版の認証職員READ・D23送信前確認を例外一覧へ同期する。

## レビュー状態

基準版の本文と関連する静的根拠を照合。独立レビュー・統合リンク検査は統合担当で実施する。仕様記述を実装完了、受入PASS、実環境変更承認へ変換しない。

## 操作・状態・例外の全域点検

- ログイン／refresh／logout／忘れたpassword申請／再設定／ログイン後変更／管理者reset／既存職員へのaccount追加は異なる経路。公開auth routeと保護staff routeを一律resource gateで扱わない。
- 再設定リンクは有効期限・単回使用を守る。未登録メール応答から存在を漏らさない。レート制限はRetry-Afterと再試行契約に従い、無期限連打をしない。メールの実到達・設定は未検証。
- password変更・所属解除・職員無効化後の次requestで現在authorityを評価。実環境でcacheが有効な場合のTTL遅延はarchitecture正本の既知条件であり、全環境即時失効と誤記しない。
- 別端末のUIは権限を古く表示し得るがAPIの最終認可は別。UI同期未完を認可迂回の理由にしない。
- 院切替中の保存は元の対象医院・入力を拘束し、旧画面の入力を新医院に自動保存しない。復旧後のGETと失敗writeの再送を分離。未保存離脱確認の共通UI方式はC01、並行保存・監査はC05へ接続する。
- group作成／metadata変更／rules変更／有効無効／並替／削除と職員割当変更を区別。いずれも現在の対象grant・scope、自己ロックアウト保護を確認する。実行時の権限表は既存architecture／permission modelが正本。
- Supportの共有READ・医院mutation・個人history、manualの全員公開READ案、global company、LIFF本人認証、横断一覧は明示例外。例外を全GETや全clinicへの無制限アクセスへ一般化しない。
- 旧移行staffの複数行は権威identity mapで扱う。氏名一致だけの統合や自己申告メールによる特権付与はしない。最後のsystem_admin／自己無効化の既存保護を緩和しない。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../domains/SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../domains/owners-pets/lifecycle.md#common-owner-pet-20261017)。共通基本情報・確認済み自院利用と、院別の診療資源認可を区別する。医院所属・操作権限・患者照合・本人LINE認証・区画別公開は維持し、他院資源の無権限操作／公開を許可しない。旧院別患者ID検査と共通患者対応は設計で接続する。
