package main

import (
	"bytes"
	"errors"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structtag"
)

const _argLen = 2

type TagInfo struct {
	tags           string
	line, pos, end int
}

func main() {
	// golint: nomnd
	if len(os.Args) < _argLen {
		log.Panic("need input file path")
	}
	args := os.Args[1:]
	for _, arg := range args {
		excute(arg)
	}
}

func excute(input string) {
	p, err := filepath.Abs(input)
	if err != nil {
		log.Panic("file path invalid")
	}
	reader, err := os.OpenFile(p, os.O_RDWR, 0o666)
	if err != nil {
		log.Panicf("openfile: %v", err)
	}
	defer reader.Close()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", reader, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	tagInfos := make([]*TagInfo, 0)
	ast.Inspect(f, func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if t.Type == nil {
			return true
		}

		x, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}

		for _, field := range x.Fields.List {
			if field.Tag == nil {
				continue
			}

			if field.Comment.Text() == "" && field.Doc.Text() == "" {
				continue
			}

			var taginfo TagInfo
			taginfo.tags = buildTags(field.Tag.Value, field.Comment.Text(), field.Doc.Text())

			taginfo.pos = int(field.Tag.Pos())
			taginfo.end = int(field.Tag.End())
			taginfo.line = fset.Position(token.Pos(taginfo.pos)).Line
			tagInfos = append(tagInfos, &taginfo)
		}
		return true
	},
	)

	var buf bytes.Buffer
	rewriteTag(reader, &buf, tagInfos)
	bts, err := format.Source(buf.Bytes())
	if err != nil {
		log.Panic(err)
	}

	if err := reader.Truncate(0); err != nil {
		log.Panic(err)
	}
	_, _ = reader.Seek(0, 0)
	_, _ = reader.Write(bts)
}

func buildTags(tag string, docs ...string) string {
	tags, err := structtag.Parse(tag[1 : len(tag)-1])
	if err != nil {
		log.Panic(err)
	}
	handleDocs(tags, docs...)
	return tags.String()
}

func handleDocs(tags *structtag.Tags, docs ...string) {
	if len(docs) == 0 {
		return
	}
	for _, doc := range docs {
		comments := strings.Split(doc, "\n")
		for _, comment := range comments {
			comment = strings.TrimSpace(comment)
			if !strings.HasPrefix(comment, "@tag") {
				continue
			}
			comment = strings.TrimPrefix(comment, "@tag")
			comment = strings.TrimSpace(comment)

			if !strings.HasPrefix(comment, ":") {
				continue
			}

			comment = comment[1:]

			ts := strings.Split(comment, ";")

			for _, t := range ts {
				if !strings.Contains(t, ":") {
					v, _ := tags.Get(strings.TrimSpace(t))
					if v != nil {
						continue
					}
					if v == nil {
						jsonTag, _ := tags.Get("json")
						jsonName := jsonTag.Name
						_ = tags.Set(&structtag.Tag{
							Key:  t,
							Name: jsonName,
						})
					}
				} else {
					v := strings.SplitN(t, ":", 2)
					if len(v) <= 1 {
						continue
					}
					key := strings.TrimSpace(v[0])
					values := strings.Split(v[1], ",")
					if len(values) == 0 {
						continue
					}
					temp := make([]string, 0, len(values))
					var flag bool
					for i := 0; i < len(values); i++ {
						values[i] = strings.TrimSpace(values[i])
						if strings.HasPrefix(values[i], "override") {
							flag = true
							continue
						}
						temp = append(temp, values[i])
					}
					values = temp
					var (
						name string
						opts []string
					)

					name = values[0]

					if len(values) > 1 {
						opts = values[1:]
					}

					tag, _ := tags.Get(key)
					if tag == nil {
						newTag := &structtag.Tag{
							Key:     key,
							Name:    name,
							Options: opts,
						}

						_ = tags.Set(newTag)
					} else {
						if flag {
							newTag := &structtag.Tag{
								Key:     key,
								Name:    name,
								Options: opts,
							}
							_ = tags.Set(newTag)
						} else {
							t, _ := tags.Get(key)
							t.Options = append(t.Options, values...)
						}
					}
				}
			}
		}
	}
}


// firstChar split input string once.
// func firstChar(str string, step string) []string {
// 	for i := 0; i < len(str); i++ {
// 		if str[i] == step[0] {
// 			if i == len(str)-1 {
// 				return []string{str[:i], ""}
// 			}
// 			return []string{str[:i], str[i+1:]}
// 		}
// 	}
// 	return []string{str, ""}
// }

func rewriteTag(reader *os.File, writer *bytes.Buffer, tags []*TagInfo) {
	var begin int
	for _, tag := range tags {
		writeTo(begin, tag.pos, reader, writer, tag.tags)
		begin = tag.end - 2
	}

	buf := make([]byte, 1024)
	for {
		n, err := reader.ReadAt(buf, int64(begin))
		if errors.Is(err, io.EOF) {
			writer.Write(buf[:n])
			break
		}
		begin += 1024
		writer.Write(buf)
	}
}

func writeTo(begin, end int, reader *os.File, temp *bytes.Buffer, tag string) {
	buf := make([]byte, end-begin)
	n, err := reader.ReadAt(buf, int64(begin))
	if err != nil && !errors.Is(err, io.EOF) {
		log.Panic(err)
	}
	temp.Write(buf[:n])
	temp.WriteString(tag)
}
