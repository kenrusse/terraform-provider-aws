package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2TransitGatewayConnect() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2TransitGatewayConnectRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
      "transport_transit_gateway_attachment_id": {
        Type:     schema.TypeString,
        Computed: true,
      },
			"tags": tagsSchemaComputed(),
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
      "protocol": {
        Type:     schema.TypeString,
        Computed: true,
      },
		},
	}
}

func dataSourceAwsEc2TransitGatewayConnectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayConnectsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayAttachmentIds = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway Connects: %s", input)
	output, err := conn.DescribeTransitGatewayConnects(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connects: %w", err)
	}

	if output == nil || len(output.TransitGatewayConnects) == 0 {
		return errors.New("error reading EC2 Transit Connects: no results found")
	}

	if len(output.TransitGatewayConnects) > 1 {
		return errors.New("error reading EC2 Transit Gateway Connects: multiple results found, try adjusting search criteria")
	}

	transitGatewayConnect := output.TransitGatewayConnects[0]

	if transitGatewayConnect == nil {
		return errors.New("error reading EC2 Transit Gateway Connect: empty result")
	}

	if transitGatewayConnect.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect (%s): missing options", d.Id())
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(transitGatewayConnect.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

  d.Set("Protocol", transitGatewayConnect.Options.Protocol)
	d.Set("transit_gateway_id", transitGatewayConnect.TransitGatewayId)
	d.SetId(aws.StringValue(transitGatewayConnect.TransitGatewayAttachmentId))

	return nil
}
