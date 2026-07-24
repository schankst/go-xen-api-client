/*
Package xmlrpc implements the XML-RPC specification:
http://xmlrpc.scripting.com/spec.html

Vendored from github.com/amfranz/go-xmlrpc-client (commit 76858463955d,
2019-06-12) into this fork so github.com/schankst/go-xen-api-client has no
external module dependencies of its own. No LICENSE file was present in the
upstream repository at the time of vendoring; this notice preserves
attribution to the original author.

Used internally by github.com/schankst/go-xen-api-client via NewClient and
Client.Call; not intended to be used directly by consumers of that package.
*/
package xmlrpc
