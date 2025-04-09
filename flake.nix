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
        packages = rec {
          ess = pkgs.callPackage ./. {inherit pkgs;};
          default = ess;
        };
        CGO_ENABLED = 0;

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

                scripts = {
                  prepare-release = {
                    description = "prepare a release";
                    exec = ''
                      git config --global user.email 'actions@github.com'
                      git config --global user.name 'Github Actions'
                      OLD_TAG=$(svu current | sed 's/^v//g')
                      NEW_TAG=$(svu next | sed 's/^v//g')
                      [ "$OLD_TAG" == "$NEW_TAG" ] && echo "no version bump" && exit 0
                      echo default.nix README.md | xargs sed -i "s/$OLD_TAG/$NEW_TAG/g"
                      echo default.nix README.md | xargs sed -i "s/v$OLD_TAG/v$NEW_TAG/g"
                      git add default.nix README.md
                      git commit -m "bump release version" --allow-empty
                      git tag v$NEW_TAG
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
