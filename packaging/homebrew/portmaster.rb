class Portmaster < Formula
  desc "Fast, cross-platform port and process management for developers"
  homepage "https://github.com/RichardFlp/portmaster"
  url "https://github.com/RichardFlp/portmaster/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "503b4fd86c8c4b220644d57018ee60cfcc1534a5701f5947a6d2335fd293253a"
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
