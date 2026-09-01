#!/bin/sh
set -eu

repo_dir=$(unset CDPATH; cd -- "$(dirname -- "$0")/.." && pwd)
clinic_id=${1:-}
allowed_origin=${2:-}
if [ -z "$clinic_id" ] || [ -z "$allowed_origin" ]; then
  echo "Usage: $0 <clinic-id> <allowed-https-origin>" >&2
  exit 2
fi
case "$clinic_id" in
  *[!0-9]*) echo "Clinic ID must contain digits only" >&2; exit 2 ;;
esac
if ! allowed_origin=$(node "$repo_dir/scripts/canonicalize-lab-device-origin.mjs" "$allowed_origin"); then
  echo "Allowed origin must use lowercase https with canonical IPv4, non-mapped IPv6, or strict ASCII DNS and an optional numeric port" >&2
  exit 2
fi
install_dir="$HOME/Library/Application Support/AnimalEkarte"
launch_agents_dir="$HOME/Library/LaunchAgents"
binary_path="$install_dir/lab-device-agent"
binary_tmp="$install_dir/.lab-device-agent-bin.$$"
plist_path="$launch_agents_dir/com.animalekarte.lab-device-agent.plist"
ports_file="$install_dir/lab-device-agent-ports"
container_binary="/tmp/animalekarte-lab-device-agent"
service_target="gui/$(id -u)/com.animalekarte.lab-device-agent"
device_dir=${LAB_DEVICE_AGENT_DEVICE_DIR:-/dev}

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
for port in "$device_dir"/cu.usbserial-*; do
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
"$repo_dir/packaging/macos/configure-lab-device-agent-plist.sh" "$plist_tmp" "$binary_path" "$clinic_id" "$ports_file" "$allowed_origin" \
  "$install_dir/lab-device-agent.log" "$install_dir/lab-device-agent.error.log"
chmod 600 "$plist_tmp"

backup_dir=$(mktemp -d "$install_dir/.lab-device-agent-rollback.XXXXXX")
had_binary=0
had_ports=0
had_plist=0
service_was_loaded=0
service_was_running=0
if [ -e "$binary_path" ]; then cp -p "$binary_path" "$backup_dir/binary"; had_binary=1; fi
if [ -e "$ports_file" ]; then cp -p "$ports_file" "$backup_dir/ports"; had_ports=1; fi
if [ -e "$plist_path" ]; then cp -p "$plist_path" "$backup_dir/plist"; had_plist=1; fi
if launchctl print "$service_target" > "$backup_dir/launchctl.print" 2>/dev/null; then
  service_was_loaded=1
  if grep -q 'state = running' "$backup_dir/launchctl.print"; then service_was_running=1; fi
fi

restore_previous_install() {
  rollback_failed=0

  launchctl bootout "$service_target" 2>/dev/null || true
  if [ "$had_binary" -eq 1 ]; then
    if ! cp -p "$backup_dir/binary" "$binary_path"; then
      echo "Rollback failed: could not restore the previous agent binary at $binary_path" >&2
      rollback_failed=1
    fi
  elif ! rm -f "$binary_path"; then
    echo "Rollback failed: could not remove the new agent binary at $binary_path" >&2
    rollback_failed=1
  fi
  if [ "$had_ports" -eq 1 ]; then
    if ! cp -p "$backup_dir/ports" "$ports_file"; then
      echo "Rollback failed: could not restore the previous serial-port configuration at $ports_file" >&2
      rollback_failed=1
    fi
  elif ! rm -f "$ports_file"; then
    echo "Rollback failed: could not remove the new serial-port configuration at $ports_file" >&2
    rollback_failed=1
  fi
  if [ "$had_plist" -eq 1 ]; then
    if ! cp -p "$backup_dir/plist" "$plist_path"; then
      echo "Rollback failed: could not restore the previous LaunchAgent plist at $plist_path" >&2
      rollback_failed=1
    fi
  elif ! rm -f "$plist_path"; then
    echo "Rollback failed: could not remove the new LaunchAgent plist at $plist_path" >&2
    rollback_failed=1
  fi
  if [ "$service_was_loaded" -eq 1 ] && [ "$had_plist" -eq 1 ]; then
    if ! launchctl bootstrap "gui/$(id -u)" "$plist_path"; then
      echo "Rollback failed: could not bootstrap the previous LaunchAgent at $plist_path" >&2
      rollback_failed=1
    elif [ "$service_was_running" -eq 1 ] && ! launchctl kickstart -k "$service_target"; then
      echo "Rollback failed: could not restart the previous LaunchAgent $service_target" >&2
      rollback_failed=1
    fi
  fi

  return "$rollback_failed"
}

mv "$ports_tmp" "$ports_file"
mv "$binary_tmp" "$binary_path"
mv "$plist_tmp" "$plist_path"

launchctl bootout "$service_target" 2>/dev/null || true
if ! launchctl bootstrap "gui/$(id -u)" "$plist_path"; then
  if ! restore_previous_install; then
    echo "Installation activation failed and the previous installation was not fully recovered." >&2
  fi
  exit 1
fi
if ! launchctl kickstart -k "$service_target"; then
  if ! restore_previous_install; then
    echo "Installation activation failed and the previous installation was not fully recovered." >&2
  fi
  exit 1
fi
rm -rf "$backup_dir"

echo "Lab device agent installed. Open http://localhost:3003/lab-device"
