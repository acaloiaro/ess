{
  description = "ess (env sampple sync)";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    devenv.url = "github:cachix/devenv";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    devenv,
  } @ inputs:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {
          inherit system;
        };
      in {
        packages.default = pkgs.callPackage ./. {};
        CGO_ENABLED = 0;
        defaultPackage = self.packages.${system}.default;
        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };

        devShells = let
          pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = devenv.lib.mkShell {
            inherit inputs pkgs;
            modules = [
              {
                packages = with pkgs; [
                  act
                  automake
                  go_1_24
                  gotools
                  golangci-lint
                  go-tools
                  gopls
                  svu
                ];

                scripts = with pkgs; {
                  prepare-release = {
                    description = "prepare a release";
                    exec = ''
                      git config --global user.email 'actions@github.com'
                      git config --global user.name 'Github Actions'
                      OLD_TAG=$(svu current)
                      NEW_TAG=$(svu next)
                      [ "$OLD_TAG" == "$NEW_TAG" ] && echo "no version bump" && exit 0
                      echo default.nix README.md | xargs sed -i "s/$(svu current)/$(svu next)/g"
                      go mod vendor
                      sed -i "s|vendorHash = \".*\"|vendorHash = \"$(nix hash path ./vendor)\"|g" default.nix
                      git add default.nix main.go README.md
                      git commit -m "bump release version" --allow-empty
                      git tag $NEW_TAG
                      git push
                      git push --tags
                    '';
                  };
                };
              }
            ];
          };
        };
      }
    );
}
