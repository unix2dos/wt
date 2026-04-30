class Ww < Formula
  desc "Worktree primitive your AI agents and you share"
  homepage "https://github.com/unix2dos/ww"
  version "0.11.0"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/unix2dos/ww/releases/download/v0.11.0/ww-v0.11.0-darwin-arm64.tar.gz"
    sha256 "03fee0f522ae4d7d028e0ab4413487099fec0287843fd8b71b53e3919cbc0ed7"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/unix2dos/ww/releases/download/v0.11.0/ww-v0.11.0-darwin-amd64.tar.gz"
    sha256 "08264be6b7a7c658a68c9c239700aecd2488fc54ab123b2d0ea006289325fdf3"
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/unix2dos/ww/releases/download/v0.11.0/ww-v0.11.0-linux-arm64.tar.gz"
    sha256 "4a1939d29c640625fd1ddf5d3cc97d3d8017119bdf9d9102f2a99b3c7c8f6ae1"
  elsif OS.linux? && Hardware::CPU.intel?
    url "https://github.com/unix2dos/ww/releases/download/v0.11.0/ww-v0.11.0-linux-amd64.tar.gz"
    sha256 "39b851a0dec59674484b83bac5e4d30809d64903b3cfdc543429cd7165254a22"
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
