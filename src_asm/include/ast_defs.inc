# ==============================================================================
# Shared AST Node Definitions for Parser, Checker, and Codegen
# ==============================================================================

.equ AST_PROGRAM,    1
.equ AST_REQUIRE,    2
.equ AST_INTENT,     3
.equ AST_CAP,        4
.equ AST_FN,         5
.equ AST_ACTOR,      6
.equ AST_LET,        7
.equ AST_RETURN,     8
.equ AST_CALL,       9
.equ AST_EXPR_BIN,   10
.equ AST_INT,        11
.equ AST_STRING,     12
.equ AST_IDENT,      13
.equ AST_BLOCK,      14
.equ AST_IF,         15
.equ AST_WHILE,      16
.equ AST_FOR,        17
.equ AST_ASSIGN,     18
.equ AST_UNARY,      19
.equ AST_PARAM,      20
.equ AST_ARRAY,      21
.equ AST_INDEX,      22
.equ AST_STRUCT,     23
.equ AST_STRUCT_LIT, 24
.equ AST_FIELD,      25

.equ AST_OFF_TYPE,      0
.equ AST_OFF_NAME_PTR,  8
.equ AST_OFF_NAME_LEN,  16
.equ AST_OFF_CHILD_PTR, 24
.equ AST_OFF_NEXT_PTR,  32
.equ AST_OFF_RIGHT_PTR, 40
.equ AST_OFF_VAL,       48
.equ AST_OFF_LINE,      56
.equ AST_NODE_SIZE,     64
