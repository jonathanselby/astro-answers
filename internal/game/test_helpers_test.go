package game

import (
	"fmt"
	"strconv"
)

func fmtSscanf(display string, a, b *int) (int, error) {
	return fmt.Sscanf(display, "%d x %d", a, b)
}

func intString(v int) string { return strconv.Itoa(v) }
