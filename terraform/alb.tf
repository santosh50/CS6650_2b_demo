data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "alb" {
  name        = "${var.service_name}-alb-sg"
  description = "Allow HTTP from anywhere to the ALB"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP from anywhere"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }
}

resource "aws_lb" "this" {
  name               = "${var.service_name}-alb"
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = module.network.subnet_ids
}

resource "aws_lb_target_group" "this" {
  name        = "${var.service_name}-tg"
  target_type = "ip"
  protocol    = "HTTP"
  port        = var.container_port
  vpc_id      = data.aws_vpc.default.id

  deregistration_delay = 30

  health_check {
    path                = "/health"
    protocol            = "HTTP"
    port                = "traffic-port"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
    matcher             = "200"
  }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.this.arn
  }
}

output "alb_dns_name" {
  description = "Use http://<this> (port 80) as the Locust host"
  value       = aws_lb.this.dns_name
}
