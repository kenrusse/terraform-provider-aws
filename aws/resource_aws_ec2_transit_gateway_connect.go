package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TransitGatewayConnect() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayConnectCreate,
		Read:   resourceAwsEc2TransitGatewayConnectRead,
		Delete: resourceAwsEc2TransitGatewayConnectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// "tags":     tagsSchema(),
			"tags":     tagsSchemaForceNew(),
			"tags_all": tagsSchemaComputed(),
			"transport_transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayConnectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	TransitGatewayConnect, err := ec2DescribeTransitGatewayConnect(conn, d.Id())

	if isAWSErr(err, "InvalidTransitGatewayConnectID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway Connect (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect: %s", err)
	}

	if TransitGatewayConnect == nil {
		log.Printf("[WARN] EC2 Transit Gateway Connect (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if TransitGatewayConnect.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect (%s): missing options", d.Id())
	}

	d.Set("protocol", TransitGatewayConnect.Options.Protocol)

	tags := keyvaluetags.Ec2KeyValueTags(TransitGatewayConnect.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if TransitGatewayConnect.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment (%s): missing options", d.Id())
	}

	d.Set("protocol", TransitGatewayConnect.Options.Protocol)

	d.Set("transit_gateway_id", TransitGatewayConnect.TransitGatewayId)
	d.Set("transport_transit_gateway_attachment_id", TransitGatewayConnect.TransportTransitGatewayAttachmentId)

	return nil
}

func resourceAwsEc2TransitGatewayConnectCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	// peerAccountId := meta.(*AWSClient).accountid
	// if v, ok := d.GetOk("peer_account_id"); ok {
	// 	peerAccountId = v.(string)
	// }

	transportTransitGatewayAttachmentId := d.Get("transport_transit_gateway_attachment_id").(string)

	input := &ec2.CreateTransitGatewayConnectInput{
		Options: &ec2.CreateTransitGatewayConnectRequestOptions{
			Protocol: aws.String(d.Get("protocol").(string)),
		},
		TransportTransitGatewayAttachmentId: aws.String(transportTransitGatewayAttachmentId),
		TagSpecifications:                   ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayAttachment),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Connect Attachment: %s", input)
	output, err := conn.CreateTransitGatewayConnect(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Connect Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayConnect.TransitGatewayAttachmentId))

	// if err := waitForEc2TransitGatewayConnectCreation(conn, d.Id()); err != nil {
	// 	return fmt.Errorf("error waiting for EC2 Transit Gateway Connect Attachment (%s) availability: %s", d.Id(), err)
	// }

	return resourceAwsEc2TransitGatewayConnectRead(d, meta)
}

func resourceAwsEc2TransitGatewayConnectDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteTransitGatewayConnectInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Connect (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayConnect(input)

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Connnect: %s", err)
	}

	if err := waitForEc2TransitGatewayConnectDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Connect Attachment (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
