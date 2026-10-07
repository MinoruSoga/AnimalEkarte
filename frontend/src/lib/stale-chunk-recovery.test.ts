import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  isChunkLoadError,
  reloadForStaleChunk,
  shouldRecoverByReload,
} from "./stale-chunk-recovery";

const CHUNK_ERROR_CHROME = new TypeError(
  "Failed to fetch dynamically imported module: https://stg.noah-karte.com/assets/medical-records-OLD.js",
);
const CHUNK_ERROR_FIREFOX = new TypeError(
  "error loading dynamically imported module: https://stg.noah-karte.com/assets/x.js",
);
const CHUNK_ERROR_SAFARI = new TypeError("Importing a module script failed.");
const CHUNK_ERROR_MIME = new TypeError(
  'Failed to load module script: Expected a JavaScript-or-Wasm module script but the server responded with a MIME type of "text/html".',
);

function stubReload() {
  const reload = vi.fn();
  vi.stubGlobal("window", { location: { reload } });
  return reload;
}

describe("isChunkLoadError", () => {
  it.each([
    ["Chrome", CHUNK_ERROR_CHROME],
    ["Firefox", CHUNK_ERROR_FIREFOX],
    ["Safari", CHUNK_ERROR_SAFARI],
    ["MIME type", CHUNK_ERROR_MIME],
  ])("returns true for %s dynamic import failure", (_name, error) => {
    expect(isChunkLoadError(error)).toBe(true);
  });

  it("returns true for webpack-style ChunkLoadError name", () => {
    const error = new Error("Loading chunk 42 failed.");
    error.name = "ChunkLoadError";
    expect(isChunkLoadError(error)).toBe(true);
  });

  it.each([
    ["非Error", "Failed to fetch dynamically imported module"],
    ["null", null],
    ["無関係なError", new Error("Network Error")],
    ["通常のTypeError", new TypeError("Cannot read properties of undefined")],
  ])("returns false for %s", (_name, error) => {
    expect(isChunkLoadError(error)).toBe(false);
  });
});

describe("shouldRecoverByReload", () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("チャンク失敗かつ未リロードなら回復対象", () => {
    expect(shouldRecoverByReload(CHUNK_ERROR_CHROME)).toBe(true);
  });

  it("チャンク失敗以外は回復対象外", () => {
    expect(shouldRecoverByReload(new Error("boom"))).toBe(false);
    expect(shouldRecoverByReload(null)).toBe(false);
  });

  it("直近にリロード済みなら再試行しない(ループ防止)", () => {
    sessionStorage.setItem("emr:stale-chunk-reload-at", String(Date.now()));
    expect(shouldRecoverByReload(CHUNK_ERROR_CHROME)).toBe(false);
  });

  it("クールダウン経過後は再び回復対象", () => {
    sessionStorage.setItem("emr:stale-chunk-reload-at", String(Date.now() - 61_000));
    expect(shouldRecoverByReload(CHUNK_ERROR_CHROME)).toBe(true);
  });

  it("マーカーが壊れている場合はリロードしない", () => {
    sessionStorage.setItem("emr:stale-chunk-reload-at", "not-a-number");
    expect(shouldRecoverByReload(CHUNK_ERROR_CHROME)).toBe(false);
  });

  it("オフライン時はリロードしない", () => {
    vi.spyOn(navigator, "onLine", "get").mockReturnValue(false);
    expect(shouldRecoverByReload(CHUNK_ERROR_CHROME)).toBe(false);
  });

  it("sessionStorage が読めない環境ではリロードしない", () => {
    vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new DOMException("denied");
    });
    expect(shouldRecoverByReload(CHUNK_ERROR_CHROME)).toBe(false);
  });
});

describe("reloadForStaleChunk", () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("マーカーを記録してフルリロードする", () => {
    const reload = stubReload();
    expect(reloadForStaleChunk()).toBe(true);
    expect(reload).toHaveBeenCalledTimes(1);
    const marker = sessionStorage.getItem("emr:stale-chunk-reload-at");
    expect(marker).not.toBeNull();
    expect(Number(marker)).toBeGreaterThan(0);
  });

  it("マーカーを書けない環境ではリロードしない", () => {
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new DOMException("denied");
    });
    const reload = stubReload();
    expect(reloadForStaleChunk()).toBe(false);
    expect(reload).not.toHaveBeenCalled();
  });
});
