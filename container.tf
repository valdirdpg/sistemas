# Criação do container para implantação
resource "docker_container" "reservas1" {
  name         = "reservas1"
  image        = docker_image.reserva.image_id
  ports {
    internal = 8080
    external = 8085
  }

  volumes {
    host_path      = "/home/valdirdpg/workspace/aluno/reserva-salas/dados"
    container_path = "/app/dados"
  }

  restart = "always"
}
resource "docker_container" "prometheus" {
  name         = "prometheus"
  image        = "prom/prometheus"
  ports {
    internal = 9090
    external = 9095
 }

  volumes {
    host_path      = "/home/aluno/reserva-salas/prometheus"
    container_path = "/etc/prometheus"
  }

  volumes {
    host_path      = "/home/aluno/reserva-salas/prometheus/data"
    container_path = "/prometheus"
  }

  restart = "always"
}


resource "docker_container" "grafana" {
  name         = "grafana"
  image        = "grafana/grafana"
  ports {
    internal = 3000
    external = 3005
  }

  volumes {
    host_path      = "/home/aluno/reserva-salas/grafana/data"
    container_path = "/var/lib/grafana"
  }

  user = "1001:1001"

  env = ["GF_INSTALL_PLUGINS=grafana-clock-panel,grafana-simple-json-datasource", " GF_SECURITY_ADMIN_USER=admin",
          "GF_SECURITY_ADMIN_PASSWORD=admin", "GF_USERS_ALLOW_SIGN_UP=false"]

  depends_on = [docker_container.prometheus]

  restart = "always"
}
