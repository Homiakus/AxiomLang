package ast

import (
	"AxiomLang/pkg/lexer"
	"bytes"
	"fmt"
	"strings"
)

// Node base interface
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement interface
type Statement interface {
	Node
	statementNode()
}

// Expression interface
type Expression interface {
	Node
	expressionNode()
}

// Program is the root AST node
type Program struct {
	Imports      []*ImportStatement
	Enums        []*EnumStatement
	Capabilities []*CapabilityStatement
	Structs      []*StructStatement
	Functions    []*FunctionStatement
	Intents      []*IntentStatement
	Statements   []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, imp := range p.Imports {
		out.WriteString(imp.String() + "\n")
	}
	for _, e := range p.Enums {
		out.WriteString(e.String() + "\n")
	}
	for _, c := range p.Capabilities {
		out.WriteString(c.String() + "\n")
	}
	for _, s := range p.Structs {
		out.WriteString(s.String() + "\n")
	}
	for _, f := range p.Functions {
		out.WriteString(f.String() + "\n")
	}
	for _, i := range p.Intents {
		out.WriteString(i.String() + "\n")
	}
	for _, s := range p.Statements {
		out.WriteString(s.String() + "\n")
	}
	return out.String()
}

// ImportStatement represents: import "path/to/module.axl"
type ImportStatement struct {
	Token lexer.Token // IMPORT token
	Path  string
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string       { return fmt.Sprintf("import %q", is.Path) }

// EnumVariant represents variant in enum
type EnumVariant struct {
	Name *Identifier
	Type string // optional payload type, e.g. "int" or ""
}

// EnumStatement represents: enum Result { Ok(string), Err(int) }
type EnumStatement struct {
	Token    lexer.Token // ENUM token
	Name     *Identifier
	Variants []*EnumVariant
}

func (es *EnumStatement) statementNode()       {}
func (es *EnumStatement) TokenLiteral() string { return es.Token.Literal }
func (es *EnumStatement) String() string {
	var vars []string
	for _, v := range es.Variants {
		if v.Type != "" {
			vars = append(vars, fmt.Sprintf("%s(%s)", v.Name.String(), v.Type))
		} else {
			vars = append(vars, v.Name.String())
		}
	}
	return fmt.Sprintf("enum %s { %s }", es.Name.String(), strings.Join(vars, ", "))
}

// CapabilityStatement represents: require cap[Resource.Action(args)]
type CapabilityStatement struct {
	Token    lexer.Token // REQUIRE token
	Resource string      // e.g. "Console.Out" or "FileSystem.Write"
	Arg      Expression  // e.g. "/data/user.json"
}

func (cs *CapabilityStatement) statementNode()       {}
func (cs *CapabilityStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *CapabilityStatement) String() string {
	if cs.Arg != nil {
		return fmt.Sprintf("require cap[%s(%s)]", cs.Resource, cs.Arg.String())
	}
	return fmt.Sprintf("require cap[%s]", cs.Resource)
}

// StructField represents: name: type
type StructField struct {
	Name *Identifier
	Type string
}

// StructStatement represents: struct Name { field1: type1, field2: type2 }
type StructStatement struct {
	Token  lexer.Token // STRUCT token
	Name   *Identifier
	Fields []*StructField
}

func (ss *StructStatement) statementNode()       {}
func (ss *StructStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *StructStatement) String() string {
	var fields []string
	for _, f := range ss.Fields {
		fields = append(fields, fmt.Sprintf("%s: %s", f.Name.String(), f.Type))
	}
	return fmt.Sprintf("struct %s { %s }", ss.Name.String(), strings.Join(fields, ", "))
}

// IntentStatement represents: intent "description" { ... }
type IntentStatement struct {
	Token       lexer.Token // INTENT token
	Description string
	Body        *BlockStatement
}

func (is *IntentStatement) statementNode()       {}
func (is *IntentStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IntentStatement) String() string {
	return fmt.Sprintf("intent %q %s", is.Description, is.Body.String())
}

// BlockStatement represents { stmt; stmt; }
type BlockStatement struct {
	Token      lexer.Token // { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{\n")
	for _, s := range bs.Statements {
		out.WriteString("  " + s.String() + "\n")
	}
	out.WriteString("}")
	return out.String()
}

// Parameter represents a typed function argument: name: type
type Parameter struct {
	Name *Identifier
	Type string
}

// FunctionStatement represents: fn name(param1: type, ...) -> returnType { ... }
type FunctionStatement struct {
	Token      lexer.Token // FN token
	Name       *Identifier
	Parameters []*Parameter
	ReturnType string
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var params []string
	for _, p := range fs.Parameters {
		params = append(params, fmt.Sprintf("%s: %s", p.Name.String(), p.Type))
	}
	ret := ""
	if fs.ReturnType != "" {
		ret = " -> " + fs.ReturnType
	}
	return fmt.Sprintf("fn %s(%s)%s %s", fs.Name.String(), strings.Join(params, ", "), ret, fs.Body.String())
}

// ReturnStatement represents: return value;
type ReturnStatement struct {
	Token       lexer.Token // RETURN token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	if rs.ReturnValue != nil {
		return fmt.Sprintf("return %s;", rs.ReturnValue.String())
	}
	return "return;"
}

// LetStatement represents: let [mut] name [: type] = value;
type LetStatement struct {
	Token   lexer.Token // LET token
	Name    *Identifier
	Mutable bool
	Type    string
	Value   Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	mut := ""
	if ls.Mutable {
		mut = "mut "
	}
	typeStr := ""
	if ls.Type != "" {
		typeStr = ": " + ls.Type
	}
	return fmt.Sprintf("let %s%s%s = %s;", mut, ls.Name.String(), typeStr, ls.Value.String())
}

// SpawnStatement represents: spawn actor_id(payload); (Erlang pattern)
type SpawnStatement struct {
	Token   lexer.Token // SPAWN token
	ActorID *Identifier
	Payload Expression
}

func (ss *SpawnStatement) statementNode()       {}
func (ss *SpawnStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SpawnStatement) String() string {
	if ss.Payload != nil {
		return fmt.Sprintf("spawn %s(%s);", ss.ActorID.String(), ss.Payload.String())
	}
	return fmt.Sprintf("spawn %s();", ss.ActorID.String())
}

// SendStatement represents: send target <- message; (Erlang/Go pattern)
type SendStatement struct {
	Token   lexer.Token // SEND token
	Target  *Identifier
	Message Expression
}

func (ss *SendStatement) statementNode()       {}
func (ss *SendStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SendStatement) String() string {
	return fmt.Sprintf("send %s <- %s;", ss.Target.String(), ss.Message.String())
}

// MatchCase represents a single arm: Pattern => { ... }
type MatchCase struct {
	Pattern Expression
	Body    *BlockStatement
}

// MatchStatement represents: match expr { pattern => { ... } }
type MatchStatement struct {
	Token lexer.Token // MATCH token
	Value Expression
	Cases []*MatchCase
}

func (ms *MatchStatement) statementNode()       {}
func (ms *MatchStatement) TokenLiteral() string { return ms.Token.Literal }
func (ms *MatchStatement) String() string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("match %s {\n", ms.Value.String()))
	for _, c := range ms.Cases {
		out.WriteString(fmt.Sprintf("    %s => %s\n", c.Pattern.String(), c.Body.String()))
	}
	out.WriteString("}")
	return out.String()
}

// IfStatement represents: if condition { consequence } else { alternative }
type IfStatement struct {
	Token       lexer.Token // IF token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("if ")
	out.WriteString(is.Condition.String())
	out.WriteString(" ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

// ForInStatement represents: for item in collection { ... }
type ForInStatement struct {
	Token      lexer.Token // FOR token
	Item       *Identifier
	Collection Expression
	Body       *BlockStatement
}

func (fs *ForInStatement) statementNode()       {}
func (fs *ForInStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForInStatement) String() string {
	return fmt.Sprintf("for %s in %s %s", fs.Item.String(), fs.Collection.String(), fs.Body.String())
}

// ExpressionStatement wraps an expression as a statement
type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// Identifier represents an identifier expression
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// StringLiteral represents a string
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return fmt.Sprintf("%q", sl.Value) }

// BooleanLiteral represents true / false
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// ArrayLiteral represents [elem1, elem2, ...]
type ArrayLiteral struct {
	Token    lexer.Token // '['
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var elements []string
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

// BinaryExpression represents left op right (e.g. x + 5, a > b)
type BinaryExpression struct {
	Token    lexer.Token // operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", be.Left.String(), be.Operator, be.Right.String())
}

// CallExpression represents fn(arg1, arg2) or obj.method(arg1)
type CallExpression struct {
	Token     lexer.Token // '(' token or identifier
	Function  Expression  // Identifier or MemberAccessExpression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), strings.Join(args, ", "))
}

// MemberAccessExpression represents object.field
type MemberAccessExpression struct {
	Token  lexer.Token // '.'
	Object Expression
	Member string
}

func (me *MemberAccessExpression) expressionNode()      {}
func (me *MemberAccessExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MemberAccessExpression) String() string {
	if me == nil {
		return ""
	}
	if me.Object == nil {
		return me.Member
	}
	return fmt.Sprintf("%s.%s", me.Object.String(), me.Member)
}

// MapLiteral represents { "key": value, ... }
type MapLiteral struct {
	Token lexer.Token // '{'
	Pairs map[Expression]Expression
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *MapLiteral) String() string {
	var pairs []string
	for key, value := range ml.Pairs {
		pairs = append(pairs, key.String()+": "+value.String())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

// PrefixExpression represents !expr or -expr
type PrefixExpression struct {
	Token    lexer.Token // '!' or '-'
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

// RangeExpression represents start..end (e.g. 0..10)
type RangeExpression struct {
	Token    lexer.Token // '..'
	Start    Expression
	End      Expression
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token.Literal }
func (re *RangeExpression) String() string {
	startStr := ""
	if re.Start != nil {
		startStr = re.Start.String()
	}
	endStr := ""
	if re.End != nil {
		endStr = re.End.String()
	}
	return fmt.Sprintf("%s..%s", startStr, endStr)
}

// PipeExpression represents value |> function(args...)
type PipeExpression struct {
	Token    lexer.Token // '|>'
	Left     Expression
	Right    Expression  // CallExpression or Identifier
}

func (pe *PipeExpression) expressionNode()      {}
func (pe *PipeExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PipeExpression) String() string {
	return fmt.Sprintf("(%s |> %s)", pe.Left.String(), pe.Right.String())
}

// TryExpression represents expr?
type TryExpression struct {
	Token    lexer.Token // '?'
	Value    Expression
}

func (te *TryExpression) expressionNode()      {}
func (te *TryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TryExpression) String() string {
	return fmt.Sprintf("%s?", te.Value.String())
}

// IndexExpression represents expr[index]
type IndexExpression struct {
	Token lexer.Token // '['
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return fmt.Sprintf("%s[%s]", ie.Left.String(), ie.Index.String())
}




