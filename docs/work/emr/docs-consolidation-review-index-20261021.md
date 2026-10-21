# docs PR #784 集約・レビュー引継ぎ台帳（2026-10-21）

本人の「docs変更は1PR、既存PRへ追加」という最新指示に基づき、#784を唯一の集約先とする。これは本文保全・文書差分集約の記録であり、レビュー対応完了・マージ可能・runtime完成の宣言ではない。試験はNOT_RUN。元PRは親の統合確認までclose/deleteせず、旧レビューのresolved状態を変更しない。

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
