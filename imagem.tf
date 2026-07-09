# Configuração do provider Docker
provider "docker" {
  host = "unix:///var/run/docker.sock"
}

# Criação da imagem Docker
resource "docker_image" "reserva" {
  name         = "reserva"
  build {
    context = "."
    dockerfile   = "./Dockerfile"
  }
  keep_locally = false
}
