package portabletext

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

type Block struct {
	Type     string `json:"_type"`
	Style    string `json:"style,omitempty"`
	Children []Span `json:"children,omitempty"`
	MarkDefs []Mark `json:"markDefs,omitempty"`
}

type Span struct {
	Type  string   `json:"_type"`
	Text  string   `json:"text"`
	Marks []string `json:"marks,omitempty"`
}

type Mark struct {
	Type string `json:"_type"`
	Key  string `json:"_key"`
	Href string `json:"href,omitempty"`
}

func Render(blocks []interface{}) template.HTML {
	var buf bytes.Buffer

	for _, block := range blocks {
		if blockMap, ok := block.(map[string]interface{}); ok {
			renderBlock(&buf, blockMap)
		}
	}

	return template.HTML(buf.String())
}

func renderBlock(buf *bytes.Buffer, block map[string]interface{}) {
	blockType, _ := block["_type"].(string)
	style, _ := block["style"].(string)

	switch blockType {
	case "block":
		tag := "p"
		if style != "" {
			switch style {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				tag = style
			case "blockquote":
				tag = "blockquote"
			}
		}

		buf.WriteString(fmt.Sprintf("<%s>", tag))

		if children, ok := block["children"].([]interface{}); ok {
			for _, child := range children {
				if childMap, ok := child.(map[string]interface{}); ok {
					renderSpan(buf, childMap, block)
				}
			}
		}

		buf.WriteString(fmt.Sprintf("</%s>", tag))

	case "image":
		if asset, ok := block["asset"].(map[string]interface{}); ok {
			if ref, ok := asset["_ref"].(string); ok {
				imgUrl := fmt.Sprintf("https://cdn.sanity.io/images/5uh7zmgn/production/%s",
					strings.Replace(ref, "image-", "", 1))
				buf.WriteString(fmt.Sprintf("<img src=\"%s\" alt=\"\" />", imgUrl))
			}
		}

	case "code":
		language := "text"
		if lang, ok := block["language"].(string); ok {
			language = lang
		}

		code := ""
		if c, ok := block["code"].(string); ok {
			code = c
		}

		buf.WriteString(fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>",
			template.HTMLEscapeString(language), template.HTMLEscapeString(code)))
	}
}

func getMarkDefs(parentBlock map[string]interface{}) []map[string]interface{} {
	var markDefs []map[string]interface{}
	if mdefs, ok := parentBlock["markDefs"].([]interface{}); ok {
		for _, mdef := range mdefs {
			if mdefMap, ok := mdef.(map[string]interface{}); ok {
				markDefs = append(markDefs, mdefMap)
			}
		}
	}
	return markDefs
}

func processCustomMark(buf *bytes.Buffer, markKey string, markDefs []map[string]interface{}, isClosing bool) bool {
	for _, markDef := range markDefs {
		if key, ok := markDef["_key"].(string); ok && key == markKey {
			markType, _ := markDef["_type"].(string)

			switch markType {
			case "link":
				if isClosing {
					buf.WriteString("</a>")
				} else if href, ok := markDef["href"].(string); ok {
					buf.WriteString(fmt.Sprintf("<a href=\"%s\">", href))
				}
				return true
			}
		}
	}
	return false
}

func processStandardMark(buf *bytes.Buffer, markKey string, isClosing bool) {
	var tag string
	switch markKey {
	case "strong":
		tag = "strong"
	case "em":
		tag = "em"
	case "code":
		tag = "code"
	case "underline":
		tag = "u"
	case "strike-through":
		tag = "s"
	default:
		return
	}

	if isClosing {
		buf.WriteString(fmt.Sprintf("</%s>", tag))
	} else {
		buf.WriteString(fmt.Sprintf("<%s>", tag))
	}
}

func renderSpan(buf *bytes.Buffer, span map[string]interface{}, parentBlock map[string]interface{}) {
	text, _ := span["text"].(string)
	marks, hasMarks := span["marks"].([]interface{})

	if hasMarks && len(marks) > 0 {
		markDefs := getMarkDefs(parentBlock)

		for _, mark := range marks {
			if markKey, ok := mark.(string); ok {
				if !processCustomMark(buf, markKey, markDefs, false) {
					processStandardMark(buf, markKey, false)
				}
			}
		}
	}

	buf.WriteString(template.HTMLEscapeString(text))

	if hasMarks && len(marks) > 0 {
		markDefs := getMarkDefs(parentBlock)

		for i := len(marks) - 1; i >= 0; i-- {
			if markKey, ok := marks[i].(string); ok {
				if !processCustomMark(buf, markKey, markDefs, true) {
					processStandardMark(buf, markKey, true)
				}
			}
		}
	}
}
