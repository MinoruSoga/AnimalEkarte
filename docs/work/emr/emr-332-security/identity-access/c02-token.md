# U02-C02 — トークン検証 (token verification)

Campaign: emr-332-security-defense / Unit: U02-AUTHN-AUTHZ-CLINIC / Attempt: att-u02-1
Format: `docs/work/emr/emr-332-security/baseline/case-verdict-template.md` の 12 フィールド構成。

## ケース判定レコード

```text
ケース ID:            C02
対象リビジョン:        de28e9765a29091339a872a85f1875a947e07e5b
実施方式 (method):     static + simulated
正常条件:
  期待 — HS256 で署名され、purpose (access: Subject 空 / refresh:
         Subject=="refresh")・必須 claims (jti, iat, exp, 有効 lifetime)
         を満たすトークンのみ受理。発行側は epoch>0・JTI エントロピー
         必須で fail-closed。
  観測 — TestTokenService_IssueAndVerify PASS、TestTokenService_
         VerifyAccessToken 正常系 PASS、TestTokenService_
         ParseRefreshTokenClaims 正常系 PASS、TestTokenMatchesAccountEpoch
         (現行 epoch 一致 + bounded legacy 許可) PASS
         (identity-access/evidence/token-service.txt)。static: token_service.go:227-241
         (VerifyAccessToken), :243-293 (VerifyRefreshToken + blacklist),
         :95-129 (IssueAccessToken — epoch<=0 拒否 :102-103, JTI entropy
         必須 :105-108)。
拒否条件:
  期待 — 期限切れ・lifetime 超過・署名改竄・不正 secret・alg!=HS256
         (alg confusion 対策)・purpose 不一致 (access で refresh を
         使う/その逆)・必須 claims 欠落・空/超長入力・epoch 不一致は
         全て拒否。
  観測 — identity-access/evidence/token-service.txt で以下すべて PASS:
         期限切れ/超過 lifetime 拒否 (TestTokenService_VerifyAccessToken/
         expired, overlong lifetime; TestLegacyTokenWindowValid_
         ExactTTLBoundaries)、署名改竄 (tampered: 署名不正で拒否)、
         purpose 不一致 (access_token: subject 不一致で拒否 /
         refresh 側も同様)、HS384 alg 拒否 (alg_reject: HS384_は拒否する
         — jwtKeyFunc が HS256 以外に ErrSignatureInvalid,
         token_service.go:77-82)、必須 claims 欠落拒否 (ID/iat/exp/
         epoch 境界 — :235-239, :254-258, legacyTokenWindowValid
         :297-311)、入力サイズ上限 4096 bytes (maxJWTBytes :24,
         parseClaims :85-87 — TestTokenService_RejectsJWTInputOver
         MaximumSize)、JTI entropy 欠如時発行拒否 (TestTokenService_
         IssueFailsClosedWhenJTIEntropyIsUnavailable — :58-75)、
         epoch 不一致/負 epoch 拒否 (TokenMatchesAccountEpoch :318-329)。
         refresh 経路は nil blacklist で 503 fail-closed (:263-265)。
         ミドルウェア経由でも TestAuth/expired_token_returns_401,
         token_signed_with_HS384_is_rejected, token_signed_with_
         wrong_secret_returns_401, TestAuth_RejectsRefreshTokenSubject,
         TestAuth_RejectsAccountEpochMismatchAndResolverFailure が
         PASS (identity-access/evidence/auth-middleware.txt)。
防御判定:            PASS (static+simulated の範囲)
副作用呼出回数:       0 — 検証失敗は claims 返却前に早期 return。拒否時に
                     blacklist 書込・audit 書込なし (VerifyAccessToken
                     :227-241 は純粋検証; VerifyRefreshToken の
                     family 失効書込 :283-289 は revoked JTI 検出時の
                     意図的副作用であり正常拒否ケースでは 0)。
                     実測 DB/redis 副作用は未検証 (simulated のみ)。
検出判定:            PASS — 拒否は呼出側で 401 として記録・応答
                     (middleware/auth.go:80-84)。トークン値・claims
                     内容をログへ書かないことを static 確認
                     (token_service.go, middleware/auth.go に tokenStr/
                     claims 内容を出力するログ行なし; WARN/ERROR は
                     staff_id と error kind のみ — auth.go:245-284)。
                     証跡: evidence-t01 において WARN/ERROR 行は
                     "staff_id=123" とエラー種別のみでトークン文字列
                     を含まない。
安全な証跡参照:       backend/internal/auth/token_service.go:24, :58-75,
                     :77-93, :95-129, :227-241, :243-293, :297-311,
                     :318-329 / backend/internal/authjwt/claims.go:6-14 /
                     docs/work/emr/emr-332-security/identity-access/evidence/token-service.txt /
                     docs/work/emr/emr-332-security/identity-access/evidence/auth-middleware.txt
制約・未検証範囲:     staging/production 実測は BLOCKED。実ブラックリスト
                     (redis/DB) の fail-closed 性は static + 既存 mock 検証
                     のみ; 実プロダクション鍵・発行経路の検証は対象外。
                     clock skew 境界は legacyTokenClockSkew (定数)
                     の静的検証 + テストのみ。
修復案:              none applied — ギャップ証拠なし。
再検証結果:          修復なしのため未実施。simulated 再実行全 PASS。
```

## 補足: alg/purpose/size の強制ポイント

- `iss` claim: **N/A-by-design** — access token は `iss` を発行・要求しない (`authjwt.Claims` に Issuer field なし、token_service.go:111-122 / :227-241)。issuer 検証ケースは設計上非該当として記録 (拒否ケースではなく未実装)。検証は署名 (HS256)・purpose (Subject)・登録 claims (jti/exp/iat)・epoch に限定される。
- `jwtKeyFunc` (token_service.go:77-82) — `token.Method != jwt.SigningMethodHS256` の場合 `jwt.ErrSignatureInvalid`。`alg=none`・HS384・RS256 いずれも受理しない (テストで HS384 拒否を確認)。
- `VerifyAccessToken` (:232-234) — `claims.Subject != ""` → "invalid token type"。access/refresh 混用を purpose チェックで拒否 (refresh は :251 で `refreshTokenSubject` 必須)。
- `parseClaims` (:84-93) — 空文字/4096 bytes 超を `ErrTokenMalformed` で拒否してから parse。
- `legacyTokenWindowValid` (:297-311) — `0 < lifetime <= maxTTL` かつ `iat <= now+skew` を強制 (iat/exp 必須含む)。
- `TokenMatchesAccountEpoch` (:318-329) — epoch>0 の両方で一致必須。legacy (epoch==0) は `iat > accountMutation` の厳格比較で bounded 許可。負 epoch・nil claims は拒否。

## 検証コマンド

```sh
docker run --rm --pull never --network none \
  -v <worktree>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine \
  go test -count=1 -v -run '^(TestTokenService_|TestLegacyTokenWindowValid_|TestTokenMatchesAccountEpoch)' ./internal/auth/
# => PASS / ok github.com/animal-ekarte/backend/internal/auth 0.008s
#    (identity-access/evidence/token-service.txt, exit 0)
```
