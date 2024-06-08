// Package types holds common API types.
package types

// TextComponent is a text component used throughout the API.
// It should be JSON-marshalled and written as a String.
// https://wiki.vg/Text_formatting#Text_components
type TextComponent struct {
	// Text is the text.
	Text string `json:"text"`
}
