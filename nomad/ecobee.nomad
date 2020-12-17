job "ecobee" {
  datacenters = ["proxmox"]
  type = "batch"

  periodic {
    cron             = "*/15 * * * * *"
    prohibit_overlap = true
  }

  group "ecobee" {
    count = 1
    
    restart {
      attempts = 1
      interval = "5m"
      delay = "15s"
      mode = "fail"
    }

    task "ecobee" {
      driver = "exec"
    
      // This may change depending on permissions of the config file dir and files within it
      user = "ecobee"

      config {
          // Binary located here, nomad provides /bin/ to the chroot environment
          // Config and token file are in /etc/ecobee/
          command = "/bin/ecobee"
      }


      resources {
        cpu    = 10
        memory = 100
        }
      }
    }
  }
