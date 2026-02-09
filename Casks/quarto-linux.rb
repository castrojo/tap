cask "quarto-linux" do
  version "1.8.27"
  sha256 "bdf689b5589789a1f21d89c3b83d78ed02a97914dd702e617294f2cc1ea7387d"

  url "https://github.com/quarto-dev/quarto-cli/releases/download/v#{version}/quarto-#{version}-linux-amd64.tar.gz"
  name "Quarto"
  desc "Open-source scientific and technical publishing system built on Pandoc"
  homepage "https://quarto.org/"

  binary "quarto-#{version}/bin/quarto"
end
