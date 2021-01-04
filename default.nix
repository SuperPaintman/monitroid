{ buildGoModule }:

{
  # Packages.
  monitroid = buildGoModule rec {
    pname = "monitroid";
    version = "0.0.0";

    subPackages = [ "cmd/monitroid" "cmd/monitroidctl" ];

    src = ./.;

    vendorSha256 = "0v38jcj7immgf0xyy1bjcwrvp0q5s68zzhr88ijwrkkss7qkcvj9";

    postInstall = ''
      mkdir -p "$out/share/dbus-1/system.d"

      cat << EOF > "$out/share/dbus-1/system.d/monitroid.conf"
      <!DOCTYPE busconfig PUBLIC
        "-//freedesktop//DTD D-BUS Bus Configuration 1.0//EN"
        "http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">

      <busconfig>
        <policy user="root">
          <allow own="com.github.SuperPaintman.monitroid"/>
          <allow send_destination="com.github.SuperPaintman.monitroid"/>
          <allow receive_sender="com.github.SuperPaintman.monitroid"/>
        </policy>

        <policy context="default">
          <allow send_destination="com.github.SuperPaintman.monitroid"/>
          <allow receive_sender="com.github.SuperPaintman.monitroid"/>
        </policy>
      </busconfig>
      EOF
    '';
  };

  # Nixos.
  nixos = import ./nixos;

  # Path.
  path = ./.;
}
