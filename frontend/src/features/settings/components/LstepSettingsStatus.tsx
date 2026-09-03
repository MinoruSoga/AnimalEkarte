import { C } from "@/lib/design-tokens";
import { formatJSTDate } from "@/lib/jst-date";
import type { LstepSettingsResponse } from "../hooks/use-lstep-settings";

function formatSyncDate(iso: string): string {
  const [year, month, day] = formatJSTDate(iso).split("-");
  return `${Number(year)}年${Number(month)}月${Number(day)}日`;
}

interface LstepSettingsStatusHeaderProps {
  isConfigured: boolean;
  isSyncEnabled: boolean;
}

export function LstepSettingsStatusHeader({
  isConfigured,
  isSyncEnabled,
}: LstepSettingsStatusHeaderProps) {
  return (
    <div className="flex items-center justify-between mb-6">
      <h2 className={`text-base font-semibold ${C.text}`}>Lステップ連携設定</h2>
      <div className="flex items-center gap-2">
        <StatusPill active={isConfigured} activeLabel="設定済み" inactiveLabel="未設定" />
        {isConfigured ? (
          <StatusPill active={isSyncEnabled} activeLabel="同期中" inactiveLabel="同期停止中" />
        ) : null}
      </div>
    </div>
  );
}

interface StatusPillProps {
  active: boolean;
  activeLabel: string;
  inactiveLabel: string;
}

function StatusPill({ active, activeLabel, inactiveLabel }: StatusPillProps) {
  if (active) {
    return (
      <span
        className={`inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full ${C.bgStatusGreen} ${C.textStatusGreen} border ${C.borderStatusGreen}`}
      >
        <span className={`inline-block w-1.5 h-1.5 rounded-full ${C.bgStatusGreenDot}`} />
        {activeLabel}
      </span>
    );
  }

  return (
    <span
      className={`inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full ${C.bgStatusGray} ${C.textStatusGray} border ${C.borderMuted}`}
    >
      <span className={`inline-block w-1.5 h-1.5 rounded-full ${C.bgStatusGrayMedium}`} />
      {inactiveLabel}
    </span>
  );
}

interface LstepConfiguredSummaryProps {
  settings: LstepSettingsResponse | undefined;
  isConfigured: boolean;
}

export function LstepConfiguredSummary({ settings, isConfigured }: LstepConfiguredSummaryProps) {
  if (!isConfigured) return null;

  return (
    <div
      className={`mb-6 p-3 rounded-xs ${C.bgPage} border ${C.borderLight} text-sm ${C.text60} flex flex-col gap-1`}
    >
      {settings?.lstep_api_key_masked ? (
        <span>
          Lステップ APIキー:{" "}
          <span className={`font-mono ${C.text}`}>••••{settings.lstep_api_key_masked}</span>
        </span>
      ) : null}
      {settings?.line_channel_access_token_masked ? (
        <span>
          LINE Channel Access Token:{" "}
          <span className={`font-mono ${C.text}`}>
            ••••{settings.line_channel_access_token_masked}
          </span>
        </span>
      ) : null}
      {settings?.line_channel_secret_masked ? (
        <span>
          LINE Channel Secret:{" "}
          <span className={`font-mono ${C.text}`}>••••{settings.line_channel_secret_masked}</span>
        </span>
      ) : null}
      {settings?.liff_id ? (
        <span>
          LIFF ID: <span className={`font-mono ${C.text}`}>{settings.liff_id}</span>
        </span>
      ) : null}
      {settings?.lstep_base_url ? (
        <span>
          LステップベースURL:{" "}
          <span className={`font-mono ${C.text}`}>{settings.lstep_base_url}</span>
        </span>
      ) : null}
      {settings?.sync_enabled_at ? (
        <span>
          同期有効化日時:{" "}
          <span className={`font-mono ${C.text}`}>{formatSyncDate(settings.sync_enabled_at)}</span>
        </span>
      ) : null}
    </div>
  );
}
