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
                    exec =
                      # bash
                      ''
                        git config --global user.email 'actions@github.com'
                        git config --global user.name 'Github Actions'
                        OLD_VERSION=$(svu current --tag.prefix "")
                        NEW_VERSION=$(svu next --tag.prefix "" --always)
                        [ "$OLD_VERSION" == "$NEW_VERSION" ] && echo "no version bump" && exit 0
                        sed -i "s/$OLD_VERSION/$NEW_VERSION/g" README.md
                        echo "$NEW_VERSION" >version.txt
                        git add version.txt README.md
                        git commit -m "bump release version" --allow-empty
                        git tag v$NEW_VERSION
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
