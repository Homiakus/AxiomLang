.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/mca_memory.inc"

.section .text
.global parser_init
.global parser_parse_decl

# AST Node Type IDs
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

# AST Node structure offsets (MCA Allocated)
.equ AST_OFF_TYPE,      0
.equ AST_OFF_NAME_PTR,  8
.equ AST_OFF_NAME_LEN,  16
.equ AST_OFF_CHILD_PTR, 24
.equ AST_OFF_NEXT_PTR,  32
.equ AST_NODE_SIZE,     40

# ==============================================================================
# parser_create_node(arena: %rdi, node_type: %rsi) -> %rax (Node Pointer)
# ==============================================================================
parser_create_node:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12

    movq %rsi, %r12          # type
    movq $AST_NODE_SIZE, %rsi
    call mca_arena_alloc
    testq %rax, %rax
    jz create_node_fail

    movq %r12, AST_OFF_TYPE(%rax)
    movq $0, AST_OFF_NAME_PTR(%rax)
    movq $0, AST_OFF_NAME_LEN(%rax)
    movq $0, AST_OFF_CHILD_PTR(%rax)
    movq $0, AST_OFF_NEXT_PTR(%rax)

create_node_fail:
    popq %r12
    leave
    ret
