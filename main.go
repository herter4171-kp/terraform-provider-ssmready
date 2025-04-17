
package main

import (
    "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
    "ssmready/provider"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        ProviderFunc: provider.Provider,
    })
}
