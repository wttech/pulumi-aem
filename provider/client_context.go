package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/wttech/pulumi-aem/provider/client"
)

type ClientContext[T interface{}] struct {
	cl   *client.Client
	log  p.Logger
	data T
}
