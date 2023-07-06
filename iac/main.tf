resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    App = "movie-review"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "main"
    App  = "movie-review"
  }
}

resource "aws_route_table" "main" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name = "main"
    App  = "movie-review"
  }
}

resource "aws_subnet" "subnet1" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "eu-central-1a"

  tags = {
    Name = "subnet1"
    App  = "movie-review"
  }
}

resource "aws_subnet" "subnet2" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "eu-central-1b"

  tags = {
    Name = "subnet2"
    App  = "movie-review"
  }
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.subnet1.id
  route_table_id = aws_route_table.main.id
}

resource "aws_route_table_association" "b" {
  subnet_id      = aws_subnet.subnet2.id
  route_table_id = aws_route_table.main.id
}

resource "aws_db_subnet_group" "main" {
  name       = "main"
  subnet_ids = [aws_subnet.subnet1.id, aws_subnet.subnet2.id]

  tags = {
    App = "movie-review"
  }
}

resource "aws_security_group" "allow_postgres" {
  name        = "allow_postgres"
  description = "Allow inbound traffic from EC2 instances and my personal computer to RDS"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "allow_postgres"
  }
}

resource "aws_security_group" "allow_ssh" {
  name        = "allow_ssh"
  description = "Allow SSH inbound traffic"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "allow_ssh"
  }
}

resource "aws_security_group" "allow_web" {
  name        = "allow_web"
  description = "Allow inbound traffic on port 80"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "allow_web"
  }
}

resource "aws_iam_role" "ec2_role" {
  name               = "movie-review-ec2-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "ec2_policy" {
  name   = "movie-review-ec2-policy"
  role   = aws_iam_role.ec2_role.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameter"
      ],
      "Resource": "arn:aws:ssm:*:*:parameter/movie-review/*"
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "ec2_profile" {
  name = "movie-review-ec2-profile"
  role = aws_iam_role.ec2_role.name
}

resource "aws_instance" "host_instance" {
  ami                         = "ami-09a13963819e32919"
  instance_type               = "t3.micro"
  key_name                    = "movie-review"
  iam_instance_profile        = aws_iam_instance_profile.ec2_profile.name
  vpc_security_group_ids      = [aws_security_group.allow_ssh.id, aws_security_group.allow_web.id]
  user_data_replace_on_change = true
  subnet_id                   = aws_subnet.subnet1.id
  associate_public_ip_address = true

  user_data = <<-EOF
              #!/bin/bash
              sudo apt-get update
              sudo apt-get install -y docker.io awscli jq
              sudo systemctl start docker
              sudo systemctl enable docker
              REGION=eu-central-1
              JWT_SECRET=$(aws ssm get-parameter --name "/movie-review/jwt-secret" --region $REGION --with-decryption --output json | jq -r .Parameter.Value)
              ADMIN_NAME=$(aws ssm get-parameter --name "/movie-review/admin/name" --region $REGION --with-decryption --output json | jq -r .Parameter.Value)
              ADMIN_EMAIL=$(aws ssm get-parameter --name "/movie-review/admin/email" --region $REGION --with-decryption --output json | jq -r .Parameter.Value)
              ADMIN_PASSWORD=$(aws ssm get-parameter --name "/movie-review/admin/password" --region $REGION --with-decryption --output json | jq -r .Parameter.Value)
              DB_URL=$(aws ssm get-parameter --name "/movie-review/db-url" --region $REGION --with-decryption --output json | jq -r .Parameter.Value)
              sudo docker run -d \
                --name movie-reviews \
                -e JWT_SECRET=$JWT_SECRET \
                -e ADMIN_NAME=$ADMIN_NAME \
                -e ADMIN_EMAIL=$ADMIN_EMAIL \
                -e ADMIN_PASSWORD=$ADMIN_PASSWORD \
                -e DB_URL=$DB_URL \
                -p 80:8080 \
                boichique/movie-reviews:latest
              sudo docker run -d \
                --name watchtower \
                -v /var/run/docker.sock:/var/run/docker.sock \
                containrrr/watchtower \
                boichique/movie-reviews:latest \
                --schedule "0/30 * * * * *"
              EOF

  tags = {
    Name = "host-instance"
    App  = "movie-review"
  }
}

resource "aws_db_instance" "postgres_db" {
  allocated_storage           = 20
  engine                      = "postgres"
  engine_version              = "15"
  instance_class              = "db.t3.micro"
  db_name                     = "movie_review"
  username                    = "movie_review"
  password                    = "CHANGE_ME"
  parameter_group_name        = "default.postgres15"
  vpc_security_group_ids      = [aws_security_group.allow_postgres.id]
  db_subnet_group_name        = aws_db_subnet_group.main.name
  publicly_accessible         = true
  identifier                  = "movie-review-db"
  allow_major_version_upgrade = true
  skip_final_snapshot         = true

  tags = {
    Name = "postgres-db"
    App  = "movie-review"
  }
}

resource "aws_ssm_parameter" "db_url" {
  name  = "/movie-review/db-url"
  type  = "SecureString"
  value = "postgresql://${aws_db_instance.postgres_db.username}:${aws_db_instance.postgres_db.password}@${aws_db_instance.postgres_db.endpoint}/${aws_db_instance.postgres_db.db_name}?sslmode=prefer"

  lifecycle {
    ignore_changes = [value]
  }
}

variable "secrets" {
  type = set(string)
  default = [
    "/movie-review/jwt-secret",
    "/movie-review/admin/name",
    "/movie-review/admin/email",
    "/movie-review/admin/password",
  ]
}

resource "aws_ssm_parameter" "secrets" {
  for_each = toset(var.secrets)
  name     = each.value
  type     = "SecureString"
  value    = "CHANGE_ME"

  lifecycle {
    ignore_changes = [value]
  }
}
