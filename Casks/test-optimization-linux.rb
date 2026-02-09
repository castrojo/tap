# typed: strict
# frozen_string_literal: true

cask "test-optimization-linux" do
  version "1.0.0"
  sha256 "abc123def456abc123def456abc123def456abc123def456abc123def456abcd"

  url "https://github.com/BurntSushi/ripgrep/releases/download/#{version}/ripgrep-#{version}-x86_64-unknown-linux-musl.tar.gz"
  name "Test Optimization"
  desc "Test cask for CI optimization experiment"
  homepage "https://github.com/castrojo/tap"

  # Linux-only cask
  depends_on formula: "bash"

  binary "rg", target: "test-rg"

  zap trash: [
    "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/test-optimization",
    "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/test-optimization",
    "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/test-optimization",
  ]
end
# Test run 2
