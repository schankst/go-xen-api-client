package xmlrpc

// Struct is the Go representation of an XML-RPC <struct>: an unordered set
// of named members, each itself an XML-RPC value.
type Struct map[string]interface{}

// Base64 marks a string as XML-RPC <base64> data rather than <string> data.
type Base64 string
