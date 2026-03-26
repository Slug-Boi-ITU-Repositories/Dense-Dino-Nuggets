variable "do_token" {
  sensitive = true
}

variable "postgres_user" {
  sensitive = true
}

variable "postgres_password" {
  sensitive = true
}

variable "postgres_db" {
  default = "minitwit_db"
}

variable "volume_mount" {
  default = "/mnt/volume_fra1_01/pgdata"
}