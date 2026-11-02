#!/usr/bin/env bash
set -euo pipefail
version=$1
checksums=$2
base="https://github.com/runtimeconditions/go-rc-cli/releases/download/v$version"
url() { printf '%s/rc_%s_%s.tar.gz' "$base" "$version" "$1"; }
sum() { awk -v f="rc_${version}_$1.tar.gz" '$2 == f { print $1 }' "$checksums"; }
platform() {
  printf '    on_%s do\n' "$1"
  printf '      url "%s"\n      sha256 "%s"\n    end\n' "$(url "$2")" "$(sum "$2")"
}
cat <<RUBY
class Rc < Formula
  desc "Runtime Conditions command line"
  homepage "https://github.com/runtimeconditions/go-rc-cli"
  version "$version"
  license "MIT"

  depends_on "rc-profiler-go"

  on_macos do
$(platform arm darwin_arm64)
$(platform intel darwin_amd64)
  end

  on_linux do
$(platform arm linux_arm64)
$(platform intel linux_amd64)
  end

  def install
    bin.install "rc"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/rc --version")
  end
end
RUBY
