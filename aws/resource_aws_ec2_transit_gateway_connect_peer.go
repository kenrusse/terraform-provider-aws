package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TransitGatewayConnectPeer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayConnectPeerCreate,
		Read:   resourceAwsEc2TransitGatewayConnectPeerRead,
		Delete: resourceAwsEc2TransitGatewayConnectPeerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: SetTagsDiff,

		Schema: map[string]*schema.Schema{
			// "tags":     tagsSchema(),
			"tags":     tagsSchemaForceNew(),
			"tags_all": tagsSchemaComputed(),
			"transport_transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"peer_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"peer_asn": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"inside_cidr_blocks": {
				Type:     schema.TypeList,
				Elem:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayConnectPeerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	TransitGatewayConnectPeer, err := ec2DescribeTransitGatewayConnectPeer(conn, d.Id())

	if isAWSErr(err, "InvalidTransitGatewayConnectPeerID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway Connect peer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect peer: %s", err)
	}

	if TransitGatewayConnectPeer == nil {
		log.Printf("[WARN] EC2 Transit Gateway Connect peer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if TransitGatewayConnectPeer.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect peer (%s): missing options", d.Id())
	}

	tags := keyvaluetags.Ec2KeyValueTags(TransitGatewayConnectPeer.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

  d.Set("transport_transit_gateway_attachment_id", TransitGatewayConnectPeer.TransportTransitGatewayAttachmentId)
	d.Set("transit_gateway_address", TransitGatewayConnectPeer.TransitGatewayAddress)
	d.Set("peer_address", TransitGatewayConnectPeer.PeerAddress)

	d.Set("inside_cidr_blocks", TransitGatewayConnectPeer.InsideCidrBlocks)
  d.Set("peer_asn", TransitGatewayConnectPeer.ConnectPeerConfiguration.PeerAsn)

	return nil
}

func resourceAwsEc2TransitGatewayConnectPeerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	transportTransitGatewayAttachmentId := d.Get("transport_transit_gateway_attachment_id").(string)
  transitGatewayAddress := d.Get("transit_gateway_address").(string)
	peerAddress := d.Get("peer_address").(string)
	insideCidrBlocks := d.Get("inside_cidr_blocks").(string)
	peerAsn := d.Get("peer_asn").(int)

	input := &ec2.CreateTransitGatewayConnectPeerInput{
		BgpOptions: &ec2.CreateTransitGatewayConnectPeerRequestOptions{
			PeerAsn: aws.Int64(int64(peerAsn.(int))),
		},
		TransportTransitGatewayAttachmentId: aws.String(transportTransitGatewayAttachmentId),
		TransitGatewayAddress: aws.String(transitGatewayAddress),
		PeerAddress: aws.String(peerAddress),
		InsideCidrBlocks: aws.String(insideCidrBlocks),
		TagSpecifications:                   ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayAttachment),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Connect peer: %s", input)
	output, err := conn.CreateTransitGatewayConnectPeer(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Connect peer: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayConnectPeer.TransitGatewayConnectPeerId))

	if err := waitForEc2TransitGatewayConnectPeerCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Connect Peer (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsEc2TransitGatewayConnectPeerRead(d, meta)
}

func resourceAwsEc2TransitGatewayConnectPeerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteTransitGatewayConnectPeerInput{
		TransitGatewayConnectPeerId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Connect Peer (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayConnectPeer(input)

	if isAWSErr(err, "InvalidTransitGatewayConnnectPeerID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Connnect Peer: %s", err)
	}

	if err := waitForEc2TransitGatewayConnectPeerDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Connect Peer (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
