job "influx" {
  datacenters = ["proxmox"]
  type = "service"
  update {
    max_parallel = 1
    min_healthy_time = "10s"
    healthy_deadline = "3m"
    progress_deadline = "10m"
    health_check = "task_states"
    auto_revert = false
    canary = 0
  }
  migrate {
    max_parallel = 1
    health_check = "task_states"
    min_healthy_time = "10s"
    healthy_deadline = "5m"
  }

  group "influx" {
    count = 1
 
    volume "influx" {
      type = "host"
      source = "influx"
      read_only = false
    }

    
    restart {
      attempts = 1
      interval = "30m"
      delay = "15s"
      mode = "fail"
    }

    task "influxdb" {
      driver = "docker"

      volume_mount {
        volume      = "influx"
        destination = "/var/lib/influxdb"
        read_only = false
      }

      config {
        image = "quay.io/influxdb/influxdb:v2.0.3"
        port_map {
          db = 8086
        }
      }

      service {
        name = "influxdb"
        tags = ["influxdb", "influx", "db"]
        port = "db"
        check {
          name     = "alive"
          type     = "http"
          port = "db"
          address_mode = "driver"
          interval = "5s"
          timeout  = "2s"
          path = "/ui"
        }
    }

      resources {
        cpu    = 100
        memory = 256
        network {
          port "db" {
            static = 8086
          }
        }
      }
    }
  }
}
