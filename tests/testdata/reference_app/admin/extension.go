package refadmin

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/forgego/forge/admin"
)

//go:embed assets
var assetsFS embed.FS

// Register registers the reference admin extension with Forge Admin.
func Register() error {
	sub, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		return fmt.Errorf("reference extension assets: %w", err)
	}
	return admin.RegisterExtension(admin.Extension{
		Name:    "reference",
		Assets:  sub,
		Version: admin.ExtensionAPIVersion,
	})
}
