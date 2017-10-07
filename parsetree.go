// The UXToolkit Project
// Copyright (c) Wirecog, LLC. All rights reserved.
// Use of this source code is governed by a BSD-style
// license, which can be found in the LICENSE file.

package reconcile

import (
	"bytes"
	"errors"
	"io"

	"golang.org/x/net/html"
)

// Declare considered empty HTML tags
// Reference: https://developer.mozilla.org/en-US/docs/Glossary/Empty_element
var emptyTags = map[string]int{"input": 1, "area": 1, "base": 1, "br": 1, "col": 1, "embed": 1, "hr": 1, "img": 1, "keygen": 1, "link": 1, "meta": 1, "param": 1, "source": 1, "track": 1, "wbr": 1}

func IsEmptyElement(tag string) bool {
	if _, ok := emptyTags[tag]; ok == true {
		return true
	} else {
		return false
	}
}

type ParseTree struct {
	reader     io.Reader
	src        []byte
	ChildNodes []*DOMNode
}

func NewParseTree(src []byte) (*ParseTree, error) {
	r := bytes.NewReader(src)
	z := html.NewTokenizer(r)
	tree := &ParseTree{src: src, reader: r}
	var node *DOMNode = nil
	for tokenType := z.Next(); tokenType != html.ErrorToken; tokenType = z.Next() {
		token := z.Token()
		if next, err := tree.parse(token, node, z); err != nil {
			return nil, err
		} else {
			node = next
		}
	}
	return tree, nil
}

func (pt *ParseTree) parse(token html.Token, current *DOMNode, tokenizer *html.Tokenizer) (next *DOMNode, err error) {

	emptyElement := IsEmptyElement(token.Data)
	offset := tokenizer.CurrentOffset()
	var result *DOMNode
	if token.Type == html.StartTagToken || token.Type == html.SelfClosingTagToken {

		node := &DOMNode{}
		node.NodeType = ElementNodeType
		node.Name = token.Data
		node.tree = pt

		for _, attribute := range token.Attr {
			node.Attributes = append(node.Attributes, Attribute{
				Name:  attribute.Key,
				Value: attribute.Val,
			})
		}

		if current != nil {
			node.Position = make([]int, len(current.Position)+1)
			copy(node.Position, current.Position)
			node.Position[len(current.Position)] = len(current.ChildNodes)
			node.ParentNode = current
			current.ChildNodes = append(current.ChildNodes, node)
		} else {
			node.Position = []int{len(pt.ChildNodes)}
		}

		node.startPosition, err = pt.ReverseFind(0, offset[0]-1, '<')

		if err != nil {
			return nil, err
		}

		node.innerStartPosition = offset[0]
		next = node
		result = node

		if token.Type == html.SelfClosingTagToken || emptyElement == true {
			current = node
		}
	}

	if token.Type == html.EndTagToken || token.Type == html.SelfClosingTagToken || emptyElement == true {

		offset := tokenizer.CurrentOffset()

		if token.Type == html.SelfClosingTagToken || emptyElement == true {
			current.IsSelfClosing = true

			if current.ParentNode != nil {
				next = current.ParentNode
			} else {
				next = nil
			}

		} else {
			current.endPosition = offset[0]
			closingTagLength := len(current.Name) + 3
			current.innerEndPosition = offset[0] - closingTagLength

			if current.ParentNode != nil {
				next = current.ParentNode
			} else {
				next = nil
			}

		}
	}

	if token.Type == html.TextToken {

		charData := token.Data
		text := &DOMNode{}
		text.NodeType = TextNodeType
		text.Value = []byte(charData)
		if current != nil {
			text.Position = make([]int, len(current.Position)+1)
			copy(text.Position, current.Position)
			text.Position[len(current.Position)] = len(current.ChildNodes)
			text.ParentNode = current
			current.ChildNodes = append(current.ChildNodes, text)
		} else {
			text.Position = []int{len(pt.ChildNodes)}
		}
		result = text
		next = current
	} else if token.Type == html.CommentToken {

		htmlComment := token.Data
		comment := &DOMNode{}
		comment.NodeType = CommentNodeType
		comment.Value = []byte(htmlComment)
		if current != nil {

			comment.Position = make([]int, len(current.Position)+1)
			copy(comment.Position, current.Position)
			comment.Position[len(current.Position)] = len(current.ChildNodes)
			comment.ParentNode = current
			current.ChildNodes = append(current.ChildNodes, comment)
		} else {

			comment.Position = []int{len(pt.ChildNodes)}
		}
		result = comment
		next = current
	}

	if result != nil && current == nil {

		pt.ChildNodes = append(pt.ChildNodes, result)
	}
	return next, nil
}

func (t *ParseTree) Compare(other *ParseTree) (Changes, error) {

	changes := []Reconciler{}
	if err := compareNodes(&changes, t.ChildNodes, other.ChildNodes); err != nil {
		return nil, err
	}

	return changes, nil
}

func compareAttributes(changes *[]Reconciler, a, b *DOMNode) {

	bAttributes := b.AttributesMap()
	aAttributes := a.AttributesMap()

	for attributeName := range aAttributes {
		if _, found := bAttributes[attributeName]; !found {

			*changes = append(*changes, Reconciler{ActionType: RemoveNodeAttributeAction, ExistingNode: a, AttributeName: attributeName})
		}
	}

	for name, bValue := range bAttributes {
		value, found := aAttributes[name]
		if !found {

			*changes = append(*changes, Reconciler{ActionType: SetNodeAttributeAction, ExistingNode: a, AttributeName: name, AttributeValue: bValue})

		} else if value != bValue {

			*changes = append(*changes, Reconciler{ActionType: SetNodeAttributeAction, ExistingNode: a, AttributeName: name, AttributeValue: bValue})

		}
	}
}

func compareNodes(changes *[]Reconciler, a, b []*DOMNode) error {

	bCount := len(b)
	aCount := len(a)
	minCount := bCount

	if bCount > aCount {
		for _, bNode := range b[aCount:] {

			*changes = append(*changes, Reconciler{ActionType: AppendChildNodeAction, ParentNode: bNode.ParentNode, ChildNode: bNode})
		}
		minCount = aCount
	} else if aCount > bCount {
		for _, node := range a[bCount:] {

			*changes = append(*changes, Reconciler{ActionType: RemoveNodeAction, ExistingNode: node})

		}
		minCount = bCount
	}
	for i := 0; i < minCount; i++ {
		bNode := b[i]
		n := a[i]
		if n.IsEqual(*bNode) == false {

			*changes = append(*changes, Reconciler{ActionType: ReplaceNodeAction, ExistingNode: n, NewNode: bNode})

			continue
		}
		if bNode.NodeType == ElementNodeType {
			node := n
			compareAttributes(changes, node, bNode)
			compareNodes(changes, n.ChildNodes, bNode.ChildNodes)
		}
	}
	return nil
}

func (pt *ParseTree) ReverseFind(m int, n int, b byte) (int, error) {
	if n >= len(pt.src) || n < m {
		return -1, errors.New("invalid value for n")
	}
	if m >= len(pt.src) || m < 0 {
		return -1, errors.New("invalid value for m")
	}
	for i := n; i >= m; i-- {
		if pt.src[i] == b {
			return i, nil
		}
	}
	return -1, nil
}
