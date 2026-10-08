/**
 * デプロイ後に古いタブが残っていると、content-hash 付きチャンク(/assets/*.js)が
 * 新しいデプロイの asset manifest に存在せず 404 になる。frontend Worker の
 * SPA fallback は GET の 404 を index.html(200/text/html)で返すため、
 * lazy route / React.lazy の import() が HTML を受け取り reject する。
 * そのまま RouteErrorBoundary の「エラーが発生しました」画面に到達し、
 * 旧タブのままでは何度遷移しても同じ失敗を繰り返す(EMR-271)。
 * import() の再試行は同じ旧チャンク名を要求するだけで回復しないため、
 * 新しい manifest を取得するフルリロードだけが回復手段になる。
 */

const RELOAD_MARKER_KEY = "emr:stale-chunk-reload-at";
const RELOAD_COOLDOWN_MS = 60_000;

/**
 * 動的 import 失敗の既知シグネチャ。
 * Chrome/Edge: "Failed to fetch dynamically imported module: <url>"
 * Firefox:     "error loading dynamically imported module: <url>"
 * Safari:      "Importing a module script failed."
 * MIME 不一致: "Failed to load module script: ... MIME type of \"text/html\""
 */
const DYNAMIC_IMPORT_FAILURE = /dynamically imported module|module script|importing a module/i;

/** lazy import の拒否がチャンク読み込み失敗によるものかを判定する。 */
export function isChunkLoadError(error: unknown): boolean {
  if (error instanceof Error === false) return false;
  if (error.name === "ChunkLoadError") return true;
  return DYNAMIC_IMPORT_FAILURE.test(error.message);
}

function readReloadMarker(): number | null {
  try {
    const raw = sessionStorage.getItem(RELOAD_MARKER_KEY);
    return raw === null ? 0 : Number(raw);
  } catch {
    return null;
  }
}

function writeReloadMarker(): boolean {
  try {
    sessionStorage.setItem(RELOAD_MARKER_KEY, String(Date.now()));
    return true;
  } catch {
    return false;
  }
}

/**
 * フルリロードで回復を試みるべき失敗かを判定する(副作用なし)。
 * - チャンク読み込み失敗以外は対象外(API エラー等はリロードで直らない)
 * - オフライン時はリロードしても回復しないため除外
 * - sessionStorage が使えない環境ではループ防止が効かないため除外
 * - 直近にリロード済みなら再試行しない(原因がデプロイ乖離でない場合の無限ループ防止)
 */
export function shouldRecoverByReload(error: unknown): boolean {
  if (!isChunkLoadError(error)) return false;
  if (typeof navigator !== "undefined" && navigator.onLine === false) return false;
  const last = readReloadMarker();
  if (last === null || Number.isNaN(last)) return false;
  return Date.now() - last >= RELOAD_COOLDOWN_MS;
}

/**
 * リロードマーカーを記録してフルリロードする。
 * マーカーを書けない環境ではリロードせず false を返す。
 */
export function reloadForStaleChunk(): boolean {
  if (!writeReloadMarker()) return false;
  window.location.reload();
  return true;
}
