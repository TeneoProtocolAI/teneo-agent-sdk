package modules

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// safeGetFloat navigates nested JSON from CoinGecko safely
func safeGetFloat(m map[string]interface{}, path ...string) float64 {
	cur := interface{}(m)
	for i, p := range path {
		if mm, ok := cur.(map[string]interface{}); ok {
			v, exists := mm[p]
			if !exists {
				return 0
			}
			if i == len(path)-1 {
				switch x := v.(type) {
				case float64:
					return x
				case int:
					return float64(x)
				case int64:
					return float64(x)
				case string:
					f, _ := strconv.ParseFloat(x, 64)
					return f
				default:
					return 0
				}
			}
			cur = v
		} else {
			return 0
		}
	}
	return 0
}

func DebugLog(label string, v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Printf("[DEBUG] %s: %s\n", label, string(b))
}
