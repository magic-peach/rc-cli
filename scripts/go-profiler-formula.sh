#!/usr/bin/env bash
set -euo pipefail
version=$1
checksums=$2
base="https://github.com/runtimeconditions/go-rc-profiler/releases/download/v$version"
platform() {
  printf '    on_%s do\n' "$1"
  printf '      url "%s/go-rc-profiler_v%s_%s.tar.gz"\n' "$base" "$version" "$2"
  printf '      sha256 "%s"\n    end\n' "$(awk -v f="go-rc-profiler_v${version}_$2.tar.gz" '$2 == f { print $1 }' "$checksums")"
}
cat <<RUBY
class RcProfilerGo < Formula
  desc "Runtime Conditions profiler for Go codebases"
  homepage "https://github.com/runtimeconditions/go-rc-profiler"
  version "$version"

  on_macos do
$(platform arm darwin_arm64)
$(platform intel darwin_amd64)
  end

  on_linux do
$(platform arm linux_arm64)
$(platform intel linux_amd64)
  end

  def install
    bin.install "go-rc-profiler"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/go-rc-profiler --version")
  end
end
RUBY
