// package config contains helpers for getting LORIS config
// variables from a LORIS database (or config.xml file)
package config

// All Configs are assumed to be strings, because we can't
// infer type from the XML
type Key string
type Value string

// A Values represents a map from key to value for a LORIS
// connection.
type Values map[Key]Value
