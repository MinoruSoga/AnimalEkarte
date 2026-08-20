# テスト・品質保証ディレクトリ (Testing & Quality Assurance)

> **目的**: システムの信頼性、パフォーマンス、およびコード品質を担保するための戦略と記録の管理。  
> **読者**: AI エージェント(Claude Code / Grok)。  
> **タイミング**: docs/ops/testing/ 配下編集時。

---

## 📂 ディレクトリ構成

ファイル索引の正本は [README.md](README.md)（二重管理を避けるため本書には索引を置かない）。

**必読順（受入タスク）**:

1. [TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md)
2. [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md) + `scripts/check-uat-env.sh`
3. [scenarios/README.md](scenarios/README.md)
4. フォームなら [scenarios/FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md) + [scenarios/FORM-FIELD-INVENTORY.md](scenarios/FORM-FIELD-INVENTORY.md)

---

## ✅ 品質基準の原則

1. **自動化の優先**: 回帰バグを即座に検知するため、臨床フローの主要パスを E2E でカバーする（詳細は E2E_TESTING_GUIDE）。
2. **受入は scenarios/**: 納品前の業務・フォーム項目単位の証明は L4（TEST_ARCHITECTURE）。E2E で代替しない。
3. **項目単位まで**: V シリーズ完了条件は inventory 全 fieldKey への F プロトコル適用。
4. **実データに近い検証**: seed `003_demo`（local）/ `004_staging`（STG）。
5. **証跡**: ローカル `reports/uat-YYYY-MM-DD/`（gitignore。コミットしない）。シナリオ md に結果を書かない。FAIL は Linear。
6. **秘密を書かない**: パスワード・トークンは env のみ。

---
