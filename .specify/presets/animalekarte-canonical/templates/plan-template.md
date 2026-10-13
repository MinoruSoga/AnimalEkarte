# Implementation Plan: [変更名]

**Change Reference**: [spec.md]
**Canonical Requirements**: [domain正本のFR-ID／GWTへのリンク]
**Status**: [計画状態・実機状態を区別]

## Technical Context
既存architecture・ADR・API契約・DB正本へリンクし、変更対象と制約だけを記す。臨床・会計・権限の未決事項を技術調査で勝手に決定しない。

## Constitution Check
参照口 `.specify/memory/constitution.md` を通じて既存正本の臨床安全・医院分離・Git安全・Docker限定検証・文書規約を確認する。

## Phase 0: Research and Blocking Questions
技術調査とユーザー合意が必要な業務の未決事項を分ける。未決事項の依存作業はBLOCKEDのまま。

## Phase 1: Design and Contracts
実装方式と既存正本/API/DB/受入条件への影響を記載。業務ルールの複製はしない。

## Verification Plan
| FR-ID／条件ID | 確認方法 | 対象環境 | 証跡 | 未確認の理由 |
|:---|:---|:---|:---|:---|
| [ID] | [限定検証] | [対象] | [予定／実施済みを区別] | [不足] |

## Phase 2: Task Handoff
tasks.mdに正本IDと条件IDを引き継ぐ。実装・公開は今回の計画作成だけでは許可されない。
