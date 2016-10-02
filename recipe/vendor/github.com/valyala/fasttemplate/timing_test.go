package fasttemplate

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"
	"text/template"
)

var (
	source        = "http://{{uid}}.foo.bar.com/?cb={{cb}}{{width}}&width={{width}}&height={{height}}&timeout={{timeout}}&uid={{uid}}&subid={{subid}}&ref={{ref}}"
	result        = "http://aaasdf.foo.bar.com/?cb=12341232&width=1232&height=123&timeout=123123&uid=aaasdf&subid=asdfds&ref=http://google.com/aaa/bbb/ccc"
	resultEscaped = "http://aaasdf.foo.bar.com/?cb=12341232&width=1232&height=123&timeout=123123&uid=aaasdf&subid=asdfds&ref=http%3A%2F%2Fgoogle.com%2Faaa%2Fbbb%2Fccc"

	resultBytes        = []byte(result)
	resultEscapedBytes = []byte(resultEscaped)

	m = map[string]interface{}{
		"cb":      []byte("1234"),
		"width":   []byte("1232"),
		"height":  []byte("123"),
		"timeout": []byte("123123"),
		"uid":     []byte("aaasdf"),
		"subid":   []byte("asdfds"),
		"ref":     []byte("http://google.com/aaa/bbb/ccc"),
	}
)

func map2slice(m map[string]interface{}) []string {
	var a []string
	for k, v := range m {
		a = append(a, "{{"+k+"}}", string(v.([]byte)))
	}
	return a
}

func BenchmarkFmtFprintf(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var w bytes.Buffer
		for pb.Next() {
			fmt.Fprintf(&w,
				"http://%[5]s.foo.bar.com/?cb=%[1]s%[2]s&width=%[2]s&height=%[3]s&timeout=%[4]s&uid=%[5]s&subid=%[6]s&ref=%[7]s",
				m["cb"], m["width"], m["height"], m["timeout"], m["uid"], m["subid"], m["ref"])
			x := w.Bytes()
			if !bytes.Equal(x, resultBytes) {
				b.Fatalf("Unexpected result\n%q\nExpected\n%q\n", x, result)
			}
			w.Reset()
		}
	})
}

func BenchmarkStringsReplace(b *testing.B) {
	mSlice := map2slice(m)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			x := source
			for i := 0; i < len(mSlice); i += 2 {
				x = strings.Replace(x, mSlice[i], mSlice[i+1], -1)
			}
			if x != result {
				b.Fatalf("Unexpected result\n%q\nExpected\n%q\n", x, result)
			}
		}
	})
}

func BenchmarkStringsReplacer(b *testing.B) {
	mSlice := map2slice(m)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := strings.NewReplacer(mSlice...)
			x := r.Replace(source)
			if x != result {
				b.Fatalf("Unexpected result\n%q\nExpected\n%q\n", x, result)
			}
		}
	})
}

func BenchmarkTextTemplate(b *testing.B) {
	s := strings.Replace(source, "{{", "{{.", -1)
	t, err := template.New("test").Parse(s)
	if err != nil {
		b.Fatalf("Error when parsing template: %s", err)
	}

	mm := make(map[string]string)
	for k, v := range m {
		mm[k] = string(v.([]byte))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var w bytes.Buffer
		for pb.Next() {
			if err := t.Execute(&w, mm); err != nil {
				b.Fatalf("error when executing template: %s", err)
			}
			x := w.Bytes()
			if !bytes.Equal(x, resultBytes) {
				b.Fatalf("unexpected result\n%q\nExpected\n%q\n", x, resultBytes)
			}
			w.Reset()
		}
	})
}

func BenchmarkFastTemplateExecuteFunc(b *testing.B) {
	t, err := NewTemplate(source, "{{", "}}")
	if err != nil {
		b.Fatalf("error in template: %s", err)
	}

	f := func(w io.Writer, tag string) (int, error) {
		return w.Write(m[tag].([]byte))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var w bytes.Buffer
		for pb.Next() {
			if _, err := t.ExecuteFunc(&w, f); err != nil {
				b.Fatalf("unexpected error: %s", err)
			}
			x := w.Bytes()
			if !bytes.Equal(x, resultBytes) {
				b.Fatalf("unexpected result\n%q\nExpected\n%q\n", x, resultBytes)
			}
			w.Reset()
		}
	})
}

func BenchmarkFastTemplateExecute(b *testing.B) {
	t, err := NewTemplate(source, "{{", "}}")
	if err != nil {
		b.Fatalf("error in template: %s", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var w bytes.Buffer
		for pb.Next() {
			if _, err := t.Execute(&w, m); err != nil {
				b.Fatalf("unexpected error: %s", err)
			}
			x := w.Bytes()
			if !bytes.Equal(x, resultBytes) {
				b.Fatalf("unexpected result\n%q\nExpected\n%q\n", x, resultBytes)
			}
			w.Reset()
		}
	})
}

func BenchmarkFastTemplateExecuteFuncString(b *testing.B) {
	t, err := NewTemplate(source, "{{", "}}")
	if err != nil {
		b.Fatalf("error in template: %s", err)
	}

	f := func(w io.Writer, tag string) (int, error) {
		return w.Write(m[tag].([]byte))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			x := t.ExecuteFuncString(f)
			if x != result {
				b.Fatalf("unexpected result\n%q\nExpected\n%q\n", x, result)
			}
		}
	})
}

func BenchmarkFastTemplateExecuteString(b *testing.B) {
	t, err := NewTemplate(source, "{{", "}}")
	if err != nil {
		b.Fatalf("error in template: %s", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			x := t.ExecuteString(m)
			if x != result {
				b.Fatalf("unexpected result\n%q\nExpected\n%q\n", x, result)
			}
		}
	})
}

func BenchmarkFastTemplateExecuteTagFunc(b *testing.B) {
	t, err := NewTemplate(source, "{{", "}}")
	if err != nil {
		b.Fatalf("error in template: %s", err)
	}

	mm := make(map[string]interface{})
	for k, v := range m {
		if k == "ref" {
			vv := v.([]byte)
			v = TagFunc(func(w io.Writer, tag string) (int, error) { return w.Write([]byte(url.QueryEscape(string(vv)))) })
		}
		mm[k] = v
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var w bytes.Buffer
		for pb.Next() {
			if _, err := t.Execute(&w, mm); err != nil {
				b.Fatalf("unexpected error: %s", err)
			}
			x := w.Bytes()
			if !bytes.Equal(x, resultEscapedBytes) {
				b.Fatalf("unexpected result\n%q\nExpected\n%q\n", x, resultEscapedBytes)
			}
			w.Reset()
		}
	})
}
