#!/bin/sh
set -eu

repo_dir=$(unset CDPATH; cd -- "$(dirname -- "$0")/.." && pwd)
clinic_id=${1:-}
allowed_origin=${2:-}
output_dir=${3:-}
if [ -z "$clinic_id" ] || [ -z "$allowed_origin" ] || [ -z "$output_dir" ]; then
  echo "Usage: $0 <clinic-id> <allowed-https-origin> <new-output-directory>" >&2
  exit 2
fi
case "$clinic_id" in
  *[!0-9]*) echo "Clinic ID must contain digits only" >&2; exit 2 ;;
esac
if ! node - "$allowed_origin" <<'NODE'
const raw = process.argv[2].trim();
if (!/^https:\/\/[^/?#]+$/.test(raw)) process.exit(1);
try {
  const parsed = new URL(raw);
  const authority = raw.slice("https://".length);
  const port = parsed.port;
  if (
    parsed.protocol !== "https:" ||
    parsed.username !== "" ||
    parsed.password !== "" ||
    parsed.hostname === "" ||
    parsed.hostname.includes("*") ||
    authority.endsWith(":") ||
    (port !== "" && (Number(port) < 1 || Number(port) > 65535))
  ) process.exit(1);
} catch {
  process.exit(1);
}
NODE
then
  echo "Allowed origin must be an exact https origin with no credentials and a valid optional numeric port" >&2
  exit 2
fi
if [ -e "$output_dir" ]; then
  echo "Output path already exists: $output_dir" >&2
  exit 1
fi

work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT HUP INT TERM
build_id=$(basename "$work_dir" | tr -cd '[:alnum:]')
container_arm64="/tmp/animalekarte-lab-device-agent-$build_id-arm64"
container_amd64="/tmp/animalekarte-lab-device-agent-$build_id-amd64"

cd "$repo_dir"
docker compose exec -T backend env GOOS=darwin GOARCH=arm64 go build -o "$container_arm64" ./cmd/lab-device-agent
docker compose exec -T backend env GOOS=darwin GOARCH=amd64 go build -o "$container_amd64" ./cmd/lab-device-agent
docker compose cp "backend:$container_arm64" "$work_dir/lab-device-agent-arm64"
docker compose cp "backend:$container_amd64" "$work_dir/lab-device-agent-amd64"

mkdir -p "$output_dir"
lipo -create \
  "$work_dir/lab-device-agent-arm64" \
  "$work_dir/lab-device-agent-amd64" \
  -output "$output_dir/lab-device-agent"
chmod 700 "$output_dir/lab-device-agent"

cp "$repo_dir/packaging/macos/com.animalekarte.lab-device-agent.plist" "$output_dir/com.animalekarte.lab-device-agent.plist"
printf '%s\n%s\n' "$clinic_id" "$allowed_origin" > "$output_dir/lab-device-agent.conf"
chmod 600 "$output_dir/lab-device-agent.conf"
cp "$repo_dir/packaging/macos/configure-lab-device-agent-plist.sh" "$output_dir/configure-plist.sh"
chmod 700 "$output_dir/configure-plist.sh"
cp "$repo_dir/packaging/macos/install-lab-device-agent.sh" "$output_dir/install.sh"
chmod 700 "$output_dir/install.sh"
cp "$repo_dir/packaging/macos/diagnose-lab-device-agent.sh" "$output_dir/diagnose.sh"
chmod 700 "$output_dir/diagnose.sh"

(
  cd "$output_dir"
  shasum -a 256 lab-device-agent lab-device-agent.conf com.animalekarte.lab-device-agent.plist configure-plist.sh install.sh diagnose.sh > SHA256SUMS
)
manifest_sha=$(shasum -a 256 "$output_dir/SHA256SUMS" | awk '{print $1}')
echo "Created client bundle: $output_dir"
echo "Send this manifest SHA-256 through a separate trusted channel: $manifest_sha"
