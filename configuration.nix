{ config, pkgs, ... }:
{
  imports = [
    ./nomad.nix
    ./consul.nix
  ];

  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;
  system.stateVersion = "25.05";

  # Ephemeral writable root with tmpfs overlay
  boot.initrd.postMountCommands = ''
    mkdir -p /run/overlay-workdir /run/overlay-upper
  '';

  fileSystems."/var/lib" = {
    device = "overlay";
    fsType = "overlay";
    options = [
      "lowerdir=/nix/store"
      "upperdir=/run/overlay-upper"
      "workdir=/run/overlay-workdir"
    ];
    depends = [ "/nix/store" ];
  };

  fileSystems."/run/overlay-upper" = {
    device = "tmpfs";
    fsType = "tmpfs";
    options = [
      "mode=0755"
      "size=10%"
    ];
  };

  system.autoUpgrade.channel = "https://nixos.org/channels/nixos-25.05";
  nixpkgs.config.allowUnfree = true;
  security.sudo.wheelNeedsPassword = false; # TODO parameterize dev / debug build

  users.users.eva = {
    isNormalUser = true;
    extraGroups = [ "wheel" ];
    hashedPassword = "$y$j9T$t3ZgFM5FEG0AUqqILwpP1/$ylY03EGd/R0ZTuuFM.9bCtx0C9DYkleyN0UqKtuOsY.";

    openssh.authorizedKeys.keys = pkgs.lib.splitString "\n" (
      builtins.readFile (
        pkgs.fetchurl {
          url = "https://github.com/gastrodon.keys";
          sha256 = "I6oFQ8CW+VF9/Us1uXwv9/q2LcS5OMhizbSRXagmA64=";
        }
      )
    );
  };

  services.openssh = {
    enable = true;
    ports = [ 22 ];
    settings = {
      PasswordAuthentication = false;
      PubkeyAuthentication = true;
      PermitRootLogin = "no";
    };
  };

  systemd.services."autovt@tty50" = {
    enable = true;
    overrideStrategy = "asDropin";
    serviceConfig = {
      ExecStart = [
        "${pkgs.util-linux}/sbin/agetty --autologin eva --login-program ${pkgs.shadow}/bin/login tty50 $TERM"
      ];
    };
  };

  # ip addr show -> print interface: ip-addr
  # generates list of interface:ip pairs
  programs.bash.loginShellInit = ''
    ${pkgs.iproute2}/bin/ip addr show \
      | ${pkgs.gawk}/bin/awk '/^[0-9]+:/ { iface=$2; sub(/:$/,"",iface) } /inet / && iface != "lo" { \
        split($2,a,"/"); \
        print iface ": " a[1] \
      }'
  '';

  environment.systemPackages = with pkgs; [ ];
}
