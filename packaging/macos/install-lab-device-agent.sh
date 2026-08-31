#!/bin/sh
set -eu

bundle_dir=$(unset CDPATH; cd -- "$(dirname -- "$0")" && pwd)
install_dir="$HOME/Library/Application Support/AnimalEkarte"
launch_agents_dir="$HOME/Library/LaunchAgents"
binary_path="$install_dir/lab-device-agent"
binary_tmp="$install_dir/.lab-device-agent-bin.$$"
plist_path="$launch_agents_dir/com.animalekarte.lab-device-agent.plist"
ports_file="$install_dir/lab-device-agent-ports"

cd "$bundle_dir"
shasum -a 256 -c SHA256SUMS >/dev/null
clinic_id=$(sed -n '1p' "$bundle_dir/lab-device-agent.conf")
allowed_origin=$(sed -n '2p' "$bundle_dir/lab-device-agent.conf")
config_lines=$(wc -l < "$bundle_dir/lab-device-agent.conf" | tr -d ' ')
if [ -z "$clinic_id" ] || [ -z "$allowed_origin" ] || [ "$config_lines" -ne 2 ]; then
  echo "Bundle configuration is invalid" >&2
  exit 1
fi

case "$(uname -m)" in
  arm64|x86_64) lipo lab-device-agent -verify_arch "$(uname -m)" ;;
  *) echo "このMacのCPUには対応していません。" >&2; exit 1 ;;
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
  echo "NX600/AU10V専用USBを2本接続してから、もう一度実行してください。" >&2
  exit 1
fi
chmod 600 "$ports_tmp"

cp "$bundle_dir/lab-device-agent" "$binary_tmp"
chmod 700 "$binary_tmp"
cp "$bundle_dir/com.animalekarte.lab-device-agent.plist" "$plist_tmp"
"$bundle_dir/configure-plist.sh" "$plist_tmp" "$binary_path" "$clinic_id" "$ports_file" "$allowed_origin" \
  "$install_dir/lab-device-agent.log" "$install_dir/lab-device-agent.error.log"
chmod 600 "$plist_tmp"

mv "$ports_tmp" "$ports_file"
mv "$binary_tmp" "$binary_path"
mv "$plist_tmp" "$plist_path"
launchctl bootout "gui/$(id -u)/com.animalekarte.lab-device-agent" 2>/dev/null || true
launchctl bootstrap "gui/$(id -u)" "$plist_path"
launchctl kickstart -k "gui/$(id -u)/com.animalekarte.lab-device-agent"

echo "インストールが完了しました。検査機器画面を開いてください。"
