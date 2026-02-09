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
end
