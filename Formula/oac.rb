class Oac < Formula
  desc "Guided macOS cleaner for leftover AI assistant files"
  homepage "https://github.com/carlisle0615/OpenAgentCleaner"
  license "MIT"
  version "0.2.0"

  on_arm do
    url "https://github.com/carlisle0615/OpenAgentCleaner/releases/download/v#{version}/oac_0.2.0_darwin_arm64.tar.gz"
    sha256 "46451bfe861f9e9d0c50f824519382477697a3e9c28e46e1f2dccb737b429206"
  end

  on_intel do
    url "https://github.com/carlisle0615/OpenAgentCleaner/releases/download/v#{version}/oac_0.2.0_darwin_amd64.tar.gz"
    sha256 "7070ba96ed1268279ef2bc833ae4f6ffa47dc8a1d77dbf345ceb49a53e64a278"
  end

  def install
    bin.install "oac"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/oac version")
  end
end
