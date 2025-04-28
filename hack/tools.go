//go:build tools

package hack

import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
)
