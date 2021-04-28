package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestEKSNodeGroup_default(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "example" {
		cluster_name    = "test aws_eks_node_group"
		node_group_name = "example"
		node_role_arn   = "node_role_arn"
		subnet_ids      = ["subnet_id"]

		scaling_config {
			desired_size = 1
			max_size     = 1
			min_size     = 1
		}
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "CPU credits",
					PriceHash:       "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:             "Storage (general purpose SSD, gp2)",
					PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)

}

func TestEKSNodeGroup_defaultCpuCredits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "example" {
		cluster_name    = "test aws_eks_node_group"
		node_group_name = "example"
		node_role_arn   = "node_role_arn"
		subnet_ids      = ["subnet_id"]

		scaling_config {
			desired_size = 3
			max_size     = 3
			min_size     = 1
		}
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_eks_node_group.example": map[string]interface{}{
			"cpu_credit_hrs":    350,
			"virtual_cpu_count": 2,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
				{
					Name:            "CPU credits",
					PriceHash:       "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2 * 350 * 3)),
				},
				{
					Name:             "Storage (general purpose SSD, gp2)",
					PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)

}

func TestEKSNodeGroup_disk_size_instance_type(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "example" {
		cluster_name    = "test aws_eks_node_group"
		node_group_name = "example"
		instance_types  = ["t2.medium"]
		node_role_arn   = "node_role_arn"
		disk_size 			= 30
		subnet_ids      = ["subnet_id"]

		scaling_config {
			desired_size = 1
			max_size     = 1
			min_size     = 1
		}
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, t2.medium)",
					PriceHash:       "250382a8c0da495d6048e6fc57e526bc-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage (general purpose SSD, gp2)",
					PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)

}

func TestEKSNodeGroup_launch_template(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "example_with_launch_template" {
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
			id      = aws_launch_template.foo.id
			version = "default_version"
		}
	}

	resource "aws_launch_template" "foo" {
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

		instance_type = "m5.xlarge"

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
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.example_with_launch_template",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "aws_launch_template.foo",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Instance usage (Linux/UNIX, on-demand, m5.xlarge)",
							PriceHash:       "fc1dbb5469f07f2758e25e083d0effda-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
						},
						{
							Name:            "EBS-optimized usage",
							PriceHash:       "cd2c995d58f38ce65bd0740371c3e06d-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
						},
						{
							Name:            "Inference accelerator (eia1.medium)",
							PriceHash:       "3a42dad03b09f630bdba373e7ed51c0f-66d0d770bee368b4f2a8f2f597eeb417",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
						},
					},

					SubResourceChecks: []testutil.ResourceCheck{
						{
							Name: "root_block_device",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage (general purpose SSD, gp2)",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(24)),
								},
							},
						},
						{
							Name: "block_device_mapping[0]",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage (general purpose SSD, gp2)",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(60)),
								},
							},
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)

}

func TestEKSNodeGroup_launch_template_by_name(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "example_with_launch_template" {
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
			name      = aws_launch_template.foo.name
			version = "default_version"
		}
	}

	resource "aws_launch_template" "foo" {
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

		instance_type = "m5.xlarge"

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
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.example_with_launch_template",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "aws_launch_template.foo",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Instance usage (Linux/UNIX, on-demand, m5.xlarge)",
							PriceHash:       "fc1dbb5469f07f2758e25e083d0effda-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
						},
						{
							Name:            "EBS-optimized usage",
							PriceHash:       "cd2c995d58f38ce65bd0740371c3e06d-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
						},
						{
							Name:            "Inference accelerator (eia1.medium)",
							PriceHash:       "3a42dad03b09f630bdba373e7ed51c0f-66d0d770bee368b4f2a8f2f597eeb417",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
						},
					},

					SubResourceChecks: []testutil.ResourceCheck{
						{
							Name: "root_block_device",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage (general purpose SSD, gp2)",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(24)),
								},
							},
						},
						{
							Name: "block_device_mapping[0]",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage (general purpose SSD, gp2)",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(60)),
								},
							},
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)

}

func TestEKSNodeGroup_with_instance_launch_template_without_instance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "example_with_launch_template" {
		cluster_name    = "test aws_eks_node_group"
		node_group_name = "example"
		node_role_arn   = "node_role_arn"
		subnet_ids      = ["subnet_id"]

		instance_types  = ["t3.medium"]

		scaling_config {
			desired_size = 3
			max_size     = 1
			min_size     = 1
		}

		launch_template {
			id      = aws_launch_template.foo.id
			version = "default_version"
		}
	}

	resource "aws_launch_template" "foo" {
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
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_eks_node_group.example_with_launch_template": map[string]interface{}{
			"cpu_credit_hrs":    350,
			"virtual_cpu_count": 2,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.example_with_launch_template",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "aws_launch_template.foo",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Instance usage (Linux/UNIX, on-demand, t3.medium)",
							PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
						},
						{
							Name:            "CPU credits",
							PriceHash:       "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(350 * 2 * 3)),
						},
						{
							Name:            "Inference accelerator (eia1.medium)",
							PriceHash:       "3a42dad03b09f630bdba373e7ed51c0f-66d0d770bee368b4f2a8f2f597eeb417",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
						},
					},

					SubResourceChecks: []testutil.ResourceCheck{
						{
							Name: "root_block_device",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage (general purpose SSD, gp2)",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(24)),
								},
							},
						},
						{
							Name: "block_device_mapping[0]",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage (general purpose SSD, gp2)",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(60)),
								},
							},
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)

}

func TestEKSNodeGroup_spot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "example" {
		cluster_name    = "test aws_eks_node_group"
		node_group_name = "example"
		node_role_arn   = "node_role_arn"
		subnet_ids      = ["subnet_id"]
		capacity_type   = "SPOT"

		scaling_config {
			desired_size = 1
			max_size     = 1
			min_size     = 1
		}
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, spot, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-803d7f1cd2f621429b63f791730e7935",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "CPU credits",
					PriceHash:       "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:             "Storage (general purpose SSD, gp2)",
					PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestEKSNodeGroup_reserved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "reserved" {
		cluster_name    = "test aws_eks_node_group"
		node_group_name = "example"
		node_role_arn   = "node_role_arn"
		subnet_ids      = ["subnet_id"]

		scaling_config {
			desired_size = 1
			max_size     = 1
			min_size     = 1
		}
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_eks_node_group.reserved": map[string]interface{}{
			"reserved_instance_type":           "standard",
			"reserved_instance_term":           "1_year",
			"reserved_instance_payment_option": "no_upfront",
			"cpu_credit_hrs":                   350,
			"virtual_cpu_count":                2,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.reserved",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, reserved, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-354de5028123250997d97c05d011fe1c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "CPU credits",
					PriceHash:       "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(350 * 2 * 1)),
				},
				{
					Name:             "Storage (general purpose SSD, gp2)",
					PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}

func TestEKSNodeGroup_windows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "windows" {
		cluster_name    = "test aws_eks_node_group"
		node_group_name = "example"
		node_role_arn   = "node_role_arn"
		subnet_ids      = ["subnet_id"]

		scaling_config {
			desired_size = 1
			max_size     = 1
			min_size     = 1
		}
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_eks_node_group.windows": map[string]interface{}{
			"operating_system": "windows",
			"cpu_credit_hrs": 100,
			"virtual_cpu_count": 2,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.windows",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Windows, on-demand, t3.medium)",
					PriceHash:       "9b141b37e9cd40b8225ed689fbdb0963-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "CPU credits",
					PriceHash:       "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 2 * 1)),
				},
				{
					Name:             "Storage (general purpose SSD, gp2)",
					PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
