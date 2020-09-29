// Copyright 2020 Sam Uong
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ast

type Expr interface {
	Visit(v ExprVisitor) (string, error)
}

type ExprVisitor interface {
	VisitString(s String) (string, error)
	VisitTilde(t Tilde) (string, error)
	VisitVar(v Var) (string, error)
	VisitWord(w Word) (string, error)
}

type String struct {
	Text string
}

func (s String) Visit(v ExprVisitor) (string, error) {
	return v.VisitString(s)
}

type Tilde struct {
	Text string
}

func (t Tilde) Visit(v ExprVisitor) (string, error) {
	return v.VisitTilde(t)
}

type Var struct {
	Identifier string
}

func (v Var) Visit(visit ExprVisitor) (string, error) {
	return visit.VisitVar(v)
}

type Word struct {
	SubExprs []Expr
}

func (w Word) Visit(v ExprVisitor) (string, error) {
	return v.VisitWord(w)
}

