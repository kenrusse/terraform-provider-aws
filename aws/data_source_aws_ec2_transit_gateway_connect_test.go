package aws

import (
	// "fmt"
	"testing"

	// "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEc2TransitGatewayConnectDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_connect.test"
	resourceName := "aws_ec2_transit_gateway_connect.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "protocol", dataSourceName, "protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transport_transit_gateway_attachment_id", dataSourceName, "transport_transit_gateway_attachment_id"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayConnectDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_connect.test"
	resourceName := "aws_ec2_transit_gateway_connect.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectDataSourceConfigID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "protocol", dataSourceName, "protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transport_transit_gateway_attachment_id", dataSourceName, "transport_transit_gateway_attachment_id"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayConnectDataSourceConfigFilter() string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() + `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  protocol = "gre"
  transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

data "aws_ec2_transit_gateway_connect" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_connect.test.id]
  }
}
`
}

func testAccAWSEc2TransitGatewayConnectDataSourceConfigID() string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() + `

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  protocol = "gre"
  transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

data "aws_ec2_transit_gateway_connect" "test" {
  id = aws_ec2_transit_gateway_connect.test.id
}
`
}

// func testAccCheckAWSEc2TransitGatewayConnectDestroy(s *terraform.State) error {
// 	conn := testAccProvider.Meta().(*AWSClient).ec2conn
//
// 	for _, rs := range s.RootModule().Resources {
// 		if rs.Type != "aws_ec2_transit_gateway_connect" {
// 			continue
// 		}
//
// 		transitGatewayConnect, err := ec2DescribeTransitGatewayConnect(conn, rs.Primary.ID)
//
// 		if isAWSErr(err, "InvalidTransitGatewayConnectID.NotFound", "") {
// 			continue
// 		}
//
// 		if err != nil {
// 			return err
// 		}
//
// 		if transitGatewayConnect == nil {
// 			continue
// 		}
//
// 		if aws.StringValue(transitGatewayConnect.State) != ec2.TransitGatewayAttachmentStateDeleted {
// 			return fmt.Errorf("EC2 Transit Gateway Connect (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(transitGatewayConnect.State))
// 		}
// 	}
//
// 	return nil
// }
