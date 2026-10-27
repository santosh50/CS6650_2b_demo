variable "cpu_target" {
  type    = number
  default = 70
}
variable "min_tasks" {
  type    = number
  default = 2
}
variable "max_tasks" {
  type    = number
  default = 4
}
variable "scale_out_cooldown" {
  type    = number
  default = 300
}
variable "scale_in_cooldown" {
  type    = number
  default = 300
}

resource "aws_appautoscaling_target" "ecs" {
  service_namespace  = "ecs"
  scalable_dimension = "ecs:service:DesiredCount"
  resource_id        = "service/${module.ecs.cluster_name}/${module.ecs.service_name}"
  min_capacity       = var.min_tasks
  max_capacity       = var.max_tasks
}

resource "aws_appautoscaling_policy" "cpu" {
  name               = "${var.service_name}-cpu-target-tracking"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  resource_id        = aws_appautoscaling_target.ecs.resource_id

  target_tracking_scaling_policy_configuration {
    target_value       = var.cpu_target
    scale_out_cooldown = var.scale_out_cooldown
    scale_in_cooldown  = var.scale_in_cooldown

    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
  }
}
