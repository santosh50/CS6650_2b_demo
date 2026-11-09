variable "region" {
  type    = string
  default = "us-east-1"
}

variable "worker_count" {
  type    = number
  default = 1
}

variable "enable_lambda" {
  type    = bool
  default = false # only true during the Part III curl tests
}
