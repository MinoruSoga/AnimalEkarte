#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
clinic_id=${1:-}
output_dir=${2:-}
if [ -z "$clinic_id" ] || [ -z "$output_dir" ]; then
  echo "Usage: $0 <clinic-id> <new-output-directory>" >&2
  exit 2
fi
case "$clinic_id" in
  *[!0-9]*) echo "Clinic ID must contain digits only" >&2; exit 2 ;;
esac
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
plutil -replace ProgramArguments.2 -string "$clinic_id" "$output_dir/com.animalekarte.lab-device-agent.plist"
plutil -lint "$output_dir/com.animalekarte.lab-device-agent.plist" >/dev/null
cp "$repo_dir/packaging/macos/install-lab-device-agent.sh" "$output_dir/install.sh"
chmod 700 "$output_dir/install.sh"
cp "$repo_dir/packaging/macos/diagnose-lab-device-agent.sh" "$output_dir/diagnose.sh"
chmod 700 "$output_dir/diagnose.sh"

(
  cd "$output_dir"
  shasum -a 256 lab-device-agent com.animalekarte.lab-device-agent.plist install.sh diagnose.sh > SHA256SUMS
)
manifest_sha=$(shasum -a 256 "$output_dir/SHA256SUMS" | awk '{print $1}')
echo "Created client bundle: $output_dir"
echo "Send this manifest SHA-256 through a separate trusted channel: $manifest_sha"
