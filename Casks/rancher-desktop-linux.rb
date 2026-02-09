# typed: strict
# frozen_string_literal: true

cask "rancher-desktop-linux" do
  version "1.22.0"
  sha256 "081bc82ac988b1467f6445dddb483395ca7b1aac2164594fd5f4e2cb7344ba6d"

  url "https://github.com/rancher-sandbox/rancher-desktop/releases/download/" \
      "v#{version}/rancher-desktop-linux-v#{version}.zip"
  name "Rancher Desktop"
  desc "Container management and Kubernetes on the desktop"
  homepage "https://rancherdesktop.io/"

  depends_on formula: "bash"

  preflight do
    # Create required directories
    xdg_data_home = ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
    [
      "#{xdg_data_home}/applications",
      "#{xdg_data_home}/icons/hicolor/512x512/apps",
      "#{Dir.home}/.local/bin",
    ].each { |dir| FileUtils.mkdir_p(dir) }

    # Fix desktop file paths
    desktop_file = "#{staged_path}/resources/resources/linux/rancher-desktop.desktop"
    if File.exist?(desktop_file)
      content = File.read(desktop_file)
      # Fix Exec path to point to user binary location
      content = content.gsub(
        "Exec=rancher-desktop",
        "Exec=#{Dir.home}/.local/bin/rancher-desktop"
      )
      # Fix Icon path to point to installed icon
      content = content.gsub(
        "Icon=rancher-desktop",
        "Icon=#{xdg_data_home}/icons/hicolor/512x512/apps/rancher-desktop.png"
      )
      File.write(desktop_file, content)
    end
  end

  # Install binary
  binary "rancher-desktop", target: "#{Dir.home}/.local/bin/rancher-desktop"

  # Install desktop file
  artifact "resources/resources/linux/rancher-desktop.desktop",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/rancher-desktop.desktop"

  # Install icon
  artifact "resources/resources/icons/logo-square-512.png",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/" \
                   "icons/hicolor/512x512/apps/rancher-desktop.png"

  zap trash: [
    "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/rancher-desktop",
    "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/rancher-desktop",
    "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/rancher-desktop",
  ]
end
