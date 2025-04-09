{pkgs ? import <nixpkgs> {}, ...}:
pkgs.buildGoModule {
  pname = "ess";
  version = "2.16.4";
  src = ./.;
  vendorHash = null;

  meta = {
    description = "ess (env-sample-sync): automatically and safely synchronize env.sample files with .env";
    license = pkgs.lib.licenses.bsd2;
  };
}
