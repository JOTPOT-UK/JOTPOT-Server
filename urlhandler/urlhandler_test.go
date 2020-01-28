package urlhandler

import (
	"errors"
	"net/url"
	"testing"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

type testingURLRequestInterface struct {
	URLp *url.URL
	jps.Request
}

func (i *testingURLRequestInterface) URL() *url.URL {
	return i.URLp
}

func testingURLServerFrame(u *url.URL) *jps.ServerFrame {
	return &jps.ServerFrame{
		Request: jps.IncomingRequest{
			Request: &testingURLRequestInterface{
				URLp:    u,
				Request: nil,
			},
		},
	}
}

func TestStripPort(t *testing.T) {
	u80 := url.URL{
		Host: "www.hello.world:80",
	}
	f := testingURLServerFrame(&u80)
	StripPort(f)
	if u80.Host != "www.hello.world" {
		t.Error("StripPort failed to strip www.hello.world:80, result:", u80.Host)
		return
	}
	ip, err := url.Parse("http://1.2.3.4:8080")
	if err != nil {
		panic(err)
	}
	f = testingURLServerFrame(ip)
	StripPort(f)
	if ip.Host != "1.2.3.4" {
		t.Error("StripPort failed to strip 1.2.3.4:8080, result:", ip.Host)
		return
	}
}

func TestStripHTTPPort(t *testing.T) {
	u80 := url.URL{
		Host: "www.hello.world:80",
	}
	f := testingURLServerFrame(&u80)
	StripHTTPPort(f)
	if u80.Host != "www.hello.world" {
		t.Error("StripHTTPPort failed to strip www.hello.world:80, result:", u80.Host)
		return
	}
	ip, err := url.Parse("http://1.2.3.4:8080")
	if err != nil {
		panic(err)
	}
	f = testingURLServerFrame(ip)
	StripHTTPPort(f)
	if ip.Host != "1.2.3.4:8080" {
		t.Error("StripHTTPPort failed to not strip 1.2.3.4:8080, result:", ip.Host)
		return
	}
	u443, err := url.Parse("https://jotpot.uk:443/server")
	if err != nil {
		panic(err)
	}
	f = testingURLServerFrame(u443)
	StripHTTPPort(f)
	if u443.Host != "jotpot.uk" {
		t.Error("StripHTTPPort failed to strip jotpot.uk:443, result:", u443.Host)
		return
	}
	u444, err := url.Parse("https://jotpot.uk:444/server")
	if err != nil {
		panic(err)
	}
	f = testingURLServerFrame(u444)
	StripHTTPPort(f)
	if u444.Host != "jotpot.uk:444" {
		t.Error("StripHTTPPort failed to not strip jotpot.uk:444, result:", u444.Host)
		return
	}
}

func poe(u *url.URL, err error) *url.URL {
	if err != nil {
		panic(err)
	}
	return u
}

func TestHostSplitter(t *testing.T) {
	ErrTest := errors.New("test error to be returned")
	correctFrame := true
	helloWorld80Frame := testingURLServerFrame(poe(url.Parse("http://www.helloworld.com:80/")))
	helloWorld80Counter := 0
	helloWorld443Frame := testingURLServerFrame(poe(url.Parse("https://www.helloworld.com:443/hello/world?Oooo")))
	helloWorld443Counter := 0
	helloWorldNPFrame := testingURLServerFrame(poe(url.Parse("https://www.helloworld.com")))
	helloWorldNPCounter := 0
	localhostFrame := testingURLServerFrame(poe(url.Parse("http://localhost/index.html#menu")))
	localhostCounter := 0
	handler := HostSplitter(map[string]*jps.ServerFrameHandler{
		"helloworld.com:80": &jps.ServerFrameHandler{
			//Protocol: "http",
			F: func(f *jps.ServerFrame) (bool, error) {
				if f != helloWorld80Frame {
					correctFrame = false
				}
				helloWorld80Counter++
				return true, nil
			},
		},
		"helloworld.com:443": &jps.ServerFrameHandler{
			Protocol: "https",
			F: func(f *jps.ServerFrame) (bool, error) {
				if f != helloWorld443Frame {
					correctFrame = false
				}
				helloWorld443Counter++
				return true, nil
			},
		},
		"localhost": &jps.ServerFrameHandler{
			Protocol: "http",
			F: func(f *jps.ServerFrame) (bool, error) {
				if f != localhostFrame {
					correctFrame = false
				}
				localhostCounter++
				return false, ErrTest
			},
		},
		"helloworld.com": &jps.ServerFrameHandler{
			Protocol: "http",
			F: func(f *jps.ServerFrame) (bool, error) {
				if f != helloWorldNPFrame {
					correctFrame = false
				}
				helloWorldNPCounter++
				return true, nil
			},
		},
	}).Handle

	b, e := handler(helloWorld80Frame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com:80 handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com:80 handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com:80 handler")
		return
	}
	if helloWorld80Counter != 1 {
		t.Error("Incorrect handler called for helloworld.com:80")
		return
	}

	b, e = handler(helloWorld80Frame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com:80 handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com:80 handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com:80 handler")
		return
	}
	if helloWorld80Counter != 2 {
		t.Error("Incorrect handler called for helloworld.com:80")
		return
	}

	b, e = handler(helloWorld443Frame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com:443 handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com:443 handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com:443 handler")
		return
	}
	if helloWorld443Counter != 1 {
		t.Error("Incorrect handler called for helloworld.com:443")
		return
	}

	b, e = handler(helloWorld80Frame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com:80 handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com:80 handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com:80 handler")
		return
	}
	if helloWorld80Counter != 3 {
		t.Error("Incorrect handler called for helloworld.com:80")
		return
	}

	b, e = handler(helloWorld443Frame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com:443 handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com:443 handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com:443 handler")
		return
	}
	if helloWorld443Counter != 2 {
		t.Error("Incorrect handler called for helloworld.com:443")
		return
	}

	b, e = handler(helloWorld80Frame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com:80 handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com:80 handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com:80 handler")
		return
	}
	if helloWorld80Counter != 4 {
		t.Error("Incorrect handler called for helloworld.com:80")
		return
	}

	b, e = handler(helloWorldNPFrame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com handler")
		return
	}
	if helloWorldNPCounter != 1 {
		t.Error("Incorrect handler called for helloworld.com")
		return
	}

	b, e = handler(helloWorldNPFrame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com handler")
		return
	}
	if helloWorldNPCounter != 2 {
		t.Error("Incorrect handler called for helloworld.com")
		return
	}

	b, e = handler(helloWorld80Frame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com:80 handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com:80 handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com:80 handler")
		return
	}
	if helloWorld80Counter != 5 {
		t.Error("Incorrect handler called for helloworld.com:80")
		return
	}

	b, e = handler(localhostFrame)
	if e != nil {
		t.Error("Incorrect error returned from localhost handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from localhost handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to localhost handler")
		return
	}
	if localhostCounter != 1 {
		t.Error("Incorrect handler called for localhost")
		return
	}

	b, e = handler(helloWorldNPFrame)
	if e != nil {
		t.Error("Incorrect error returned from helloworld.com handler, expecring nil, got", e)
		return
	}
	if !b {
		t.Error("Incorrect bool returned from helloworld.com handler, expecring true, got", b)
		return
	}
	if !correctFrame {
		t.Error("Correct frame not passed to helloworld.com handler")
		return
	}
	if helloWorldNPCounter != 3 {
		t.Error("Incorrect handler called for helloworld.com")
		return
	}
}
