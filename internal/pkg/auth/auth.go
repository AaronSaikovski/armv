/*
MIT License

# Copyright (c) 2024 Aaron Saikovski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package auth

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// }
func GetLogin(ctx context.Context, subscriptionID string) bool {
	cred, err := GetAzureDefaultCredential()
	if err != nil {
		return false
	}

	client, err := SubscriptionClientCred(cred)
	if err != nil {
		return false
	}

	return GetSubscriptionClient(ctx, client, subscriptionID) == nil
}

// }
func GetAzureDefaultCredential() (*azidentity.DefaultAzureCredential, error) {
	return azidentity.NewDefaultAzureCredential(nil)

}

func NewResourceClient(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*armresources.Client, error) {
	clientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	return clientFactory.NewClient(), nil
}

//		return client, nil
//	}
func SubscriptionClientCred(cred *azidentity.DefaultAzureCredential) (*armsubscription.SubscriptionsClient, error) {
	// Azure SDK Resource Management clients accept the credential as a parameter.
	// The client will authenticate with the credential as necessary.
	return armsubscription.NewSubscriptionsClient(cred, nil)

}

func GetSubscriptionClient(ctx context.Context, client *armsubscription.SubscriptionsClient, subscriptionID string) error {
	_, err := client.Get(ctx, subscriptionID, nil)
	return err
}
