variable "linode_token" {
  description = "Linode API key"
  type        = string
	sensitive   = true
}

variable "root_pass" {
  description = "A root pass for each machine deployment"
  type        = string
	sensitive   = true
}

variable "region" {
  description = "The region to deploy resources"
  type        = string
  default     = "us-southeast"
}

variable "interface_public_key" {
  description = "Public key of the interface"
  type        = string
  default     = "svySoZDKZZcmfc5lOaSsbBtyiBYW3ho0EhBUSD0oD3o="
}

variable "endpoint" {
  description = "client public ip_address"
  type        = string
}
