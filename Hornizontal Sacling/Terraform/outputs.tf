output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}

# NEW: ALB endpoint to access the service
output "load_balancer_url" {
  description = "URL of the Application Load Balancer"
  value       = "http://${module.network.alb_dns_name}"
}

# NEW: Instructions for testing
output "test_endpoint" {
  description = "Example curl command to test the service"
  value       = "curl http://${module.network.alb_dns_name}/search?q=electronics"
}

output "health_check_endpoint" {
  description = "Health check endpoint for monitoring"
  value       = "http://${module.network.alb_dns_name}/health"
}