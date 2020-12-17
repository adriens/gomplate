//+build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

type NestedTemplatesSuite struct {
	tmpDir *fs.Dir
	srv    *http.Server
}

var _ = Suite(&NestedTemplatesSuite{})

func (s *NestedTemplatesSuite) SetUpSuite(c *C) {

	one := `{{ . }}`
	two := `{{ range $n := (seq 2) }}{{ $n }}: {{ $ }} {{ end }}`

	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("hello.t", `Hello {{ . }}!`),
		fs.WithDir("templates",
			fs.WithFile("one.t", one),
			fs.WithFile("two.t", two),
		),
	)

	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	handle(c, err)
	h := http.NewServeMux()
	h.HandleFunc("/one.t", typeHandler("text/plain", one))
	h.HandleFunc("/two.t", typeHandler("text/plain", two))
	s.srv = &http.Server{Handler: h}
	go s.srv.Serve(l)
}

func (s *NestedTemplatesSuite) TearDownSuite(c *C) {
	s.tmpDir.Remove()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	s.srv.Shutdown(ctx)
}

func (s *NestedTemplatesSuite) TestNestedTemplates(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-t", "hello="+s.tmpDir.Join("hello.t"),
		"-i", `{{ template "hello" "World"}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "Hello World!"})

	result = icmd.RunCmd(icmd.Cmd{
		Command: []string{
			GomplateBin,
			"-t", "hello.t",
			"-i", `{{ template "hello.t" "World"}}`,
		},
		Dir: s.tmpDir.Path(),
	},
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "Hello World!"})

	result = icmd.RunCmd(icmd.Cmd{
		Command: []string{
			GomplateBin,
			"-t", "templates/",
			"-i", `{{ template "templates/one.t" "one"}}
{{ template "templates/two.t" "two"}}`,
		},
		Dir: s.tmpDir.Path(),
	},
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `one
1: two 2: two`})

	result = icmd.RunCmd(icmd.Cmd{
		Command: []string{
			GomplateBin,
			"-t", fmt.Sprintf("file://%s/templates/", s.tmpDir.Path()),
			"-i", `{{ template "templates/one.t" "one"}}
{{ template "templates/two.t" "two"}}`,
		},
		Dir: s.tmpDir.Path(),
	},
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `one
1: two 2: two`})
}
