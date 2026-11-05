package xmlrpc

import "testing"

// FuzzDecodeResponse checks that no input makes the decoder panic, hang or
// overflow the stack.
func FuzzDecodeResponse(f *testing.F) {
	for _, seed := range []string{
		wrapAsResponse("<value><string>x</string></value>"),
		wrapAsResponse("<value><struct><member><name>a</name><value><array><data><value><int>1</int></value></data></array></value></member></struct></value>"),
		wrapAsResponse("<value>\n <dateTime.iso8601>20260101T00:00:00</dateTime.iso8601>\n</value>"),
		`<methodResponse><fault><value><struct><member><name>faultCode</name><value><int>1</int></value></member></struct></value></fault></methodResponse>`,
		`<methodResponse><params><param><value><array><data>`,
		``,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, body []byte) {
		_, _ = decodeResponse(body)
	})
}
