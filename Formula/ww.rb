class Ww < Formula
  desc "Worktree primitive your AI agents and you share"
  homepage "https://github.com/unix2dos/ww"
  version "0.11.2"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/unix2dos/ww/releases/download/v0.11.2/ww-v0.11.2-darwin-arm64.tar.gz"
    sha256 "fb5a5bef11f3954518af98cb955a345b573f66535843bb244b86aa27f74de072"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/unix2dos/ww/releases/download/v0.11.2/ww-v0.11.2-darwin-amd64.tar.gz"
    sha256 "d1644e013deda08a782a6482bec64082df2a8d72577bbe7947c4dfcb94936040"
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/unix2dos/ww/releases/download/v0.11.2/ww-v0.11.2-linux-arm64.tar.gz"
    sha256 "fbc88623c56b3789df874ce0d07bf02be6f0cb89f763ff730141af5aaef5c28e"
  elsif OS.linux? && Hardware::CPU.intel?
    url "https://github.com/unix2dos/ww/releases/download/v0.11.2/ww-v0.11.2-linux-amd64.tar.gz"
    sha256 "97ce85eaf792302bcb7714c450aa2ac09ddc51761b2317d0675baac460d18127"
  end

  def install
    bin.install "bin/ww-helper"
    libexec.install "shell/ww.sh"
    doc.install "README.md"
  end

  def caveats
    <<~EOS
      `ww` changes the current shell directory, so Homebrew installs the helper and shell library
      but leaves shell activation to you.

      Add one line to your shell rc file:

      For zsh:
        eval "$("#{opt_bin}/ww-helper" init zsh)"

      For bash:
        eval "$("#{opt_bin}/ww-helper" init bash)"
    EOS
  end

  test do
    assert_path_exists libexec/"ww.sh"
    assert_match "Usage: ww-helper", shell_output("#{bin}/ww-helper help")
  end
end
