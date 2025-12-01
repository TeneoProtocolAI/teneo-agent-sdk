package modules

import "fmt"

func RunWatch(args []string) (string, error) {
    if len(args) == 0 {
        return "Usage: watch [KOL]", nil
    }
    kol := args[0]
    return fmt.Sprintf("Now watching KOL: %s (mock). Alerts will be generated on significant activity.", kol), nil
}
