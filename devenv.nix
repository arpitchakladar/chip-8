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
    golines
    golangci-lint
  ] ++ libraries;

  env = {
    CGO_ENABLED = "1";
    LD_LIBRARY_PATH = "${pkgs.lib.makeLibraryPath libraries}";
  };

  scripts.chip-8.exec = "go run cmd/chip-8 $@";
  scripts.test-all.exec = "go test ./...";

  pre-commit.hooks = {
    gotest.enable = true;
    golangci-lint.enable = false;
    golangci-lint-native = {
      enable = true;
      name = "golangci-lint (native)";
      entry = "${pkgs.golangci-lint}/bin/golangci-lint run";
      files = "\\.go$";
      pass_filenames = false;
    };
    golangci-lint-wasm = {
      enable = true;
      name = "golangci-lint (wasm)";
      entry = "env GOARCH=wasm GOOS=js ${pkgs.golangci-lint}/bin/golangci-lint run --build-tags js,wasm";
      files = "\\.go$";
      pass_filenames = false;
    };
    golines = {
      enable = true;
      name = "golines";
      entry = "${pkgs.golines}/bin/golines -w --max-len=80";
      files = "\\.go$";
      pass_filenames = true;
    };
  };
  pre-commit.hooks.editorconfig-checker.enable = true;

  enterShell = ''
    export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath libraries}
  '';
}
