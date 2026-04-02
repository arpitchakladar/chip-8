{ pkgs, ... }:

let
  libraries = [
    pkgs.xorg.libX11
    pkgs.xorg.libXcursor
    pkgs.xorg.libXinerama
    pkgs.xorg.libXrandr
    pkgs.xorg.libXi
    pkgs.mesa
    pkgs.SDL2
    pkgs.SDL2_ttf
    pkgs.SDL2_gfx
  ];
in
{
  languages.go = {
    enable = true;
    lsp = {
      enable = true;
      package = pkgs.gopls;
    };
  };

  languages.javascript = {
    enable = true;
    npm.enable = true;
    lsp = {
      enable = true;
      package = pkgs.typescript-language-server;
    };
  };

  packages = with pkgs; [
    git
    pkg-config
  ] ++ libraries;

  env = {
    CGO_ENABLED = "1";
    LD_LIBRARY_PATH = "${pkgs.lib.makeLibraryPath libraries}";
  };

  scripts.chip-8.exec = "go run cmd/chip-8/main.go $@";
  scripts.test-all.exec = "go test ./...";

  pre-commit.hooks = {
    gotest.enable = true;
    golangci-lint.enable = true;
  };
  pre-commit.hooks.editorconfig-checker.enable = true;

  enterShell = ''
    export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath libraries}
  '';
}
