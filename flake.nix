{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    devenv.url = "github:cachix/devenv/v0.6.3";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    devenv,
    systems,
    gomod2nix,
    ...
  } @ inputs: let
    forEachSystem = nixpkgs.lib.genAttrs (import systems);
  in {
    packages = forEachSystem (system: let
      callPackage = nixpkgs.darwin.apple_sdk_11_0.callPackage or nixpkgs.legacyPackages.${system}.callPackage;
    in {
      default = callPackage ./. {
        inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
      };
    });

    devShells = forEachSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      default = devenv.lib.mkShell {
        inherit inputs pkgs;
        modules = [
          {
            packages = with pkgs; [
              automake
              go_1_21
              gomod2nix.legacyPackages.${system}.gomod2nix
              gotools
              golangci-lint
              go-tools
              gopls
              pre-commit
              flyctl
            ];

            pre-commit.hooks.gomod2nix = {
              enable = true;
              pass_filenames = false;
              name = "gomod2nix";
              description = "Run gomod2nix before commit";
              entry = "${gomod2nix.legacyPackages.${system}.gomod2nix}/bin/gomod2nix";
            };
          }
        ];
      };
    });
  };
}
