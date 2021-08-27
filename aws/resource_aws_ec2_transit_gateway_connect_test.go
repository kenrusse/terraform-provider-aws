package aws

import (
	// "errors"
	"fmt"
	// "log"
	// "regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// func init() {
// 	resource.AddTestSweepers("aws_ec2_transit_gateway_connect", &resource.Sweeper{
// 		Name: "aws_ec2_transit_gateway_connect",
// 		F:    testSweepEc2TransitGatewayConnects,
// 	})
// }
//
// func testSweepEc2TransitGatewayConnects(region string) error {
// 	client, err := sharedClientForRegion(region)
// 	if err != nil {
// 		return fmt.Errorf("error getting client: %s", err)
// 	}
// 	conn := client.(*AWSClient).ec2conn
// 	input := &ec2.DescribeTransitGatewayConnectsInput{}
//
// 	for {
// 		output, err := conn.DescribeTransitGatewayConnects(input)
//
// 		if testSweepSkipSweepError(err) {
// 			log.Printf("[WARN] Skipping EC2 Transit Gateway connect sweep for %s: %s", region, err)
// 			return nil
// 		}
//
// 		if err != nil {
// 			return fmt.Errorf("error retrieving EC2 Transit Gateway connects: %s", err)
// 		}
//
// 		for _, attachment := range output.TransitGatewayConnects {
// 			// if aws.StringValue(attachment.ResourceType) != ec2.TransitGatewayAttachmentResourceTypeVpc {
// 			// 	continue
// 			// }
//
// 			if aws.StringValue(attachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
// 				continue
// 			}
//
// 			id := aws.StringValue(attachment.TransitGatewayAttachmentId)
//
// 			input := &ec2.DeleteTransitGatewayConnectInput{
// 				TransitGatewayAttachmentId: aws.String(id),
// 			}
//
// 			log.Printf("[INFO] Deleting EC2 Transit Gateway connect: %s", id)
// 			_, err := conn.DeleteTransitGatewayConnect(input)
//
// 			if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
// 				continue
// 			}
//
// 			if err != nil {
// 				return fmt.Errorf("error deleting EC2 Transit Gateway Connect (%s): %s", id, err)
// 			}
//
// 			if err := waitForEc2TransitGatewayConnectDeletion(conn, id); err != nil {
// 				return fmt.Errorf("error waiting for EC2 Transit Gateway Connect (%s) deletion: %s", id, err)
// 			}
// 		}
//
// 		if aws.StringValue(output.NextToken) == "" {
// 			break
// 		}
//
// 		input.NextToken = output.NextToken
// 	}
//
// 	return nil
// }

func TestAccAWSEc2TransitGatewayConnect_basic(t *testing.T) {
	var transitGatewayConnect1 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayVPCAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	// vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "gre"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transport_transit_gateway_attachment_id", transitGatewayVPCAttachmentResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSEc2TransitGatewayConnectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_connect" {
			continue
		}

		transitGatewayConnect, err := ec2DescribeTransitGatewayConnect(conn, rs.Primary.ID)

		if isAWSErr(err, "InvalidTransitGatewayConnectID.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if transitGatewayConnect == nil {
			continue
		}

		if aws.StringValue(transitGatewayConnect.State) != ec2.TransitGatewayAttachmentStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway Connect (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(transitGatewayConnect.State))
		}
	}

	return nil
}

func testAccAWSEc2TransitGatewayConnectConfig() string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() + `

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
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
`
}

func testAccCheckAWSEc2TransitGatewayConnectExists(resourceName string, transitGatewayConnect *ec2.TransitGatewayConnect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Connect ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		attachment, err := ec2DescribeTransitGatewayConnect(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if attachment == nil {
			return fmt.Errorf("EC2 Transit Gateway Connect not found")
		}

		if aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStateAvailable {
			return fmt.Errorf("EC2 Transit Gateway VPC connect (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(attachment.State))
		}

		*transitGatewayConnect = *attachment

		return nil
	}
}
