# ---------------- Part III: Lambda ----------------

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/order-processor-lambda"
  retention_in_days = 7
}

resource "aws_lambda_function" "processor" {
  function_name    = "order-processor-lambda"
  role             = data.aws_iam_role.lab.arn
  runtime          = "provided.al2" # if AWS rejects it, use "provided.al2023" and note it in the report
  handler          = "bootstrap"
  architectures    = ["x86_64"]
  memory_size      = 512
  timeout          = 10 # default 3 s would cut off the 3 s payment
  filename         = "${path.module}/../src/lambda/function.zip"
  source_code_hash = filebase64sha256("${path.module}/../src/lambda/function.zip")
  depends_on       = [aws_cloudwatch_log_group.lambda]
}

# Only exists when enable_lambda = true
resource "aws_lambda_permission" "sns" {
  count         = var.enable_lambda ? 1 : 0
  statement_id  = "AllowSNSInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.processor.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.orders.arn
}

# Only exists when enable_lambda = true
resource "aws_sns_topic_subscription" "to_lambda" {
  count      = var.enable_lambda ? 1 : 0
  topic_arn  = aws_sns_topic.orders.arn
  protocol   = "lambda"
  endpoint   = aws_lambda_function.processor.arn
  depends_on = [aws_lambda_permission.sns]
}

output "lambda_function_name" {
  value = aws_lambda_function.processor.function_name
}
