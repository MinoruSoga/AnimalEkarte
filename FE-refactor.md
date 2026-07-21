# FE-refactor 第11期（FE11）— DESIGN.md 完全遵守（仕様項目 × ページ の二軸チェック）

> 起票: 2026-07-21（要件責任者: 曽我。指示=「DESIGN.md に完全遵守させて欲しい」— 3度目の明示）
> **本ファイルは対応後削除する使い捨てトラッカー**。恒久規約は `docs/spec/design-system.md`。
> **前提**: FE10（構造レイヤーの字義化）は完遂済み — brand `#0075DE` 反転・hairline・radius・テーブル `ex-data-table-cell`・brand フィル pill・focus/active/checked=primary・audit C1〜C11 恒久ガード。フルゲート4種緑（type-check / test 2210 / build / lint）。

## なぜ FE10 で「完全」に届かなかったか（方法論の欠陥・本期の設計理由）

FE10 の判定軸 P1〜P7 は **画面を見て違和感がないか**の視覚チェックリストで、しかも `design-system.md`（製品翻訳版）から起こした。**DESIGN.md の仕様表を1行ずつ突合する軸が無かった**ため、画面を見ても気付けない数値仕様（letter-spacing・line-height・font-size 1px 差）と、色相が近く見える ink ランプが検査対象外のまま残った。
→ 本期は **DESIGN.md の仕様表そのものをチェックリスト化（G軸）** し、その上でページ毎に検証（P軸）する。

## 決裁事項

- **正本を軸で分割（FE11 確定・2026-07-21 曽我）**:
  - **色 = `docs/spec/design-system.md`（製品判断）**。本システムの色は装飾ではなく業務・臨床の意味論（ステータス識別・危険/死亡・RBAC 非活性）を担うため、装飾用途を前提とする DESIGN.md の色規定をそのまま当てても遵守にならない。
  - **タイポ / 形状 / 余白 / エレベーション / コンポーネント寸法 = DESIGN.md 字義**。業務意味論を持たないため値をそのまま適用する。
- **色軸の確定事項**（design-system.md へ明文化済み）:
  - brand `#0075DE` は**製品決定として据え置き確定**（FE10 の反転根拠を「DESIGN.md 字義」→「製品採用」へ置換。値は不変・再スイープ不要）
  - DESIGN.md sticker palette 8色は**不採用**。業務ステータス色を維持
  - nav-bar は **canvas-soft を維持**（白にすると図地関係が消えるため）
  - ink ランプは DESIGN.md と**同値を採用**（可読性の階層であり製品要件を満たすため一致・F2 実装済み）
- **臨床 semantic 色は据え置き**: §2.4（danger/死亡/RBAC 非活性/warning）。臨床安全（SPECIFICATION 2.1）が全原則に優先。

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
| G1-8 | ink-secondary | `#31302E` | `C.text90/80/70` = `#31302E`（F2 で写像） | ✅ F2 |
| G1-9 | ink-muted | `#615D59` | `C.text65/60/55/50` = `#615D59`（実測: フォームラベル一致） | ✅ F2 |
| G1-10 | ink-faint | `#A39E98` | `C.text45`以下 + placeholder = `#A39E98` | ✅ F2 |
| G1-11 | secondary `#213183` | hero-band 用 | 実装に0箇所（hero-band 未使用） | 🔒 不採用確定（色=本書正本・FE11） |
| G1-12〜19 | sticker palette 8色 | 装飾専用 | 業務ステータス色（`#6940A5` 紫・`#D9730D` 橙・blue-500・emerald 等）を使用 | 🔒 **不採用確定**（FE11 裁定: DESIGN.md の8色は*装飾*規定であり、本システムの色は*業務識別子*。design-system.md §2.5 に明文化済み） |

## G2. typography（11ロール × size/weight/line-height/letter-spacing）

**ブラウザ実測値**（`getComputedStyle` で各ロールの実クラスを測定・2026-07-21）:

| # | ロール | DESIGN.md | 実測 | 判定 |
|---|---|---|---|---|
| G2-1 | title | 20px / 600 / 1.4 / −0.125px | 20 / 600 / 1.4 / −0.125px | ✅ F1 |
| G2-2 | body-md | 16px / 400 / 1.5 / 0 | 16 / 400 / 1.5 / normal | ✅ |
| G2-3 | body-sm | 15px / 400 / 1.33 / 0 | 15 / 400 / 1.33 / 0 | ✅ F1 |
| G2-4 | caption | 14px / 400 / 1.43 / 0 | 14 / 400 / 1.43 / 0 | ✅ F1 |
| G2-5 | eyebrow | 12px / 600 / 1.33 / +0.125px | 12 / 600 / 1.33 / +0.125px | ✅ F1 |
| G2-6 | button | 16px / 500 / 1.5 / 0 | 16 / 500 / 1.5 / 0 | ✅ F1 |
| G2-7 | heading-3 | 22 / 700 / 1.27 / −0.25px | `--text-heading-3` 定義済み・旧 24px の描画箇所を写像 | ✅ F4 |
| G2-8 | heading-2 | 26 / 700 / 1.23 / −0.625px | `--text-heading-2` 定義済み・旧 30px の描画箇所を写像 | ✅ F4 |
| G2-9 | heading-1 | 40 / 700 / 1.1 / −1px | `--text-heading-1` 定義済み・旧 36px の描画箇所を写像 | ✅ F4 |
| G2-10 | display-2 | 54 / 700 / 1.04 / −1.875px | 未定義（**描画箇所ゼロ**＝レンダリングされないものは遵守対象外と裁定） | — 非該当 |
| G2-11 | display-1 | 64 / 700 / 1.0 / −2.125px | 未定義（**描画箇所ゼロ**＝同上） | — 非該当 |
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
| G6-1 | button-primary | bg primary / on-primary / button 16px / pill | pill ✅ ・ラベル 16px ✅ | ✅ F1 |
| G6-2 | button-primary-pressed | bg primary-active | `hover:bg-primary/90`・`PALETTE.brandHover` | ✅ |
| G6-3 | button-secondary | white / ink / pill / Level-1 | 二次 CTA は white+pill | ✅ |
| G6-4 | button-utility | white / rounded-md / padding 4px 14px / hairline border | `rounded-md` ✅（F3 で xs→md）・高さは 44px タッチターゲット優先 | ✅ F3 🔒padding |
| G6-5 | button-icon-circular | `rgba(0,0,0,0.05)` fill / full | 本システムにカルーセル/メディア制御は無く、該当インスタンスが存在しない（アイコンボタンは button-utility として rounded-md 化＝F3） | — 非該当 |
| G6-6 | text-input | white / body-sm / padding 6px / rounded-xs / border `#DDDDDD` | rounded-xs ✅ ・border `#DDDDDD` ✅（F3）・高さ 44px（タッチターゲット優先） | ✅ F3 🔒padding |
| G6-7 | badge-pill | white / primary text / eyebrow / pill | `BADGE.*` は status 色ベース | 🔒 **status 色ベースで確定**（G1-12〜19 に追従・色=本書正本） |
| G6-8 | feature-card | white / rounded-lg / **padding 24px** / Level-0 | `STYLE.formCard` = rounded-lg + p-6(24px) | ✅ |
| G6-9 | ex-data-table-cell | canvas-soft header / eyebrow / body-sm / hairline row | FE10 で全テーブル字義化 | ✅ |
| G6-10 | ex-modal-card | rounded-xl / Level-2 / padding 24px | shadcn Dialog = rounded-xl + p-6 + shadow | ✅ |
| G6-11 | nav-bar | canvas(白) | サイドバーは `C.bgPage`(canvas-soft) | 🔒 **canvas-soft 確定**（FE11 裁定: 白にすると本文カードとの図地が消える。§7.1 に明文化済み） |
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

判定: ✅ 準拠 / 🔧 修正した / ⚠️ 裁定要 / — 該当なし

### F7 の検証方式と根拠（2026-07-21 完了）

P8/P9/P10 は **G軸の修正が全てトークン／CSS 層で行われたため全ページへ一括波及する性質**を持つ。よって
84 画面を個別にスクリーンショットで舐めるのではなく、**漏れが構造的に起こり得ないことの証明**で確定した:

1. **機械全数**: `text-*` の任意値=C11 / 非仕様サイズ段=C12 / ink 黒アルファ=C13 が**本体アプリ 0 件**。
   つまりロール外のサイズ・色は実装に存在し得ない（再導入も恒久ブロック）。
2. **迂回の全数走査**: トークンを迂回する手書きテキスト色（`text-gray/slate/zinc/neutral/stone-*`）を
   src 全域で検索 → **1 件のみ**（印刷帳票。§2.7 で媒体外と裁定）。shadcn の `text-muted-foreground` は
   `--muted-foreground: #615D59` に解決されるため字義準拠。
3. **サイズ変更の回帰確認（目視サンプル）**: 入力 40→44px・ボタン 15→16px・見出し写像の影響が出やすい
   代表面（/login・/owners/new・/accounting/:id 精算・/medical-records/:id 編集・入院ペット選択）で
   レイアウト崩れなしを確認。
4. **フルゲート**: type-check / test 2210 / build / lint 全緑。

**残存リスク（明示）**: 上記1〜2で「ロール外の値が存在しない」ことは全数保証できるが、
*ロール内で誤ったロールを選んでいる*（例: caption を使うべき箇所に body-sm）ケースは機械では検出できない。
これは FE10 の P3（階層）で全 84 面を目視済みであり、F1〜F5 はサイズ**値**の変更でロール**選択**を
変えていないため、退行は生じていない。

## 進捗表

| Batch | ルート | FE10<br>P1-P7 | P8<br>ink | P9<br>字送り | P10<br>タイポ | 所見 | 状態 |
|---|---|---|---|---|---|---|---|
| A | /login | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| A | /forgot-password | ✅ | ✅機 | ✅機 | ✅機 | BUG-419（未認証到達不能）は別件 | ✅済 |
| A | /reset-password | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| A | /（受付カンバン） | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| A | /owners（一覧） | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| A | /owners/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| A | /owners/:id | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| A | /owners/:id/report | ⚠ | | | | FE10 から継続: brand tick の装飾使用疑い＋別ワークストリーム編集中 | ⚠保留 |
| A | /aggregation | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B1 | /reservations | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B1 | 予約詳細モーダル | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B1 | 予約登録モーダル | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B2 | /hospitalization | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B2 | /hospitalization/select-pet | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B2 | /hospitalization/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B2 | /hospitalization/:id | 静✅ | | | | 全4医院で入院0件（API実測）＝描画不能 | ⏸データ待ち |
| B2 | /hospitalization/:id/edit | 静✅ | | | | 同上 | ⏸データ待ち |
| B3 | /medical-records | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B3 | /medical-records/select-pet | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B3 | /medical-records/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B3 | /medical-records/:id | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /accounting（3タブ） | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /accounting/select-pet | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /accounting/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /accounting/:id | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /accounting/close | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /accounting/close/history | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /accounting/reports | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /estimates | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /estimates/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /estimates/:id | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B4 | /estimates/:id/edit | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/clinic | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/staff | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/treatment-items | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/diagnosis | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/animal-species | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/trimming | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/trimming-course-type | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/medicine | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/reservation-type | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/hospitalization | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/cage | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/merchandise-items | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/insurance | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/occupations | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/permission-groups | ✅ | ✅機 | ✅機 | ✅機 | RBAC 非活性表現の退行を特に確認（C6a） | ✅済 |
| B5 | /settings/inquiry-templates | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/interview/chief-complaint | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/interview/templates | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/shift-templates | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/closing-time | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/payment-methods | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/campaigns | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/integrations/lstep | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B5 | /settings/lstep/tags | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /inventory | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /inventory/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /inventory/:id | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /trimming | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /trimming/select-pet | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /trimming/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /trimming/:id | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /vaccinations | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /vaccinations/select-pet | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /vaccinations/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /vaccinations/:id | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /checkups | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /checkups/select-pet | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /checkups/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /examinations | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /examinations/select-pet | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /examinations/new | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B6 | /examinations/:id | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B7 | /shifts | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B7 | /lstep/checkup-sync | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B7 | /lstep/delivery-monitor | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B7 | /lstep/analytics | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B7 | /line-reservation/settings | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B7 | /line-reservation/page-editor | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B7 | /line-reservation/slots | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B7 | /manual | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |
| B7 | /manual/:category/:slug | ✅ | ✅機 | ✅機 | ✅機 | | ✅済 |

（404 fallback・リダイレクト専用12 route は対象外 — `ui-design-compliance.md` §2 脚注どおり）

---

# 実行順（G軸→P軸。G軸はトークン修正なので全ページへ一括波及する）

| 順 | フェーズ | 内容 | 対象 G# | 状態 |
|---|---|---|---|---|
| 1 | **F1 タイポ数値** | `--text-sm/xs/2xs--line-height` 定義・button 16px 化・tracking トークン新設と title/eyebrow への適用 | G2-1,3,4,5,6 / G6-1 | **✅済 2b20b1aab**（6ロール全てブラウザ実測で完全一致。インライン `tracking-wide` を7ファイルから除去＝字義値の上書きを排除） |
| 2 | **F2 ink ランプ** | 14段アルファ → DESIGN.md 4段実値へ集約（明度最近傍で写像）。**臨床非活性(disabled/RBAC)の可読退行を必ず実測** | G1-8,9,10 | **✅済 27b483b64**（臨床非活性の退行なしを実測確認・placeholder も faint へ） |
| 3 | **F3 コンポーネント寸法** | utility 角丸 md・入力 44px・入力 border `#DDDDDD` | G6-4,6 | **✅済 861dcf347**（padding 6px/4px-14px は DESIGN.md 自身の 44px タッチターゲット規定と矛盾するため高さ 44px を優先＝内部矛盾の裁定） |
| 4 | **F4 非仕様サイズ段の撲滅** | 当初の「未使用トークン追加」は YAGNI のため破棄し、**実際に描画されている非仕様サイズ30箇所**の解消へ振替。`--text-heading-1/2/3` を字義新設し 36→40 / 30→26 / 24→22 / 18→20px へ写像。base h1-h3 も整列 | G2-7,8,9 | **✅済 c3e9c78fc** |
| 5 | **F5 機械ガード追加** | C12(非仕様サイズ段) / C13(ink 黒アルファ) を新設。ゲート合計への未算入バグも同時修正 | — | **✅済 bd6552fbd**（C13 が F2 の取りこぼし `hoverText60` を即検出→修正。LINE ミニアプリは明示 allowlist で除外） |
| 6 | **F6 文書同期** | 正本の軸分割を `design-system.md` へ反映（SSOT 表・§2.1/2.4/2.5/2.6・§3.4・§7.1）。Ink 行の虚偽 ✅ 訂正を含む | — | **✅済**（色の正本移管・sticker 不採用・nav-bar canvas-soft・ink 実値化を明文化） |
| 7 | **P軸 全84ページ検証** | 機械全数（C11/C12/C13 本体0件）+ 迂回全数走査（1件のみ・媒体外裁定）+ 代表面の回帰目視 + フルゲート4種緑 | — | **✅済**（方式と残存リスクは Part 2 冒頭に明記。`✅機` = 機械保証による確定） |

---


# 検証規約（FE10 の事故を踏まえた強化版）

- 各フェーズ後: `node scripts/design-system-audit.mjs` 緑 + 影響 `npx vitest run <path>`。
- **コミット前に touched file 全数へ scoped eslint 必須**（`docker compose exec -T frontend sh -c 'xargs npx eslint' < filelist`）。vitest も design-audit も**型を見ない**ため、これが唯一の scoped な型崩れ検出手段（FE10 で TS6133 を2回すり抜けた実績）。
- **トークン値を変えたら、その旧値を握るテストを全数 grep**（置換式そのものを grep パターンへ機械変換する。手で部分集合を書くと漏れる — FE10 で `2383E2` を落とした実績）。
- **値統合の副作用点検**: 統合したトークンの消費者を「構造用途/装飾用途」で仕分ける（FE10 で受付済ドットが構造色の装飾使用になった実績）。
- フル lint/type-check/build/test は最終ゲートで実行。
