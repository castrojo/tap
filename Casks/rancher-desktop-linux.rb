cask "rancher-desktop-linux" do
  version "1.22.0"
  sha256 "081bc82ac988b1467f6445dddb483395ca7b1aac2164594fd5f4e2cb7344ba6d"

  url "https://github.com/rancher-sandbox/rancher-desktop/releases/download/v#{version}/rancher-desktop-linux-v#{version}.zip"
  name "Rancher Desktop"
  desc "Kubernetes and container management on the desktop"
  homepage "https://rancherdesktop.io/"

  binary "rancher-desktop"
  artifact "resources/resources/linux/rancher-desktop.desktop",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/rancher-desktop.desktop"
  artifact "resources/resources/icons/logo-square-512.png",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/rancher-desktop.png"

  preflight do
    xdg_data_home = ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
    FileUtils.mkdir_p "#{xdg_data_home}/applications"
    FileUtils.mkdir_p "#{xdg_data_home}/icons"

    desktop_file = "#{staged_path}/resources/resources/linux/rancher-desktop.desktop"
    if File.exist?(desktop_file)
      content = File.read(desktop_file)
      updated_content = content.gsub(/^Exec=rancher-desktop$/, "Exec=#{HOMEBREW_PREFIX}/bin/rancher-desktop")
      updated_content = updated_content.gsub(/^Icon=rancher-desktop$/, "Icon=#{xdg_data_home}/icons/rancher-desktop.png")
      File.write(desktop_file, updated_content)
    end
  end

  zap trash: [
    "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/rancher-desktop",
    "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/rancher-desktop",
    "#{Dir.home}/.local/share/rancher-desktop",
  ]
end
