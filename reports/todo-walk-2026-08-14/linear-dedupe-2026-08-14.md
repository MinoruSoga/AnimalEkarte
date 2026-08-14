# Linear 重複調査・統合 — 電カル（2026-08-14）

## 結論

| 項目 | 数 |
|------|-----|
| 整理前 Needs Human（ノア関連 概算） | ~32 |
| **整理後 Needs Human（ノア）** | **20** |
| 新規統合 | **BRT-68**（UAT 人間レーン） |
| Duplicate | BRT-38 · BRT-58〜65 |
| Blocked（メタ/終了） | BRT-7 · BRT-8 · BRT-66 |

## 実施した統合

| 操作 | 内容 |
|------|------|
| **BRT-37 拡充** | OPS-1 **#89+#97** 一体。タイトル更新 |
| **BRT-38 → Duplicate** | 正本 BRT-37 |
| **BRT-68 新規** | P1+H1〜H7 チェックリスト1枚 |
| **BRT-58〜65 → Duplicate** | 正本 BRT-68 |
| **BRT-66 → Blocked** | 送付メタ廃止。本体は BRT-39/47/49/55/56 + Fable §E |
| **BRT-7 → Blocked** | 旧 USER gate pack · 現行チケットに置換済 |
| **BRT-8 → Blocked** | bug.md loop 終了（Open 0） |
| **BRT-45 注記** | close ゲート専用。実施は BRT-68 |

## 残 Needs Human（20 · これ以上むやみに潰さない）

| Linear | 役割 | まとめない理由 |
|--------|------|----------------|
| [BRT-37](https://linear.app/baritechllc/issue/BRT-37) | credential ローテ | 実行系 USER |
| [BRT-39](https://linear.app/baritechllc/issue/BRT-39) | #201 臨床 | 臨床専権 |
| [BRT-40](https://linear.app/baritechllc/issue/BRT-40) | #211 | 臨床+OPS |
| [BRT-41](https://linear.app/baritechllc/issue/BRT-41) | #249 | 臨床 |
| [BRT-42](https://linear.app/baritechllc/issue/BRT-42) | #250 | 外部 |
| [BRT-43](https://linear.app/baritechllc/issue/BRT-43) | #252 | 会計投入 |
| [BRT-44](https://linear.app/baritechllc/issue/BRT-44) | #253 | PROD |
| [BRT-45](https://linear.app/baritechllc/issue/BRT-45) | #254 close | **判定ゲート**（実施≠close） |
| [BRT-46](https://linear.app/baritechllc/issue/BRT-46) | #255 | roster |
| [BRT-47](https://linear.app/baritechllc/issue/BRT-47) | #256 U13 | 1語判断 |
| [BRT-48](https://linear.app/baritechllc/issue/BRT-48) | #257 Go-live | gate |
| [BRT-49](https://linear.app/baritechllc/issue/BRT-49) | #258 | 契約 |
| [BRT-50](https://linear.app/baritechllc/issue/BRT-50) | #259 | 外部 |
| [BRT-51](https://linear.app/baritechllc/issue/BRT-51) | #261 | #201 後 hub |
| [BRT-52](https://linear.app/baritechllc/issue/BRT-52) | #284 | phase2 実機 |
| [BRT-55](https://linear.app/baritechllc/issue/BRT-55) | #299 | staging merge |
| [BRT-56](https://linear.app/baritechllc/issue/BRT-56) | PO-008 | クライアント |
| [BRT-57](https://linear.app/baritechllc/issue/BRT-57) | PO-10 | STG presence |
| [BRT-67](https://linear.app/baritechllc/issue/BRT-67) | OPS-13 | migrate USER |
| [BRT-68](https://linear.app/baritechllc/issue/BRT-68) | UAT 人間レーン | **実施束** |

## これ以上まとめなかったもの

- **GH 1 Issue = 1 Linear**（#201 と #261 等）— close 条件・担当が違う
- **BRT-45 と BRT-68** — close 判定と作業チェックリストを分離（Fable: local FAIL0 で閉じない）
- **BRT-55 と BRT-68 の H1** — H1 は #299 後依存として BRT-68 内チェックのまま残置

## 任意の次（さらに薄くするなら · 未実施）

| 案 | 効果 | リスク |
|----|------|--------|
| BRT-51 を BRT-39 の子だけにして NH から外す | -1 | hub 可視性↓ |
| BRT-57+67 を「STG OPS 束」に | -1 | 実行タイミングが違う |
| BRT-48 を BRT-44/45 の親ゲートに | 構造化 | 起票コスト |

現状 20 は **1 担当・1 成果物** 単位として妥当。

## 参照

- Fable: `reports/fable-needs-human-answer-2026-08-14.md`
- 依頼: `reports/fable-needs-human-request-2026-08-14.md`
