cask "sublime-text" do
  version "4200"
  sha256 "36f69c551ad18ee46002be4d9c523fe545d93b67fea67beea731e724044b469f"

  url "https://download.sublimetext.com/sublime_text_build_#{version}_x64.tar.xz"
  name "Sublime Text"
  desc "Sophisticated text editor for code, markup and prose"
  homepage "https://www.sublimetext.com/"

  binary "sublime_text/sublime_text", target: "subl"
end
