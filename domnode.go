// The UXToolkit Project
// Copyright (c) Wirecog, LLC. All rights reserved.
// Use of this source code is governed by a BSD-style
// license, which can be found in the LICENSE file.

package reconcile

import (
	"html"
	"strconv"

	"honnef.co/go/js/dom"
)

const (
	ElementNodeType = iota
	CommentNodeType
	TextNodeType
)

type Positions struct {
	startPosition      int
	innerStartPosition int
	endPosition        int
	innerEndPosition   int
}

type Attribute struct {
	Name  string
	Value string
}

type DOMNode struct {
	Positions
	NodeType      int
	Position      []int
	ParentNode    *DOMNode
	ChildNodes    []*DOMNode
	Contents      []byte
	Name          string
	Value         []byte
	IsSelfClosing bool
	Attributes    []Attribute
	tree          *ParseTree
}

func NewDOMNode(nodeType int) *DOMNode {

	result := &DOMNode{}
	result.innerEndPosition = -1
	if nodeType == ElementNodeType {
		result.NodeType = ElementNodeType
	} else if nodeType == CommentNodeType {
		result.NodeType = CommentNodeType
	} else if nodeType == TextNodeType {
		result.NodeType = TextNodeType
	}

	return result
}

func (dn *DOMNode) Create() dom.Node {

	var result dom.Node = nil
	d := dom.GetWindow().Document()

	if dn.NodeType == CommentNodeType {

		element := d.Underlying().Call("createComment", string(dn.Value))
		result = dom.WrapNode(element)

	} else if dn.NodeType == TextNodeType {

		result = d.CreateTextNode(string(dn.Value))

	} else if dn.NodeType == ElementNodeType {

		element := d.CreateElement(dn.Name)
		for _, attribute := range dn.Attributes {
			element.SetAttribute(attribute.Name, attribute.Value)
		}

		element.SetInnerHTML(string(dn.GetHTMLContents(true)))
		result = element
	}

	return result
}

func (dn *DOMNode) Locate(rootElement dom.Element) dom.Node {

	result := rootElement.ChildNodes()[dn.Position[0]]
	for _, index := range dn.Position[1:] {
		result = result.ChildNodes()[index]
	}
	return result
}

func (dn *DOMNode) IsEqual(b DOMNode) bool {

	if dn.NodeType != b.NodeType {
		return false
	}

	if dn.NodeType == TextNodeType || dn.NodeType == CommentNodeType {
		if string(dn.Value) != string(b.Value) {
			return false
		} else {
			return true
		}
	}

	if dn.NodeType == ElementNodeType {

		if dn.Name != b.Name {
			return false
		}

		attrs := dn.Attributes
		otherAttrs := b.Attributes
		if len(attrs) != len(otherAttrs) {
			return false
		}
		for i, attr := range attrs {
			otherAttr := otherAttrs[i]
			if attr != otherAttr {
				return false
			}
		}
		return true
	}

	return false
}

func (dn *DOMNode) GetHTMLContents(innerContentsOnly bool) []byte {
	if dn.IsSelfClosing == true {
		return nil
	} else {
		contents := string(dn.tree.src[dn.innerStartPosition:dn.innerEndPosition])
		return []byte(html.UnescapeString(contents))
	}
}

func (dn *DOMNode) AttributesMap() map[string]string {

	var result map[string]string

	if dn.NodeType == ElementNodeType {
		result = make(map[string]string)

		for _, attribute := range dn.Attributes {
			result[attribute.Name] = attribute.Value
		}
	} else {
		result = nil
	}

	return result

}

func (dn *DOMNode) String() string {

	s := "### DOMNODE ###\n"
	s += "name: " + dn.Name + "\n"
	s += "innerStartPosition: " + strconv.Itoa(dn.innerStartPosition) + "\n"
	s += "innerEndPosition: " + strconv.Itoa(dn.innerEndPosition) + "\n"
	return s
}
