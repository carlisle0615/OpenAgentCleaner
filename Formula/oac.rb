class Oac < Formula
  desc "Guided macOS cleaner for leftover AI assistant files"
  homepage "https://github.com/carlisle0615/OpenAgentCleaner"
  version "0.1.1"

  on_arm do
    url "https://github.com/carlisle0615/OpenAgentCleaner/releases/download/v#{version}/oac_0.1.1_darwin_arm64.tar.gz"
    sha256 "152832f69676fa51ea04e9024f56d6c6dd58de88941089cfde16dcbd3f8a76c2"
  end

  on_intel do
    url "https://github.com/carlisle0615/OpenAgentCleaner/releases/download/v#{version}/oac_0.1.1_darwin_amd64.tar.gz"
    sha256 "d70466d4d98c6b56e9974b7e7fe740b93af1689bebee49851e6c23f972b4ec12"
  end

  def install
    bin.install "oac"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/oac version")
  end
end
