# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-west-2"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "ecr_service"
}

variable "service_name" {
  type    = string
  default = "CS6650L2"
}

variable "container_port" {
  type    = number
  default = 8080
}

# REMOVED: ecs_count (replaced by auto-scaling min/max)
# variable "ecs_count" {
#   type    = number
#   default = 1
# }

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# Fargate CPU units (256 = 0.25 vCPU)
variable "fargate_cpu" {
  type    = string
  default = "256"
}

# Fargate Memory in MB
variable "fargate_memory" {
  type    = string
  default = "512"
}

# ===== NEW VARIABLES FOR ALB =====
# Health check configuration
variable "health_check_path" {
  type    = string
  default = "/health"  # Your Go app needs this endpoint
}

# ===== NEW VARIABLES FOR AUTO SCALING =====
# Minimum number of tasks (replaces ecs_count)
variable "min_instances" {
  type    = number
  default = 2  # Start with 2 for redundancy
}

# Maximum number of tasks
variable "max_instances" {
  type    = number
  default = 6  # Scale up to 4 tasks max
}

# Target CPU percentage for scaling
variable "target_cpu" {
  type    = number
  default = 50  # Scale when CPU hits 70%
}

# Cooldown period in seconds
variable "scale_cooldown" {
  type    = number
  default = 150  # 5 minutes
}