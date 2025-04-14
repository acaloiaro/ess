{
  self,
  pkgs ? import <nixpkgs> {},
  ...
}:
pkgs.buildGoModule rec {
  pname = "ess";
  version = pkgs.lib.strings.removeSuffix "\n" (builtins.readFile ./version.txt);
  src = ./.;
  vendorHash = null;
  ldflags = [
    "-X 'main.version=${version}-nix'"
    "-X 'main.commit=${self.rev or "dev"}'"
  ];
  CGO_ENABLED = 0;

  meta = {
    description = "ess (env-sample-sync): automatically and safely synchronize env.sample files with .env";
    homepage = "https://github.com/acaloiaro/ess";
    license = pkgs.lib.licenses.bsd2;
  };
}
