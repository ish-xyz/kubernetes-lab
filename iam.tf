resource "aws_iam_role" "kube_controllers_role" {
  name = "kube-controllers"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "kube_controllers_role_attachment" {
  role       = aws_iam_role.kube_controllers_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
}

resource "aws_iam_instance_profile" "kube_controllers_profile" {
  name = "kube-controllers"
  role = aws_iam_role.kube_controllers_role.name
}