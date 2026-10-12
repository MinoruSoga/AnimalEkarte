#!/usr/bin/env bash

ci_py_probe() {
  "$1" -c 'import sys; sys.exit(0 if sys.version_info >= (3, 11) else 1)' >/dev/null 2>&1
}

ci_py_resolve() {
  if [[ -n "${CI_PYTHON:-}" ]]; then
    if ! command -v "$CI_PYTHON" >/dev/null 2>&1; then
      echo "ERROR: CI_PYTHON='$CI_PYTHON' is not executable. CI_PYTHON には Python 3.11+ へのパスを指定してください。" >&2
      return 1
    fi
    if ! ci_py_probe "$CI_PYTHON"; then
      echo "ERROR: CI_PYTHON='$CI_PYTHON' is not Python >= 3.11 (make ci の Python スクリプトは tomllib 等 3.11+ を要求)。" >&2
      return 1
    fi
    printf '%s\n' "$CI_PYTHON"
    return 0
  fi
  local c
  for c in python3 python3.14 python3.13 python3.12 python3.11; do
    command -v "$c" >/dev/null 2>&1 || continue
    if ci_py_probe "$c"; then
      printf '%s\n' "$c"
      return 0
    fi
  done
  echo "ERROR: Python >= 3.11 が見つかりません（make ci の Python スクリプトは tomllib 等 3.11+ を要求）。" >&2
  echo "       python3.11+ を導入するか、CI_PYTHON=/path/to/python3.x で明示指定してください。" >&2
  return 1
}

ci_py() {
  local py
  py="$(ci_py_resolve)" || return 1
  "$py" "$@"
}
