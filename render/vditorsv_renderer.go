// Lute - 一款对中文语境优化的 Markdown 引擎，支持 Go 和 JavaScript
// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package render

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/88250/lute/html"

	"github.com/88250/lute/ast"
	"github.com/88250/lute/lex"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/util"
)

// VditorSVRenderer 描述了 Vditor Split-View DOM 渲染器。
type VditorSVRenderer struct {
	*BaseRenderer
	nodeWriterStack        []*bytes.Buffer // 节点输出缓冲栈
	needRenderFootnotesDef bool
	LastOut                []byte // 最新输出的 newline 长度个字节
}

var newline = []byte("<span><br /><span style=\"display: none\">\n</span></span>")

func (r *VditorSVRenderer) WriteByte(c byte) {
	r.Writer.WriteByte(c)
	r.LastOut = append(r.LastOut, c)
	if len(newline) < len(r.LastOut) {
		r.LastOut = r.LastOut[len(r.LastOut)-len(newline):]
	}
}

func (r *VditorSVRenderer) Write(content []byte) {
	if length := len(content); 0 < length {
		r.Writer.Write(content)
		r.LastOut = append(r.LastOut, content...)
		if len(newline) < len(r.LastOut) {
			r.LastOut = r.LastOut[len(r.LastOut)-len(newline):]
		}
	}
}

func (r *VditorSVRenderer) WriteString(content string) {
	if length := len(content); 0 < length {
		r.Writer.WriteString(content)
		r.LastOut = append(r.LastOut, content...)
		if len(newline) < len(r.LastOut) {
			r.LastOut = r.LastOut[len(r.LastOut)-len(newline):]
		}
	}
}

func (r *VditorSVRenderer) Newline() {
	if !bytes.Equal(newline, r.LastOut) {
		r.Writer.Write(newline)
		r.LastOut = newline
	}
}

// NewVditorSVRenderer 创建一个 Vditor Split-View DOM 渲染器
func NewVditorSVRenderer(tree *parse.Tree) *VditorSVRenderer {
	ret := &VditorSVRenderer{BaseRenderer: NewBaseRenderer(tree)}
	ret.RendererFuncs[ast.NodeDocument] = ret.renderDocument
	ret.RendererFuncs[ast.NodeParagraph] = ret.renderParagraph
	ret.RendererFuncs[ast.NodeText] = ret.renderText
	ret.RendererFuncs[ast.NodeCodeSpan] = ret.renderCodeSpan
	ret.RendererFuncs[ast.NodeCodeSpanOpenMarker] = ret.renderCodeSpanOpenMarker
	ret.RendererFuncs[ast.NodeCodeSpanContent] = ret.renderCodeSpanContent
	ret.RendererFuncs[ast.NodeCodeSpanCloseMarker] = ret.renderCodeSpanCloseMarker
	ret.RendererFuncs[ast.NodeCodeBlock] = ret.renderCodeBlock
	ret.RendererFuncs[ast.NodeCodeBlockFenceOpenMarker] = ret.renderCodeBlockOpenMarker
	ret.RendererFuncs[ast.NodeCodeBlockFenceInfoMarker] = ret.renderCodeBlockInfoMarker
	ret.RendererFuncs[ast.NodeCodeBlockCode] = ret.renderCodeBlockCode
	ret.RendererFuncs[ast.NodeCodeBlockFenceCloseMarker] = ret.renderCodeBlockCloseMarker
	ret.RendererFuncs[ast.NodeMathBlock] = ret.renderMathBlock
	ret.RendererFuncs[ast.NodeMathBlockOpenMarker] = ret.renderMathBlockOpenMarker
	ret.RendererFuncs[ast.NodeMathBlockContent] = ret.renderMathBlockContent
	ret.RendererFuncs[ast.NodeMathBlockCloseMarker] = ret.renderMathBlockCloseMarker
	ret.RendererFuncs[ast.NodeInlineMath] = ret.renderInlineMath
	ret.RendererFuncs[ast.NodeInlineMathOpenMarker] = ret.renderInlineMathOpenMarker
	ret.RendererFuncs[ast.NodeInlineMathContent] = ret.renderInlineMathContent
	ret.RendererFuncs[ast.NodeInlineMathCloseMarker] = ret.renderInlineMathCloseMarker
	ret.RendererFuncs[ast.NodeEmphasis] = ret.renderEmphasis
	ret.RendererFuncs[ast.NodeEmA6kOpenMarker] = ret.renderEmAsteriskOpenMarker
	ret.RendererFuncs[ast.NodeEmA6kCloseMarker] = ret.renderEmAsteriskCloseMarker
	ret.RendererFuncs[ast.NodeEmU8eOpenMarker] = ret.renderEmUnderscoreOpenMarker
	ret.RendererFuncs[ast.NodeEmU8eCloseMarker] = ret.renderEmUnderscoreCloseMarker
	ret.RendererFuncs[ast.NodeStrong] = ret.renderStrong
	ret.RendererFuncs[ast.NodeStrongA6kOpenMarker] = ret.renderStrongA6kOpenMarker
	ret.RendererFuncs[ast.NodeStrongA6kCloseMarker] = ret.renderStrongA6kCloseMarker
	ret.RendererFuncs[ast.NodeStrongU8eOpenMarker] = ret.renderStrongU8eOpenMarker
	ret.RendererFuncs[ast.NodeStrongU8eCloseMarker] = ret.renderStrongU8eCloseMarker
	ret.RendererFuncs[ast.NodeBlockquote] = ret.renderBlockquote
	ret.RendererFuncs[ast.NodeBlockquoteMarker] = ret.renderBlockquoteMarker
	ret.RendererFuncs[ast.NodeHeading] = ret.renderHeading
	ret.RendererFuncs[ast.NodeHeadingC8hMarker] = ret.renderHeadingC8hMarker
	ret.RendererFuncs[ast.NodeList] = ret.renderList
	ret.RendererFuncs[ast.NodeListItem] = ret.renderListItem
	ret.RendererFuncs[ast.NodeThematicBreak] = ret.renderThematicBreak
	ret.RendererFuncs[ast.NodeHardBreak] = ret.renderHardBreak
	ret.RendererFuncs[ast.NodeSoftBreak] = ret.renderSoftBreak
	ret.RendererFuncs[ast.NodeHTMLBlock] = ret.renderHTML
	ret.RendererFuncs[ast.NodeInlineHTML] = ret.renderInlineHTML
	ret.RendererFuncs[ast.NodeLink] = ret.renderLink
	ret.RendererFuncs[ast.NodeImage] = ret.renderImage
	ret.RendererFuncs[ast.NodeBang] = ret.renderBang
	ret.RendererFuncs[ast.NodeOpenBracket] = ret.renderOpenBracket
	ret.RendererFuncs[ast.NodeCloseBracket] = ret.renderCloseBracket
	ret.RendererFuncs[ast.NodeOpenParen] = ret.renderOpenParen
	ret.RendererFuncs[ast.NodeCloseParen] = ret.renderCloseParen
	ret.RendererFuncs[ast.NodeLinkText] = ret.renderLinkText
	ret.RendererFuncs[ast.NodeLinkSpace] = ret.renderLinkSpace
	ret.RendererFuncs[ast.NodeLinkDest] = ret.renderLinkDest
	ret.RendererFuncs[ast.NodeLinkTitle] = ret.renderLinkTitle
	ret.RendererFuncs[ast.NodeStrikethrough] = ret.renderStrikethrough
	ret.RendererFuncs[ast.NodeStrikethrough1OpenMarker] = ret.renderStrikethrough1OpenMarker
	ret.RendererFuncs[ast.NodeStrikethrough1CloseMarker] = ret.renderStrikethrough1CloseMarker
	ret.RendererFuncs[ast.NodeStrikethrough2OpenMarker] = ret.renderStrikethrough2OpenMarker
	ret.RendererFuncs[ast.NodeStrikethrough2CloseMarker] = ret.renderStrikethrough2CloseMarker
	ret.RendererFuncs[ast.NodeTaskListItemMarker] = ret.renderTaskListItemMarker
	ret.RendererFuncs[ast.NodeTable] = ret.renderTable
	ret.RendererFuncs[ast.NodeTableHead] = ret.renderTableHead
	ret.RendererFuncs[ast.NodeTableRow] = ret.renderTableRow
	ret.RendererFuncs[ast.NodeTableCell] = ret.renderTableCell
	ret.RendererFuncs[ast.NodeEmoji] = ret.renderEmoji
	ret.RendererFuncs[ast.NodeEmojiUnicode] = ret.renderEmojiUnicode
	ret.RendererFuncs[ast.NodeEmojiImg] = ret.renderEmojiImg
	ret.RendererFuncs[ast.NodeEmojiAlias] = ret.renderEmojiAlias
	ret.RendererFuncs[ast.NodeFootnotesDef] = ret.renderFootnotesDef
	ret.RendererFuncs[ast.NodeFootnotesRef] = ret.renderFootnotesRef
	ret.RendererFuncs[ast.NodeToC] = ret.renderToC
	ret.RendererFuncs[ast.NodeBackslash] = ret.renderBackslash
	ret.RendererFuncs[ast.NodeBackslashContent] = ret.renderBackslashContent
	ret.RendererFuncs[ast.NodeHTMLEntity] = ret.renderHtmlEntity
	return ret
}

func (r *VditorSVRenderer) Render() (output []byte) {
	output = r.BaseRenderer.Render()
	if 1 > len(r.Tree.Context.LinkRefDefs) || r.needRenderFootnotesDef {
		return
	}

	// 将链接引用定义添加到末尾
	r.WriteString("<div data-block=\"0\" data-type=\"link-ref-defs-block\">")
	for _, node := range r.Tree.Context.LinkRefDefs {
		label := node.LinkRefLabel
		dest := node.ChildByType(ast.NodeLinkDest).Tokens
		destStr := string(dest)
		r.WriteString("[" + string(label) + "]:")
		if util.Caret != destStr {
			r.WriteString(" ")
		}
		r.WriteString(destStr + "\n")
	}
	r.WriteString("</div>")
	output = r.Writer.Bytes()
	return
}

func (r *VditorSVRenderer) RenderFootnotesDefs(context *parse.Context) []byte {
	r.WriteString("<div data-block=\"0\" data-type=\"footnotes-block\">")
	for _, def := range context.FootnotesDefs {
		r.WriteString("<div data-type=\"footnotes-def\">")
		tree := &parse.Tree{Name: "", Context: context}
		tree.Context.Tree = tree
		tree.Root = &ast.Node{Type: ast.NodeDocument}
		tree.Root.AppendChild(def)
		defRenderer := NewVditorIRRenderer(tree)
		def.FirstChild.PrependChild(&ast.Node{Type: ast.NodeText, Tokens: []byte("[" + string(def.Tokens) + "]: ")})
		defRenderer.needRenderFootnotesDef = true
		defContent := defRenderer.Render()
		r.Write(defContent)
		r.WriteString("</div>")
	}
	r.WriteString("</div>")
	return r.Writer.Bytes()
}

func (r *VditorSVRenderer) renderHtmlEntity(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderSpanNode(node)
		r.tag("code", [][]string{{"data-newline", "1"}, {"class", "vditor-sv__marker--pre"}, {"data-type", "html-entity"}}, false)
		r.Write(html.EscapeHTML(html.EscapeHTML(node.Tokens)))
		r.tag("/code", nil, false)
		r.tag("span", [][]string{{"class", "vditor-sv__preview"}, {"data-render", "2"}}, false)
		r.tag("code", nil, false)
		r.Write(html.UnescapeHTML(node.HtmlEntityTokens))
		r.tag("/code", nil, false)
		r.tag("/span", nil, false)
	} else {
		r.tag("/span", nil, false)
	}
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderBackslashContent(node *ast.Node, entering bool) ast.WalkStatus {
	r.Write(html.EscapeHTML(node.Tokens))
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderBackslash(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.WriteString("<span data-type=\"backslash\">")
		r.WriteString("<span>")
		r.WriteByte(lex.ItemBackslash)
		r.WriteString("</span>")
	} else {
		r.WriteString("</span>")
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderToC(node *ast.Node, entering bool) ast.WalkStatus {
	headings := r.headings()
	length := len(headings)
	r.WriteString("<div class=\"vditor-toc\" data-block=\"0\" data-type=\"toc-block\" contenteditable=\"false\">")
	if 0 < length {
		for _, heading := range headings {
			spaces := (heading.HeadingLevel - 1) * 2
			r.WriteString(strings.Repeat("&emsp;", spaces))
			r.WriteString("<span data-type=\"toc-h\">")
			r.WriteString(heading.Text() + "</span><br>")
		}
	} else {
		r.WriteString("[toc]<br>")
	}
	r.WriteString("</div>")
	caretInDest := bytes.Contains(node.Tokens, []byte(util.Caret))
	r.WriteString("<p data-block=\"0\">")
	if caretInDest {
		r.WriteString(util.Caret)
	}
	r.WriteString("</p>")
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderFootnotesDef(node *ast.Node, entering bool) ast.WalkStatus {
	if !r.needRenderFootnotesDef {
		return ast.WalkStop
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderFootnotesRef(node *ast.Node, entering bool) ast.WalkStatus {
	previousNodeText := node.PreviousNodeText()
	previousNodeText = strings.ReplaceAll(previousNodeText, util.Caret, "")
	if "" == previousNodeText {
		r.WriteString(parse.Zwsp)
	}
	idx, def := r.Tree.Context.FindFootnotesDef(node.Tokens)
	idxStr := strconv.Itoa(idx)
	label := def.Text()
	attrs := [][]string{{"data-type", "footnotes-ref"}}
	attrs = append(attrs, []string{"class", "vditor-tooltipped vditor-tooltipped__s"})
	attrs = append(attrs, []string{"aria-label", html.EscapeString(label)})
	attrs = append(attrs, []string{"data-footnotes-label", string(node.FootnotesRefLabel)})
	r.tag("sup", attrs, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bracket"}}, false)
	r.WriteByte(lex.ItemOpenBracket)
	r.tag("/span", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker--link"}}, false)
	r.Write(node.Tokens)
	r.tag("/span", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker--hide"}, {"data-render", "1"}}, false)
	r.WriteString(idxStr)
	r.tag("/span", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bracket"}}, false)
	r.WriteByte(lex.ItemCloseBracket)
	r.tag("/span", nil, false)
	r.WriteString("</sup>" + parse.Zwsp)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderCodeBlockCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	r.tag("span", [][]string{{"data-type", "code-block-close-marker"}, {"class", "vditor-sv__marker"}}, false)
	r.Write(node.Tokens)
	r.tag("/span", nil, false)
	r.Newline()
	if !r.isLastNode(r.Tree.Root, node) {
		r.Write(newline)
	}
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderCodeBlockInfoMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--info"}, {"data-type", "code-block-info"}}, false)
	r.Write(node.CodeBlockInfo)
	r.tag("/span", nil, false)
	r.Newline()
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderCodeBlockOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"data-type", "code-block-open-marker"}, {"class", "vditor-sv__marker"}}, false)
	r.Write(node.Tokens)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderCodeBlock(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderDivNode(node)
	} else {
		r.WriteString("</div>")
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderCodeBlockCode(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("pre", [][]string{{"class", "vditor-sv__marker--pre"}}, false)
	r.tag("code", nil, false)
	r.Write(html.EscapeHTML(bytes.TrimSpace(node.Tokens)))
	r.WriteString("</code></pre>")
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderEmojiAlias(node *ast.Node, entering bool) ast.WalkStatus {
	r.Write(node.Tokens)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderEmojiImg(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderEmojiUnicode(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderEmoji(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderInlineMathCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteByte(lex.ItemDollar)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderInlineMathContent(node *ast.Node, entering bool) ast.WalkStatus {
	tokens := html.EscapeHTML(node.Tokens)
	r.Write(tokens)
	r.tag("/code", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderInlineMathOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteByte(lex.ItemDollar)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderInlineMath(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderSpanNode(node)
	} else {
		r.tag("/span", nil, false)
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderMathBlockCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"data-type", "math-block-close-marker"}, {"class", "vditor-sv__marker"}}, false)
	r.WriteString("$$")
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderMathBlockContent(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("pre", [][]string{{"class", "vditor-sv__marker--pre"}}, false)
	r.tag("code", [][]string{{"data-type", "math-block"}, {"class", "language-math"}}, false)
	r.Write(html.EscapeHTML(bytes.TrimSpace(node.Tokens)))
	r.WriteString("</code></pre>")
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderMathBlockOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"data-type", "math-block-open-marker"}, {"class", "vditor-sv__marker"}}, false)
	r.WriteString("$$")
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderMathBlock(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderDivNode(node)
	} else {
		r.Newline()
		if !entering && !r.isLastNode(r.Tree.Root, node) {
			r.Write(newline)
		}
		r.WriteString("</div>")
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderTableCell(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderTableRow(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderTableHead(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderTable(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.tag("div", [][]string{{"data-block", "0"}, {"data-type", "table"}}, false)
		r.Write(node.Tokens)
	} else {
		r.Newline()
		if !r.isLastNode(r.Tree.Root, node) {
			r.Write(newline)
		}
		r.tag("/div", nil, false)
	}
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderStrikethrough(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderSpanNode(node)
	} else {
		r.tag("/span", nil, false)
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderStrikethrough1OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteString("~")
	r.tag("/span", nil, false)
	r.tag("s", [][]string{{"data-newline", "1"}}, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderStrikethrough1CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("/s", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteString("~")
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderStrikethrough2OpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteString("~~")
	r.tag("/span", nil, false)
	r.tag("s", [][]string{{"data-newline", "1"}}, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderStrikethrough2CloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("/s", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteString("~~")
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderLinkTitle(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--title"}}, false)
	r.WriteByte(lex.ItemDoublequote)
	r.Write(node.Tokens)
	r.WriteByte(lex.ItemDoublequote)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderLinkDest(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--link"}}, false)
	r.Write(node.Tokens)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderLinkSpace(node *ast.Node, entering bool) ast.WalkStatus {
	r.WriteByte(lex.ItemSpace)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderLinkText(node *ast.Node, entering bool) ast.WalkStatus {
	if ast.NodeImage == node.Parent.Type {
		r.tag("span", [][]string{{"class", "vditor-sv__marker--bracket"}}, false)
	} else {
		if 3 == node.Parent.LinkType {
			r.tag("span", nil, false)
		} else {
			r.tag("span", [][]string{{"class", "vditor-sv__marker--bracket"}}, false)
		}
	}
	r.Write(node.Tokens)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderCloseParen(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--paren"}}, false)
	r.WriteByte(lex.ItemCloseParen)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderOpenParen(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--paren"}}, false)
	r.WriteByte(lex.ItemOpenParen)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderCloseBracket(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bracket"}}, false)
	r.WriteByte(lex.ItemCloseBracket)
	r.tag("/span", nil, false)

	if 3 == node.Parent.LinkType {
		linkText := node.Parent.ChildByType(ast.NodeLinkText)
		if !bytes.EqualFold(node.Parent.LinkRefLabel, linkText.Tokens) {
			r.tag("span", [][]string{{"class", "vditor-sv__marker--link"}}, false)
			r.WriteByte(lex.ItemOpenBracket)
			r.Write(node.Parent.LinkRefLabel)
			r.WriteByte(lex.ItemCloseBracket)
			r.tag("/span", nil, false)
		}
	}
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderOpenBracket(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bracket"}}, false)
	r.WriteByte(lex.ItemOpenBracket)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderBang(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteByte(lex.ItemBang)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderImage(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.tag("span", nil, false)
	} else {
		r.tag("/span", nil, false)
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderLink(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if 3 == node.LinkType {
			node.ChildByType(ast.NodeOpenParen).Unlink()
			node.ChildByType(ast.NodeLinkDest).Unlink()
			if linkSpace := node.ChildByType(ast.NodeLinkSpace); nil != linkSpace {
				linkSpace.Unlink()
				node.ChildByType(ast.NodeLinkTitle).Unlink()
			}
			node.ChildByType(ast.NodeCloseParen).Unlink()
		}

		r.renderSpanNode(node)
	} else {
		r.tag("/span", nil, false)
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderHTML(node *ast.Node, entering bool) ast.WalkStatus {
	r.renderDivNode(node)
	tokens := bytes.TrimSpace(node.Tokens)
	r.WriteString("<pre class=\"vditor-sv__marker--pre vditor-sv__marker\">")
	r.tag("code", [][]string{{"data-type", "html-block"}}, false)
	r.Write(html.EscapeHTML(tokens))
	r.WriteString("</code></pre>")
	r.Newline()
	if !r.isLastNode(r.Tree.Root, node) {
		r.Write(newline)
	}
	r.WriteString("</div>")
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderInlineHTML(node *ast.Node, entering bool) ast.WalkStatus {
	r.renderSpanNode(node)
	r.tag("code", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.Write(html.EscapeHTML(node.Tokens))
	r.tag("/code", nil, false)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderDocument(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Writer = &bytes.Buffer{}
		r.nodeWriterStack = append(r.nodeWriterStack, r.Writer)
	} else {
		r.nodeWriterStack = r.nodeWriterStack[:len(r.nodeWriterStack)-1]
		buf := bytes.Trim(r.Writer.Bytes(), " \t\n")
		r.Writer.Reset()
		r.Write(buf)
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderParagraph(node *ast.Node, entering bool) ast.WalkStatus {
	rootParent := ast.NodeDocument == node.Parent.Type
	if entering {
		if rootParent {
			r.tag("div", [][]string{{"data-type", "p"}, {"data-block", "0"}}, false)
		}
	} else  {
		if !node.ParentIs(ast.NodeTableCell) {
			r.Newline()
		}
		inTightList := false
		lastListItemLastPara := false
		if parent := node.Parent; nil != parent {
			if ast.NodeListItem == parent.Type { // ListItem.Paragraph
				listItem := parent
				if nil != listItem.Parent && nil != listItem.Parent.ListData {
					// 必须通过列表（而非列表项）上的紧凑标识判断，因为在设置该标识时仅设置了 List.Tight
					// 设置紧凑标识的具体实现可参考函数 List.Finalize()
					inTightList = listItem.Parent.Tight

					if nextItem := listItem.Next; nil == nextItem {
						nextPara := node.Next
						lastListItemLastPara = nil == nextPara
					}
				} else {
					inTightList = true
				}
			}
		}

		if (!inTightList || (lastListItemLastPara)) && !node.ParentIs(ast.NodeTableCell) {
			r.Write(newline)
		}
		if rootParent {
			r.tag("/div", nil, false)
		}
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) listItemParent(node *ast.Node) (inListItem, looseListItemFirstChild bool) {
	grandparent := node.Parent.Parent
	if nil != grandparent && ast.NodeList == grandparent.Type { // List.ListItem.Paragraph
		inListItem = true
		if !grandparent.Tight {
			looseListItemFirstChild = node.Parent.FirstChild == node
		}
	}
	return
}

func (r *VditorSVRenderer) renderText(node *ast.Node, entering bool) ast.WalkStatus {
	if r.Option.AutoSpace {
		r.Space(node)
	}
	if r.Option.FixTermTypo {
		r.FixTermTypo(node)
	}
	if r.Option.ChinesePunct {
		r.ChinesePunct(node)
	}

	node.Tokens = bytes.TrimRight(node.Tokens, "\n")
	r.Write(html.EscapeHTML(node.Tokens))
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderCodeSpan(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderSpanNode(node)
	} else {
		r.tag("/span", nil, false)
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderCodeSpanOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteString(strings.Repeat("`", node.Parent.CodeMarkerLen))
	r.tag("/span", nil, false)
	r.tag("span", [][]string{{"data-newline", "1"}}, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderCodeSpanContent(node *ast.Node, entering bool) ast.WalkStatus {
	r.Write(html.EscapeHTML(node.Tokens))
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderCodeSpanCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("/span", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteString(strings.Repeat("`", node.Parent.CodeMarkerLen))
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderEmphasis(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderSpanNode(node)
	} else {
		r.tag("/span", nil, false)
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderEmAsteriskOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bi"}}, false)
	r.WriteByte(lex.ItemAsterisk)
	r.tag("/span", nil, false)
	r.tag("em", [][]string{{"data-newline", "1"}}, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderEmAsteriskCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("/em", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bi"}}, false)
	r.WriteByte(lex.ItemAsterisk)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderEmUnderscoreOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bi"}}, false)
	r.WriteByte(lex.ItemUnderscore)
	r.tag("/span", nil, false)
	r.tag("em", [][]string{{"data-newline", "1"}}, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderEmUnderscoreCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("/em", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bi"}}, false)
	r.WriteByte(lex.ItemUnderscore)
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderStrong(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.renderSpanNode(node)
	} else {
		r.tag("/span", nil, false)
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderStrongA6kOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bi"}}, false)
	r.WriteString("**")
	r.tag("/span", nil, false)
	r.tag("strong", [][]string{{"data-newline", "1"}}, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderStrongA6kCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("/strong", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bi"}}, false)
	r.WriteString("**")
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderStrongU8eOpenMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bi"}}, false)
	r.WriteString("__")
	r.tag("/span", nil, false)
	r.tag("strong", [][]string{{"data-newline", "1"}}, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderStrongU8eCloseMarker(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("/strong", nil, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker--bi"}}, false)
	r.WriteString("__")
	r.tag("/span", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderBlockquote(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Writer = &bytes.Buffer{}
		r.nodeWriterStack = append(r.nodeWriterStack, r.Writer)
	} else {
		writer := r.nodeWriterStack[len(r.nodeWriterStack)-1]
		r.nodeWriterStack = r.nodeWriterStack[:len(r.nodeWriterStack)-1]

		blockquoteLines := bytes.Buffer{}
		buf := writer.Bytes()
		lines := bytes.Split(buf, newline)
		length := len(lines)
		if 1 == len(r.nodeWriterStack) { // 已经是根这一层
			length = len(lines)
			if 1 < length && lex.IsBlank(lines[length-1]) {
				lines = lines[:length-1]
			}
		}

		inListItem, looseListItemFirstP := r.listItemParent(node)
		if inListItem && !looseListItemFirstP {
			blockquoteLines.Write(newline)
		}
		for _, line := range lines {
			if 0 == len(line) {
				continue
			}
			blockquoteLines.WriteString(`<span class="vditor-sv__marker">&gt; </span>`)
			blockquoteLines.Write(line)
			blockquoteLines.Write(newline)
		}
		buf = blockquoteLines.Bytes()
		writer.Reset()
		bq := node.ParentIs(ast.NodeBlockquote)
		if !bq && !inListItem {
			writer.WriteString(`<div data-block="0" data-type="blockquote">`)
		} else {
			if inListItem {
				writer.WriteString(`<span data-type="blockquote">`)
			} else {
				writer.Write(newline)
			}
		}

		writer.Write(buf)
		r.nodeWriterStack[len(r.nodeWriterStack)-1].Write(writer.Bytes())
		r.Writer = r.nodeWriterStack[len(r.nodeWriterStack)-1]
		buf = r.Writer.Bytes()
		r.Writer.Reset()
		r.Write(buf)
		r.Newline()
		if !bq && !inListItem {
			r.WriteString("</div>")
		} else {
			if inListItem {
				r.WriteString("</span>")
			}
		}
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderBlockquoteMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderHeading(node *ast.Node, entering bool) ast.WalkStatus {
	rootParent := ast.NodeDocument == node.Parent.Type
	if entering {
		if rootParent {
			r.tag("div", [][]string{{"data-block", "0"}, {"data-type", "heading"}}, false)
			r.tag("span", [][]string{{"class", "h" + headingLevel[node.HeadingLevel:node.HeadingLevel+1]}}, false)
		}
		r.tag("span", [][]string{{"class", "vditor-sv__marker--heading"}, {"data-type", "heading-marker"}}, false)
		r.WriteString(strings.Repeat("#", node.HeadingLevel) + " ")
		r.tag("/span", nil, false)
	} else {
		if rootParent {
			r.tag("/span", nil, false)
		}
		r.Newline()
		if rootParent {
			r.Write(newline)
			r.WriteString("</div>")
		}
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderHeadingC8hMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderList(node *ast.Node, entering bool) ast.WalkStatus {
	blockContainerParent := node.ParentIs(ast.NodeListItem, ast.NodeBlockquote)
	if blockContainerParent {
		return ast.WalkContinue
	}

	if entering {
		var attrs [][]string
		if node.Tight {
			attrs = append(attrs, []string{"data-tight", "true"})
		}
		if 0 == node.BulletChar {
			if 1 != node.Start {
				attrs = append(attrs, []string{"start", strconv.Itoa(node.Start)})
			}
		}
		switch node.ListData.Typ {
		case 0:
			attrs = append(attrs, []string{"data-type", "ul"})
			attrs = append(attrs, []string{"data-marker", string(node.Marker)})
		case 1:
			attrs = append(attrs, []string{"data-type", "ol"})
			attrs = append(attrs, []string{"data-marker", strconv.Itoa(node.Num) + string(node.ListData.Delimiter)})
		case 3:
			attrs = append(attrs, []string{"data-type", "task"})
			if 0 == node.ListData.BulletChar {
				attrs = append(attrs, []string{"data-marker", strconv.Itoa(node.Num) + string(node.ListData.Delimiter)})
			} else {
				attrs = append(attrs, []string{"data-marker", string(node.Marker)})
			}
		}
		attrs = append(attrs, []string{"data-block", "0"})
		r.tag("div", attrs, false)
	} else {
		r.Newline()
		r.tag("/div", nil, false)
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderListItem(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		r.Writer = &bytes.Buffer{}
		r.nodeWriterStack = append(r.nodeWriterStack, r.Writer)
	} else {
		writer := r.nodeWriterStack[len(r.nodeWriterStack)-1]
		r.nodeWriterStack = r.nodeWriterStack[:len(r.nodeWriterStack)-1]
		indent := len(node.Marker) + 1
		if 1 == node.ListData.Typ || (3 == node.ListData.Typ && 0 == node.ListData.BulletChar) {
			indent++
		}
		indentSpaces := bytes.Repeat([]byte{lex.ItemSpace}, indent)
		indentedLines := bytes.Buffer{}
		buf := writer.Bytes()
		lines := bytes.Split(buf, newline)
		for _, line := range lines {
			if 0 == len(line) {
				indentedLines.Write(newline)
				continue
			}
			indentedLines.Write(indentSpaces)
			indentedLines.Write(line)
			indentedLines.Write(newline)
		}
		buf = indentedLines.Bytes()
		if indent < len(buf) {
			buf = buf[indent:]
		}

		listItemBuf := bytes.Buffer{}
		var marker string
		if 1 == node.ListData.Typ || (3 == node.ListData.Typ && 0 == node.ListData.BulletChar) {
			marker = strconv.Itoa(node.Num) + string(node.ListData.Delimiter)
		} else {
			marker = string(node.Marker)
		}
		listItemBuf.WriteString(`<span data-type="li" data-marker="` + marker + `" class="vditor-sv__marker--bi">` + marker + " </span>")
		buf = append(listItemBuf.Bytes(), buf...)
		if node.ParentIs(ast.NodeTableCell) {
			buf = bytes.ReplaceAll(buf, newline, nil)
		}
		writer.Reset()
		writer.Write(buf)
		buf = writer.Bytes()
		if node.ParentIs(ast.NodeTableCell) {
			buf = bytes.ReplaceAll(buf, newline, nil)
		}
		r.nodeWriterStack[len(r.nodeWriterStack)-1].Write(buf)
		r.Writer = r.nodeWriterStack[len(r.nodeWriterStack)-1]
		buf = bytes.TrimSuffix(r.Writer.Bytes(), newline)
		r.Writer.Reset()
		r.Write(buf)
		if !node.ParentIs(ast.NodeTableCell) {
			r.Newline()
		}
	}
	return ast.WalkContinue
}

func (r *VditorSVRenderer) renderTaskListItemMarker(node *ast.Node, entering bool) ast.WalkStatus {
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderThematicBreak(node *ast.Node, entering bool) ast.WalkStatus {
	r.tag("div", [][]string{{"data-type", "thematic-break"}, {"class", "vditor-sv__marker"}}, false)
	r.tag("span", [][]string{{"class", "vditor-sv__marker"}}, false)
	r.WriteString("---\n\n")
	r.tag("/span", nil, false)
	r.tag("/div", nil, false)
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderHardBreak(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	return ast.WalkStop
}

func (r *VditorSVRenderer) renderSoftBreak(node *ast.Node, entering bool) ast.WalkStatus {
	r.Newline()
	return ast.WalkStop
}

func (r *VditorSVRenderer) tag(name string, attrs [][]string, selfclosing bool) {
	if r.DisableTags > 0 {
		return
	}

	r.WriteString("<")
	r.WriteString(name)
	if 0 < len(attrs) {
		for _, attr := range attrs {
			r.WriteString(" " + attr[0] + "=\"" + attr[1] + "\"")
		}
	}
	if selfclosing {
		r.WriteString(" /")
	}
	r.WriteString(">")
}

func (r *VditorSVRenderer) renderSpanNode(node *ast.Node) {
	var attrs [][]string
	switch node.Type {
	case ast.NodeEmphasis:
		attrs = append(attrs, []string{"data-type", "em"})
	case ast.NodeStrong:
		attrs = append(attrs, []string{"data-type", "strong"})
	case ast.NodeStrikethrough:
		attrs = append(attrs, []string{"data-type", "s"})
	case ast.NodeLink:
		if 3 != node.LinkType {
			attrs = append(attrs, []string{"data-type", "a"})
		} else {
			attrs = append(attrs, []string{"data-type", "link-ref"})
		}
	case ast.NodeImage:
		attrs = append(attrs, []string{"data-type", "img"})
	case ast.NodeCodeSpan:
		attrs = append(attrs, []string{"data-type", "code"})
	default:
		attrs = append(attrs, []string{"data-type", "inline-node"})
	}
	r.tag("span", attrs, false)
	return
}

func (r *VditorSVRenderer) renderDivNode(node *ast.Node) {
	attrs := [][]string{{"data-block", "0"}}
	switch node.Type {
	case ast.NodeCodeBlock:
		attrs = append(attrs, []string{"data-type", "code-block"})
	case ast.NodeHTMLBlock:
		attrs = append(attrs, []string{"data-type", "html-block"})
	case ast.NodeMathBlock:
		attrs = append(attrs, []string{"data-type", "math-block"})
	}
	r.tag("div", attrs, false)
	return
}

func (r *VditorSVRenderer) Text(node *ast.Node) (ret string) {
	ast.Walk(node, func(n *ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch n.Type {
			case ast.NodeText, ast.NodeLinkText, ast.NodeLinkDest, ast.NodeLinkTitle, ast.NodeCodeBlockCode, ast.NodeCodeSpanContent, ast.NodeInlineMathContent, ast.NodeMathBlockContent, ast.NodeHTMLBlock, ast.NodeInlineHTML:
				ret += string(n.Tokens)
			case ast.NodeCodeBlockFenceInfoMarker:
				ret += string(n.CodeBlockInfo)
			case ast.NodeLink:
				if 3 == n.LinkType {
					ret += string(n.LinkRefLabel)
				}
			}
		}
		return ast.WalkContinue
	})
	return
}
