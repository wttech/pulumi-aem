package main

import (
	p "github.com/pulumi/pulumi-go-provider"

	aem "github.com/wttech/pulumi-provider-aem/provider"
)

// Serve the provider against Pulumi's Provider protocol.
func main() { p.RunProvider(aem.Name, aem.Version, aem.Provider()) }
