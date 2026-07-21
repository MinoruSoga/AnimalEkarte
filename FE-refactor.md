# FE-refactor 第11期（FE11）— DESIGN.md 完全遵守（仕様項目 × ページ の二軸チェック）

> 起票: 2026-07-21（要件責任者: 曽我。指示=「DESIGN.md に完全遵守させて欲しい」— 3度目の明示）
> **本ファイルは対応後削除する使い捨てトラッカー**。恒久規約は `docs/spec/design-system.md`。
> **前提**: FE10（構造レイヤーの字義化）は完遂済み — brand `#0075DE` 反転・hairline・radius・テーブル `ex-data-table-cell`・brand フィル pill・focus/active/checked=primary・audit C1〜C11 恒久ガード。フルゲート4種緑（type-check / test 2210 / build / lint）。

## なぜ FE10 で「完全」に届かなかったか（方法論の欠陥・本期の設計理由）

FE10 の判定軸 P1〜P7 は **画面を見て違和感がないか**の視覚チェックリストで、しかも `design-system.md`（製品翻訳版）から起こした。**DESIGN.md の仕様表を1行ずつ突合する軸が無かった**ため、画面を見ても気付けない数値仕様（letter-spacing・line-height・font-size 1px 差）と、色相が近く見える ink ランプが検査対象外のまま残った。
→ 本期は **DESIGN.md の仕様表そのものをチェックリスト化（G軸）** し、その上でページ毎に検証（P軸）する。

## 決裁事項

- **ink ランプ = 選択肢A（字義値へ集約）** を採用。ユーザー「完全遵守」指示に従う。実装の14段アルファ（`/90`〜`/15`・実使用 1664 箇所）を DESIGN.md の4段実値へ畳む。
- **唯一の例外は据え置き**: §2.4 臨床 semantic 色（danger/死亡/RBAC 非活性/warning）。臨床安全（SPECIFICATION 2.1）が全原則に優先。
- **G1-c/G6-d は裁定待ち**（下記「要裁定」節）。それ以外は字義が一意なので裁定不要で実施する。

---

# Part 1 — G軸: DESIGN.md 仕様項目チェックリスト（実測値ベース）

判定: ✅ 字義一致 / ❌ 不一致 / △ 未定義（DESIGN.md にあるが実装に無い） / 🔒 例外（裁定済み）

## G1. colors（フロントマター20値）

| # | トークン | DESIGN.md | 実装の現状（実測） | 判定 |
|---|---|---|---|---|
| G1-1 | primary | `#0075DE` | `PALETTE.brand` / `--primary` | ✅ |
| G1-2 | primary-active | `#005BAB` | `PALETTE.brandHover` | ✅ |
| G1-3 | on-primary | `#FFFFFF` | `--primary-foreground` | ✅ |
| G1-4 | canvas / surface | `#FFFFFF` | `C.bgWhite` / `--card` | ✅ |
| G1-5 | canvas-soft | `#F6F5F4` | `C.bgPage` / `--background` | ✅ |
| G1-6 | hairline | `#E6E6E6` | `C.borderLight` / `--border` | ✅ |
| G1-7 | ink | `#000000` | `C.text` | ✅ |
| G1-8 | **ink-secondary** | `#31302E` | **実装に0箇所**。`text-[#000000]/80`等の黒アルファで代用 | ❌ |
| G1-9 | **ink-muted** | `#615D59` | `--muted-foreground` のみ字義値。`C.text60/55/50` は黒アルファ | ❌ |
| G1-10 | **ink-faint** | `#A39E98` | **実装に0箇所**。`C.text40`等の黒アルファで代用 | ❌ |
| G1-11 | **secondary** | `#213183` | **実装に0箇所**（hero-band 未使用のため） | △ |
| G1-12〜19 | **sticker palette 8色**<br>sky `#62AEF0` / purple `#D6B6F6` / purple-deep `#391C57` / pink `#FF64C8` / orange `#DD5B00` / orange-deep `#793400` / teal `#2A9D99` / green `#1AAE39` / brown `#523410` | 装飾専用 | **全て実装に0箇所**。装飾/status は独自色（`#6940A5` purple・`#D9730D` orange・`bg-blue-500` 等） | △ **要裁定 G1-c** |

## G2. typography（11ロール × size/weight/line-height/letter-spacing）

**ブラウザ実測値**（`getComputedStyle` で各ロールの実クラスを測定・2026-07-21）:

| # | ロール | DESIGN.md | 実測 | 判定 |
|---|---|---|---|---|
| G2-1 | title | 20px / 600 / 1.4 / **−0.125px** | 20 / 600 / 1.4 / **normal** | ❌ tracking |
| G2-2 | body-md | 16px / 400 / 1.5 / 0 | 16 / 400 / 1.5 / normal | ✅ |
| G2-3 | body-sm | 15px / 400 / **1.33** / 0 | 15 / 400 / **1.429** / normal | ❌ line-height |
| G2-4 | caption | 14px / 400 / **1.43** / 0 | 14 / 400 / **1.333** / normal | ❌ line-height |
| G2-5 | eyebrow | 12px / 600 / **1.33** / **+0.125px** | 12 / 600 / **1.45** / **+0.3px** | ❌ line-height・tracking |
| G2-6 | button | **16px** / 500 / 1.5 / 0 | **15px** / 500 / 1.429 / normal | ❌ size・line-height |
| G2-7 | heading-3 | 22 / 700 / 1.27 / −0.25px | 未定義（アプリ未使用） | △ |
| G2-8 | heading-2 | 26 / 700 / 1.23 / −0.625px | 未定義（アプリ未使用） | △ |
| G2-9 | heading-1 | 40 / 700 / 1.1 / −1px | 未定義（アプリ未使用） | △ |
| G2-10 | display-2 | 54 / 700 / 1.04 / −1.875px | 未定義（アプリ未使用） | △ |
| G2-11 | display-1 | 64 / 700 / 1.0 / −2.125px | 未定義（アプリ未使用） | △ |
| G2-12 | font-family | Inter 代替可（DESIGN.md 明記） | `'Inter','Noto Sans JP'` | ✅ |

> **body-sm と caption の line-height が入れ替わっている**のは、Tailwind の `--text-*--line-height` 既定が元サイズ基準のまま残り、size だけ上書きしたため。

## G3〜G5（形状・余白・エレベーション）

| # | 項目 | DESIGN.md | 実装 | 判定 |
|---|---|---|---|---|
| G3 | rounded xs4/sm5/md8/lg12/xl16/full | 6段 | `--radius-*` 全段一致（xxs は xs へ整列） | ✅ |
| G4 | spacing 4/8/12/16/24/28/32 | 7段 | Tailwind スケールで全段表現可 | ✅ |
| G5 | elevation Level0/1/2 | hairline のみ / 4段極薄 / 5段 | `--shadow-level1` `--shadow-level2` + Level0=hairline | ✅ |

## G6. components（14 + ex-* 10）

| # | 項目 | DESIGN.md | 実装の現状 | 判定 |
|---|---|---|---|---|
| G6-1 | button-primary | bg primary / on-primary / **button 16px** / pill | pill ✅ ・**ラベル 15px** | ❌（G2-6 と同根） |
| G6-2 | button-primary-pressed | bg primary-active | `hover:bg-primary/90`・`PALETTE.brandHover` | ✅ |
| G6-3 | button-secondary | white / ink / pill / Level-1 | 二次 CTA は white+pill | ✅ |
| G6-4 | button-utility | white / rounded-md / **padding 4px 14px** / hairline border | rounded-md ✅ ・padding は h-11 系で別値 | ❌ padding |
| G6-5 | button-icon-circular | `rgba(0,0,0,0.05)` fill / full | アイコンボタンは透明+hover tint | ❌ fill 未一致 |
| G6-6 | text-input | white / body-sm / **padding 6px** / rounded-xs / border `#DDDDDD` | rounded-xs ✅ ・border ✅ ・**padding は h-11 系** | ❌ padding |
| G6-7 | badge-pill | white surface / **primary text** / eyebrow / pill / 4px 8px | `BADGE.*` は status 色ベース | ❌（要裁定 G6-d） |
| G6-8 | feature-card | white / rounded-lg / **padding 24px** / Level-0 | `STYLE.formCard` = rounded-lg + p-6(24px) | ✅ |
| G6-9 | ex-data-table-cell | canvas-soft header / eyebrow / body-sm / hairline row | FE10 で全テーブル字義化 | ✅ |
| G6-10 | ex-modal-card | rounded-xl / Level-2 / padding 24px | shadcn Dialog = rounded-xl + p-6 + shadow | ✅ |
| G6-11 | nav-bar | **canvas(白)** / ink / body-sm / 16px | サイドバーは `C.bgPage`(canvas-soft) | ❌（要裁定 G6-d） |
| G6-12 | hero-band / footer | marketing 専用 | アプリ本体に該当面なし | — |

---

# Part 2 — P軸: ページ毎 × チェックリスト毎（84ルート）

## 判定軸

```
[FE10] P1図地 / P2境界 / P3階層 / P4余白 / P5アクセント / P6テーブル / P7状態
       → FE10 で全ページ判定済み。本期は G軸修正後の【回帰のみ】確認する。
[新設] P8  ink   : 本文/補助テキストが DESIGN.md ink 4段（#000000/#31302E/#615D59/#A39E98）で描画される
       P9  字送り: 見出し(title)に −0.125px、eyebrow に +0.125px が効いている
       P10 タイポ: ロール別 size/line-height が G2 表と一致（特に button 16px・body-sm 1.33・caption 1.43・eyebrow 1.33）
```

判定: ✅ 準拠 / 🔧 修正した / ⚠️ 裁定要 / — 該当なし / （空欄）未着手

## 進捗表

| Batch | ルート | FE10<br>P1-P7 | P8<br>ink | P9<br>字送り | P10<br>タイポ | 所見 | 状態 |
|---|---|---|---|---|---|---|---|
| A | /login | ✅ | | | | | 未 |
| A | /forgot-password | ✅ | | | | BUG-419（未認証到達不能）は別件 | 未 |
| A | /reset-password | ✅ | | | | | 未 |
| A | /（受付カンバン） | ✅ | | | | | 未 |
| A | /owners（一覧） | ✅ | | | | | 未 |
| A | /owners/new | ✅ | | | | | 未 |
| A | /owners/:id | ✅ | | | | | 未 |
| A | /owners/:id/report | ⚠ | | | | FE10 から継続: brand tick の装飾使用疑い＋別ワークストリーム編集中 | ⚠保留 |
| A | /aggregation | ✅ | | | | | 未 |
| B1 | /reservations | ✅ | | | | | 未 |
| B1 | 予約詳細モーダル | ✅ | | | | | 未 |
| B1 | 予約登録モーダル | ✅ | | | | | 未 |
| B2 | /hospitalization | ✅ | | | | | 未 |
| B2 | /hospitalization/select-pet | ✅ | | | | | 未 |
| B2 | /hospitalization/new | ✅ | | | | | 未 |
| B2 | /hospitalization/:id | 静✅ | | | | 全4医院で入院0件（API実測）＝描画不能 | ⏸データ待ち |
| B2 | /hospitalization/:id/edit | 静✅ | | | | 同上 | ⏸データ待ち |
| B3 | /medical-records | ✅ | | | | | 未 |
| B3 | /medical-records/select-pet | ✅ | | | | | 未 |
| B3 | /medical-records/new | ✅ | | | | | 未 |
| B3 | /medical-records/:id | ✅ | | | | | 未 |
| B4 | /accounting（3タブ） | ✅ | | | | | 未 |
| B4 | /accounting/select-pet | ✅ | | | | | 未 |
| B4 | /accounting/new | ✅ | | | | | 未 |
| B4 | /accounting/:id | ✅ | | | | | 未 |
| B4 | /accounting/close | ✅ | | | | | 未 |
| B4 | /accounting/close/history | ✅ | | | | | 未 |
| B4 | /accounting/reports | ✅ | | | | | 未 |
| B4 | /estimates | ✅ | | | | | 未 |
| B4 | /estimates/new | ✅ | | | | | 未 |
| B4 | /estimates/:id | ✅ | | | | | 未 |
| B4 | /estimates/:id/edit | ✅ | | | | | 未 |
| B5 | /settings | ✅ | | | | | 未 |
| B5 | /settings/clinic | ✅ | | | | | 未 |
| B5 | /settings/staff | ✅ | | | | | 未 |
| B5 | /settings/treatment-items | ✅ | | | | | 未 |
| B5 | /settings/diagnosis | ✅ | | | | | 未 |
| B5 | /settings/animal-species | ✅ | | | | | 未 |
| B5 | /settings/trimming | ✅ | | | | | 未 |
| B5 | /settings/trimming-course-type | ✅ | | | | | 未 |
| B5 | /settings/medicine | ✅ | | | | | 未 |
| B5 | /settings/reservation-type | ✅ | | | | | 未 |
| B5 | /settings/hospitalization | ✅ | | | | | 未 |
| B5 | /settings/cage | ✅ | | | | | 未 |
| B5 | /settings/merchandise-items | ✅ | | | | | 未 |
| B5 | /settings/insurance | ✅ | | | | | 未 |
| B5 | /settings/occupations | ✅ | | | | | 未 |
| B5 | /settings/permission-groups | ✅ | | | | RBAC 非活性表現の退行を特に確認（C6a） | 未 |
| B5 | /settings/inquiry-templates | ✅ | | | | | 未 |
| B5 | /settings/interview/chief-complaint | ✅ | | | | | 未 |
| B5 | /settings/interview/templates | ✅ | | | | | 未 |
| B5 | /settings/shift-templates | ✅ | | | | | 未 |
| B5 | /settings/closing-time | ✅ | | | | | 未 |
| B5 | /settings/payment-methods | ✅ | | | | | 未 |
| B5 | /settings/campaigns | ✅ | | | | | 未 |
| B5 | /settings/integrations/lstep | ✅ | | | | | 未 |
| B5 | /settings/lstep/tags | ✅ | | | | | 未 |
| B6 | /inventory | ✅ | | | | | 未 |
| B6 | /inventory/new | ✅ | | | | | 未 |
| B6 | /inventory/:id | ✅ | | | | | 未 |
| B6 | /trimming | ✅ | | | | | 未 |
| B6 | /trimming/select-pet | ✅ | | | | | 未 |
| B6 | /trimming/new | ✅ | | | | | 未 |
| B6 | /trimming/:id | ✅ | | | | | 未 |
| B6 | /vaccinations | ✅ | | | | | 未 |
| B6 | /vaccinations/select-pet | ✅ | | | | | 未 |
| B6 | /vaccinations/new | ✅ | | | | | 未 |
| B6 | /vaccinations/:id | ✅ | | | | | 未 |
| B6 | /checkups | ✅ | | | | | 未 |
| B6 | /checkups/select-pet | ✅ | | | | | 未 |
| B6 | /checkups/new | ✅ | | | | | 未 |
| B6 | /examinations | ✅ | | | | | 未 |
| B6 | /examinations/select-pet | ✅ | | | | | 未 |
| B6 | /examinations/new | ✅ | | | | | 未 |
| B6 | /examinations/:id | ✅ | | | | | 未 |
| B7 | /shifts | ✅ | | | | | 未 |
| B7 | /lstep/checkup-sync | ✅ | | | | | 未 |
| B7 | /lstep/delivery-monitor | ✅ | | | | | 未 |
| B7 | /lstep/analytics | ✅ | | | | | 未 |
| B7 | /line-reservation/settings | ✅ | | | | | 未 |
| B7 | /line-reservation/page-editor | ✅ | | | | | 未 |
| B7 | /line-reservation/slots | ✅ | | | | | 未 |
| B7 | /manual | ✅ | | | | | 未 |
| B7 | /manual/:category/:slug | ✅ | | | | | 未 |

（404 fallback・リダイレクト専用12 route は対象外 — `ui-design-compliance.md` §2 脚注どおり）

---

# 実行順（G軸→P軸。G軸はトークン修正なので全ページへ一括波及する）

| 順 | フェーズ | 内容 | 対象 G# | 状態 |
|---|---|---|---|---|
| 1 | **F1 タイポ数値** | `--text-sm/xs/2xs--line-height` 定義・button 16px 化・tracking トークン新設と title/eyebrow への適用 | G2-1,3,4,5,6 / G6-1 | 未 |
| 2 | **F2 ink ランプ** | 14段アルファ → DESIGN.md 4段実値へ集約（明度最近傍で写像）。**臨床非活性(disabled/RBAC)の可読退行を必ず実測** | G1-8,9,10 | 未 |
| 3 | **F3 コンポーネント寸法** | text-input padding 6px・button-utility padding 4px/14px・button-icon-circular fill | G6-4,5,6 | 未 |
| 4 | **F4 未定義トークンの定義** | secondary・sticker palette 8色・display/heading 系をトークンとして定義（アプリ未使用でも語彙を揃える） | G1-11〜19 / G2-7〜11 | 未 |
| 5 | **F5 機械ガード追加** | 黒アルファ text 直値の再混入を audit で禁止（C12）・tracking/line-height の退行検知 | — | 未 |
| 6 | **F6 文書同期** | `design-system.md` §2.6 の **Ink 行の虚偽 ✅ を訂正**・§3.1 に実装値を反映 | — | 未 |
| 7 | **P軸 全84ページ検証** | 上記波及後に P8/P9/P10 を全ページ実測（+ P1〜P7 の回帰確認） | — | 未 |

---

# 要裁定（着手前にユーザー判断が必要な2件）

## G1-c: スティッカーパレット8色をどう扱うか

DESIGN.md は装飾色として8色を規定するが、実装は独自の status 色（`#6940A5` purple・`#D9730D` orange・`bg-blue-500`・emerald 等）を使っており**1色も一致しない**。ただしこれらは「装飾」ではなく**業務ステータスの識別子**として機能しており、スタッフは色で状態を読んでいる。

- **案A（字義優先）**: status 色を DESIGN.md の8色へ全面置換。→ 完全遵守だが**現場の色記憶を壊す**。
- **案B（現状維持＋明文化）**: status 色は §2.4 semantic と同じく「臨床/業務運用のための例外」として文書化。→ DESIGN.md との差分は残るが、業務影響ゼロ。
- **推奨 = B**。DESIGN.md の sticker palette は「イラスト・カテゴリドット」という*装飾*用途の規定であり、業務状態の識別に色を使う本システムの用途とは前提が異なるため。

## G6-d: nav-bar（サイドバー）と badge-pill

- nav-bar: DESIGN.md は白 canvas を規定、実装は canvas-soft。→ 白にすると本文カードとの図地が消える（P1 と衝突）。**推奨 = 現状維持＋文書化**。
- badge-pill: DESIGN.md は「白地＋primary 文字」、実装は status 色ベース。→ G1-c と同じ論点。**推奨 = G1-c の裁定に追従**。

---

# 検証規約（FE10 の事故を踏まえた強化版）

- 各フェーズ後: `node scripts/design-system-audit.mjs` 緑 + 影響 `npx vitest run <path>`。
- **コミット前に touched file 全数へ scoped eslint 必須**（`docker compose exec -T frontend sh -c 'xargs npx eslint' < filelist`）。vitest も design-audit も**型を見ない**ため、これが唯一の scoped な型崩れ検出手段（FE10 で TS6133 を2回すり抜けた実績）。
- **トークン値を変えたら、その旧値を握るテストを全数 grep**（置換式そのものを grep パターンへ機械変換する。手で部分集合を書くと漏れる — FE10 で `2383E2` を落とした実績）。
- **値統合の副作用点検**: 統合したトークンの消費者を「構造用途/装飾用途」で仕分ける（FE10 で受付済ドットが構造色の装飾使用になった実績）。
- フル lint/type-check/build/test は最終ゲートで実行。
