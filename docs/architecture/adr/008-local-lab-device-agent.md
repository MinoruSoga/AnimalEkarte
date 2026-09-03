# ADR-008: Macローカル検査機器受信agent

- Status: Accepted
- Implementation: code complete; physical-device UAT and release gates pending
- Date: 2026-08-20
- Supersedes: ADR-007の「メニューバー常駐／デーモンは今は作らない」判断のみ

## Context

Web Serialは初回許可時にブラウザ標準のポート選択を要求する。また、同型USBシリアル変換器が複数ある場合、Web Serialが公開するVID/PIDだけでは、その先がNX600、AU10V、PU-4010のどれかを判定できない。実運用者はUSB識別子から機器を判断できず、機器別の選択操作は業務上成立しない。

Mac側では`/dev/cu.usbserial-*`を個別ポートとして列挙できる。現在確認済みのNX600とAU10Vはともに9600 8N1であり、受信電文は既存backend decoderが内容から分類できる。

## Decision

ユーザー権限のLaunchAgentとして`lab-device-agent`を常駐させる。

1. インストール時に接続済みのNX600/AU10V用2ポートを許可リストへ固定し、そのポートだけを9600 8N1で自動監視する。
2. 2秒の無通信で区切った生バイトを変更せず、pending 100件・reject 100件の独立した上限付きメモリキューへ保持する。
3. agentは`127.0.0.1:17654`だけでHTTPを待ち受ける。
4. `/lab-device`がキューをpollし、既存の認証済み`POST /v1/lab-device/frames`へ`device_hint=auto`で送る。
5. backend成功後だけACKする。400の不正電文はagent内のreject領域に生バイトを保持し、その他の失敗は再試行する。
6. agentへJWT、APIキー、患者情報を保存しない。

運用者はAPIの`LAB_DEVICE_AGENT_CONSUMER_TOKEN`と、bundle/installへ渡すconsumer tokenに同一の値を供給する。bundleはtokenを生成せず、4th引数または`LAB_DEVICE_AGENT_CONSUMER_TOKEN`が空ならfail-closedとする。その値を`lab-device-agent.conf`のL3へ保存し、mode 0600で保護する。LaunchAgentは同じ値を`--consumer-token`としてagentへ渡す。認証済みFrontendは`GET /v1/lab-device/agent-consumer`からAPI envの同じ値を取得し、claim・frame取得・ACK/rejectごとに`X-Lab-Device-Consumer-Token`で提示する。JWTはagentへ保存しない。token未設定または不一致ではprotected operationをfail-closedにする。`/health`は診断のため認証不要のままとする。

医院IDは秘密情報ではないローカルscope設定として保存する。Frontendは医院IDに一致する単一consumer leaseを取得し、consumer tokenとowner token付きでのみframe取得・ACK/rejectできる。これにより別医院のタブや同時consumerへの二重配送を拒否する。

機器分類の正本は既存backend decoderとし、agentに分類ロジックを複製しない。

## Safety boundaries

- loopback bind、Host完全一致、Frontend Origin allowlist、Private Network Access preflightを必須とする。
- 生電文、USBシリアル、デバイスパス、患者情報をログへ出さない。
- pendingまたはrejectキュー満杯時は既存フレームを捨てず、新規受付またはreject移動を失敗させてdegraded状態にする。
- PU-4010は2400 8E1のため初期agentで推測接続しない。
- IDEXX PIMS 応答（ACK+A+IM/SM）は `lab-device-agent --pims-reply` の明示オプトインのみ。既定は読取専用。医院 VetLab ケーブルでは有効にしない。
- 許可リストへ登録したUSBシリアル変換器はNX600/AU10V専用配線として扱う。同じ変換器をPU-4010、IDEXX、または未確認機器へ差し替える運用は信頼境界外とし、agent稼働中は行わない。
- 停止手段は`launchctl bootout gui/$UID/com.animalekarte.lab-device-agent`とする。

### Local trust boundary

Macのログイン済みOSユーザーと、そのユーザー権限で動くプロセスを信頼境界内とする。同一ユーザーの任意プロセスはagent HTTPを経由せず`/dev/cu.usbserial-*`を直接読めるため、consumer tokenは同一OSユーザー侵害を防ぐ境界ではない。端末のOSアカウント分離、画面ロック、マルウェア対策を前提とする。HTTPのclinic binding、consumer token、consumer leaseは、正規Frontendの別医院セッション・複数タブ間の誤配送防止を目的とする。

## Consequences

- 日常のUSB選択操作はゼロになる。
- ページを閉じている間もMac側で受信し、メモリ上限までは保持できる。ただしMac再起動では未配送キューが消える。
- PU-4010とIDEXXは別の制御された実機検証が完了するまで対象外として表示する。
- 将来ページ非依存でbackendへ直接送る場合は、clinic限定・失効可能なagent credentialを別ADRで設計する。
