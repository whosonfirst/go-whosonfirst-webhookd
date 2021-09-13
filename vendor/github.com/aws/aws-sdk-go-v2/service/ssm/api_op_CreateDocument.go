// Code generated by smithy-go-codegen DO NOT EDIT.

package ssm

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Creates a Amazon Web Services Systems Manager (SSM document). An SSM document
// defines the actions that Systems Manager performs on your managed instances. For
// more information about SSM documents, including information about supported
// schemas, features, and syntax, see Amazon Web Services Systems Manager Documents
// (https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-ssm-docs.html)
// in the Amazon Web Services Systems Manager User Guide.
func (c *Client) CreateDocument(ctx context.Context, params *CreateDocumentInput, optFns ...func(*Options)) (*CreateDocumentOutput, error) {
	if params == nil {
		params = &CreateDocumentInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "CreateDocument", params, optFns, c.addOperationCreateDocumentMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*CreateDocumentOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type CreateDocumentInput struct {

	// The content for the new SSM document in JSON or YAML format. We recommend
	// storing the contents for your new document in an external JSON or YAML file and
	// referencing the file in a command. For examples, see the following topics in the
	// Amazon Web Services Systems Manager User Guide.
	//
	// * Create an SSM document
	// (Amazon Web Services API)
	// (https://docs.aws.amazon.com/systems-manager/latest/userguide/create-ssm-document-api.html)
	//
	// *
	// Create an SSM document (Amazon Web Services CLI)
	// (https://docs.aws.amazon.com/systems-manager/latest/userguide/create-ssm-document-cli.html)
	//
	// *
	// Create an SSM document (API)
	// (https://docs.aws.amazon.com/systems-manager/latest/userguide/create-ssm-document-api.html)
	//
	// This member is required.
	Content *string

	// A name for the SSM document. You can't use the following strings as document
	// name prefixes. These are reserved by Amazon Web Services for use as document
	// name prefixes:
	//
	// * aws-
	//
	// * amazon
	//
	// * amzn
	//
	// This member is required.
	Name *string

	// A list of key-value pairs that describe attachments to a version of a document.
	Attachments []types.AttachmentsSource

	// An optional field where you can specify a friendly name for the SSM document.
	// This value can differ for each version of the document. You can update this
	// value at a later time using the UpdateDocument operation.
	DisplayName *string

	// Specify the document format for the request. The document format can be JSON,
	// YAML, or TEXT. JSON is the default format.
	DocumentFormat types.DocumentFormat

	// The type of document to create.
	DocumentType types.DocumentType

	// A list of SSM documents required by a document. This parameter is used
	// exclusively by AppConfig. When a user creates an AppConfig configuration in an
	// SSM document, the user must also specify a required document for validation
	// purposes. In this case, an ApplicationConfiguration document requires an
	// ApplicationConfigurationSchema document for validation purposes. For more
	// information, see What is AppConfig?
	// (https://docs.aws.amazon.com/appconfig/latest/userguide/what-is-appconfig.html)
	// in the AppConfig User Guide.
	Requires []types.DocumentRequires

	// Optional metadata that you assign to a resource. Tags enable you to categorize a
	// resource in different ways, such as by purpose, owner, or environment. For
	// example, you might want to tag an SSM document to identify the types of targets
	// or the environment where it will run. In this case, you could specify the
	// following key-value pairs:
	//
	// * Key=OS,Value=Windows
	//
	// *
	// Key=Environment,Value=Production
	//
	// To add tags to an existing SSM document, use
	// the AddTagsToResource operation.
	Tags []types.Tag

	// Specify a target type to define the kinds of resources the document can run on.
	// For example, to run a document on EC2 instances, specify the following value:
	// /AWS::EC2::Instance. If you specify a value of '/' the document can run on all
	// types of resources. If you don't specify a value, the document can't run on any
	// resources. For a list of valid resource types, see Amazon Web Services resource
	// and property types reference
	// (https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html)
	// in the CloudFormation User Guide.
	TargetType *string

	// An optional field specifying the version of the artifact you are creating with
	// the document. For example, "Release 12, Update 6". This value is unique across
	// all versions of a document, and can't be changed.
	VersionName *string

	noSmithyDocumentSerde
}

type CreateDocumentOutput struct {

	// Information about the SSM document.
	DocumentDescription *types.DocumentDescription

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationCreateDocumentMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpCreateDocument{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpCreateDocument{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpCreateDocumentValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opCreateDocument(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opCreateDocument(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "ssm",
		OperationName: "CreateDocument",
	}
}
