// The UXToolkit Project
// Copyright (c) Wirecog, LLC. All rights reserved.
// Use of this source code is governed by a BSD-style
// license, which can be found in the LICENSE file.

package reconcile

import (
	"errors"

	"honnef.co/go/js/dom"
)

const (
	SetNodeAttributeAction = iota
	RemoveNodeAttributeAction
	RemoveNodeAction
	ReplaceNodeAction
	AppendChildNodeAction
)

type ReconcileParams struct {
}

type Reconciler struct {
	ActionType     int
	ParentNode     *DOMNode
	ChildNode      *DOMNode
	ExistingNode   *DOMNode
	NewNode        *DOMNode
	AttributeName  string
	AttributeValue string
}

func (r *Reconciler) ApplyChange(rootElement dom.Element) error {

	switch r.ActionType {

	case SetNodeAttributeAction:
		r.SetNodeAttribute(rootElement)
	case RemoveNodeAttributeAction:
		r.RemoveNodeAttribute(rootElement)
	case RemoveNodeAction:
		r.RemoveNode(rootElement)
	case ReplaceNodeAction:
		r.ReplaceNode(rootElement)
	case AppendChildNodeAction:
		r.AppendChildNode(rootElement)
	default:
		return errors.New("Unknown action")
	}
	return nil
}

type Changes []Reconciler

func (r *Reconciler) AppendChildNode(rootElement dom.Element) {
	var parent dom.Node
	if r.ParentNode != nil {
		parent = r.ParentNode.Locate(rootElement)
	} else {
		parent = rootElement
	}
	child := r.ChildNode.Create()
	parent.AppendChild(child)
}

func (r *Reconciler) ReplaceNode(rootElement dom.Element) {
	var parent dom.Node
	if r.ExistingNode.ParentNode != nil {
		parent = r.ExistingNode.ParentNode.Locate(rootElement)
	} else {
		parent = rootElement
	}
	existingNode := r.ExistingNode.Locate(rootElement)
	newNode := r.NewNode.Create()
	parent.ReplaceChild(newNode, existingNode)
}

func (r *Reconciler) RemoveNode(rootElement dom.Element) {
	var parent dom.Node
	if r.ExistingNode.ParentNode != nil {
		parent = r.ExistingNode.ParentNode.Locate(rootElement)
	} else {
		parent = rootElement
	}
	self := r.ExistingNode.Locate(rootElement)
	parent.RemoveChild(self)
	if r.ExistingNode.ParentNode != nil {
		lastIndex := r.ExistingNode.Position[len(r.ExistingNode.Position)-1]
		for _, sibling := range r.ExistingNode.ParentNode.ChildNodes[lastIndex:] {
			switch sibling.NodeType {
			case ElementNodeType:
				sibling.Position[len(sibling.Position)-1] = sibling.Position[len(sibling.Position)-1] - 1
			case TextNodeType:
				sibling.Position[len(sibling.Position)-1] = sibling.Position[len(sibling.Position)-1] - 1
			case CommentNodeType:
				sibling.Position[len(sibling.Position)-1] = sibling.Position[len(sibling.Position)-1] - 1
			default:
				panic("Undetermined Node Type!")
			}
		}
	}

}

func (r *Reconciler) RemoveNodeAttribute(rootElement dom.Element) {
	self := r.ExistingNode.Locate(rootElement).(dom.Element)
	self.RemoveAttribute(r.AttributeName)
}

func (r *Reconciler) SetNodeAttribute(rootElement dom.Element) {
	self := r.ExistingNode.Locate(rootElement).(dom.Element)
	self.SetAttribute(r.AttributeName, r.AttributeValue)
}

func (c Changes) ApplyChanges(rootElement dom.Element) {

	for _, reconciler := range c {
		reconciler.ApplyChange(rootElement)
	}

}
