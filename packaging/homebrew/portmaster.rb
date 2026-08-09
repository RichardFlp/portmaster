class Portmaster < Formula
  desc "Fast, cross-platform port and process management for developers"
  homepage "https://github.com/RichardFlp/portmaster"
  url "https://github.com/RichardFlp/portmaster/archive/refs/tags/v1.1.1.tar.gz"
  sha256 "9aab2b2c3c315dafb9b5b685ad259ea4e70ba75a0e13a1c99197b1767715219e"
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
