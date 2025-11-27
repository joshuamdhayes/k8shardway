package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an instance of our component with the same files as before:
		website, err := NewAwsS3Website(ctx, "my-website", AwsS3WebsiteArgs{
			Files: []string{"index.html"},
		})
		if err != nil {
			return err
		}

		// And export its autoassigned URL:
		ctx.Export("url", website.Url)
		return nil
	})
}
