{
  description = "chip-8 emulator";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    devenv = {
      url = "github:cachix/devenv";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  nixConfig = {
    extra-trusted-public-keys = "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw=";
    extra-substituters = "https://devenv.cachix.org";
  };

  outputs =
    {
      nixpkgs,
      devenv,
      ...
    }@inputs:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      devShells.${system}.default = devenv.lib.mkShell {
        inherit inputs pkgs;
        modules = [
          (
            { pkgs, ... }:
            let
              libraries = with pkgs; [
                libX11
                libXcursor
                libXinerama
                libXrandr
                libXi
                mesa
                SDL2
                SDL2_ttf
                SDL2_gfx
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
              scripts.chip-8.exec = "go run ./cmd/chip-8 $@";
              scripts.test-all.exec = "go test ./...";
              git-hooks.hooks = {
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
                editorconfig-checker.enable = true;
              };
              enterShell = ''
                export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath libraries}
              '';
            }
          )
        ];
      };
    };
}
