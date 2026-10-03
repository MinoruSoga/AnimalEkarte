/**
 * BugReportTab — 「バグを報告」タブ
 *
 * タブを開いた時点で画面を自動キャプチャし、件名・詳細と一緒に
 * POST /v1/support/bug-reports（multipart）へ送信する。
 * スクショはプレビュー確認 → 撮り直し / 画像差し替え / 削除が可能。
 * 個人情報が写り得るため、送信前に必ずプレビューを確認させる。
 */
import { useCallback, useEffect, useRef, useState } from "react";
import { useLocation } from "react-router";
import { Camera, CheckCircle2, ImagePlus, Loader2, RefreshCw, Send, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, STYLE } from "@/lib/design-tokens";

import { useCreateBugReport } from "../api/create-bug-report";
import { capturePageScreenshot } from "../lib/capture-screenshot";

const SCREENSHOT_MAX_BYTES = 8 * 1024 * 1024; // BE の multipart 上限と揃える
const ACCEPTED_IMAGE_TYPES = new Set(["image/png", "image/jpeg", "image/webp"]);

interface BugReportTabProps {
  onClose: () => void;
}

export function BugReportTab({ onClose }: BugReportTabProps) {
  const location = useLocation();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [title, setTitle] = useState("");
  const [detail, setDetail] = useState("");
  const [screenshot, setScreenshot] = useState<File | null>(null);
  const [screenshotUrl, setScreenshotUrl] = useState<string | null>(null);
  const screenshotUrlRef = useRef<string | null>(null);
  const [capturing, setCapturing] = useState(true);
  const [submitted, setSubmitted] = useState(false);

  const createReport = useCreateBugReport();

  // スクショの差し替え・削除を一元化し、旧 object URL を必ず解放する。
  // URL 生成はイベントハンドラ/非同期コールバックでのみ行い、render/effect の同期処理には置かない。
  const applyScreenshot = useCallback((file: File | null) => {
    if (screenshotUrlRef.current !== null) {
      URL.revokeObjectURL(screenshotUrlRef.current);
    }
    const url = file === null ? null : URL.createObjectURL(file);
    screenshotUrlRef.current = url;
    setScreenshot(file);
    setScreenshotUrl(url);
  }, []);

  // アンマウント時に残っている object URL を解放
  useEffect(() => {
    return () => {
      if (screenshotUrlRef.current !== null) {
        URL.revokeObjectURL(screenshotUrlRef.current);
      }
    };
  }, []);

  const capture = useCallback(async () => {
    setCapturing(true);
    const file = await capturePageScreenshot();
    setCapturing(false);
    if (file !== null) {
      applyScreenshot(file);
    }
  }, [applyScreenshot]);

  // タブを開いた時点（このタブの初回マウント）で現在画面を自動キャプチャ。
  // Radix Tabs は非アクティブな Content を unmount するため、ここは「バグを報告」
  // を選んだ直前の画面状態を写す。
  // capture() 内の setState を effect 本体から同期的に呼ばないよう microtask に逃がす。
  useEffect(() => {
    queueMicrotask(() => void capture());
  }, [capture]);

  const handleFileChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0];
      event.target.value = "";
      if (!file) return;
      if (!ACCEPTED_IMAGE_TYPES.has(file.type)) {
        toast.error("PNG / JPEG / WebP 形式の画像を選択してください");
        return;
      }
      if (file.size > SCREENSHOT_MAX_BYTES) {
        toast.error("画像サイズが上限（8MB）を超えています");
        return;
      }
      applyScreenshot(file);
    },
    [applyScreenshot],
  );

  const canSubmit = title.trim().length > 0 && !createReport.isPending && !capturing;

  const handleSubmit = useCallback(
    (event: React.FormEvent) => {
      event.preventDefault();
      const trimmedTitle = title.trim();
      if (trimmedTitle.length === 0) return;
      createReport.mutate(
        {
          title: trimmedTitle,
          detail: detail.trim(),
          pageUrl: window.location.href,
          routePath: location.pathname,
          userAgent: navigator.userAgent,
          viewport: `${window.innerWidth}x${window.innerHeight}`,
          appVersion: (import.meta.env.VITE_APP_VERSION as string | undefined) ?? "",
          screenshot,
        },
        {
          onSuccess: () => {
            setSubmitted(true);
          },
        },
      );
    },
    [title, detail, location.pathname, screenshot, createReport],
  );

  if (submitted) {
    return (
      <div className="flex flex-col items-center gap-3 p-6 text-center">
        <CheckCircle2 className={`size-8 ${C.textStatusGreen}`} aria-hidden="true" />
        <p className={`text-base font-medium ${C.text}`}>報告を送信しました</p>
        <p className={`text-sm ${C.text60}`}>
          ご協力ありがとうございます。内容を確認のうえ対応します。
        </p>
        <Button variant="outline" size="sm" onClick={onClose}>
          閉じる
        </Button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-3 p-3">
      <div className="flex flex-col gap-1">
        <label htmlFor="bug-report-title" className={STYLE.formLabel}>
          件名{" "}
          <span className={C.textRequired} aria-hidden="true">
            *
          </span>
        </label>
        <input
          id="bug-report-title"
          type="text"
          required
          maxLength={200}
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="例: 会計画面でボタンが押せない"
          className={STYLE.formInput}
        />
      </div>

      <div className="flex flex-col gap-1">
        <label htmlFor="bug-report-detail" className={STYLE.formLabel}>
          詳細・再現手順
        </label>
        <Textarea
          id="bug-report-detail"
          rows={4}
          maxLength={4000}
          value={detail}
          onChange={(e) => setDetail(e.target.value)}
          placeholder={
            "いつ・どの画面で・何をしたときに起きたか\n例: 会計確定ボタンを押すと真っ白になる"
          }
        />
      </div>

      <fieldset className="flex flex-col gap-1.5">
        <legend className={STYLE.formLabel}>スクリーンショット（現在の画面）</legend>
        <div
          className={`flex items-center justify-center overflow-hidden rounded-xxs border ${C.borderMedium} ${C.bgPage} min-h-[120px]`}
        >
          {capturing ? (
            <span className={`flex items-center gap-2 py-8 text-sm ${C.text50}`}>
              <Loader2 className="size-4 animate-spin" aria-hidden="true" />
              画面を撮影しています…
            </span>
          ) : screenshotUrl !== null ? (
            <img
              src={screenshotUrl}
              alt="送信されるスクリーンショットのプレビュー"
              className="max-h-[180px] w-full object-contain"
            />
          ) : (
            <span className={`flex items-center gap-2 py-8 text-sm ${C.text60}`}>
              <Camera className="size-4" aria-hidden="true" />
              スクリーンショットなし
            </span>
          )}
        </div>
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => void capture()}
            disabled={capturing}
          >
            <RefreshCw className="size-3.5" aria-hidden="true" />
            撮り直す
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => fileInputRef.current?.click()}
          >
            <ImagePlus className="size-3.5" aria-hidden="true" />
            画像を差し替える
          </Button>
          {screenshot !== null ? (
            <Button
              type="button"
              variant="ghost-danger"
              size="sm"
              onClick={() => applyScreenshot(null)}
            >
              <Trash2 className="size-3.5" aria-hidden="true" />
              削除
            </Button>
          ) : null}
        </div>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/png,image/jpeg,image/webp"
          className="hidden"
          onChange={handleFileChange}
          aria-label="スクリーンショット画像を選択"
        />
        <p className={`text-2xs leading-relaxed ${C.text60}`}>
          スクリーンショットには患者・飼主の個人情報が写り込む場合があります。
          報告は全医院のスタッフに共有され、AI・チケット管理ツールへ送信される場合があります。
          内容を確認してから送信してください。
        </p>
      </fieldset>

      <PrimaryButton type="submit" disabled={!canSubmit} className="mt-1">
        <Send className="size-4" aria-hidden="true" />
        {createReport.isPending ? "送信中…" : "報告を送信"}
      </PrimaryButton>
    </form>
  );
}
