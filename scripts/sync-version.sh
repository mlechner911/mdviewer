#!/usr/bin/env bash
# Writes the version from VERSION into every file that carries it, so the
# app, the installers, the packages and the product page agree. Run through
# `task version:sync` (also part of build and release) or
# `task version:set -- 1.7.0`. Idempotent; prints the files it changed.
# With --check nothing is written; it fails when a file is out of step.
#
# Not touched on purpose: CHANGELOG.md (written by hand per release), the
# npm packages in packages/ and the Go module mlcext (their own versions).
set -euo pipefail
cd "$(dirname "$0")/.."

V=$(tr -d ' \t\r\n' < VERSION)
if ! [[ "$V" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "VERSION must be MAJOR.MINOR.PATCH, got '$V'" >&2
  exit 1
fi

CHECK=0
[ "${1:-}" = "--check" ] && CHECK=1

changed=()
# edit FILE PERL-EXPRESSION: apply (or, with --check, compare), remember
# the file when it changed.
edit() {
  local f=$1 expr=$2
  [ -f "$f" ] || return 0
  local out
  out=$(V="$V" perl -0p -e "$expr" "$f")
  if [ "$out" != "$(cat "$f")" ]; then
    changed+=("$f")
    [ $CHECK -eq 1 ] || printf '%s\n' "$out" > "$f"
  fi
}

# Wails project config (wails3 update build-assets reads it).
edit build/config.yml 's/(\n  version: )"[^"]*"/$1"$ENV{V}"/'
# Windows: resource info of the exe and the NSIS installer.
edit build/windows/info.json 's/("file_version": )"[^"]*"/$1"$ENV{V}"/; s/("ProductVersion": )"[^"]*"/$1"$ENV{V}"/'
edit build/windows/nsis/wails_tools.nsh 's/(!define INFO_PRODUCTVERSION )"[^"]*"/$1"$ENV{V}"/'
# macOS bundle.
for f in build/darwin/Info.plist build/darwin/Info.dev.plist; do
  edit "$f" 's#(<key>CFBundleShortVersionString</key>\s*<string>)[^<]*#$1$ENV{V}#; s#(<key>CFBundleVersion</key>\s*<string>)[^<]*#$1$ENV{V}#'
done
# Linux packages (deb, rpm, AppImage via nfpm).
edit build/linux/nfpm/nfpm.yaml 's/(\nversion: )"[^"]*"/$1"$ENV{V}"/'
# Frontend package (shown nowhere, kept in step).
edit frontend/package.json 's/("version": )"[^"]*"/$1"$ENV{V}"/'
# README: versioned file names of the downloads.
edit README.md 's/^(# MarkSafe )\d+\.\d+\.\d+/$1$ENV{V}/m; s/marksafe-\d+\.\d+\.\d+-/marksafe-$ENV{V}-/g'
# Product page (own repository in mlcprodweb/, if checked out).
edit mlcprodweb/meta.yaml 's/^(version: )"[^"]*"/$1"$ENV{V}"/m'

if [ ${#changed[@]} -eq 0 ]; then
  echo "version $V: all files up to date"
elif [ $CHECK -eq 1 ]; then
  echo "version $V missing in (run: task version:sync):" >&2
  printf '  %s\n' "${changed[@]}" >&2
  exit 1
else
  echo "version $V written to:"
  printf '  %s\n' "${changed[@]}"
fi
