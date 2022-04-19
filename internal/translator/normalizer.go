package translator

import "knflows/sw/internal/types"

// Normalize the SW spec, e.g. make defaults explicits
func Normalize(sw *types.SW) {
    if sw.Start == nil && len(sw.States) > 0 {
        sw.Start = sw.States[0]
    }
}
