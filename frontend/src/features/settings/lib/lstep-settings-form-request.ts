import type { LstepSettingsRequest } from "../hooks/use-lstep-settings";

const SECRET_FIELDS = [
  "lstep_api_key",
  "line_channel_access_token",
  "line_channel_secret",
] as const;

const TEXT_FIELDS = ["liff_id", "lstep_base_url"] as const;

const NUMERIC_FIELDS = [
  "dormant_prevention_180_days",
  "dormant_prevention_210_days",
  "dormant_prevention_240_days",
  "dormant_prevention_365_days",
  "health_prevention_lookback_days",
  "vaccine_deadline_days",
  "cpm_v2_coming_threshold",
  "cpm_v2_good_threshold",
  "cpm_v2_family_threshold",
  "cpm_v2_noah_threshold",
  "cpm_v1_dormant_days",
  "cpm_v1_noah_days",
  "cpm_v1_noah_annual_visits",
  "cpm_v1_noah_ltv",
  "cpm_v1_core_days",
  "cpm_v1_core_annual_visits",
  "cpm_v1_core_ltv",
  "cpm_v1_spot_min_amount",
  "cpm_v1_spot_inactive_days",
  "cpm_v1_growing_max_days",
  "cpm_v1_growing_min_visits",
  "cpm_v1_growing_max_visits",
  "cpm_v1_ltv_break_low",
] as const;

type SecretField = (typeof SECRET_FIELDS)[number];
type TextField = (typeof TEXT_FIELDS)[number];
type NumericField = (typeof NUMERIC_FIELDS)[number];

export function buildLstepSettingsRequest(
  formData: FormData,
  isSyncEnabled: boolean,
): LstepSettingsRequest {
  const req: LstepSettingsRequest = {};

  for (const field of SECRET_FIELDS) {
    setTrimmedString(req, field, formData, true);
  }
  for (const field of TEXT_FIELDS) {
    setTrimmedString(req, field, formData, field !== "liff_id");
  }

  const cpmVersion = formData.get("cpm_version");
  if (cpmVersion === "v1" || cpmVersion === "v2") {
    req.cpm_version = cpmVersion;
  }

  for (const field of NUMERIC_FIELDS) {
    setPositiveInteger(req, field, formData);
  }

  req.is_sync_enabled = isSyncEnabled;
  return req;
}

function setTrimmedString(
  req: LstepSettingsRequest,
  field: SecretField | TextField,
  formData: FormData,
  skipEmpty: boolean,
) {
  const value = formData.get(field);
  if (typeof value !== "string") return;
  const trimmed = value.trim();
  if (skipEmpty && trimmed === "") return;
  req[field] = trimmed;
}

function setPositiveInteger(req: LstepSettingsRequest, field: NumericField, formData: FormData) {
  const value = formData.get(field);
  if (typeof value !== "string" || value === "") return;
  const parsed = parseInt(value, 10);
  if (!isNaN(parsed) && parsed >= 1) {
    req[field] = parsed;
  }
}
