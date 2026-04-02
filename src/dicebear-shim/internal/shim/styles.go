package shim

// GravatarToDiceBear maps Gravatar `d=` values to DICEbear style names.
var GravatarToDiceBear = map[string]string{
	"identicon": "identicon",
	"retro":     "pixel-art",
	"monsterid": "bottts",
	"wavatar":   "adventurer",
	"robohash":  "bottts",
	"mp":        "shapes",
	"blank":     "shapes",
}

// ResolveDiceBearStyle maps a Gravatar default param to a DICEbear style,
// falling back to the provided default style.
func ResolveDiceBearStyle(gravatarDefault, fallback string) string {
	if style, ok := GravatarToDiceBear[gravatarDefault]; ok {
		return style
	}
	return fallback
}
