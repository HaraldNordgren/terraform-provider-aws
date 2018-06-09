package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsIamUserPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamUserPolicyAttachmentCreate,
		Read:   resourceAwsIamUserPolicyAttachmentRead,
		Delete: resourceAwsIamUserPolicyAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"user": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"policy_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsIamUserPolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	user := d.Get("user").(string)
	arn := d.Get("policy_arn").(string)

	err := attachPolicyToUser(conn, user, arn)
	if err != nil {
		return fmt.Errorf("[WARN] Error attaching policy %s to IAM User %s: %v", arn, user, err)
	}

	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", user)))
	return resourceAwsIamUserPolicyAttachmentRead(d, meta)
}

func resourceAwsIamUserPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	user := d.Get("user").(string)
	arn := d.Get("policy_arn").(string)

	_, err := conn.GetUser(&iam.GetUserInput{
		UserName: aws.String(user),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NoSuchEntity" {
				log.Printf("[WARN] No such entity found for Policy Attachment (%s)", user)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	attachedPolicies, err := conn.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(user),
	})
	if err != nil {
		return err
	}

	var policy string
	for _, p := range attachedPolicies.AttachedPolicies {
		if *p.PolicyArn == arn {
			policy = *p.PolicyArn
		}
	}

	if policy == "" {
		log.Printf("[WARN] No such User found for Policy Attachment (%s)", user)
		d.SetId("")
	}
	return nil
}

func resourceAwsIamUserPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	print("@@@@@@@@@@@@@@@@ resourceAwsIamUserPolicyAttachmentDelete 10\n")
	conn := meta.(*AWSClient).iamconn
	userRaw := d.Get("user")
	print("@@@@@@@@@@@@@@@@ resourceAwsIamUserPolicyAttachmentDelete 11 ", userRaw == nil, "\n")
	user := userRaw.(string)
	print("@@@@@@@@@@@@@@@@ resourceAwsIamUserPolicyAttachmentDelete 12\n")
	print("@@@@@@@@@@@@@@@@ resourceAwsIamUserPolicyAttachmentDelete 13 ", d.Get("policy_arn"), "\n")
	arn := d.Get("policy_arn").(string)

	err := detachPolicyFromUser(conn, user, arn)
	if err != nil {
		print("?&?&?&?&?&?&?& 13\n")
		return fmt.Errorf("[WARN] Error removing policy %s from IAM User %s: %v", arn, user, err)
	}
	print("?&?&?&?&?&?&?& 14\n")
	return nil
}

func attachPolicyToUser(conn *iam.IAM, user string, arn string) error {
	_, err := conn.AttachUserPolicy(&iam.AttachUserPolicyInput{
		UserName:  aws.String(user),
		PolicyArn: aws.String(arn),
	})
	if err != nil {
		return err
	}
	return nil
}

func detachPolicyFromUser(conn *iam.IAM, user string, arn string) error {
	_, err := conn.DetachUserPolicy(&iam.DetachUserPolicyInput{
		UserName:  aws.String(user),
		PolicyArn: aws.String(arn),
	})
	if err != nil {
		return err
	}
	return nil
}
