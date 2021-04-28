provider "aws" {
  region = "us-east-1"
  profile = "robh"
}

/*
resource "aws_eks_node_group" "example" {
  cluster_name    = "test aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]
  capacity_type   = "SPOT"
  //instance_types = ["t3.medium", "m5.large"]

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }
}


resource "aws_eks_node_group" "example_with_launch_template" {
  cluster_name    = "test aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  instance_types = ["m5.large"]

  disk_size = 300

  scaling_config {
    desired_size = 3
    max_size     = 3
    min_size     = 1
  }

  launch_template {
    id      = aws_launch_template.foo.id
    version = "default_version"
  }
}

resource "aws_launch_template" "foo" {
  name = "foo"

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  cpu_options {
    core_count       = 4
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "standard"
  }

  disable_api_termination = true

  ebs_optimized = true

  elastic_gpu_specifications {
    type = "test"
  }

  elastic_inference_accelerator {
    type = "eia1.medium"
  }

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "terminate"

  kernel_id = "test"

  key_name = "test"

  license_specification {
    license_configuration_arn = "arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  network_interfaces {
    associate_public_ip_address = true
  }

  placement {
    availability_zone = "us-west-2a"
  }

  ram_disk_id = "test"

  vpc_security_group_ids = ["example"]

}

resource "aws_eks_node_group" "example_with_launch_template_instance" {
  cluster_name    = "test aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 3
    max_size     = 1
    min_size     = 1
  }

  launch_template {
    id      = aws_launch_template.instance.id
    version = "default_version"
  }
}

resource "aws_launch_template" "instance" {
  name = "foo"

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 20
    }
  }

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  cpu_options {
    core_count       = 4
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "unlimited"
  }

  disable_api_termination = true

  ebs_optimized = true

  elastic_gpu_specifications {
    type = "test"
  }

  elastic_inference_accelerator {
    type = "eia1.medium"
  }

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "terminate"

  instance_type = "t3.large"

  kernel_id = "test"

  key_name = "test"

  license_specification {
    license_configuration_arn = "arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  network_interfaces {
    associate_public_ip_address = true
  }

  placement {
    availability_zone = "us-west-2a"
  }

  ram_disk_id = "test"

  vpc_security_group_ids = ["example"]

}
*/

resource "aws_launch_template" "lt1" {
  image_id      = "fake_ami"
  instance_type = "t2.medium"

  block_device_mappings {
    device_name = "xvdf"
    ebs {
      volume_size = 10
    }
  }

  block_device_mappings {
    device_name = "xvfa"
    ebs {
      volume_size = 20
      volume_type = "io1"
      iops        = 200
    }
  }
}

resource "aws_autoscaling_group" "asg1" {
  launch_template {
    id = aws_launch_template.lt1.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
}