# docs PR #784 集約・レビュー引継ぎ台帳（2026-10-21）

本人の「docs変更は1PR、既存PRへ追加」という最新指示に基づき、#784を唯一の集約先とする。これは本文保全・文書差分集約の記録であり、レビュー対応完了・マージ可能・runtime完成の宣言ではない。試験はNOT_RUN。親の統合保全確認後、元10PR（#785–790・#792–795）へ保全先コメントを付けてclose済み・未マージをreadback確認した。branch/worktreeを保持し、旧レビューのresolved状態は変更しない。

## 固定した入力

各PRの最新headを下表へ固定し、mainとのmerge-baseからの変更がdocs配下だけであることを確認した。20文書のpath重複はない。元本文20文書は固定headとバイト一致する。残差10文書と本台帳を加えた集約差分もdocsのみ。#791の実装差分やコードcommitをcherry-pickしない。最新main e71ddf034（#791統合済み）は比較対象としてのみ読み、#790に必要な既存mainの証明書文書baseline差分だけを保持して文書patchを適用した。

| 元PR | 固定head | 文書数 | 全thread / 未解決 |
|---|---|---:|---:|
| [#784](https://github.com/MinoruSoga/AnimalEkarte/pull/784) | `5a1e00204457084b159346b765cf0002cc8ed164` | 1 | 16 / 3 |
| [#785](https://github.com/MinoruSoga/AnimalEkarte/pull/785) | `b1d281f6bee6cbf4928d051a2d04ced5868e4877` | 1 | 1 / 0 |
| [#786](https://github.com/MinoruSoga/AnimalEkarte/pull/786) | `2ffcd9d22cda71d5786f2bcfdf8dfc16424e6531` | 1 | 7 / 7 |
| [#787](https://github.com/MinoruSoga/AnimalEkarte/pull/787) | `e17d8e42e5fa6ba7972eac1ecaeb05aaf080c34a` | 1 | 10 / 9 |
| [#788](https://github.com/MinoruSoga/AnimalEkarte/pull/788) | `8b4cbbf73fd1c8ccb1abd20a7577a8dc4e3a6c9b` | 1 | 24 / 12 |
| [#789](https://github.com/MinoruSoga/AnimalEkarte/pull/789) | `d173c847148222e9405de0e7149b0aa65ceef541` | 1 | 3 / 3 |
| [#790](https://github.com/MinoruSoga/AnimalEkarte/pull/790) | `66b0b2ae85b5f94635b2382c31734f49065d727f` | 6 | 2 / 2 |
| [#792](https://github.com/MinoruSoga/AnimalEkarte/pull/792) | `d27fcbae3ba43efe8287b7732d2d793a50eb09e6` | 2 | 9 / 9 |
| [#793](https://github.com/MinoruSoga/AnimalEkarte/pull/793) | `fd7761ad9c0bc6e0a353b4bbed87d3df647adc43` | 3 | 0 / 0 |
| [#794](https://github.com/MinoruSoga/AnimalEkarte/pull/794) | `acd882e65a52becc345f869d439318765729eb82` | 2 | 0 / 0 |
| [#795](https://github.com/MinoruSoga/AnimalEkarte/pull/795) | `8b2ccde265e5147fc60445cad77395e7d55e8caf` | 1 | 0 / 0 |

## 本文の所在

| 元PR | 集約先path |
|---|---|
| #784 | [docs/work/emr/emr-503-504-selected-wire-contract-20261021.md](emr-503-504-selected-wire-contract-20261021.md) |
| #785 | [docs/ops/infra/production/runbook.md](../../ops/infra/production/runbook.md) |
| #786 | [docs/work/emr/emr-507-508-selected-wire-contract-20261021.md](emr-507-508-selected-wire-contract-20261021.md) |
| #787 | [docs/work/emr/emr-490-492-selected-wire-contract-20261021.md](emr-490-492-selected-wire-contract-20261021.md) |
| #788 | [docs/work/emr/emr-496-497-selected-wire-contract-20261021.md](emr-496-497-selected-wire-contract-20261021.md) |
| #789 | [docs/work/emr/emr-368-selected-author-attribution-contract-20261021.md](emr-368-selected-author-attribution-contract-20261021.md) |
| #790 | [docs/design/domains/checkups/technical-spec.md](../../design/domains/checkups/technical-spec.md) |
| #790 | [docs/design/domains/vaccinations/technical-spec.md](../../design/domains/vaccinations/technical-spec.md) |
| #790 | [docs/spec/domains/checkups/lifecycle.md](../../spec/domains/checkups/lifecycle.md) |
| #790 | [docs/spec/domains/medical-records/pages/medical-record-form.md](../../spec/domains/medical-records/pages/medical-record-form.md) |
| #790 | [docs/spec/domains/vaccinations/lifecycle.md](../../spec/domains/vaccinations/lifecycle.md) |
| #790 | [docs/work/emr/emr-486-certificate-api-contract-20261020.md](emr-486-certificate-api-contract-20261020.md) |
| #792 | [docs/work/emr/emr-495-consultation-correction-contract-20261021.md](emr-495-consultation-correction-contract-20261021.md) |
| #792 | [docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md](emr-516-517-plan-owner-link-contract-20261021.md) |
| #793 | [docs/work/emr/emr-458-option-assignment-selected-contract-20261021.md](emr-458-option-assignment-selected-contract-20261021.md) |
| #793 | [docs/work/emr/emr-480-booked-duration-selected-contract-20261021.md](emr-480-booked-duration-selected-contract-20261021.md) |
| #793 | [docs/work/emr/emr-493-staff-leave-selected-contract-20261021.md](emr-493-staff-leave-selected-contract-20261021.md) |
| #794 | [docs/work/emr/emr-482-trimming-price-source-selected-20261021.md](emr-482-trimming-price-source-selected-20261021.md) |
| #794 | [docs/work/emr/emr-483-price-acknowledgement-selected-20261021.md](emr-483-price-acknowledgement-selected-20261021.md) |
| #795 | [docs/work/emr/emr-485-storage-cas-contract-20261021.md](emr-485-storage-cas-contract-20261021.md) |

## 未commitだった正本残差10文書

入力patch SHA-256は `cd95362a5bedd221e54f4c4d98e5deec0f29f35b59665a35bc557384cadb84dc`。転送4部分を連結・HTML entity復元後に一致を確認し、本文変更を加えず適用した。元の未commit tree/原patchは保全したまま。trimming NULLはno-course専用、D03は494採用契約参照、旧セット移行不要を保持、CF-only復旧、LSTEP/DMのPhase1内作業順、通知再開のstart > server nowを同期する。写真の未確定採否は追加しない。
- [docs/delivery/MIGRATION_SCOPE.md](../../delivery/MIGRATION_SCOPE.md)
- [docs/design/domains/medical-records/technical-spec.md](../../design/domains/medical-records/technical-spec.md)
- [docs/design/domains/trimming/technical-spec.md](../../design/domains/trimming/technical-spec.md)
- [docs/ops/deploy/runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md](../../ops/deploy/runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md)
- [docs/ops/infra/architecture.md](../../ops/infra/architecture.md)
- [docs/ops/testing/scenarios/README.md](../../ops/testing/scenarios/README.md)
- [docs/ops/testing/scenarios/V05-auth-line-forms.md](../../ops/testing/scenarios/V05-auth-line-forms.md)
- [docs/spec/domains/SPECIFICATION_PROGRESS.md](../../spec/domains/SPECIFICATION_PROGRESS.md)
- [docs/spec/line/README.md](../../spec/line/README.md)
- [docs/spec/line/lstep-integration.md](../../spec/line/lstep-integration.md)

## 既存レビューと返信への引継ぎ

固定取得時点で72スレッド、うち未解決45件（784:3、785:0、786:7、787:9、788:12、789:3、790:2、792:9、793–795:0）。親側の3担当が判定・修正案を準備中であり、ここでは既存本文を保全し原指摘と返信を追跡する。原スレッドが解決済みでも集約後の再検証完了を意味しない。未解決は未解決として引き継ぎ、本文取り込みだけを理由にresolveしない。新しいレビューはこの固定一覧とは別に集約PRで扱う。

| 元PR | 原指摘 | 元状態 | 対象path | 既存返信 |
|---|---|---|---|---|
| #784 | [4332795852](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332795852) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332905846](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332905846) |
| #784 | [4332795853](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332795853) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332906527](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332906527) |
| #784 | [4332795854](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332795854) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332907187](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332907187) |
| #784 | [4332795859](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332795859) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332908043](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332908043) |
| #784 | [4332795862](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332795862) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332908733](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332908733) |
| #784 | [4332795869](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332795869) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332909359](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332909359) |
| #784 | [4332795876](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332795876) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332910078](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332910078) |
| #784 | [4332795887](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332795887) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332910875](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332910875) |
| #784 | [4332795892](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332795892) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332911484](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332911484) |
| #784 | [4332938262](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332938262) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4332956303](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332956303) |
| #784 | [4332982320](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332982320) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4333011698](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4333011698) |
| #784 | [4332982324](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332982324) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4333247492](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4333247492) |
| #784 | [4332982327](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4332982327) | 原スレッド解決済み | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | [4333249235](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4333249235) |
| #784 | [4333034886](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4333034886) | 未解決・引継ぎ | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | なし |
| #784 | [4333052761](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4333052761) | 未解決・引継ぎ | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | なし |
| #784 | [4333052762](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4333052762) | 未解決・引継ぎ | `docs/work/emr/emr-503-504-selected-wire-contract-20261021.md` | なし |
| #785 | [4332802710](https://github.com/MinoruSoga/AnimalEkarte/pull/785#discussion_r4332802710) | 原スレッド解決済み | `docs/ops/infra/production/runbook.md` | [4332815486](https://github.com/MinoruSoga/AnimalEkarte/pull/785#discussion_r4332815486) |
| #786 | [4332998766](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998766) | 未解決・引継ぎ | `docs/work/emr/emr-507-508-selected-wire-contract-20261021.md` | なし |
| #786 | [4332998776](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998776) | 未解決・引継ぎ | `docs/work/emr/emr-507-508-selected-wire-contract-20261021.md` | なし |
| #786 | [4332998783](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998783) | 未解決・引継ぎ | `docs/work/emr/emr-507-508-selected-wire-contract-20261021.md` | なし |
| #786 | [4332998790](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998790) | 未解決・引継ぎ | `docs/work/emr/emr-507-508-selected-wire-contract-20261021.md` | なし |
| #786 | [4332998796](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998796) | 未解決・引継ぎ | `docs/work/emr/emr-507-508-selected-wire-contract-20261021.md` | なし |
| #786 | [4332998801](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998801) | 未解決・引継ぎ | `docs/work/emr/emr-507-508-selected-wire-contract-20261021.md` | なし |
| #786 | [4332998805](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998805) | 未解決・引継ぎ | `docs/work/emr/emr-507-508-selected-wire-contract-20261021.md` | なし |
| #787 | [4332980580](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4332980580) | 原スレッド解決済み | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | [4333005377](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333005377) / [4333023651](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333023651) |
| #787 | [4332996978](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4332996978) | 未解決・引継ぎ | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | なし |
| #787 | [4332996979](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4332996979) | 未解決・引継ぎ | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | なし |
| #787 | [4332996982](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4332996982) | 未解決・引継ぎ | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | なし |
| #787 | [4332996984](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4332996984) | 未解決・引継ぎ | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | なし |
| #787 | [4333039061](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039061) | 未解決・引継ぎ | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | なし |
| #787 | [4333039067](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039067) | 未解決・引継ぎ | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | なし |
| #787 | [4333039070](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039070) | 未解決・引継ぎ | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | なし |
| #787 | [4333039073](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039073) | 未解決・引継ぎ | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | なし |
| #787 | [4333039077](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039077) | 未解決・引継ぎ | `docs/work/emr/emr-490-492-selected-wire-contract-20261021.md` | なし |
| #788 | [4332982448](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4332982448) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333049352](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333049352) / [4333052110](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333052110) |
| #788 | [4332982451](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4332982451) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333049675](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333049675) / [4333090556](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333090556) |
| #788 | [4332982458](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4332982458) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333050049](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333050049) / [4333085037](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333085037) |
| #788 | [4332987632](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4332987632) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333050507](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333050507) |
| #788 | [4332987636](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4332987636) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333050879](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333050879) |
| #788 | [4332987644](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4332987644) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333051245](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333051245) |
| #788 | [4332987653](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4332987653) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333051615](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333051615) |
| #788 | [4332987662](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4332987662) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333052081](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333052081) |
| #788 | [4333035535](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333035535) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333241557](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333241557) / [4333264954](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333264954) |
| #788 | [4333075253](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333075253) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333283663](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333283663) / [4333332514](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333332514) |
| #788 | [4333075257](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333075257) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333284052](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333284052) / [4333328563](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333328563) |
| #788 | [4333075262](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333075262) | 原スレッド解決済み | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | [4333284387](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333284387) / [4333320389](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333320389) |
| #788 | [4333278069](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278069) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333278078](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278078) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333278083](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278083) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333278087](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278087) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333278092](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278092) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333310878](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333310878) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333323628](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323628) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333323635](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323635) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333323640](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323640) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333323646](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323646) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333323652](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323652) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #788 | [4333323657](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323657) | 未解決・引継ぎ | `docs/work/emr/emr-496-497-selected-wire-contract-20261021.md` | なし |
| #789 | [4333333724](https://github.com/MinoruSoga/AnimalEkarte/pull/789#discussion_r4333333724) | 未解決・引継ぎ | `docs/work/emr/emr-368-selected-author-attribution-contract-20261021.md` | なし |
| #789 | [4333333732](https://github.com/MinoruSoga/AnimalEkarte/pull/789#discussion_r4333333732) | 未解決・引継ぎ | `docs/work/emr/emr-368-selected-author-attribution-contract-20261021.md` | なし |
| #789 | [4333333735](https://github.com/MinoruSoga/AnimalEkarte/pull/789#discussion_r4333333735) | 未解決・引継ぎ | `docs/work/emr/emr-368-selected-author-attribution-contract-20261021.md` | なし |
| #790 | [4333359305](https://github.com/MinoruSoga/AnimalEkarte/pull/790#discussion_r4333359305) | 未解決・引継ぎ | `docs/design/domains/checkups/technical-spec.md` | なし |
| #790 | [4333359307](https://github.com/MinoruSoga/AnimalEkarte/pull/790#discussion_r4333359307) | 未解決・引継ぎ | `docs/spec/domains/vaccinations/lifecycle.md` | なし |
| #792 | [4333431638](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333431638) | 未解決・引継ぎ | `docs/work/emr/emr-495-consultation-correction-contract-20261021.md` | なし |
| #792 | [4333431649](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333431649) | 未解決・引継ぎ | `docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md` | なし |
| #792 | [4333432895](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432895) | 未解決・引継ぎ | `docs/work/emr/emr-495-consultation-correction-contract-20261021.md` | なし |
| #792 | [4333432900](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432900) | 未解決・引継ぎ | `docs/work/emr/emr-495-consultation-correction-contract-20261021.md` | なし |
| #792 | [4333432905](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432905) | 未解決・引継ぎ | `docs/work/emr/emr-495-consultation-correction-contract-20261021.md` | なし |
| #792 | [4333432909](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432909) | 未解決・引継ぎ | `docs/work/emr/emr-495-consultation-correction-contract-20261021.md` | なし |
| #792 | [4333432913](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432913) | 未解決・引継ぎ | `docs/work/emr/emr-495-consultation-correction-contract-20261021.md` | なし |
| #792 | [4333432916](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432916) | 未解決・引継ぎ | `docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md` | なし |
| #792 | [4333432919](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432919) | 未解決・引継ぎ | `docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md` | なし |

コード/API/DDL/設定/資格/DB/実送信/試験/マージは本集約で実行しない。本文中の未実装依存・未確認根拠・NOT_RUNを保持する。

## 固定45件の文書対応結果（旧thread状態は未変更）

親の独立判定を反映し、44件を文書修正、4333039070は旧DDL前提不成立と記録する。以下は固定commitの本文根拠であり、runtime検証・レビュー返信／resolve完了ではない。新規レビューや後続owner実装は別途残る。

| 元PR・指摘 | 文書対応／判断 | 固定修正根拠 |
|---|---|---|
| [#784 / 4333034886](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4333034886) | dispatch開始CASとreport親lock。開始済み外部効果の回収保証は追加しない | [本文L79](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-503-504-selected-wire-contract-20261021.md#L79) |
| [#784 / 4333052761](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4333052761) | claim/start/resultと各必須監査を同TX | [本文L83](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-503-504-selected-wire-contract-20261021.md#L83) |
| [#784 / 4333052762](https://github.com/MinoruSoga/AnimalEkarte/pull/784#discussion_r4333052762) | 未開始失効CAS、開始済みはprovider終結＋不存在証拠の両方 | [本文L87](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-503-504-selected-wire-contract-20261021.md#L87) |
| [#786 / 4332998766](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998766) | invoice専用番号生成・初回固定 | [本文L167](https://github.com/MinoruSoga/AnimalEkarte/blob/ff03c02deaf3ed9701a5fb660c95a192fd547535/docs/work/emr/emr-507-508-selected-wire-contract-20261021.md#L167) |
| [#786 / 4332998776](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998776) | 原finalize/finance event＋Asia/Tokyo暦日 | [本文L173](https://github.com/MinoruSoga/AnimalEkarte/blob/ff03c02deaf3ed9701a5fb660c95a192fd547535/docs/work/emr/emr-507-508-selected-wire-contract-20261021.md#L173) |
| [#786 / 4332998783](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998783) | 旧CRUD status迂回を同sliceで閉鎖 | [本文L181](https://github.com/MinoruSoga/AnimalEkarte/blob/ff03c02deaf3ed9701a5fb660c95a192fd547535/docs/work/emr/emr-507-508-selected-wire-contract-20261021.md#L181) |
| [#786 / 4332998790](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998790) | current計算＋correct after参照。過去replayは別 | [本文L187](https://github.com/MinoruSoga/AnimalEkarte/blob/ff03c02deaf3ed9701a5fb660c95a192fd547535/docs/work/emr/emr-507-508-selected-wire-contract-20261021.md#L187) |
| [#786 / 4332998796](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998796) | title/comment固定header snapshot | [本文L193](https://github.com/MinoruSoga/AnimalEkarte/blob/ff03c02deaf3ed9701a5fb660c95a192fd547535/docs/work/emr/emr-507-508-selected-wire-contract-20261021.md#L193) |
| [#786 / 4332998801](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998801) | billing単位edition発見read | [本文L199](https://github.com/MinoruSoga/AnimalEkarte/blob/ff03c02deaf3ed9701a5fb660c95a192fd547535/docs/work/emr/emr-507-508-selected-wire-contract-20261021.md#L199) |
| [#786 / 4332998805](https://github.com/MinoruSoga/AnimalEkarte/pull/786#discussion_r4332998805) | 提示版に対応するnullable reply | [本文L205](https://github.com/MinoruSoga/AnimalEkarte/blob/ff03c02deaf3ed9701a5fb660c95a192fd547535/docs/work/emr/emr-507-508-selected-wire-contract-20261021.md#L205) |
| [#787 / 4332996978](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4332996978) | parent_idはallowlist外、既存階層を再解釈しない | [本文L95](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L95) |
| [#787 / 4332996979](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4332996979) | 現C06でprivate capture可能とはせずRequest VerifiedCommand限定deltaを明記 | [本文L219](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L219) |
| [#787 / 4332996982](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4332996982) | 旧LINE write無作用閉鎖＋canonical FE移行 | [本文L101](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L101) |
| [#787 / 4332996984](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4332996984) | group選択createは期待版1件、groupなし空集合 | [本文L159](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L159) |
| [#787 / 4333039061](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039061) | create-only本人に限定receipt回復、本文view昇格なし | [本文L186](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L186) |
| [#787 / 4333039067](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039067) | 全local ownerのbinding/候補索引coverage未検証ならactivation保留 | [本文L103](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L103) |
| [#787 / 4333039070](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039070) | 不成立：最終001 L10864でreservation_types.category/colorはDROP済み。旧category defaultを修復する前提は採用せず、新列も復活しない | [本文L99](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L99) |
| [#787 / 4333039073](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039073) | usage_resume_pending=false明示、unknown既存値はbackfillしない | [本文L99](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L99) |
| [#787 / 4333039077](https://github.com/MinoruSoga/AnimalEkarte/pull/787#discussion_r4333039077) | 同TX名称revision1＋receipt name_revision | [本文L99](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L99) |
| [#788 / 4333278069](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278069) | clinic-scoped copy provenance、private本人baselineを公開しない | [本文L301](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L301) |
| [#788 / 4333278078](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278078) | 現在有効staff_clinic_assignmentsのscope/lock検証 | [本文L314](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L314) |
| [#788 / 4333278083](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278083) | source group＋全entry childでlossless保存 | [本文L306](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L306) |
| [#788 / 4333278087](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278087) | branchごとclosed hash projection | [本文L309](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L309) |
| [#788 / 4333278092](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333278092) | current presenceでabsent表現、旧scalar保持 | [本文L301](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L301) |
| [#788 / 4333310878](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333310878) | 重複指摘：entry配列のgroup/child cardinalityを同補正で閉じる | [本文L306](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L306) |
| [#788 / 4333323628](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323628) | Request VerifiedCommand限定delta、未実装を明記 | [本文L318](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L318) |
| [#788 / 4333323635](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323635) | 現在tenure session/epoch一致 | [本文L60](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L60) |
| [#788 / 4333323640](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323640) | 版別effective_display_year、過去unknown非補完 | [本文L324](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L324) |
| [#788 / 4333323646](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323646) | inquiry refは作成前後ともrecord ID | [本文L275](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L275) |
| [#788 / 4333323652](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323652) | 当時主訴名/basis固定、現masterで過去補完なし | [本文L314](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L314) |
| [#788 / 4333323657](https://github.com/MinoruSoga/AnimalEkarte/pull/788#discussion_r4333323657) | principal-qualified unique/lookup | [本文L202](https://github.com/MinoruSoga/AnimalEkarte/blob/d81904989d084234c4e9543c77d3072437ce26e7/docs/work/emr/emr-496-497-selected-wire-contract-20261021.md#L202) |
| [#789 / 4333333724](https://github.com/MinoruSoga/AnimalEkarte/pull/789#discussion_r4333333724) | ownerごとのclosed保存値before/after、一般trimなし | [本文L87](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-368-selected-author-attribution-contract-20261021.md#L87) |
| [#789 / 4333333732](https://github.com/MinoruSoga/AnimalEkarte/pull/789#discussion_r4333333732) | 4tupleで版を含めdedup | [本文L136](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-368-selected-author-attribution-contract-20261021.md#L136) |
| [#789 / 4333333735](https://github.com/MinoruSoga/AnimalEkarte/pull/789#discussion_r4333333735) | graph advisory/bindingだけpre-actor、business rowはactor後 | [本文L89](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-368-selected-author-attribution-contract-20261021.md#L89) |
| [#790 / 4333359305](https://github.com/MinoruSoga/AnimalEkarte/pull/790#discussion_r4333359305) | 文書確認日2026-10-21 | [本文L10](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/design/domains/checkups/technical-spec.md#L10) |
| [#790 / 4333359307](https://github.com/MinoruSoga/AnimalEkarte/pull/790#discussion_r4333359307) | normal独立欄は当初値default、既知RV欄新設なし、RV4mapping資料待ち | [本文L203](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/spec/domains/vaccinations/lifecycle.md#L203) |
| [#792 / 4333431638](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333431638) | 了承snapshotはcommand、live CAS/fenceはPreconditions。提案digestへ結合しCASをhashしない | [本文L155](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-495-consultation-correction-contract-20261021.md#L155) |
| [#792 / 4333431649](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333431649) | 既存raw JSON1MiB上限をparse/deep-copy/hash/claim前へ。新500文字業務制限なし | [本文L196](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md#L196) |
| [#792 / 4333432895](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432895) | 既存target nonNULL拒否、通常編集又は実source C0/C1へ。架空source作成なし | [本文L169](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-495-consultation-correction-contract-20261021.md#L169) |
| [#792 / 4333432900](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432900) | owner-private耐久request/commandと同claim TX、現在本文で再採取しない | [本文L153](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-495-consultation-correction-contract-20261021.md#L153) |
| [#792 / 4333432905](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432905) | 既存494担当準備済みcohortに限定、GET/endでbaseline生成なし・受付へedit追加なし | [本文L171](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-495-consultation-correction-contract-20261021.md#L171) |
| [#792 / 4333432909](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432909) | 適格actual_visit_factsを訂正・receipt同outer TX、単なる転記は来院事実へ変換なし | [本文L175](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-495-consultation-correction-contract-20261021.md#L175) |
| [#792 / 4333432913](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432913) | 選択先owner明示不存在CAS・同TX初期作成 | [本文L173](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-495-consultation-correction-contract-20261021.md#L173) |
| [#792 / 4333432916](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432916) | 要求どおりreceiptへgeneration追加せず、旧新treatment tupleから既存484 GETを別read。treatment-onlyを他ownerへ拡張しない | [本文L192](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md#L192) |
| [#792 / 4333432919](https://github.com/MinoruSoga/AnimalEkarte/pull/792#discussion_r4333432919) | 既存idを保持しclosed owner_reference/plan_revision_id等を加算、非空fixtureを固定 | [本文L188](https://github.com/MinoruSoga/AnimalEkarte/blob/86640ed4f5fe232ab464b77dcf07031e4ae67adc/docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md#L188) |

## 追加受領稿と照合

すべてEND後に連結し、HTML entity復元後の原文SHA-256一致を確認した。原稿とpatchはrepo外へ保全し、旧2稿の重複追加はしない。回復・PDF・RVの冒頭だけ、親指示による採択済み技術契約／runtime未実装へ最小整合した。既存checkups/D08技術文書と486旧契約は今回選定範囲への参照を同期し、v1・見本・provider・実環境の未完を保持する。

| 受領物 | 原文SHA-256 | 反映commit |
|---|---|---|
| [481](emr-481-source-binding-contract-20261021.md) | `0bee471992cf02df8f3ca8e07efd494a79be82a9e4f2ccfbdec30dddac23d82c` | `ff03c02` |
| [474/487回復](emr-474-487-operation-recovery-contract-20261021.md) | `2d57276cef89af282dbe2a3ed6fac4e5b97bdf4cac4e9595afa748b8c440a056` | `ff03c02` |
| 786修正patch | `7d9a8454b6348a2c53ee5d851dbdb5164cb7dfb6aa637e12fd4381f314831eca` | `ff03c02` |
| 787/788修正patch | `fbb6d1b95602867ca1a798892a264f53220d55dcf04cc4cf55350d20ac83f3a8` | `d819049` |
| [510統合全文](emr-510-inventory-selected-contract-20261021.md) | `c007a4bc15d3ba77d5fadc8c30a860974956ee2c08eff711372dd0e1c8b2efa4` | `4025f7d` |
| 固定残り17件patch | `8ffc51f26927e6eb5ea62cfbb363ae9be2fdb01f43fa559aeedfa58ac43a0cd4` | `86640ed` |
| [500PDF](emr-500-outsourced-checkup-pdf-contract-20261021.md) | `57c8308e8d858b3a58803c73f59d9118b023fe7535027bacd677c51a49224595` | `194fe12` |
| [486/487 RV R3](emr-486-487-rv-batch-reissue-contract-20261021.md) | `78f83ca3650d4bef7c49ca9ed79f75d5611f016456de32ae287206b4d7238693` | `6660db7` |

490/492適用後全文SHAは `061eb1f2adad8ab30828a4a125acf44e7fb6af81664f496c2e92e623aba9955e` で親指定値と一致。496/497適用後全文SHAは `dd655f6da74acb75498576cd5e9c9fa020e599536f51f16a90218a0c40619f43`。patch末尾の空行contextだけ既存EOFへ合わせ、本文の推測補完なしで置換対象を全件一致検査した。

PDF §2のC06到達後HTTP mappingは親追加指示により `12702fe` でnot-succeeded=200／operation-in-progress=202／edit-right-lost=409へ訂正した。前段validationや媒体/サイズ/認証errorは変更していない。原稿SHAはこの公開時整合前の受領値である。

試験・fixture・DB・provider・PDF実処理・実送信はNOT_RUN。文書対応をもってレビューthread完了やマージ可能とは宣言しない。

## 横断契約の接続補正（2026-10-21 06:34 UTC親採択）

[Q共通節](../../design/shared/operation-participation.md#verified-command-selected-20261021)へRequest.VerifiedCommandの全文を集約し、指定consumerは同じQからApply入力を再構成する。owner-private耐久commandとは別で、既存Content-only/v1・CAS/fence・短Contentの契約は変更しない。489の限定Request/Runner deltaは未実装依存のまま。481/516/517/RVの認可anchor欠落・不整合・解決不能は403非開示、元targetと現在開示認可確定後の結果読取失敗だけ503へ整合した。510はschema_version文字列とedit-right-lost409を明記。PDFの公開済み3状態HTTP訂正は再適用せず保持。504 Cは未受領のため、この補正batchには含めない。実装・試験・DB・mergeなし。

### 補助author B：指定単一commitの取込

入力 `6b57254f69e16b224d2321a41da1cf7cb6671e98`（親base `e71ddf03462a0ee7a9864ad2ce6854b243591770`）は以下3docs・362行追加だけであることを確認し、`09397634e` へ単一cherry-pickした。補助branch全体やそのbaseのruntime変更は取り込まない。3文書とも元commitのfile blobとバイト一致、共同receipt/source/runtime未完の境界を保持する。

- [364/375/376 CSV列契約](emr-364-375-376-csv-column-contract-20261021.md)
- [365価格小数契約](emr-365-vaccine-price-decimal-contract-20261021.md)
- [371四区分運用手順](emr-371-four-category-operating-procedure-20261021.md)

Q横断接続は `0421039e4`、安定参照と記録は `767d5724d`。追加文書を含む静的照合は40docs限定・参照341件のpath/anchor欠落なし、git diff --check通過。試験はNOT_RUN。504 Cは引き続き原稿未受領で、別担当の到着後に指定補正を適用する。

### 補助author C：指定単一commitと504 C待ち分の完了

入力 `584884f12b41c81e9740e6a48065537076b5308f`（base `e71ddf034`）は[297機器接続](emr-297-device-mapping-bridge-contract-20261021.md)・[504 C通知契約](emr-504-c-dispatch-adapter-contract-20261021.md)の2docs・329行追加のみを確認し、`e493255c9`へ単一cherry-pickした。取込直後の2file blobは元commitとバイト一致、branch/base runtimeを混入しない。続けて親のQ1/R1保留分を504 Cへ適用し、共通Qへの接続と認可anchorの403／認可確定後の結果読取故障503を明記した。これにより前段台帳の「504 C未受領」は当時状態となり、予定追加docsの受領・取込は完了。新19指摘は親の固定snapshot判定待ちで、旧thread返信・resolveは行わない。実装・試験・DB・LLM・送信・mergeなし。
