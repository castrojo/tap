cask "sublime-text-linux" do
  version "4200"
  sha256 "36f69c551ad18ee46002be4d9c523fe545d93b67fea67beea731e724044b469f"

  # Linux x64 tarball (Priority 1 format - preferred)
  # Verified: SHA256 calculated from official download
  # Platform: Linux x86_64 only
  url "https://download.sublimetext.com/sublime_text_build_#{version}_x64.tar.xz"
  name "Sublime Text"
  desc "Sophisticated text editor for code, markup and prose"
  homepage "https://www.sublimetext.com/"

  binary "sublime_text/sublime_text", target: "subl"
  # Install desktop file and icon for GUI launcher integration
  artifact "sublime_text/sublime_text.desktop",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/sublime-text.desktop"
  artifact "sublime_text/Icon/128x128/sublime-text.png",
           target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/sublime-text.png"

  preflight do
    # Ensure directories exist using XDG environment variables
    xdg_data_home = ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
    FileUtils.mkdir_p "#{xdg_data_home}/applications"
    FileUtils.mkdir_p "#{xdg_data_home}/icons"

    # Fix Exec path in desktop file to point to Homebrew binary
    desktop_file = "#{staged_path}/sublime_text/sublime_text.desktop"
    if File.exist?(desktop_file)
      content = File.read(desktop_file)
      # Replace hardcoded /opt path with Homebrew path
      updated_content = content.gsub(%r{/opt/sublime_text/sublime_text}, "#{HOMEBREW_PREFIX}/bin/subl")
      File.write(desktop_file, updated_content)
    end
  end

  zap trash: [
    "#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/sublime-text",
    "#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/sublime-text",
  ]
end
