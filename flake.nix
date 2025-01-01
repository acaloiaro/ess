{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    devenv.url = "github:cachix/devenv";
  };

  outputs = {
    self,
    nixpkgs,
    devenv,
    systems,
    ...
  } @ inputs: let
    forEachSystem = nixpkgs.lib.genAttrs (import systems);
  in {
    packages = forEachSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
      callPackage = nixpkgs.darwin.apple_sdk_11_0.callPackage or nixpkgs.legacyPackages.${system}.callPackage;
    in {
      default = (callPackage ./default.nix {}).${system};
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
              go_1_23
              gotools
              golangci-lint
              go-tools
              gopls
              nix-update
              pre-commit
              svu
            ];
          }
        ];
      };
    });
  };
}
