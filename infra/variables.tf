variable "do_token"          { sensitive = true }
variable "postgres_user"     { sensitive = true }
variable "postgres_password" { sensitive = true }
variable "monitor_pub_key"   { sensitive = true }

variable "postgres_host"  {}
variable "monitor_ip"     {}

variable "docker_username"       { default = "flakiator" }
variable "container_name_prefix" { default = "prod" }
variable "postgres_db"           { default = "minitwit_db" }
variable "postgres_port"         { default = "5432" }
variable "db_ssl_mode"           { default = "disable" }
variable "volume_mount"          { default = "/mnt/volume_fra1_01/pgdata" }
variable "postgres_cpu_limit"    { default = "0.2" }
variable "postgres_mem_limit"    { default = "300M" }
variable "app_replicas"          { default = "2" }
variable "app_cpu_limit"         { default = "0.2" }
variable "app_mem_limit"         { default = "256M" }