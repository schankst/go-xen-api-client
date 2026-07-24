/*
Package xenapi is a Go client for the XenAPI protocol used by XCP-ng,
XenServer, and Citrix Hypervisor.

This is a fork of github.com/terra-farm/go-xen-api-client, regenerated
against a current XenAPI schema (see SchemaXAPIRelease) instead of
upstream's last release, which only covered XenAPI through XenServer 7.3.
See the README for what else differs from upstream, and GitHub Releases
(https://github.com/schankst/go-xen-api-client/releases) for what changed
release to release.

# Usage

Create a Client with NewClient, log in via Client.Session.LoginWithPassword,
then call methods through the per-class fields on Client (Client.VM,
Client.Host, Client.SR, and so on - one field per XenAPI class):

	client, err := xenapi.NewClient("https://10.0.0.10/", nil)
	session, err := client.Session.LoginWithPassword("root", "password", "", "")
	vms, err := client.VM.GetAllRecords(session)
	err = client.Session.Logout(session)

Almost all of this package (every *_gen.go file) is generated from XenAPI's
own xenapi.json schema description, one Go method per XenAPI message and
one Go type per XenAPI class/enum - see MAINTAINING.md for how and when
that happens. error.go (XenAPI error code constants) is regenerated from a
separate upstream source on the same schedule. client.go, xenapi.go (the
generator itself), and the xmlrpc subpackage are hand-maintained.
*/
package xenapi
