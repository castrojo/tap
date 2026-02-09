# typed: strict
# frozen_string_literal: true

cask "rancher-desktop-linux" do
  version "v1.22.0"
  sha256 "081bc82ac988b1467f6445dddb483395ca7b1aac2164594fd5f4e2cb7344ba6d"

  url "https://github.com/rancher-sandbox/rancher-desktop/releases/download/v1.22.0/rancher-desktop-linux-v1.22.0.zip"
  name "rancher-desktop"
  desc "Container Management and Kubernetes on the Desktop"
  homepage "https://rancherdesktop.io/"

  # Linux-only cask
  depends_on formula: "bash"

  binary "rancher-desktop", target: "rancher-desktop"

  zap trash: [
    "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/rancher-desktop",
    "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/rancher-desktop",
    "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/rancher-desktop",
  ]
end
