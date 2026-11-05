// Package xenapi is a client for the XenAPI of XCP-ng and XenServer.
//
// Create a Client with NewClient and log in with
// Client.Session.LoginWithPassword. Each XenAPI class is a field of Client
// (Client.VM, Client.Host, Client.SR, ...) with one method per XenAPI
// message. Every method takes a context.Context and, except for login, the
// session.
//
// Failures reported by XenAPI are *Error values; match them with
// errors.Is against the ErrorCode constants.
//
// Everything in *_gen.go and error.go is generated from XenAPI's schema;
// see SchemaXAPIRelease for the release it was generated from.
package xenapi
