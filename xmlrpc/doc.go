/*
Package xmlrpc implements the parts of the XML-RPC specification
(http://xmlrpc.scripting.com/spec.html) that XenAPI actually uses: method
calls with string/int/double/boolean/dateTime.iso8601/base64/struct/array
parameters and results, plus genuine <fault> responses.

Written from scratch for this fork (replacing a vendored copy of
github.com/amfranz/go-xmlrpc-client, which had no license) using
strings.Builder and encoding/xml directly instead of net/rpc's
ClientCodec plumbing - XML-RPC over HTTP is one request, one response,
with no multiplexed-connection semantics to justify net/rpc's machinery.

Used internally by github.com/schankst/go-xen-api-client via NewClient
and Client.Call; not intended to be used directly by consumers of that
package.
*/
package xmlrpc
