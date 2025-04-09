{
  pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs;
    in
      import (fetchTree nixpkgs.locked) {}
  ),
  buildGoModule ? pkgs.buildGoModule,
}:
buildGoModule {
  pname = "ess";
  version = "2.16.1";
  pwd = ./.;
  src = ./.;
  vendorHash = "sha256-ooTP3mS7AEzwJm1JKebL0V2lqVge3WnpFZcbr1f/LIg=";
  meta = {
    description = "ess (env-sample-sync): automatically and safely synchronize env.sample files with .env";
    license = pkgs.lib.licenses.bsd2;
  };
}
