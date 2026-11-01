{
  description = "stellation: branded, self-updating Minecraft modpack distributions";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAll = f: nixpkgs.lib.genAttrs systems (s: f nixpkgs.legacyPackages.${s});
    in
    {
      packages = forAll (pkgs: {
        default = pkgs.buildGoModule {
          pname = "stellation";
          version = self.shortRev or self.dirtyShortRev or "dev";
          src = ./.;
          vendorHash = "sha256-spo+YjBeDZJiQF6guHh/I643kEWFOTrc3TIO6ZRKdU0=";
          subPackages = [ "." ];
          ldflags = [
            "-s"
            "-w"
            "-X main.version=${self.shortRev or "dev"}"
          ];
        };
      });

      devShells = forAll (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            goreleaser
            cobra-cli
            just
            packwiz
          ];
        };
      });
    };
}
