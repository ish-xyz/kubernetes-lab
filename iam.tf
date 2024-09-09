resource "aws_iam_role" "kube_nodes" {
  name = "kube-nodes"

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

resource "aws_iam_role_policy_attachment" "kube_nodes_s3_read_only" {
  role       = aws_iam_role.kube_nodes.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
}

resource "aws_iam_instance_profile" "kube_nodes" {
  name = "kube-nodes-s3-read-only"
  role = aws_iam_role.kube_nodes.name
}