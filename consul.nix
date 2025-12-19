{ config, pkgs, ... }:
let
  server = true;
  bootstrap_expect = 1;

  tag_key = "consul_cluster";
  consul_cluster_key = "primary";
  consul_token = "";
in
{
  services.consul = {
    enable = true;

    webUi = true;
    forceAddrFamily = "ipv4";
    interface.bind = "0.0.0.0";
    interface.advertise = "0.0.0.0";

    extraConfig = {
      datacenter = "dc2";
      log_level = "WARN";

      server = server;
      client_addr = "0.0.0.0";
      retry_join = [ "provider=aws tag_key=${tag_key} tag_value=${consul_cluster_key}" ];
      bootstrap_expect = if server then bootstrap_expect else null;

      recursors = [
        "1.1.1.1"
        "1.0.0.1"
        "8.8.8.8"
        "8.8.4.4"
      ];

      acl = {
        enabled = true;
        default_policy = "deny";
        enable_token_persistence = true;
        enable_token_replication = true;
        down_policy = "extend-cache";
        tokens = {
          initial_management = consul_token;
          agent = consul_token;
        };
      };

      ports = {
        grpc = 8502;
      };

      connect = {
        enabled = true;
      };
    };
  };

  networking.firewall = {
    enable = true;

    allowedTCPPorts = [
      53
      8600
    ];

    allowedUDPPorts = [
      53
      8600
    ];
  };

  networking.nftables = {
    enable = true;

    ruleset = ''
      table inet consul-dns {
        chain prerouting {
          type nat hook prerouting priority dstnat; policy accept;
          ip protocol udp udp dport 53 redirect to :8600
          ip protocol tcp tcp dport 53 redirect to :8600
        }

        chain output {
          type nat hook output priority -100; policy accept;
          ip daddr 127.0.0.1 ip protocol udp udp dport 53 redirect to :8600
          ip daddr 127.0.0.1 ip protocol tcp tcp dport 53 redirect to :8600
        }
      }
    '';
  };

  services.resolved = {
    enable = true;

    extraConfig = ''
      [Resolve]
      DNS=127.0.0.1:8600
      DNSSEC=false
      Domains=~consul
    '';
  };

  systemd.services.consul = {
    enable = true;
    wantedBy = [ "multi-user.target" ];
    postStart = "systemctl start consul";
  };
}
