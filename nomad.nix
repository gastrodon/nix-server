{ config, pkgs, ... }:
let
  server = true;
  bootstrap_expect = 1;

  tag_key = "nomad_cluster";
  nomad_cluster_key = "primary";
  consul_token = "";
in
{
  services.nomad = {
    enable = true;
    enableDocker = true;
    settings = {
      datacenter = "dc1";
      bind_addr = "0.0.0.0";

      server = {
        enabled = server;
        bootstrap_expect = if server then bootstrap_expect else null;
      };

      client = {
        enabled = true;
        cni_path = "/opt/cni/bin";
        cni_config_dir = "/opt/cni/config";
        server_join = {
          retry_join = [ "provider=aws tag_key=${tag_key} tag_value=${nomad_cluster_key}" ];
          retry_max = 0;
          retry_interval = "15s";
        };
        artifact = {
          decompression_file_count_limit = 0;
        };
      };

      plugin.docker = {
        config = {
          allow_privileged = true;
          auth = {
            config = "/etc/docker-auth.json";
          };
          volumes = {
            enabled = true;
          };
        };
      };

      plugin.raw_exec = {
        config = {
          enabled = true;
        };
      };

      acl = {
        enabled = true;
      };

      telemetry = {
        collection_interval = "1s";
        disable_hostname = true;
        prometheus_metrics = true;
        publish_allocation_metrics = true;
        publish_node_metrics = true;
      };

      consul = {
        address = "127.0.0.1:8500";
        token = consul_token;
        server_auto_join = false;
        client_auto_join = false;
      };

      limits = {
        http_max_conns_per_client = 0;
      };
    };
  };

  systemd.services.nomad = {
    enable = true;
    wantedBy = [ "multi-user.target" ];
    postStart = "systemctl start nomad";
  };
}
