{pkgs ? import <nixpkgs> {}}: let
  build = pkgs.buildGoModule {
    pname = "ess";
    version = "2.14.1";
    src = ./.;
    pwd = ./.;
    vendorHash = "sha256-oIA1LTBvkA26ferZI+uOipLE2/wZr2XKjGfESuD/cPk=";
    meta = {
      description = "ess (env-sample-sync): automatically and safely synchronize env.sample files with .env";
      license = pkgs.lib.licenses.bsd2;
    };
  };
in {
  darwin = build.overrideAttrs (old:
    old
    // {
      GOOS = "darwin";
      GOARCH = "amd64";
    });
  linux32 = build.overrideAttrs (old:
    old
    // {
      GOOS = "linux";
      GOARCH = "386";
    });
  x86_64-linux = build.overrideAttrs (old:
    old
    // {
      GOOS = "linux";
      GOARCH = "amd64";
    });
  windows64 = build.overrideAttrs (old:
    old
    // {
      GOOS = "windows";
      GOARCH = "amd64";
    });
}
