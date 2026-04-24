#!/bin/sh
set -e

# Airでホットリロード起動
exec air -c .air.toml
