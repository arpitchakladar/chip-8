{ pkgs, ... }:

{
  languages.go = {
    enable = true;
    lsp = {
      enable = true;
      package = pkgs.gopls;
    };
  };

  packages = with pkgs; [
    git
    pkg-config
    SDL2
    SDL2_ttf
    mesa
  ];

  env.CGO_ENABLED = "1";

  scripts.run-emu.exec = "go run cmd/chip-8/main.go";
  scripts.test-all.exec = "go test ./...";

  pre-commit.hooks = {
    gotest.enable = true;
    golangci-lint.enable = true;
  };
  pre-commit.hooks.editorconfig-checker.enable = true;

  enterShell = ''
    echo "--- Chip-8 Go Development Environment ---"
    go version
    echo "Ready to emulate. Use 'run-emu' to start."
  '';
}
