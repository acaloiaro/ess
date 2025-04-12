{
  description = "ess (env sampple sync)";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {
          inherit system;
        };
      in {
        packages = rec {
          ess = pkgs.callPackage ./. {inherit pkgs self;};
          default = ess;
        };

        devShells = let
          pkgs = nixpkgs.legacyPackages.${system};

          prepare-release = pkgs.writeShellApplication {
            name = "prepare-release";

            runtimeInputs = with pkgs; [
              coreutils
              git
              gnused
              svu
            ];

            text =
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
                git tag "v$NEW_VERSION"
                git push
                git push --tags
              '';
          };
        in {
          default = pkgs.mkShell {
            packages =
              [
                prepare-release
              ]
              ++ (with pkgs; [
                act
                go_1_24
                gotools
                golangci-lint
                go-tools
                gopls
              ]);

            shellHook =
              # bash
              ''
                export CGO_ENABLED=0
              '';
          };
        };
      }
    );
}
