#!/bin/sh
set -eu

if [ "$#" -ne 8 ]; then
  echo "Usage: $0 <plist> <program> <clinic-id> <ports-file> <allowed-origin> <consumer-token> <stdout-path> <stderr-path>" >&2
  exit 2
fi
plist_path=$1
program_path=$2
clinic_id=$3
ports_file=$4
allowed_origin=$5
consumer_token=$6
stdout_path=$7
stderr_path=$8
plist_buddy=/usr/libexec/PlistBuddy

if [ -z "$consumer_token" ]; then
  echo "Consumer token must not be empty" >&2
  exit 2
fi

"$plist_buddy" -c "Set :ProgramArguments:0 $program_path" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:1 --clinic-id" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:2 $clinic_id" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:3 --ports-file" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:4 $ports_file" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:5 --allowed-origin" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:6 $allowed_origin" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:7 --consumer-token" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:8 $consumer_token" "$plist_path"
"$plist_buddy" -c "Set :StandardOutPath $stdout_path" "$plist_path"
"$plist_buddy" -c "Set :StandardErrorPath $stderr_path" "$plist_path"
plutil -lint "$plist_path" >/dev/null
