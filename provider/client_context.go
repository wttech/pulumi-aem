package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/wttech/pulumi-aem-native/provider/client"
)

type ClientContext[T interface{}] struct {
	cl   *client.Client
	ctx  p.Context
	data T
}
