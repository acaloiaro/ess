{pkgs ? import <nixpkgs> {}, ...}:
pkgs.buildGoModule {
  pname = "ess";
  version = pkgs.lib.strings.removeSuffix "\n" (builtins.readFile ./version.txt);
  src = ./.;
  vendorHash = null;

  meta = {
    description = "ess (env-sample-sync): automatically and safely synchronize env.sample files with .env";
    license = pkgs.lib.licenses.bsd2;
  };
}
