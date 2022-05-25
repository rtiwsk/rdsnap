// versions
terraform {
  required_version = "~> 1.1.2"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "= 4.8.0"
    }
  }
}

// variables
variable "region" {
  default     = "ap-northeast-1"
  description = "AWS region"
}

variable "cidr" {
  default     = "10.0.0.0/16"
  description = "VPC cidr"
}

// provider
provider "aws" {
  region = var.region
}

// network
data "aws_availability_zones" "available" {}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.14.0"

  name = "example"
  cidr = var.cidr

  azs = data.aws_availability_zones.available.names
  public_subnets = [
    cidrsubnet(var.cidr, 8, 11),
    cidrsubnet(var.cidr, 8, 12),
    cidrsubnet(var.cidr, 8, 13),
  ]

  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_db_subnet_group" "rds" {
  name       = "rds-postgres-subnet"
  subnet_ids = module.vpc.public_subnets
}

resource "aws_security_group" "rds" {
  name   = "rds-postgres-secgrp"
  vpc_id = module.vpc.vpc_id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "example-rds-subnet"
  }
}

// RDS postgres
resource "aws_db_parameter_group" "example" {
  name   = "example"
  family = "postgres13"
}

resource "aws_db_instance" "example" {
  identifier             = "exampledb"
  instance_class         = "db.t3.micro"
  allocated_storage      = 10
  engine                 = "postgres"
  engine_version         = "13.4"
  db_name                = "example"
  username               = "dbadmin"
  password               = "POiu123!!x"
  db_subnet_group_name   = aws_db_subnet_group.rds.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  parameter_group_name   = aws_db_parameter_group.example.name
  publicly_accessible    = true
  skip_final_snapshot    = true
  apply_immediately      = true

  tags = {
    Name = "exampledb",
  }
}

// output
output "rds_hostname" {
  value     = aws_db_instance.example.address
  sensitive = true
}

output "rds_port" {
  value     = aws_db_instance.example.port
  sensitive = true
}

output "rds_username" {
  value     = aws_db_instance.example.username
  sensitive = true
}
