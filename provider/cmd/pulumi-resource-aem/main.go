package main

import (
	p "github.com/pulumi/pulumi-go-provider"

	aem "github.com/wttech/pulumi-aem-native/provider"
)

// Serve the provider against Pulumi's Provider protocol.
func main() { p.RunProvider(aem.Name, aem.Version, aem.Provider()) }
