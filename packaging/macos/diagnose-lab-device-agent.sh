#!/bin/sh
set -eu

label="gui/$(id -u)/com.animalekarte.lab-device-agent"
if ! launchctl print "$label" >/dev/null 2>&1; then
  echo "受信機: 停止中"
  exit 2
fi

health=$(curl -fsS --max-time 2 http://127.0.0.1:17654/health 2>/dev/null || true)
if [ -z "$health" ]; then
  echo "受信機: 起動中ですが状態を取得できません"
  exit 2
fi

case "$health" in
  *'"status":"running"'*)
    echo "受信機: 正常"
    printf '%s\n' "$health"
    exit 0
    ;;
  *'"status":"degraded"'*)
    echo "受信機: 要確認"
    printf '%s\n' "$health"
    exit 1
    ;;
  *)
    echo "受信機: 不明な応答"
    exit 2
    ;;
esac
