package rc

import (
	"bytes"
	"context"
)

func rc(args ...string) (string, string, error) {
	var out, errout bytes.Buffer
	err := Run(context.Background(), "1.2.3", args, &out, &errout)
	return out.String(), errout.String(), err
}
