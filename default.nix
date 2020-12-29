{ buildGoModule }:

{
  # Packages.
  monitroid = buildGoModule rec {
    pname = "monitroid";
    version = "0.0.0";

    subPackages = [ "cmd/monitroid" "cmd/monitroidctl" ];

    src = ./.;

    vendorSha256 = "112imazpgkil2lh4vg0i18mv1fzqlpqvk259bs85m599wr106gp5";
  };

  # Nixos.
  nixos = import ./nixos;

  # Path.
  path = ./.;
}
