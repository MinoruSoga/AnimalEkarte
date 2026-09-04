import { C, STYLE } from "@/lib/design-tokens";
import type { LstepSettingsResponse } from "../hooks/use-lstep-settings";

interface PasswordFieldProps {
  id: string;
  name: string;
  label: string;
  placeholder?: string;
}

function PasswordField({ id, name, label, placeholder }: PasswordFieldProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={STYLE.formLabel}>
        {label}
      </label>
      <input
        id={id}
        type="password"
        name={name}
        placeholder={placeholder ?? "新しい値を入力（空欄で変更なし）"}
        autoComplete="new-password"
        className={`${STYLE.formInput} rounded-xs border px-3 w-full outline-none ${C.focusRingAccent}`}
      />
    </div>
  );
}

interface TextFieldProps {
  id: string;
  name: string;
  label: string;
  defaultValue?: string;
  placeholder?: string;
}

function TextField({ id, name, label, defaultValue, placeholder }: TextFieldProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={STYLE.formLabel}>
        {label}
      </label>
      <input
        id={id}
        type="text"
        name={name}
        defaultValue={defaultValue ?? ""}
        placeholder={placeholder ?? ""}
        className={`${STYLE.formInput} rounded-xs border px-3 w-full outline-none ${C.focusRingAccent}`}
      />
    </div>
  );
}

interface CPMVersionSelectFieldProps {
  id: string;
  name: string;
  label: string;
  defaultValue: string;
}

function CPMVersionSelectField({ id, name, label, defaultValue }: CPMVersionSelectFieldProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={STYLE.formLabel}>
        {label}
      </label>
      <select
        id={id}
        name={name}
        defaultValue={defaultValue}
        className={`${STYLE.formInput} rounded-xs border px-3 w-full outline-none ${C.focusRingAccent}`}
      >
        <option value="v1">V1</option>
        <option value="v2">V2</option>
      </select>
    </div>
  );
}

interface NumberInputFieldProps {
  id: string;
  name: string;
  label: string;
  defaultValue: number;
}

function NumberInputField({ id, name, label, defaultValue }: NumberInputFieldProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className={STYLE.formLabel}>
        {label}
      </label>
      <input
        id={id}
        type="number"
        name={name}
        defaultValue={defaultValue}
        min={1}
        className={`${STYLE.formInput} rounded-xs border px-3 w-full outline-none ${C.focusRingAccent}`}
      />
    </div>
  );
}
interface LstepConnectionFieldsProps {
  settings: LstepSettingsResponse | undefined;
}

export function LstepConnectionFields({ settings }: LstepConnectionFieldsProps) {
  return (
    <>
      <PasswordField
        id="lstep_api_key"
        name="lstep_api_key"
        label="Lステップ APIキー"
        placeholder={
          settings?.lstep_api_key_masked
            ? `現在: ••••${settings.lstep_api_key_masked}（変更する場合のみ入力）`
            : "Lステップ APIキーを入力"
        }
      />
      <PasswordField
        id="line_channel_access_token"
        name="line_channel_access_token"
        label="LINE Channel Access Token"
        placeholder={
          settings?.line_channel_access_token_masked
            ? `現在: ••••${settings.line_channel_access_token_masked}（変更する場合のみ入力）`
            : "LINE Channel Access Tokenを入力"
        }
      />
      <PasswordField
        id="line_channel_secret"
        name="line_channel_secret"
        label="LINE Channel Secret"
        placeholder={
          settings?.line_channel_secret_masked
            ? `現在: ••••${settings.line_channel_secret_masked}（変更する場合のみ入力）`
            : "LINE Channel Secretを入力"
        }
      />
      <TextField
        id="liff_id"
        name="liff_id"
        label="LIFF ID"
        defaultValue={settings?.liff_id ?? ""}
        placeholder="例: 1234567890-xxxxxxxx"
      />
      <TextField
        id="lstep_base_url"
        name="lstep_base_url"
        label="LステップベースURL"
        defaultValue={settings?.lstep_base_url ?? ""}
        placeholder="例: https://api.lstep.jp"
      />
    </>
  );
}

interface LstepCpmFieldsProps {
  settings: LstepSettingsResponse | undefined;
}

export function LstepCpmFields({ settings }: LstepCpmFieldsProps) {
  return (
    <>
      <CPMVersionSelectField
        id="cpm_version"
        name="cpm_version"
        label="CPMバージョン"
        defaultValue={settings?.cpm_version ?? "v1"}
      />
      <NumberInputField
        id="cpm_v2_coming_threshold"
        name="cpm_v2_coming_threshold"
        label="CPM V2 来院数閾値 (出会い→良好)"
        defaultValue={settings?.cpm_v2_coming_threshold ?? 2}
      />
      <NumberInputField
        id="cpm_v2_good_threshold"
        name="cpm_v2_good_threshold"
        label="CPM V2 来院数閾値 (良好→ファミリー)"
        defaultValue={settings?.cpm_v2_good_threshold ?? 4}
      />
      <NumberInputField
        id="cpm_v2_family_threshold"
        name="cpm_v2_family_threshold"
        label="CPM V2 来院数閾値 (ファミリー→NOAH)"
        defaultValue={settings?.cpm_v2_family_threshold ?? 8}
      />
      <NumberInputField
        id="cpm_v2_noah_threshold"
        name="cpm_v2_noah_threshold"
        label="CPM V2 来院数閾値 (NOAH)"
        defaultValue={settings?.cpm_v2_noah_threshold ?? 13}
      />
      <NumberInputField
        id="cpm_v1_dormant_days"
        name="cpm_v1_dormant_days"
        label="CPM V1 休眠判定 (日数)"
        defaultValue={settings?.cpm_v1_dormant_days ?? 240}
      />
      <NumberInputField
        id="cpm_v1_noah_days"
        name="cpm_v1_noah_days"
        label="CPM V1 NOAH 在籍日数"
        defaultValue={settings?.cpm_v1_noah_days ?? 365}
      />
      <NumberInputField
        id="cpm_v1_noah_annual_visits"
        name="cpm_v1_noah_annual_visits"
        label="CPM V1 NOAH 年間来院回数"
        defaultValue={settings?.cpm_v1_noah_annual_visits ?? 3}
      />
      <NumberInputField
        id="cpm_v1_noah_ltv"
        name="cpm_v1_noah_ltv"
        label="CPM V1 NOAH LTV (円)"
        defaultValue={settings?.cpm_v1_noah_ltv ?? 80000}
      />
      <NumberInputField
        id="cpm_v1_core_days"
        name="cpm_v1_core_days"
        label="CPM V1 コア 在籍日数"
        defaultValue={settings?.cpm_v1_core_days ?? 180}
      />
      <NumberInputField
        id="cpm_v1_core_annual_visits"
        name="cpm_v1_core_annual_visits"
        label="CPM V1 コア 年間来院回数"
        defaultValue={settings?.cpm_v1_core_annual_visits ?? 2}
      />
      <NumberInputField
        id="cpm_v1_core_ltv"
        name="cpm_v1_core_ltv"
        label="CPM V1 コア LTV (円)"
        defaultValue={settings?.cpm_v1_core_ltv ?? 50000}
      />
      <NumberInputField
        id="cpm_v1_spot_min_amount"
        name="cpm_v1_spot_min_amount"
        label="CPM V1 スポット 単回最大金額 (円)"
        defaultValue={settings?.cpm_v1_spot_min_amount ?? 30000}
      />
      <NumberInputField
        id="cpm_v1_spot_inactive_days"
        name="cpm_v1_spot_inactive_days"
        label="CPM V1 スポット 来院停止日数"
        defaultValue={settings?.cpm_v1_spot_inactive_days ?? 90}
      />
      <NumberInputField
        id="cpm_v1_growing_max_days"
        name="cpm_v1_growing_max_days"
        label="CPM V1 グローイング 在籍最大日数"
        defaultValue={settings?.cpm_v1_growing_max_days ?? 90}
      />
      <NumberInputField
        id="cpm_v1_growing_min_visits"
        name="cpm_v1_growing_min_visits"
        label="CPM V1 グローイング 来院回数 (下限)"
        defaultValue={settings?.cpm_v1_growing_min_visits ?? 2}
      />
      <NumberInputField
        id="cpm_v1_growing_max_visits"
        name="cpm_v1_growing_max_visits"
        label="CPM V1 グローイング 来院回数 (上限)"
        defaultValue={settings?.cpm_v1_growing_max_visits ?? 3}
      />
      <NumberInputField
        id="cpm_v1_ltv_break_low"
        name="cpm_v1_ltv_break_low"
        label="CPM V1 LTV 下限境界 (円)"
        defaultValue={settings?.cpm_v1_ltv_break_low ?? 20000}
      />
    </>
  );
}

interface LstepPreventionFieldsProps {
  settings: LstepSettingsResponse | undefined;
}

export function LstepPreventionFields({ settings }: LstepPreventionFieldsProps) {
  return (
    <>
      <NumberInputField
        id="dormant_prevention_180_days"
        name="dormant_prevention_180_days"
        label="休眠予防閾値 (180日)"
        defaultValue={settings?.dormant_prevention_180_days ?? 180}
      />
      <NumberInputField
        id="dormant_prevention_210_days"
        name="dormant_prevention_210_days"
        label="休眠予防閾値 (210日)"
        defaultValue={settings?.dormant_prevention_210_days ?? 210}
      />
      <NumberInputField
        id="dormant_prevention_240_days"
        name="dormant_prevention_240_days"
        label="休眠予防閾値 (240日)"
        defaultValue={settings?.dormant_prevention_240_days ?? 240}
      />
      <NumberInputField
        id="dormant_prevention_365_days"
        name="dormant_prevention_365_days"
        label="休眠予防閾値 (365日)"
        defaultValue={settings?.dormant_prevention_365_days ?? 365}
      />
      <NumberInputField
        id="health_prevention_lookback_days"
        name="health_prevention_lookback_days"
        label="健診・来院履歴ルックバック日数"
        defaultValue={settings?.health_prevention_lookback_days ?? 365}
      />
      <NumberInputField
        id="vaccine_deadline_days"
        name="vaccine_deadline_days"
        label="ワクチン期限日数"
        defaultValue={settings?.vaccine_deadline_days ?? 60}
      />
    </>
  );
}
