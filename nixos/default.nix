{ config, lib, pkgs, ... }:

with lib;
let
  cfg = config.services.monitroid;

  monitroid = (pkgs.callPackage ../. { }).monitroid;
in
{
  options.services.monitroid = {
    enable = mkOption {
      type = types.bool;
      default = false;
      description = ''
        Whether to enable monitroid.
      '';
    };

    package = mkOption {
      default = monitroid;
      defaultText = "monitroid.monitroid";
      example = "monitroid.monitroid";
      type = types.package;
      description = ''
        Monitroid package to use.
      '';
    };
  };

  config = mkIf cfg.enable {
    services.dbus.packages = [ cfg.package ];

    systemd.services.monitroid = {
      description = "Monitroid daemon";
      wantedBy = [ "multi-user.target" ];
      serviceConfig = {
        ExecStart = "${cfg.package}/bin/monitroid";
        Restart = "always";
        RestartSec = "10s";
      };
    };
  };
}
