.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/mca_memory.inc"
.include "src_asm/include/ast_defs.inc"

.section .rodata
cap_str_io:      .ascii "IO"
cap_str_console: .ascii "Console"
cap_str_memory:  .ascii "Memory"
kw_axiom_spawn:  .ascii "__axiom_spawn"

.section .bss
.align 8
parser_arena:    .quad 0
p_src_ptr:       .quad 0
p_src_len:       .quad 0
p_lex_pos:       .quad 0
p_line:          .quad 0
cur_tok_type:    .quad 0
cur_tok_start:   .quad 0
cur_tok_len:     .quad 0
p_has_error:     .quad 0

.section .text
.global parser_init
.global parser_destroy
.global parser_parse_program
.global parser_parse_decl
.global p_line

# ==============================================================================
# parser_init(src_ptr: %rdi, src_len: %rsi) -> %rax (0 = OK, 1 = ERR)
# ==============================================================================
parser_init:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13

    movq %rdi, %r12
    movq %rsi, %r13

    movq %r12, p_src_ptr(%rip)
    movq %r13, p_src_len(%rip)
    movq $0, p_lex_pos(%rip)
    movq $1, p_line(%rip)
    movq $0, p_has_error(%rip)

    # Create 2MB MCA Arena for AST nodes
    movq $2097152, %rdi
    movq $(CAP_MEM_READ | CAP_MEM_WRITE), %rsi
    call mca_arena_create
    testq %rax, %rax
    jz .L_pinit_err
    movq %rax, parser_arena(%rip)

    call next_tok
    xorq %rax, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_pinit_err:
    movq $1, %rax
    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# parser_destroy()
# ==============================================================================
parser_destroy:
    movq parser_arena(%rip), %rdi
    testq %rdi, %rdi
    jz .L_pdest_done
    call mca_arena_destroy
    movq $0, parser_arena(%rip)
.L_pdest_done:
    ret

# ==============================================================================
# next_tok(): advances lexer, counts newlines, updates cur_tok_*
# ==============================================================================
next_tok:
    pushq %rbx
    pushq %r12
    pushq %r13

    movq p_lex_pos(%rip), %r12
    movq p_src_ptr(%rip), %rdi
    movq p_src_len(%rip), %rsi
    movq %r12, %rdx
    call lexer_next_token

    movq %rax, cur_tok_type(%rip)
    movq %rdx, p_lex_pos(%rip)
    movq %r8, cur_tok_start(%rip)
    movq %r9, cur_tok_len(%rip)

    movq p_src_ptr(%rip), %rbx
.L_line_loop:
    cmpq %rdx, %r12
    jge .L_line_done
    cmpb $10, (%rbx, %r12)
    jne .L_line_next
    incq p_line(%rip)
.L_line_next:
    incq %r12
    jmp .L_line_loop
.L_line_done:
    cmpq $TOK_ILLEGAL, cur_tok_type(%rip)
    jne .L_nt_ok
    movq $1, p_has_error(%rip)
.L_nt_ok:
    popq %r13
    popq %r12
    popq %rbx
    ret

# ==============================================================================
# parser_create_node(node_type: %rdi) -> %rax (Node Pointer)
# ==============================================================================
parser_create_node:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12

    movq %rdi, %r12
    movq parser_arena(%rip), %rdi
    movq $AST_NODE_SIZE, %rsi
    call mca_arena_alloc
    testq %rax, %rax
    jz .L_cn_fail

    movq %r12, AST_OFF_TYPE(%rax)
    movq $0, AST_OFF_NAME_PTR(%rax)
    movq $0, AST_OFF_NAME_LEN(%rax)
    movq $0, AST_OFF_CHILD_PTR(%rax)
    movq $0, AST_OFF_NEXT_PTR(%rax)
    movq $0, AST_OFF_RIGHT_PTR(%rax)
    movq $0, AST_OFF_VAL(%rax)
    movq p_line(%rip), %rcx
    movq %rcx, AST_OFF_LINE(%rax)

.L_cn_fail:
    popq %r12
    leave
    ret

# Helper: skip balanced brackets [...], returns item count in %rax
skip_brackets:
    pushq %r12
    pushq %r13
    xorq %r13, %r13          # non-bracket token count
    cmpq $TOK_LBRACKET, cur_tok_type(%rip)
    jne .L_sb_done
    call next_tok
    movq $1, %r12
.L_sb_loop:
    cmpq $TOK_EOF, cur_tok_type(%rip)
    je .L_sb_done
    cmpq $TOK_LBRACKET, cur_tok_type(%rip)
    jne .L_sb_chk_r
    incq %r12
    jmp .L_sb_adv
.L_sb_chk_r:
    cmpq $TOK_RBRACKET, cur_tok_type(%rip)
    jne .L_sb_cnt
    decq %r12
    jz .L_sb_last
    jmp .L_sb_adv
.L_sb_cnt:
    incq %r13
.L_sb_adv:
    call next_tok
    jmp .L_sb_loop
.L_sb_last:
    call next_tok
.L_sb_done:
    movq %r13, %rax
    popq %r13
    popq %r12
    ret

# Helper: skip balanced parens (...)
skip_parens:
    pushq %r12
    cmpq $TOK_LPAREN, cur_tok_type(%rip)
    jne .L_sp_done
    call next_tok
    movq $1, %r12
.L_sp_loop:
    cmpq $TOK_EOF, cur_tok_type(%rip)
    je .L_sp_done
    cmpq $TOK_LPAREN, cur_tok_type(%rip)
    jne .L_sp_chk_r
    incq %r12
    jmp .L_sp_adv
.L_sp_chk_r:
    cmpq $TOK_RPAREN, cur_tok_type(%rip)
    jne .L_sp_adv
    decq %r12
    jz .L_sp_last
.L_sp_adv:
    call next_tok
    jmp .L_sp_loop
.L_sp_last:
    call next_tok
.L_sp_done:
    popq %r12
    ret

# ==============================================================================
# parser_parse_program() -> %rax (AST_PROGRAM root or 0 on error)
# ==============================================================================
parser_parse_program:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14

    movq $AST_PROGRAM, %rdi
    call parser_create_node
    testq %rax, %rax
    jz .L_prog_err
    movq %rax, %r12          # root
    xorq %r13, %r13          # head
    xorq %r14, %r14          # tail

.L_prog_loop:
    cmpq $TOK_EOF, cur_tok_type(%rip)
    je .L_prog_done
    cmpq $0, p_has_error(%rip)
    jne .L_prog_err

    call parser_parse_decl
    testq %rax, %rax
    jz .L_prog_chk_err

    testq %r13, %r13
    jnz .L_prog_append
    movq %rax, %r13
    movq %rax, %r14
    jmp .L_prog_loop

.L_prog_append:
    movq %rax, AST_OFF_NEXT_PTR(%r14)
    movq %rax, %r14
    jmp .L_prog_loop

.L_prog_chk_err:
    cmpq $0, p_has_error(%rip)
    jne .L_prog_err
    jmp .L_prog_loop

.L_prog_done:
    cmpq $0, p_has_error(%rip)
    jne .L_prog_err
    movq %r13, AST_OFF_CHILD_PTR(%r12)
    movq %r12, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_prog_err:
    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# parser_parse_decl() -> %rax (AST node or 0)
# ==============================================================================
parser_parse_decl:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13

    movq cur_tok_type(%rip), %rax
    cmpq $TOK_REQUIRE, %rax
    je .L_decl_require
    cmpq $TOK_INTENT, %rax
    je .L_decl_intent
    cmpq $TOK_CAP, %rax
    je .L_decl_cap
    cmpq $TOK_FN, %rax
    je .L_decl_fn
    cmpq $TOK_ACTOR, %rax
    je .L_decl_actor
    cmpq $TOK_STRUCT, %rax
    je .L_decl_struct

    call parse_stmt
    popq %r13
    popq %r12
    leave
    ret

.L_decl_require:
    call next_tok
    cmpq $TOK_CAP, cur_tok_type(%rip)
    jne .L_dr_br
    call next_tok
.L_dr_br:
    # Parse [...] and collect item count (%r13) + capability bitmask (%r12)
    xorq %r13, %r13
    movq $0x10000, %r12      # bit 16 = require present
    cmpq $TOK_LBRACKET, cur_tok_type(%rip)
    jne .L_dr_mk
    call next_tok
.L_dr_loop:
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_EOF, %rax
    je .L_dr_mk
    cmpq $TOK_RBRACKET, %rax
    je .L_dr_close
    incq %r13
    cmpq $TOK_IDENT, %rax
    jne .L_dr_adv
    movq p_src_ptr(%rip), %rdi
    addq cur_tok_start(%rip), %rdi
    movq cur_tok_len(%rip), %rsi
    leaq cap_str_io(%rip), %rdx
    movq $2, %rcx
    call p_slice_eq
    testq %rax, %rax
    jz .L_dr_chk_con
    orq $1, %r12
    jmp .L_dr_adv
.L_dr_chk_con:
    movq p_src_ptr(%rip), %rdi
    addq cur_tok_start(%rip), %rdi
    movq cur_tok_len(%rip), %rsi
    leaq cap_str_console(%rip), %rdx
    movq $7, %rcx
    call p_slice_eq
    testq %rax, %rax
    jz .L_dr_chk_mem
    orq $1, %r12
    jmp .L_dr_adv
.L_dr_chk_mem:
    movq p_src_ptr(%rip), %rdi
    addq cur_tok_start(%rip), %rdi
    movq cur_tok_len(%rip), %rsi
    leaq cap_str_memory(%rip), %rdx
    movq $6, %rcx
    call p_slice_eq
    testq %rax, %rax
    jz .L_dr_adv
    orq $2, %r12
.L_dr_adv:
    call next_tok
    jmp .L_dr_loop
.L_dr_close:
    call next_tok
.L_dr_mk:
    movq $AST_REQUIRE, %rdi
    call parser_create_node
    testq %rax, %rax
    jz .L_dr_ret
    movq %r13, AST_OFF_VAL(%rax)
    movq %r12, AST_OFF_RIGHT_PTR(%rax)
.L_dr_ret:
    popq %r13
    popq %r12
    leave
    ret

p_slice_eq:
    cmpq %rcx, %rsi
    jne .L_pse_ne
    xorq %r8, %r8
.L_pse_loop:
    cmpq %rsi, %r8
    jge .L_pse_eq
    movb (%rdi, %r8), %al
    cmpb (%rdx, %r8), %al
    jne .L_pse_ne
    incq %r8
    jmp .L_pse_loop
.L_pse_eq:
    movq $1, %rax
    ret
.L_pse_ne:
    xorq %rax, %rax
    ret

.L_decl_intent:
    call next_tok
    movq $AST_INTENT, %rdi
    call parser_create_node
    movq %rax, %r12
    cmpq $TOK_STRING, cur_tok_type(%rip)
    jne .L_di_chk_block
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r12)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r12)
    call next_tok
.L_di_chk_block:
    cmpq $TOK_LBRACE, cur_tok_type(%rip)
    jne .L_di_done
    call parse_block
    movq %rax, AST_OFF_CHILD_PTR(%r12)
.L_di_done:
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_decl_cap:
    call next_tok
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_dc_br
    call next_tok
.L_dc_br:
    # Parse [...] and extract Max: <N>[KB|MB] if present
    xorq %r12, %r12          # max_bytes (0 = default)
    cmpq $TOK_LBRACKET, cur_tok_type(%rip)
    jne .L_dc_mk
    call next_tok
.L_dc_loop:
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_EOF, %rax
    je .L_dc_mk
    cmpq $TOK_RBRACKET, %rax
    je .L_dc_close
    cmpq $TOK_IDENT, %rax
    jne .L_dc_next
    # Check "Max"
    cmpq $3, cur_tok_len(%rip)
    jne .L_dc_next
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    cmpb $'M', (%rcx)
    jne .L_dc_next
    cmpb $'a', 1(%rcx)
    jne .L_dc_next
    cmpb $'x', 2(%rcx)
    jne .L_dc_next
    call next_tok
    cmpq $TOK_COLON, cur_tok_type(%rip)
    jne .L_dc_loop
    call next_tok
    cmpq $TOK_INT, cur_tok_type(%rip)
    jne .L_dc_loop
    # Parse integer token
    movq p_src_ptr(%rip), %rdi
    addq cur_tok_start(%rip), %rdi
    movq cur_tok_len(%rip), %rsi
    call parse_u64_slice
    movq %rax, %r12
    call next_tok
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_dc_loop
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    cmpq $2, cur_tok_len(%rip)
    jne .L_dc_next
    cmpb $'B', 1(%rcx)
    jne .L_dc_next
    cmpb $'K', (%rcx)
    jne .L_dc_chk_mb
    shlq $10, %r12
    jmp .L_dc_next
.L_dc_chk_mb:
    cmpb $'M', (%rcx)
    jne .L_dc_next
    shlq $20, %r12
.L_dc_next:
    call next_tok
    jmp .L_dc_loop
.L_dc_close:
    call next_tok
.L_dc_mk:
    movq $AST_CAP, %rdi
    call parser_create_node
    testq %rax, %rax
    jz .L_dc_ret
    movq %r12, AST_OFF_VAL(%rax)
.L_dc_ret:
    popq %r13
    popq %r12
    leave
    ret

parse_u64_slice:
    xorq %rax, %rax
    xorq %rcx, %rcx
.L_pus_loop:
    cmpq %rsi, %rcx
    jge .L_pus_done
    movzbq (%rdi, %rcx), %rdx
    subq $'0', %rdx
    cmpq $9, %rdx
    ja .L_pus_done
    imulq $10, %rax
    addq %rdx, %rax
    incq %rcx
    jmp .L_pus_loop
.L_pus_done:
    ret

.L_decl_fn:
    call next_tok
    movq $AST_FN, %rdi
    call parser_create_node
    movq %rax, %r12
    testq %r12, %r12
    jz .L_df_err

    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_df_params
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r12)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r12)
    call next_tok

.L_df_params:
    call parse_param_list
    movq %rax, AST_OFF_RIGHT_PTR(%r12)

    # Skip optional return type annotation until '{' or EOF
.L_df_ret_type:
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_LBRACE, %rax
    je .L_df_body
    cmpq $TOK_EOF, %rax
    je .L_df_body
    call next_tok
    jmp .L_df_ret_type

.L_df_body:
    call parse_block
    movq %rax, AST_OFF_CHILD_PTR(%r12)
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret
.L_df_err:
    xorq %rax, %rax
    popq %r13
    popq %r12
    leave
    ret

# Parse (p1 [: T], p2 [: T], ...) -> %rax (linked list of AST_PARAM)
parse_param_list:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14
    xorq %r12, %r12          # head
    xorq %r13, %r13          # tail

    cmpq $TOK_LPAREN, cur_tok_type(%rip)
    jne .L_ppl_done
    call next_tok

.L_ppl_loop:
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_RPAREN, %rax
    je .L_ppl_close
    cmpq $TOK_EOF, %rax
    je .L_ppl_done
    cmpq $TOK_IDENT, %rax
    jne .L_ppl_skip

    movq $AST_PARAM, %rdi
    call parser_create_node
    movq %rax, %r14
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r14)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r14)
    call next_tok

    # Optional ': Type'
    cmpq $TOK_COLON, cur_tok_type(%rip)
    jne .L_ppl_link
    call next_tok
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_ppl_link
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq cur_tok_len(%rip), %rdx
    cmpq $3, %rdx
    je .L_ppl_chk_t3
    cmpq $6, %rdx
    je .L_ppl_chk_str
    jmp .L_ppl_tnext
.L_ppl_chk_t3:
    cmpb $'i', (%rcx)
    jne .L_ppl_chk_str
    cmpb $'n', 1(%rcx)
    jne .L_ppl_tnext
    cmpb $'t', 2(%rcx)
    jne .L_ppl_tnext
    movq $1, AST_OFF_VAL(%r14)
    jmp .L_ppl_tnext
.L_ppl_chk_str:
    cmpb $'s', (%rcx)
    jne .L_ppl_tnext
    cmpb $'t', 1(%rcx)
    jne .L_ppl_tnext
    cmpb $'r', 2(%rcx)
    jne .L_ppl_tnext
    movq $2, AST_OFF_VAL(%r14)
.L_ppl_tnext:
    call next_tok

.L_ppl_link:
    testq %r12, %r12
    jnz .L_ppl_app
    movq %r14, %r12
    movq %r14, %r13
    jmp .L_ppl_comma
.L_ppl_app:
    movq %r14, AST_OFF_NEXT_PTR(%r13)
    movq %r14, %r13

.L_ppl_comma:
    cmpq $TOK_COMMA, cur_tok_type(%rip)
    jne .L_ppl_loop
    call next_tok
    jmp .L_ppl_loop

.L_ppl_skip:
    call next_tok
    jmp .L_ppl_loop

.L_ppl_close:
    call next_tok
.L_ppl_done:
    movq %r12, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_decl_actor:
    call next_tok
    movq $AST_ACTOR, %rdi
    call parser_create_node
    movq %rax, %r12
    testq %r12, %r12
    jz .L_df_err
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_da_body
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r12)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r12)
    call next_tok
.L_da_body:
    call parse_block
    movq %rax, AST_OFF_CHILD_PTR(%r12)
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_decl_struct:
    call next_tok
    movq $AST_STRUCT, %rdi
    call parser_create_node
    movq %rax, %r12
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_ds_body
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r12)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r12)
    call next_tok
.L_ds_body:
    xorq %r13, %r13          # field head
    pushq %r13               # [0(%rsp)] = tail
    cmpq $TOK_LBRACE, cur_tok_type(%rip)
    jne .L_ds_done
    call next_tok
.L_ds_loop:
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_RBRACE, %rax
    je .L_ds_close
    cmpq $TOK_EOF, %rax
    je .L_ds_done
    cmpq $TOK_IDENT, %rax
    jne .L_ds_skip
    movq $AST_PARAM, %rdi
    call parser_create_node
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%rax)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%rax)
    pushq %rax
    call next_tok
    cmpq $TOK_COLON, cur_tok_type(%rip)
    jne .L_ds_link
    call next_tok
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_ds_link
    # Check if field type is "str" or "string" -> mark AST_OFF_VAL = 2
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq cur_tok_len(%rip), %rdx
    cmpq $3, %rdx
    je .L_ds_chk_str
    cmpq $6, %rdx
    jne .L_ds_type_next
.L_ds_chk_str:
    cmpb $'s', (%rcx)
    jne .L_ds_type_next
    cmpb $'t', 1(%rcx)
    jne .L_ds_type_next
    cmpb $'r', 2(%rcx)
    jne .L_ds_type_next
    movq (%rsp), %rcx
    movq $2, AST_OFF_VAL(%rcx)
.L_ds_type_next:
    call next_tok
.L_ds_link:
    popq %rax
    testq %r13, %r13
    jnz .L_ds_app
    movq %rax, %r13
    movq %rax, (%rsp)
    jmp .L_ds_comma
.L_ds_app:
    movq (%rsp), %rcx
    movq %rax, AST_OFF_NEXT_PTR(%rcx)
    movq %rax, (%rsp)
.L_ds_comma:
    cmpq $TOK_COMMA, cur_tok_type(%rip)
    jne .L_ds_loop
    call next_tok
    jmp .L_ds_loop
.L_ds_skip:
    call next_tok
    jmp .L_ds_loop
.L_ds_close:
    call next_tok
.L_ds_done:
    addq $8, %rsp
    movq %r13, AST_OFF_CHILD_PTR(%r12)
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# parse_block() -> %rax (first statement in linked list, or 0)
# ==============================================================================
parse_block:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13

    xorq %r12, %r12
    xorq %r13, %r13

    cmpq $TOK_LBRACE, cur_tok_type(%rip)
    jne .L_pb_done
    call next_tok

.L_pb_loop:
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_EOF, %rax
    je .L_pb_done
    cmpq $TOK_RBRACE, %rax
    je .L_pb_close
    cmpq $0, p_has_error(%rip)
    jne .L_pb_done

    call parse_stmt
    testq %rax, %rax
    jz .L_pb_loop

    testq %r12, %r12
    jnz .L_pb_append
    movq %rax, %r12
    movq %rax, %r13
    jmp .L_pb_loop
.L_pb_append:
    movq %rax, AST_OFF_NEXT_PTR(%r13)
    movq %rax, %r13
    jmp .L_pb_loop

.L_pb_close:
    call next_tok
.L_pb_done:
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# parse_stmt() -> %rax (AST node or 0)
# ==============================================================================
parse_stmt:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14

    movq cur_tok_type(%rip), %rax
    cmpq $TOK_LET, %rax
    je .L_st_let
    cmpq $TOK_RETURN, %rax
    je .L_st_return
    cmpq $TOK_IF, %rax
    je .L_st_if
    cmpq $TOK_WHILE, %rax
    je .L_st_while
    cmpq $TOK_FOR, %rax
    je .L_st_for
    cmpq $TOK_RECEIVE, %rax
    je .L_st_receive
    cmpq $TOK_SPAWN, %rax
    je .L_st_spawn
    cmpq $TOK_IDENT, %rax
    je .L_st_ident
    cmpq $TOK_SEMICOLON, %rax
    je .L_st_skip

    call next_tok
    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_skip:
    call next_tok
    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_let:
    call next_tok
    cmpq $TOK_MUT, cur_tok_type(%rip)
    jne .L_sl_name
    call next_tok
.L_sl_name:
    movq $AST_LET, %rdi
    call parser_create_node
    movq %rax, %r12
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_sl_err
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r12)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r12)
    call next_tok

    # Optional ': Type'
    cmpq $TOK_COLON, cur_tok_type(%rip)
    jne .L_sl_eq
    call next_tok
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_sl_eq
    call next_tok
.L_sl_eq:
    cmpq $TOK_ASSIGN, cur_tok_type(%rip)
    jne .L_sl_err
    call next_tok
    call parse_expr
    movq %rax, AST_OFF_CHILD_PTR(%r12)
    cmpq $TOK_SEMICOLON, cur_tok_type(%rip)
    jne .L_sl_ok
    call next_tok
.L_sl_ok:
    movq %r12, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret
.L_sl_err:
    movq $1, p_has_error(%rip)
    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_return:
    call next_tok
    movq $AST_RETURN, %rdi
    call parser_create_node
    movq %rax, %r12
    movq cur_tok_type(%rip), %rcx
    cmpq $TOK_RBRACE, %rcx
    je .L_sr_done
    cmpq $TOK_EOF, %rcx
    je .L_sr_done
    cmpq $TOK_SEMICOLON, %rcx
    je .L_sr_semi
    call parse_expr
    movq %rax, AST_OFF_CHILD_PTR(%r12)
    cmpq $TOK_SEMICOLON, cur_tok_type(%rip)
    jne .L_sr_done
.L_sr_semi:
    call next_tok
.L_sr_done:
    movq %r12, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_if:
    call next_tok
    movq $AST_IF, %rdi
    call parser_create_node
    movq %rax, %r12          # AST_IF

    call parse_expr
    movq %rax, AST_OFF_CHILD_PTR(%r12)

    movq $AST_BLOCK, %rdi
    call parser_create_node
    movq %rax, %r13          # branches container
    movq %r13, AST_OFF_RIGHT_PTR(%r12)

    call parse_block
    movq %rax, AST_OFF_CHILD_PTR(%r13)

    cmpq $TOK_ELSE, cur_tok_type(%rip)
    jne .L_sif_done
    call next_tok
    cmpq $TOK_IF, cur_tok_type(%rip)
    je .L_sif_elif
    call parse_block
    movq %rax, AST_OFF_RIGHT_PTR(%r13)
    jmp .L_sif_done
.L_sif_elif:
    call parse_stmt
    movq %rax, AST_OFF_RIGHT_PTR(%r13)
.L_sif_done:
    movq %r12, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_while:
    call next_tok
    movq $AST_WHILE, %rdi
    call parser_create_node
    movq %rax, %r12
    call parse_expr
    movq %rax, AST_OFF_CHILD_PTR(%r12)
    call parse_block
    movq %rax, AST_OFF_RIGHT_PTR(%r12)
    movq %r12, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_for:
    # for <ident> in <expr>..<expr> { ... }
    call next_tok
    movq $AST_FOR, %rdi
    call parser_create_node
    movq %rax, %r12
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_sl_err
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r12)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r12)
    call next_tok

    cmpq $TOK_IN, cur_tok_type(%rip)
    jne .L_sl_err
    call next_tok

    call parse_expr
    movq %rax, %r13          # start or collection expr
    cmpq $TOK_DOTDOT, cur_tok_type(%rip)
    jne .L_for_collection
    call next_tok
    call parse_expr
    movq %rax, %r14          # end expr

    movq $AST_EXPR_BIN, %rdi
    call parser_create_node
    movq $TOK_DOTDOT, AST_OFF_VAL(%rax)
    movq %r13, AST_OFF_CHILD_PTR(%rax)
    movq %r14, AST_OFF_RIGHT_PTR(%rax)
    movq %rax, AST_OFF_CHILD_PTR(%r12)
    jmp .L_for_body

.L_for_collection:
    movq %r13, AST_OFF_CHILD_PTR(%r12)

.L_for_body:
    call parse_block
    movq %rax, AST_OFF_RIGHT_PTR(%r12)
    movq %r12, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_receive:
    call next_tok
    call skip_parens
    call parse_block
    movq $AST_BLOCK, %rdi
    movq %rax, %r13
    call parser_create_node
    movq %r13, AST_OFF_CHILD_PTR(%rax)
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_spawn:
    call next_tok
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_ss_err
    movq $AST_IDENT, %rdi
    call parser_create_node
    movq %rax, %r13
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r13)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r13)
    call next_tok
    xorq %r14, %r14
    cmpq $TOK_LPAREN, cur_tok_type(%rip)
    jne .L_ss_semi
    call next_tok
    cmpq $TOK_RPAREN, cur_tok_type(%rip)
    je .L_ss_rparen
    call parse_expr
    movq %rax, %r14
.L_ss_rparen:
    cmpq $TOK_RPAREN, cur_tok_type(%rip)
    jne .L_ss_semi
    call next_tok
.L_ss_semi:
    cmpq $TOK_SEMICOLON, cur_tok_type(%rip)
    jne .L_ss_mk
    call next_tok
.L_ss_mk:
    movq $AST_CALL, %rdi
    call parser_create_node
    leaq kw_axiom_spawn(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%rax)
    movq $13, AST_OFF_NAME_LEN(%rax)
    movq %r13, AST_OFF_CHILD_PTR(%rax)
    movq %r14, AST_OFF_RIGHT_PTR(%rax)
    popq %r14
    popq %r13
    popq %r12
    leave
    ret
.L_ss_err:
    movq $1, p_has_error(%rip)
    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_ident:
    movq p_src_ptr(%rip), %r12
    addq cur_tok_start(%rip), %r12
    movq cur_tok_len(%rip), %r13
    call next_tok

    cmpq $TOK_LPAREN, cur_tok_type(%rip)
    je .L_st_call
    cmpq $TOK_ASSIGN, cur_tok_type(%rip)
    je .L_st_assign
    cmpq $TOK_LBRACKET, cur_tok_type(%rip)
    je .L_st_idx_assign
    cmpq $TOK_DOT, cur_tok_type(%rip)
    je .L_st_fld_assign
    cmpq $TOK_PIPE, cur_tok_type(%rip)
    je .L_st_pipe_ident

    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_idx_assign:
    call next_tok            # skip '['
    call parse_expr
    pushq %rax               # index expr
    cmpq $TOK_RBRACKET, cur_tok_type(%rip)
    jne .L_sia_eq
    call next_tok
.L_sia_eq:
    cmpq $TOK_ASSIGN, cur_tok_type(%rip)
    jne .L_sia_err
    call next_tok
    movq $AST_INDEX, %rdi
    call parser_create_node
    movq %r12, AST_OFF_NAME_PTR(%rax)
    movq %r13, AST_OFF_NAME_LEN(%rax)
    popq %rcx                # index expr
    movq %rcx, AST_OFF_RIGHT_PTR(%rax)
    pushq %rax               # target AST_INDEX
    movq $AST_ASSIGN, %rdi
    call parser_create_node
    movq %rax, %r14
    movq %r12, AST_OFF_NAME_PTR(%r14)
    movq %r13, AST_OFF_NAME_LEN(%r14)
    popq %rcx
    movq %rcx, AST_OFF_RIGHT_PTR(%r14)
    call parse_expr
    movq %rax, AST_OFF_CHILD_PTR(%r14)
    jmp .L_sa_semi
.L_sia_err:
    popq %rax
    jmp .L_sl_err

.L_st_fld_assign:
    call next_tok            # skip '.'
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_sl_err
    movq p_src_ptr(%rip), %rax
    addq cur_tok_start(%rip), %rax
    pushq %rax               # field ptr
    pushq cur_tok_len(%rip)  # field len
    call next_tok
    cmpq $TOK_ASSIGN, cur_tok_type(%rip)
    jne .L_sfa_err
    call next_tok
    movq $AST_FIELD, %rdi
    call parser_create_node
    popq %rcx                # field len
    popq %rdx                # field ptr
    movq %rdx, AST_OFF_NAME_PTR(%rax)
    movq %rcx, AST_OFF_NAME_LEN(%rax)
    pushq %rax               # target AST_FIELD
    movq $AST_ASSIGN, %rdi
    call parser_create_node
    movq %rax, %r14
    movq %r12, AST_OFF_NAME_PTR(%r14)
    movq %r13, AST_OFF_NAME_LEN(%r14)
    popq %rcx
    movq %rcx, AST_OFF_RIGHT_PTR(%r14)
    call parse_expr
    movq %rax, AST_OFF_CHILD_PTR(%r14)
    jmp .L_sa_semi
.L_sfa_err:
    addq $16, %rsp
    jmp .L_sl_err

.L_st_pipe_ident:
    movq $AST_IDENT, %rdi
    call parser_create_node
    movq %r12, AST_OFF_NAME_PTR(%rax)
    movq %r13, AST_OFF_NAME_LEN(%rax)
    movq %rax, %rdi
    call parse_pipe_tail
    movq %rax, %r14
    cmpq $TOK_SEMICOLON, cur_tok_type(%rip)
    jne .L_sc_ret
    call next_tok
    jmp .L_sc_ret

.L_st_assign:
    call next_tok
    movq $AST_ASSIGN, %rdi
    call parser_create_node
    movq %rax, %r14
    movq %r12, AST_OFF_NAME_PTR(%r14)
    movq %r13, AST_OFF_NAME_LEN(%r14)
    call parse_expr
    movq %rax, AST_OFF_CHILD_PTR(%r14)
.L_sa_semi:
    cmpq $TOK_SEMICOLON, cur_tok_type(%rip)
    jne .L_sa_ret
    call next_tok
.L_sa_ret:
    movq %r14, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_st_call:
    movq %r12, %rdi
    movq %r13, %rsi
    call parse_call_tail
    movq %rax, %r14
    cmpq $TOK_SEMICOLON, cur_tok_type(%rip)
    jne .L_sc_ret
    call next_tok
.L_sc_ret:
    movq %r14, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

# parse_call_tail(name_ptr: %rdi, name_len: %rsi) -> %rax (AST_CALL)
# Expects cur_tok_type == TOK_LPAREN
parse_call_tail:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    movq %rdi, %r12
    movq %rsi, %r13
    call next_tok            # skip '('

    movq $AST_CALL, %rdi
    call parser_create_node
    movq %rax, %r14
    movq %r12, AST_OFF_NAME_PTR(%r14)
    movq %r13, AST_OFF_NAME_LEN(%r14)

    xorq %r15, %r15          # arg head
    xorq %r12, %r12          # arg tail

    cmpq $TOK_RPAREN, cur_tok_type(%rip)
    je .L_pct_close

.L_pct_loop:
    call parse_expr
    testq %rax, %rax
    jz .L_pct_close
    testq %r15, %r15
    jnz .L_pct_app
    movq %rax, %r15
    movq %rax, %r12
    jmp .L_pct_comma
.L_pct_app:
    movq %rax, AST_OFF_NEXT_PTR(%r12)
    movq %rax, %r12
.L_pct_comma:
    cmpq $TOK_COMMA, cur_tok_type(%rip)
    jne .L_pct_close
    call next_tok
    jmp .L_pct_loop

.L_pct_close:
    cmpq $TOK_RPAREN, cur_tok_type(%rip)
    jne .L_pct_done
    call next_tok
.L_pct_done:
    movq %r15, AST_OFF_CHILD_PTR(%r14)
    movq %r14, %rax
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# Expression parser:
# parse_expr -> parse_or -> parse_and -> parse_cmp -> parse_term -> parse_factor -> parse_unary -> parse_primary
# ==============================================================================
parse_expr:
    jmp parse_pipe

parse_pipe:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    call parse_or
    movq %rax, %rdi
    call parse_pipe_tail
    popq %r12
    leave
    ret

# parse_pipe_tail(lhs: %rdi) -> %rax
parse_pipe_tail:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14
    movq %rdi, %r12          # current lhs
.L_ppt_loop:
    cmpq $TOK_PIPE, cur_tok_type(%rip)
    jne .L_ppt_done
    call next_tok
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_ppt_err
    movq p_src_ptr(%rip), %r13
    addq cur_tok_start(%rip), %r13
    movq cur_tok_len(%rip), %r14
    call next_tok
    cmpq $TOK_LPAREN, cur_tok_type(%rip)
    je .L_ppt_args
    # Bare fn: lhs |> fn_name -> AST_CALL(fn_name, [lhs])
    movq $AST_CALL, %rdi
    call parser_create_node
    movq %r13, AST_OFF_NAME_PTR(%rax)
    movq %r14, AST_OFF_NAME_LEN(%rax)
    movq %r12, AST_OFF_CHILD_PTR(%rax)
    movq %rax, %r12
    jmp .L_ppt_loop
.L_ppt_args:
    movq %r13, %rdi
    movq %r14, %rsi
    call parse_call_tail
    # Prepend lhs (%r12) as first arg
    movq AST_OFF_CHILD_PTR(%rax), %rcx
    movq %rcx, AST_OFF_NEXT_PTR(%r12)
    movq %r12, AST_OFF_CHILD_PTR(%rax)
    movq %rax, %r12
    jmp .L_ppt_loop
.L_ppt_err:
    movq $1, p_has_error(%rip)
.L_ppt_done:
    movq %r12, %rax
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

parse_or:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    call parse_and
    movq %rax, %r12
.L_por_loop:
    cmpq $TOK_OR, cur_tok_type(%rip)
    jne .L_por_done
    movq $TOK_OR, %r13
    call next_tok
    call parse_and
    pushq %rax
    movq $AST_EXPR_BIN, %rdi
    call parser_create_node
    popq %rcx
    movq %r12, AST_OFF_CHILD_PTR(%rax)
    movq %rcx, AST_OFF_RIGHT_PTR(%rax)
    movq %r13, AST_OFF_VAL(%rax)
    movq %rax, %r12
    jmp .L_por_loop
.L_por_done:
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

parse_and:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    call parse_cmp
    movq %rax, %r12
.L_pand_loop:
    cmpq $TOK_AND, cur_tok_type(%rip)
    jne .L_pand_done
    movq $TOK_AND, %r13
    call next_tok
    call parse_cmp
    pushq %rax
    movq $AST_EXPR_BIN, %rdi
    call parser_create_node
    popq %rcx
    movq %r12, AST_OFF_CHILD_PTR(%rax)
    movq %rcx, AST_OFF_RIGHT_PTR(%rax)
    movq %r13, AST_OFF_VAL(%rax)
    movq %rax, %r12
    jmp .L_pand_loop
.L_pand_done:
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

parse_cmp:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    call parse_term
    movq %rax, %r12
.L_pcmp_loop:
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_EQ, %rax
    je .L_pcmp_op
    cmpq $TOK_NEQ, %rax
    je .L_pcmp_op
    cmpq $TOK_LT, %rax
    je .L_pcmp_op
    cmpq $TOK_GT, %rax
    je .L_pcmp_op
    cmpq $TOK_LE, %rax
    je .L_pcmp_op
    cmpq $TOK_GE, %rax
    je .L_pcmp_op
    jmp .L_pcmp_done
.L_pcmp_op:
    movq %rax, %r13
    call next_tok
    call parse_term
    pushq %rax
    movq $AST_EXPR_BIN, %rdi
    call parser_create_node
    popq %rcx
    movq %r12, AST_OFF_CHILD_PTR(%rax)
    movq %rcx, AST_OFF_RIGHT_PTR(%rax)
    movq %r13, AST_OFF_VAL(%rax)
    movq %rax, %r12
    jmp .L_pcmp_loop
.L_pcmp_done:
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

parse_term:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    call parse_factor
    movq %rax, %r12
.L_pt_loop:
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_PLUS, %rax
    je .L_pt_op
    cmpq $TOK_MINUS, %rax
    je .L_pt_op
    jmp .L_pt_done
.L_pt_op:
    movq %rax, %r13
    call next_tok
    call parse_factor
    pushq %rax
    movq $AST_EXPR_BIN, %rdi
    call parser_create_node
    popq %rcx
    movq %r12, AST_OFF_CHILD_PTR(%rax)
    movq %rcx, AST_OFF_RIGHT_PTR(%rax)
    movq %r13, AST_OFF_VAL(%rax)
    movq %rax, %r12
    jmp .L_pt_loop
.L_pt_done:
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

parse_factor:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    call parse_unary
    movq %rax, %r12
.L_pf_loop:
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_ASTERISK, %rax
    je .L_pf_op
    cmpq $TOK_SLASH, %rax
    je .L_pf_op
    jmp .L_pf_done
.L_pf_op:
    movq %rax, %r13
    call next_tok
    call parse_unary
    pushq %rax
    movq $AST_EXPR_BIN, %rdi
    call parser_create_node
    popq %rcx
    movq %r12, AST_OFF_CHILD_PTR(%rax)
    movq %rcx, AST_OFF_RIGHT_PTR(%rax)
    movq %r13, AST_OFF_VAL(%rax)
    movq %rax, %r12
    jmp .L_pf_loop
.L_pf_done:
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

parse_unary:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    movq cur_tok_type(%rip), %rax
    cmpq $TOK_MINUS, %rax
    je .L_pu_op
    cmpq $TOK_BANG, %rax
    je .L_pu_op
    call parse_primary
    popq %r12
    leave
    ret
.L_pu_op:
    movq %rax, %r12
    call next_tok
    call parse_unary
    pushq %rax
    movq $AST_UNARY, %rdi
    call parser_create_node
    popq %rcx
    movq %rcx, AST_OFF_CHILD_PTR(%rax)
    movq %r12, AST_OFF_VAL(%rax)
    popq %r12
    leave
    ret

parse_primary:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13

    movq cur_tok_type(%rip), %rax
    cmpq $TOK_INT, %rax
    je .L_pp_int
    cmpq $TOK_TRUE, %rax
    je .L_pp_true
    cmpq $TOK_FALSE, %rax
    je .L_pp_false
    cmpq $TOK_STRING, %rax
    je .L_pp_str
    cmpq $TOK_IDENT, %rax
    je .L_pp_ident
    cmpq $TOK_LBRACKET, %rax
    je .L_pp_array
    cmpq $TOK_LBRACE, %rax
    je .L_pp_struct_lit
    cmpq $TOK_LPAREN, %rax
    je .L_pp_paren

    movq $1, p_has_error(%rip)
    xorq %rax, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_pp_array:
    call next_tok            # skip '['
    movq $AST_ARRAY, %rdi
    call parser_create_node
    movq %rax, %r12          # array node
    xorq %r13, %r13          # head
    xorq %r14, %r14          # tail
    xorq %r15, %r15          # count
    pushq %r14
    pushq %r15

    cmpq $TOK_RBRACKET, cur_tok_type(%rip)
    je .L_ppa_close
.L_ppa_loop:
    call parse_expr
    testq %rax, %rax
    jz .L_ppa_close
    incq (%rsp)              # count++
    testq %r13, %r13
    jnz .L_ppa_app
    movq %rax, %r13
    movq %rax, 8(%rsp)       # tail = node
    jmp .L_ppa_comma
.L_ppa_app:
    movq 8(%rsp), %rcx
    movq %rax, AST_OFF_NEXT_PTR(%rcx)
    movq %rax, 8(%rsp)
.L_ppa_comma:
    cmpq $TOK_COMMA, cur_tok_type(%rip)
    jne .L_ppa_close
    call next_tok
    jmp .L_ppa_loop
.L_ppa_close:
    cmpq $TOK_RBRACKET, cur_tok_type(%rip)
    jne .L_ppa_done
    call next_tok
.L_ppa_done:
    popq %r15                # count
    popq %r14
    movq %r13, AST_OFF_CHILD_PTR(%r12)
    movq %r15, AST_OFF_VAL(%r12)
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_pp_struct_lit:
    call next_tok            # skip '{'
    movq $AST_STRUCT_LIT, %rdi
    call parser_create_node
    movq %rax, %r12          # struct_lit node
    xorq %r13, %r13          # head
    pushq $0                 # tail [8(%rsp)]
    pushq $0                 # count [0(%rsp)]

    cmpq $TOK_RBRACE, cur_tok_type(%rip)
    je .L_psl_close
.L_psl_loop:
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_psl_close
    movq $AST_PARAM, %rdi
    call parser_create_node
    movq %rax, %r14
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r14)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r14)
    call next_tok
    cmpq $TOK_COLON, cur_tok_type(%rip)
    jne .L_psl_close
    call next_tok
    call parse_expr
    movq %rax, AST_OFF_CHILD_PTR(%r14)
    incq (%rsp)
    testq %r13, %r13
    jnz .L_psl_app
    movq %r14, %r13
    movq %r14, 8(%rsp)
    jmp .L_psl_comma
.L_psl_app:
    movq 8(%rsp), %rcx
    movq %r14, AST_OFF_NEXT_PTR(%rcx)
    movq %r14, 8(%rsp)
.L_psl_comma:
    cmpq $TOK_COMMA, cur_tok_type(%rip)
    jne .L_psl_close
    call next_tok
    jmp .L_psl_loop
.L_psl_close:
    cmpq $TOK_RBRACE, cur_tok_type(%rip)
    jne .L_psl_done
    call next_tok
.L_psl_done:
    popq %r15
    addq $8, %rsp
    movq %r13, AST_OFF_CHILD_PTR(%r12)
    movq %r15, AST_OFF_VAL(%r12)
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_pp_true:
    call next_tok
    movq $AST_INT, %rdi
    call parser_create_node
    movq $1, AST_OFF_VAL(%rax)
    popq %r13
    popq %r12
    leave
    ret

.L_pp_false:
    call next_tok
    movq $AST_INT, %rdi
    call parser_create_node
    movq $0, AST_OFF_VAL(%rax)
    popq %r13
    popq %r12
    leave
    ret

.L_pp_int:
    movq p_src_ptr(%rip), %rsi
    addq cur_tok_start(%rip), %rsi
    movq cur_tok_len(%rip), %rcx
    xorq %rax, %rax
    xorq %rdx, %rdx
.L_atoi:
    testq %rcx, %rcx
    jz .L_atoi_done
    imulq $10, %rax
    movzbq (%rsi, %rdx, 1), %r8
    subq $'0', %r8
    addq %r8, %rax
    incq %rdx
    decq %rcx
    jmp .L_atoi
.L_atoi_done:
    movq %rax, %r12
    call next_tok
    movq $AST_INT, %rdi
    call parser_create_node
    movq %r12, AST_OFF_VAL(%rax)
    popq %r13
    popq %r12
    leave
    ret

.L_pp_str:
    movq $AST_STRING, %rdi
    call parser_create_node
    movq %rax, %r12
    movq p_src_ptr(%rip), %rcx
    addq cur_tok_start(%rip), %rcx
    movq %rcx, AST_OFF_NAME_PTR(%r12)
    movq cur_tok_len(%rip), %rcx
    movq %rcx, AST_OFF_NAME_LEN(%r12)
    call next_tok
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_pp_ident:
    movq p_src_ptr(%rip), %r12
    addq cur_tok_start(%rip), %r12
    movq cur_tok_len(%rip), %r13
    call next_tok
    cmpq $TOK_LPAREN, cur_tok_type(%rip)
    jne .L_pp_ident_var
    movq %r12, %rdi
    movq %r13, %rsi
    call parse_call_tail
    popq %r13
    popq %r12
    leave
    ret
.L_pp_ident_var:
    movq $AST_IDENT, %rdi
    call parser_create_node
    movq %r12, AST_OFF_NAME_PTR(%rax)
    movq %r13, AST_OFF_NAME_LEN(%rax)
.L_pp_postfix:
    cmpq $TOK_LBRACKET, cur_tok_type(%rip)
    je .L_pp_do_idx
    cmpq $TOK_DOT, cur_tok_type(%rip)
    je .L_pp_do_field
    jmp .L_pp_iv_done
.L_pp_do_idx:
    pushq %rax
    call next_tok            # skip '['
    call parse_expr
    pushq %rax               # index expr
    cmpq $TOK_RBRACKET, cur_tok_type(%rip)
    jne .L_pp_idx_mk
    call next_tok
.L_pp_idx_mk:
    movq $AST_INDEX, %rdi
    call parser_create_node
    popq %rcx                # index expr
    popq %rdx                # base
    movq %rdx, AST_OFF_CHILD_PTR(%rax)
    movq %rcx, AST_OFF_RIGHT_PTR(%rax)
    jmp .L_pp_postfix
.L_pp_do_field:
    pushq %rax
    call next_tok            # skip '.'
    cmpq $TOK_IDENT, cur_tok_type(%rip)
    jne .L_pp_field_err
    movq p_src_ptr(%rip), %r12
    addq cur_tok_start(%rip), %r12
    movq cur_tok_len(%rip), %r13
    call next_tok
    movq $AST_FIELD, %rdi
    call parser_create_node
    popq %rdx                # base
    movq %rdx, AST_OFF_CHILD_PTR(%rax)
    movq %r12, AST_OFF_NAME_PTR(%rax)
    movq %r13, AST_OFF_NAME_LEN(%rax)
    jmp .L_pp_postfix
.L_pp_field_err:
    popq %rax
    movq $1, p_has_error(%rip)
.L_pp_iv_done:
    popq %r13
    popq %r12
    leave
    ret

.L_pp_paren:
    call next_tok
    call parse_expr
    movq %rax, %r12
    cmpq $TOK_RPAREN, cur_tok_type(%rip)
    jne .L_pp_paren_ret
    call next_tok
.L_pp_paren_ret:
    movq %r12, %rax
    popq %r13
    popq %r12
    leave
    ret
