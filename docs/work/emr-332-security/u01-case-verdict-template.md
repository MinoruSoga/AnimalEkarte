# U01 — Shared Per-Case Verdict Record Template (ケース判定レコード雛形)

Campaign: emr-332-security-defense / Unit: U01-SAFETY-BASELINE / Attempt: att-u01-1
Purpose: EMR-332 の 18 ケース (C01-C18) すべてが同一の 12 フィールド構成で判定を記録するための共有雛形。各ケースユニット (U02-U07) はこのテンプレートを複写して使う。
Format source: ticket 本文「判定・証跡フォーマット」(ケースID／対象リビジョン／実施方式／正常条件結果／拒否条件結果／防御判定／保存・送信呼出し回数／観測・検出判定／安全な証跡参照／制約・未検証範囲／修復案／再検証結果)。

## フィールド定義 (12 fields — 固定)

| # | Field | 記録内容 |
|---|-------|----------|
| 1 | ケース ID (case id) | `C01`–`C18`。ケースをまたぐ記録や新規 ID の捏造はしない |
| 2 | 対象リビジョン (target revision) | 判定対象のコード/配信リビジョン (commit SHA、Worker バージョン等)。staging 実測では稼働リビジョン照合結果も記す (u01-safety-record §2) |
| 3 | 実施方式 (method) | `static` (静的レビュー) / `simulated` (メモリ・模擬・httptest) / `staging-observed` (実測) のいずれか1つ。複数方式を実施した場合は行を分け、強い方式への丸めをしない |
| 4 | 正常条件 (normal condition) | 許可されるべき入力・状態とその期待結果、実際の観測結果 |
| 5 | 拒否条件 (refusal condition) | 拒否されるべき入力・状態とその期待結果、実際の観測結果 |
| 6 | 防御判定 (defense verdict) | 防御の成立可否: `PASS` / `FAIL` / `SKIP` / `UNKNOWN` / `BLOCKED` / `N/A` (下記「判定値」) |
| 7 | 副作用呼出回数 (side-effect call count) | 保存・送信・後段処理の呼出し回数 (数値)。拒否ケースでは `0` を期待。数えられない場合は `UNKNOWN` と理由 |
| 8 | 検出判定 (detection verdict) | 観測・検出 (マーカー記録・ログ相関・通知) の成立可否: `PASS` / `FAIL` / `SKIP` / `UNKNOWN` / `BLOCKED` / `N/A`。**防御判定と別に記録する** |
| 9 | 安全な証跡参照 (safe evidence ref) | 証跡の参照先 (本キャンペーン内のファイルパス・行番号、テスト名、コマンド出力の保存先)。秘密・PII・実ペイロードを含まない参照のみ |
| 10 | 制約・未検証範囲 (constraints / unverified scope) | その判定がカバーしない範囲・前提・手段。未検証項目は隠さず `未検証` と明記する |
| 11 | 修復案 (remediation) | ギャップ確認時の最小修復案。証拠で確認できない案は「候補」と明記 (AC4) |
| 12 | 再検証結果 (re-verification) | 修復後に同じケース ID で再確認した結果 (正常許可・不正拒否・副作用ゼロ・検出の再確認と scoped コマンド)。未実施ならその旨 |

## 判定値 (verdict enum — ticket 本文より固定)

| 値 | 意味 |
|----|------|
| `PASS` | 記録した方式・範囲で期待結果と副作用ゼロを観測した |
| `FAIL` | 当該範囲で期待防御または検出が成立しない証拠がある |
| `SKIP` | 禁止条件等により未実施 |
| `UNKNOWN` | 証拠不足 |
| `BLOCKED` | 依存条件 (安全条件・テナント・アカウント・稼働リビジョン) が未確定 |
| `N/A` | 機能非該当の根拠あり (根拠を制約欄へ) |

## 記録ルール

- 防御 PASS と検出 PASS を分ける。マーカー表示・403 応答・設定項目の存在・模擬合格だけで staging 全経路を PASS にしない。静的確認は静的確認として記録する (ticket 本文)。
- method が `static`/`simulated` の場合、staging-observed での成立を主張しない。`staging-observed` は u01-safety-record の全前提 (稼働リビジョン照合・許可テナント・リクエスト上限・停止条件) を満たした場合のみ。
- 副作用呼出回数は DB に限らず session/監査ログ/キャッシュ/キュー/ファイル/メール/外部テレメトリを対象に数える。
- 証跡参照は safe evidence ref のみ — 実行可能な攻撃コード・実ペイロード・資格情報・実データを書かない。

## 記入例 (フィールド形態の例示 — 架空の記入値であり判定結果ではない)

```text
ケース ID:            C00 (例示)
対象リビジョン:        <commit-sha / staging 稼働 revision 照合結果>
実施方式 (method):     static | simulated | staging-observed
正常条件:              <期待> / <観測>
拒否条件:              <期待> / <観測>
防御判定:              PASS | FAIL | SKIP | UNKNOWN | BLOCKED | N/A
副作用呼出回数:         <n> (期待 0 の経路では実測値)
検出判定:              PASS | FAIL | SKIP | UNKNOWN | BLOCKED | N/A
安全な証跡参照:         <path:line / test 名 / コマンド出力先>
制約・未検証範囲:       <未検証とする範囲の明示>
修復案:                <確認済みギャップに基づく最小変更 or 候補>
再検証結果:            <同ケース ID での再確認結果 or 未実施>
```
