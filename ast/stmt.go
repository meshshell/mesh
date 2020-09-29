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

type Stmt interface {
	Visit(v StmtVisitor) (int, error)
}

type StmtVisitor interface {
	VisitStmtList(s *StmtList) (int, error)
	VisitPipeline(p *Pipeline) (int, error)
	VisitCmd(c *Cmd) (int, error)
}

type StmtList struct {
	Stmts []Stmt
}

func (s *StmtList) Visit(v StmtVisitor) (int, error) {
	return v.VisitStmtList(s)
}

type Pipeline struct {
	Stmts []Stmt
}

func (p *Pipeline) Visit(v StmtVisitor) (int, error) {
	return v.VisitPipeline(p)
}

type Cmd struct {
	Argv []Expr
}

func (c *Cmd) Visit(v StmtVisitor) (int, error) {
	return v.VisitCmd(c)
}
