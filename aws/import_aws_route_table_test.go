package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSRouteTable_importBasic(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		// Expect, 1 resource in state, and route count to be 1
		v, ok := s[0].Attributes["route.#"]
		if len(s) != 1 || !ok || v != "1" {
			return fmt.Errorf("bad state: %s", s)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfig,
			},

			{
				ResourceName:     "aws_route_table.foo",
				ImportState:      true,
				ImportStateCheck: checkFn,
			},
		},
	})
}

func TestAccAWSRouteTable_importWithCreate(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		// Expect, 1 resource in state, and route count to be 1
		v, ok := s[0].Attributes["route.#"]
		if len(s) != 1 || !ok || v != "1" {
			return fmt.Errorf("bad state: %s", s)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:           testAccRouteTableConfig,
				ImportStateCheck: checkFn,
			},

			{
				ResourceName:      "aws_route_table.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},

			{
				Config:           testAccRouteTableConfig,
				ImportStateCheck: checkFn,
			},
		},
	})
}

func TestAccAWSRouteTable_importComplex(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		// Expect, 1 resource in state, and route count to be 2
		v, ok := s[0].Attributes["route.#"]
		if len(s) != 1 || !ok || v != "2" {
			return fmt.Errorf("bad state: %s", s)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfigImportComplex,
			},

			{
				ResourceName:     "aws_route_table.mod",
				ImportState:      true,
				ImportStateCheck: checkFn,
			},
		},
	})
}

const testAccRouteTableConfigImportComplex = `
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-route-table-import-complex-default"
  }
}

resource "aws_subnet" "tf_test_subnet" {
  vpc_id                  = "${aws_vpc.foo.id}"
  cidr_block              = "10.1.0.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-route-table-import-complex-default"
  }
}

resource "aws_eip" "nat" {
  vpc                       = true
  associate_with_private_ip = "10.1.0.10"
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "terraform-testacc-route-table-import-complex-default"
  }
}

resource "aws_nat_gateway" "nat" {
  allocation_id = "${aws_eip.nat.id}"
  subnet_id     = "${aws_subnet.tf_test_subnet.id}"

  tags = {
    Name = "terraform-testacc-route-table-import-complex-default"
  }
}

resource "aws_route_table" "mod" {
  vpc_id = "${aws_vpc.default.id}"

  tags = {
    Name = "tf-rt-import-test"
  }

  depends_on = ["aws_internet_gateway.ogw", "aws_internet_gateway.gw"]
}

resource "aws_route" "mod-1" {
  route_table_id         = "${aws_route_table.mod.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${aws_nat_gateway.nat.id}"
}

resource "aws_route" "mod" {
  route_table_id            = "${aws_route_table.mod.id}"
  destination_cidr_block    = "10.181.0.0/16"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.foo.id}"
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id          = "${aws_vpc.foo.id}"
  service_name    = "com.amazonaws.us-west-2.s3"
  route_table_ids = ["${aws_route_table.mod.id}"]
}

### vpc bar

resource "aws_vpc" "bar" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-import-complex-bar"
  }
}

resource "aws_internet_gateway" "ogw" {
  vpc_id = "${aws_vpc.bar.id}"

  tags = {
    Name = "terraform-testacc-route-table-import-complex-bar"
  }
}

### vpc peer connection

resource "aws_vpc_peering_connection" "foo" {
  vpc_id        = "${aws_vpc.foo.id}"
  peer_vpc_id   = "${aws_vpc.bar.id}"
  peer_owner_id = "187416307283"

  tags = {
    Name = "tf-rt-import-test"
  }

  auto_accept = true
}
`
