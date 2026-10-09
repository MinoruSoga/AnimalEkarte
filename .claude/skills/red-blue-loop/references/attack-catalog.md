# 攻撃カタログ(invariant 別)

各項目は SECURITY.md の invariant に対応する。「合格基準」を満たせば脆弱性なし、「満たさない実測」が finding 候補。実行前に SKILL.md 手順2 のとおり合格シグナルを計画に書いてから始める。

共通の前提:

- エンドポイント・パラメータ名は必ずソースから確認する(`backend/internal/**/routes*.go`、handler の bind struct)。このカタログの `<...>` は実在パスへの置き換え箇所であり、そのまま実行しない。
- 複数 clinic の staff トークンを取得し、以降は clinic A のトークンで clinic B のリソース ID を触る形が基本形。
- 連番 ID の総当たりは数十 ID・逐次(並列なし)に抑える。必要なのは「分離が破れる 1 例」であり、網羅は回帰テストの仕事。

## A. TENANT-1 — clinic 分離(最優先)

**目的**: clinic A の staff が clinic B の clinical/accounting data に触れないことを実測する。SECURITY.md の重大度で High 起点。

**手順**:

1. clinic A / clinic B それぞれで pet・owner・medical record・reservation・billing・inventory を 1 件ずつ作成し、ID を控える。
2. clinic A のトークンで、clinic B の各リソース ID を GET / PUT / DELETE。path param・body 内 ID・query param の 3 経路すべてを試す。
3. 間接参照も試す: clinic B の pet にぶら下がる medical record 一覧、preload 経由の親取得(owner → pets → records)。`backend/internal/persistence/scope.go` の helper を使って**いない**箇所が本命(ソース grep で候補を絞ってから実測する)。
4. bulk / count / export(CSV)系エンドポイントを同じく。一括系は scope 忘れが起きやすい。
5. 背景経路(import・scheduler・バッチ)が clinic scope を持つかはソース確認で判定し、実測可能なら 1 例を実行する。

**合格基準**: 全操作が 403/404 で拒否され、レスポンス差分から clinic B リソースの存在が推定できない。

## B. AUTHZ-1 — 独立した認可・RBAC

**目的**: authentication 通過後でも、permission と ownership の独立 check がない操作が通らないことを実測する。

**手順**:

1. 通常 staff アカウントでログインし、system administrator 専用と考えられる操作(staff provisioning、permission 管理、all-clinic batch、`backend/internal/auth/http_permission.go` の AllResources に対応する操作)を叩く。期待: 403。
2. 同一 clinic 内でも role 差を試す: 閲覧のみ権限の staff が write 操作を叩く。
3. ソース側で「認可判定の入力取得エラー時に空集合で続行する fail-open」を grep する(security-checklist skill の CRITICAL 実例と同型)。見つけたら該当ハンドラを実測で確認する。

**合格基準**: 全操作が 403。role や permission を body/query で上書きしても通らない。

## C. PUBLIC-1 — 公開境界

**目的**: 認証不要で到達できる surface が、それぞれの条件で fail closed していることを実測する。

**手順**:

1. LIFF 系: LINE token なし / 改ざん token / 他 clinic の token で customer 向け API を叩く。path 上の clinic・customer と token の紐付け不一致を試す。実在ルートは `backend/internal/reservation/routes.go`・`backend/internal/lstep/routes.go` で確認。
2. LINE webhook: `x-line-signature` を欠落・改ざんして POST。本文上限も試す。HMAC 検証の bypass が本命。
3. local uploads・UAT fixture: STG や本番相当の Host ヘッダで `/uploads`・UAT fixture を叩き、dev 条件 guard が実効的に効くか実測する(guard bypass は SECURITY.md 明示の評価対象)。
4. health 系は公開前提。過剰な内部情報(バージョン・DB 詳細)を返していないかだけ確認する。

**合格基準**: 全公開境界が条件不一致で 404/401/403 のいずれか。エラーメッセージに内部詳細を含まない。

## D. PRIV-1 — 特権操作の分離

**手順**: 通常 staff session で migration・scheduler 操作・CSV import・staff provisioning 系のエンドポイントを叩く。`/_internal/*` は Worker 側で internal path として分岐する前提だが、Go Container 直叩き(local では port 越しに可能)でも同様に拒否されるか確認する。

**合格基準**: すべて通常 staff 認証のままでは到達できない(404/403/認証拒否)。

## E. OBJECT-1 — object storage

**手順**:

1. clinic A の staff が画像アップロード→取得の通常フローで signed URL を発行し、その URL の構造・有効期限を確認する。
2. clinic A のトークンで clinic B の画像 key を要求し、signed URL 発行が拒否されることを実測する。
3. key の推測(連番・予測可能パターン)で bucket 直アクセスを試す。`animalekarte-stg-images` は private 前提。presigned URL なしの GET が通れば finding。
4. ソース確認: `S3_PUBLIC_BASE_URL` を前提にした公開 URL 生成が混入していないか(`s3_uploader.go` 系の既知のコメント/実装不整合に注意)。

**合格基準**: 画像は signed URL 経由でのみ取得可能で、発行には clinic scope check がある。

## F. AUTHN — session・token

**手順**:

1. refresh token rotation: login → refresh を 2 回実行し、**使い済み refresh token の再利用**が family 無効化(reuse detection)を引き起こすことを確認する。再利用が成功する場合は High。
2. staff の clinic 割当変更・無効化後、旧 access token(最長 15 分)が使い続けられないか(account epoch 再解決の実測)。
3. Set-Cookie の httpOnly / Secure / SameSite を実測確認。
4. logout 後の access token 残存は既知の設計範囲かどうかソース確認してから判定する(短命 token の前提と照合)。

**合格基準**: reuse で family 無効化、epoch 変更で旧 token が fail closed、cookie 属性が仕様どおり。

## G. 入力系 — Injection / XSS / CSRF / business logic

**手順**:

1. SQLi スポット: ソース grep で `Raw(`・`fmt.Sprintf` を SQL に使っている箇所を抽出し、そこだけ実測(`'` と `--` による挙動変化の確認で十分。高レートツールは使わない)。
2. Stored XSS: pet 名・飼い主名・メモに `<img src=x onerror=...>` 等を保存し、検索・一覧・印刷・LIFF の各画面で発火するか確認。React は escape する前提なので、本命はソース grep した `dangerouslySetInnerHTML` 箇所(DOMPurify 有無を確認)。
3. CSRF: origin ヘッダなし・`X-Requested-With` なしの state-changing POST が通るか。SameSite 属性と合わせて判定する。
4. Upload: content-type 偽装・拡張子偽装・SVG(XSS 含む)・サイズ上限超過。EICAR テスト文字列は使用可、実マルウェアは使用禁止。
5. Business logic: 負の数量・0 割引・過去日の予約・重複予約・会計額の不整合。会計系の確認は `lane:finance` 影響を起票時に明記する。

**合格基準**: 各操作が boundary validation で拒否されるか、無害化されて保存・表示される。

## H. CONFIG-1 / LIMIT-1 — 設定と上限

**手順**:

1. (ローカルのみ)release 相当 env で起動し、weak secret・loopback DB・非 TLS DB が起動拒否されることを確認する。STG では実施しない。
2. 500 エラー応答に stack trace・内部エラー文字列が漏れないか、数箇所の異常系で実測する。
3. Rate limit は process-local であることが既知の制限。**material abuse**(認証系 endpoint の実効的な総当たり成立など)になる場合のみ報告し、単なる既知制限の再報告はしない。高レート送信はしない。

**合格基準**: release が fail closed、エラー応答が汎用メッセージ。

---

## 再報告禁止リスト(SECURITY.md 受容済み・承認済み)

以下は SECURITY.md で owner 裁定済み。**起票しない**。ただし波及条件(「STG 以外への波及」「実データ混入」「同シークレットの他用途流用」など)を満たす実測が出た場合は別の finding として評価する。

- STG デモカタログの共有パスワードが全カタログ ID(執行アカウント含む)を認証すること
- Support の閲覧・作成が全医院共有であること(変更操作の clinic スコープ破れは依然報告対象)
- Support コンテンツへの患者・飼い主情報含有と無期限保持(LLM・Plane への既存経路送信も承認済み)
- RLS が FORCE されておらず dormant であること単体(app-layer scope が破れている場合の組合せ、RLS を実効 control と誤記する docs drift は報告対象)
- 2FA 未実装(既知の制限事項)
- 攻撃者が開始時に持つ正規権限内の通常動作
- Production が draft で未 live であること自体
