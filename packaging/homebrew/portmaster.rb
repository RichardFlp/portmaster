class Portmaster < Formula
  desc "Fast, cross-platform port and process management for developers"
  homepage "https://github.com/RichardFlp/portmaster"
  url "https://github.com/RichardFlp/portmaster/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_SOURCE_TARBALL_SHA256"
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/RichardFlp/portmaster/internal/version.Version=#{version}"
    system "go", "build", *std_go_args(output: bin/"portmaster", ldflags: ldflags)
  end

  test do
    assert_match "portmaster v#{version}", shell_output("#{bin}/portmaster version")
  end
end
