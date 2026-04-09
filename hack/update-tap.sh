#!/usr/bin/env bash
# hack/update-tap.sh — Update jholm117/homebrew-tap formula after a GitHub release.
# Usage: hack/update-tap.sh v0.1.3
set -euo pipefail

VERSION="${1:?Usage: hack/update-tap.sh <tag> (e.g. v0.1.3)}"
REPO="jholm117/hackerrank-cli"
TAP_DIR="${TAP_DIR:-$HOME/repos/homebrew-tap}"
FORMULA="$TAP_DIR/Formula/hr.rb"

if [[ ! -f "$FORMULA" ]]; then
    echo "Error: $FORMULA not found. Clone jholm117/homebrew-tap to ~/repos/homebrew-tap" >&2
    exit 1
fi

echo "==> Downloading checksums for $VERSION..."
CHECKSUMS=$(gh release download "$VERSION" --repo "$REPO" --pattern "checksums.txt" --output -)
if [[ -z "$CHECKSUMS" ]]; then
    echo "Error: No checksums.txt in release $VERSION" >&2
    exit 1
fi

STRIP_V="${VERSION#v}"

darwin_arm64=$(echo "$CHECKSUMS" | grep "darwin_arm64" | awk '{print $1}')
darwin_amd64=$(echo "$CHECKSUMS" | grep "darwin_amd64" | awk '{print $1}')
linux_arm64=$(echo "$CHECKSUMS" | grep "linux_arm64" | awk '{print $1}')
linux_amd64=$(echo "$CHECKSUMS" | grep "linux_amd64" | awk '{print $1}')

echo "==> Updating $FORMULA to $STRIP_V..."
cat > "$FORMULA" << EOF
class Hr < Formula
  desc "CLI for HackerRank for Work API"
  homepage "https://github.com/$REPO"
  version "$STRIP_V"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/$REPO/releases/download/$VERSION/hr_${STRIP_V}_darwin_arm64.tar.gz"
      sha256 "$darwin_arm64"
    else
      url "https://github.com/$REPO/releases/download/$VERSION/hr_${STRIP_V}_darwin_amd64.tar.gz"
      sha256 "$darwin_amd64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/$REPO/releases/download/$VERSION/hr_${STRIP_V}_linux_arm64.tar.gz"
      sha256 "$linux_arm64"
    else
      url "https://github.com/$REPO/releases/download/$VERSION/hr_${STRIP_V}_linux_amd64.tar.gz"
      sha256 "$linux_amd64"
    end
  end

  def install
    bin.install "hr"
  end

  test do
    assert_match "CLI for HackerRank", shell_output("#{bin}/hr --help")
  end
end
EOF

echo "==> Committing and pushing..."
cd "$TAP_DIR"
git add Formula/hr.rb
git commit -m "Update hr to $STRIP_V"
git push

echo "==> Done. Run: brew upgrade jholm117/tap/hr"
