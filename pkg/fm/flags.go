// Package fm wires lazygit up to CloudBees Feature Management (the Rox SDK),
// and declares the feature flags lazygit currently reads from it.
package fm

import (
	"fmt"
	"os"
	"runtime"

	roxcontext "github.com/rollout/rox-go/v6/core/context"
	"github.com/rollout/rox-go/v6/core/model"
	"github.com/rollout/rox-go/v6/server"
)

// FlagName_CommitGraphStyle is the fully qualified name of the CommitGraphStyle
// flag, for use with the dynamic API.
const FlagName_CommitGraphStyle = "ux.CommitGraphStyle"

// flags is the container for flags registered under the default namespace.
type flags struct {
	// EnableWatchmanIntegration controls whether lazygit treats filesystem
	// change notifications as if they came from a watchman-style instant
	// notifier, rather than relying purely on interval-based polling.
	//
	// Registered under the key "watchman-integration-enabled" rather than the
	// Go field name above (see the `flagName` struct tag).
	EnableWatchmanIntegration server.RoxFlag `flagName:"watchman-integration-enabled"`
}

// uxFlags is the container for flags registered under the 'ux' namespace.
type uxFlags struct {
	// CommitGraphStyle picks the characters used to draw the commit graph,
	// as a fallback/override on top of the user's own `gui.commitGraphStyle`
	// config. Fully qualified name: "ux.CommitGraphStyle".
	CommitGraphStyle server.RoxString

	// DiffContextLines overrides the number of unified-diff context lines
	// shown around each change, as a fallback/override on top of the user's
	// own `git.diffContextSize` config. Fully qualified name:
	// "ux.DiffContextLines".
	DiffContextLines server.RoxInt
}

var Flags = &flags{
	EnableWatchmanIntegration: server.NewRoxFlag(false),
}

var UxFlags = &uxFlags{
	CommitGraphStyle: server.NewRoxString("ascii", []string{"ascii", "unicode"}),
	DiffContextLines: server.NewRoxInt(3, []int{1, 3, 5, 10}),
}

var rox = server.NewRox()

// Init sets up the CloudBees Feature Management SDK and registers lazygit's
// flag containers. gitVersion is included as a computed global custom
// property so that targeting rules can vary by the user's installed git
// version.
func Init(gitVersion string) {
	rox.RegisterWithEmptyNamespace(Flags)
	rox.Register("ux", UxFlags)

	rox.SetCustomStringProperty("client.os", runtime.GOOS)
	rox.SetCustomComputedStringProperty("client.gitVersion", func(ctx roxcontext.Context) string {
		return gitVersion
	})

	sdkKey := os.Getenv("FM_SDK_KEY")
	if sdkKey == "" {
		fmt.Println("FM_SDK_KEY environment variable not set. Feature flags not initialised.")
		return
	}

	options := server.NewRoxOptions(server.RoxOptionsBuilder{})
	<-rox.Setup(sdkKey, options)
}

// DynamicAPI exposes the dynamic API for callers that look up a flag by name
// instead of through the static containers above.
func DynamicAPI() model.DynamicAPI {
	return rox.DynamicAPI()
}

// NewContext is a small convenience wrapper so call sites don't need to
// import the rox context package directly just to build one.
func NewContext(properties map[string]interface{}) roxcontext.Context {
	return roxcontext.NewContext(properties)
}
