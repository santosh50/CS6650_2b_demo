output "alb_dns_name" {
  value = aws_lb.main.dns_name
}

output "queue_url" {
  value = aws_sqs_queue.orders.url
}

output "receiver_repo_url" {
  value = aws_ecr_repository.receiver.repository_url
}

output "processor_repo_url" {
  value = aws_ecr_repository.processor.repository_url
}

output "cluster_name" {
  value = aws_ecs_cluster.main.name
}

output "processor_service_name" {
  value = aws_ecs_service.processor.name
}
