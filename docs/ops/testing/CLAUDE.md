# テスト・品質保証ディレクトリ (Testing & Quality Assurance)

> **目的**: テスト戦略・受入手順・品質記録を安全に管理する。
> **読者**: AI エージェント、開発者、QA。

## 必読順（受入タスク）

1. [TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md)
2. [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md)
3. [scenarios/README.md](scenarios/README.md)
4. フォーム受入では [FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md) と [FORM-FIELD-INVENTORY.md](scenarios/FORM-FIELD-INVENTORY.md)

ファイル索引は [README.md](README.md) を正本とする。層、環境境界、記録方針は `TEST_ARCHITECTURE.md` を正本とし、ここへ複製しない。

## エージェント向け必須事項

- migration seed は全環境で `002_master` のみ。アカウントと臨床 fixture は含まれない。local は承認済み handoff/import、STG は承認済み UAT provisioning lane を使う。手順は [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md) を参照する。
- L4 受入を E2E で代替しない。inventory には wildcard・`要実測` が残るため、列挙済み項目だけを機械的に網羅済みと扱う。
- production または未承認の共有 clinic で、作成・更新・削除・外部送信を行わない。
- 実行結果は gitignore 対象の `reports/uat-YYYY-MM-DD/` に置き、scenario 本文へ書かない。
- 確認済み UAT FAIL は `bug.md` で重複確認・記録してから Linear で追跡する。その他の新規製品欠陥は通常の Linear intake に従う。環境 BLOCKED は `bug.md` に書かない。
- secret、password、token、cookie、idToken をリポジトリ、文書、chat、ログへ書かない。秘密の保管方法は [liff-verification.md](liff-verification.md) の境界に従う。
