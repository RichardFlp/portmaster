class Portmaster < Formula
  desc "Fast, cross-platform port and process management for developers"
  homepage "https://github.com/RichardFlp/portmaster"
  url "https://github.com/RichardFlp/portmaster/archive/refs/tags/v1.1.0.tar.gz"
  sha256 "f26e84cf008ef8362e84b717b212f1ffed9ca25fe0a782f88957a9198264b843"
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
