#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
clinic_id=${1:-}
if [ -z "$clinic_id" ]; then
  echo "Usage: $0 <clinic-id>" >&2
  exit 2
fi
case "$clinic_id" in
  *[!0-9]*) echo "Clinic ID must contain digits only" >&2; exit 2 ;;
esac
install_dir="$HOME/Library/Application Support/AnimalEkarte"
launch_agents_dir="$HOME/Library/LaunchAgents"
binary_path="$install_dir/lab-device-agent"
binary_tmp="$install_dir/.lab-device-agent-bin.$$"
plist_path="$launch_agents_dir/com.animalekarte.lab-device-agent.plist"
ports_file="$install_dir/lab-device-agent-ports"
container_binary="/tmp/animalekarte-lab-device-agent"

case "$(uname -m)" in
  arm64) goarch=arm64 ;;
  x86_64) goarch=amd64 ;;
  *) echo "Unsupported Mac architecture" >&2; exit 1 ;;
esac

mkdir -p "$install_dir" "$launch_agents_dir"
chmod 700 "$install_dir"

ports_tmp=$(mktemp "$install_dir/.lab-device-agent-ports.XXXXXX")
plist_tmp=$(mktemp "$launch_agents_dir/.lab-device-agent-plist.XXXXXX")
trap 'rm -f "$ports_tmp" "$plist_tmp" "$binary_tmp"' EXIT HUP INT TERM
port_count=0
for port in /dev/cu.usbserial-*; do
  if [ -e "$port" ]; then
    printf '%s\n' "$port" >> "$ports_tmp"
    port_count=$((port_count + 1))
  fi
done
if [ "$port_count" -ne 2 ]; then
  echo "Expected exactly the two verified NX600/AU10V serial adapters; found $port_count" >&2
  exit 1
fi
chmod 600 "$ports_tmp"

cd "$repo_dir"
docker compose exec -T backend env GOOS=darwin GOARCH="$goarch" go build -o "$container_binary" ./cmd/lab-device-agent
docker compose cp "backend:$container_binary" "$binary_tmp"
chmod 700 "$binary_tmp"

cp "$repo_dir/packaging/macos/com.animalekarte.lab-device-agent.plist" "$plist_tmp"
plutil -replace ProgramArguments.0 -string "$binary_path" "$plist_tmp"
plutil -replace ProgramArguments.2 -string "$clinic_id" "$plist_tmp"
plutil -replace ProgramArguments.4 -string "$ports_file" "$plist_tmp"
plutil -replace StandardOutPath -string "$install_dir/lab-device-agent.log" "$plist_tmp"
plutil -replace StandardErrorPath -string "$install_dir/lab-device-agent.error.log" "$plist_tmp"
plutil -lint "$plist_tmp" >/dev/null
chmod 600 "$plist_tmp"
mv "$ports_tmp" "$ports_file"
mv "$binary_tmp" "$binary_path"
mv "$plist_tmp" "$plist_path"

launchctl bootout "gui/$(id -u)/com.animalekarte.lab-device-agent" 2>/dev/null || true
launchctl bootstrap "gui/$(id -u)" "$plist_path"
launchctl kickstart -k "gui/$(id -u)/com.animalekarte.lab-device-agent"

echo "Lab device agent installed. Open http://localhost:3003/lab-device"
