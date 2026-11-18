.include "src_asm/include/axiom_defs.inc"

.section .rodata
kw_require: .ascii "require"
kw_cap:     .ascii "cap"
kw_intent:  .ascii "intent"
kw_let:     .ascii "let"
kw_mut:     .ascii "mut"
kw_fn:      .ascii "fn"
kw_return:  .ascii "return"
kw_if:      .ascii "if"
kw_else:    .ascii "else"
kw_for:     .ascii "for"
kw_in:      .ascii "in"
kw_true:    .ascii "true"
kw_false:   .ascii "false"
kw_struct:  .ascii "struct"
kw_actor:   .ascii "actor"
kw_spawn:   .ascii "spawn"
kw_send:    .ascii "send"
kw_receive: .ascii "receive"
kw_while:   .ascii "while"

# Keyword table: ptr, len, tok_id
.align 8
kw_table:
    .quad kw_require, 7, TOK_REQUIRE
    .quad kw_cap,     3, TOK_CAP
    .quad kw_intent,  6, TOK_INTENT
    .quad kw_let,     3, TOK_LET
    .quad kw_mut,     3, TOK_MUT
    .quad kw_fn,      2, TOK_FN
    .quad kw_return,  6, TOK_RETURN
    .quad kw_if,      2, TOK_IF
    .quad kw_else,    4, TOK_ELSE
    .quad kw_for,     3, TOK_FOR
    .quad kw_in,      2, TOK_IN
    .quad kw_true,    4, TOK_TRUE
    .quad kw_false,   5, TOK_FALSE
    .quad kw_struct,  6, TOK_STRUCT
    .quad kw_actor,   5, TOK_ACTOR
    .quad kw_spawn,   5, TOK_SPAWN
    .quad kw_send,    4, TOK_SEND
    .quad kw_receive, 7, TOK_RECEIVE
    .quad kw_while,   5, TOK_WHILE
    .quad 0, 0, 0

.section .text
.global lexer_next_token
.global lexer_skip_whitespace

# ==============================================================================
# lexer_skip_whitespace(src_ptr: %rdi, src_len: %rsi, pos: %rdx) -> %rax (new_pos)
# ==============================================================================
lexer_skip_whitespace:
    cmpq %rsi, %rdx
    jge .L_sw_done
    movzbq (%rdi, %rdx, 1), %rax
    cmpb $' ', %al
    je .L_sw_next
    cmpb $'\t', %al
    je .L_sw_next
    cmpb $'\r', %al
    je .L_sw_next
    cmpb $'\n', %al
    je .L_sw_next
.L_sw_done:
    movq %rdx, %rax
    ret
.L_sw_next:
    incq %rdx
    jmp lexer_skip_whitespace

# ==============================================================================
# lexer_next_token(src_ptr: %rdi, src_len: %rsi, pos: %rdx) -> (%rax: tok_type, %rdx: new_pos, %r8: tok_start, %r9: tok_len)
# Pure x86_64 zero-copy token scanner
# ==============================================================================
lexer_next_token:
    pushq %rbp
    movq %rsp, %rbp
    pushq %rbx
    pushq %r12
    pushq %r13

    movq %rdi, %r12          # buffer
    movq %rsi, %r13          # len

skip_spaces:
    cmpq %r13, %rdx
    jge tok_is_eof
    movzbq (%r12, %rdx, 1), %rax

    # Skip ' ', '\t', '\r', '\n'
    cmpb $' ', %al
    je next_char
    cmpb $'\t', %al
    je next_char
    cmpb $'\r', %al
    je next_char
    cmpb $'\n', %al
    je next_char

    # Skip comments starting with '#' or "//"
    cmpb $'#', %al
    je skip_line_comment
    cmpb $'/', %al
    jne not_comment
    leaq 1(%rdx), %rbx
    cmpq %r13, %rbx
    jge not_comment
    cmpb $'/', (%r12, %rbx, 1)
    je skip_line_comment

not_comment:
    # Character classification
    movq %rdx, %r8           # tok_start = pos

    # Strings: "..."
    cmpb $'"', %al
    je scan_string

    # Identifiers / Keywords: [a-zA-Z_]
    cmpb $'a', %al
    jl check_upper
    cmpb $'z', %al
    jle scan_ident
check_upper:
    cmpb $'A', %al
    jl check_under
    cmpb $'Z', %al
    jle scan_ident
check_under:
    cmpb $'_', %al
    je scan_ident

    # Numbers: [0-9]
    cmpb $'0', %al
    jl scan_delimiter
    cmpb $'9', %al
    jle scan_number

scan_delimiter:
    incq %rdx
    movq $1, %r9

    # Two-char checks ('==', '=>', '!=', '<=', '>=', '..', '&&', '||')
    cmpb $'=', %al
    jne .L_chk_bang
    cmpq %r13, %rdx
    jge is_assign
    movzbq (%r12, %rdx, 1), %rbx
    cmpb $'=', %bl
    je .L_is_eq
    cmpb $'>', %bl
    je .L_is_arrow
    jmp is_assign

.L_chk_bang:
    cmpb $'!', %al
    jne .L_chk_lt
    cmpq %r13, %rdx
    jge is_bang
    cmpb $'=', (%r12, %rdx, 1)
    jne is_bang
    incq %rdx
    movq $2, %r9
    movq $TOK_NEQ, %rax
    jmp finish_tok

.L_chk_lt:
    cmpb $'<', %al
    jne .L_chk_gt
    cmpq %r13, %rdx
    jge is_lt
    cmpb $'=', (%r12, %rdx, 1)
    jne is_lt
    incq %rdx
    movq $2, %r9
    movq $TOK_LE, %rax
    jmp finish_tok

.L_chk_gt:
    cmpb $'>', %al
    jne .L_chk_dot
    cmpq %r13, %rdx
    jge is_gt
    cmpb $'=', (%r12, %rdx, 1)
    jne is_gt
    incq %rdx
    movq $2, %r9
    movq $TOK_GE, %rax
    jmp finish_tok

.L_chk_dot:
    cmpb $'.', %al
    jne .L_chk_and
    cmpq %r13, %rdx
    jge is_dot
    cmpb $'.', (%r12, %rdx, 1)
    jne is_dot
    incq %rdx
    movq $2, %r9
    movq $TOK_DOTDOT, %rax
    jmp finish_tok

.L_chk_and:
    cmpb $'&', %al
    jne .L_chk_or
    cmpq %r13, %rdx
    jge .L_chk_single
    cmpb $'&', (%r12, %rdx, 1)
    jne .L_chk_single
    incq %rdx
    movq $2, %r9
    movq $TOK_AND, %rax
    jmp finish_tok

.L_chk_or:
    cmpb $'|', %al
    jne .L_chk_single
    cmpq %r13, %rdx
    jge .L_chk_single
    movzbq (%r12, %rdx, 1), %rbx
    cmpb $'|', %bl
    je .L_is_or
    cmpb $'>', %bl
    je .L_is_pipe
    jmp .L_chk_single

.L_is_or:
    incq %rdx
    movq $2, %r9
    movq $TOK_OR, %rax
    jmp finish_tok

.L_is_pipe:
    incq %rdx
    movq $2, %r9
    movq $TOK_PIPE, %rax
    jmp finish_tok

.L_is_eq:
    incq %rdx
    movq $2, %r9
    movq $TOK_EQ, %rax
    jmp finish_tok

.L_is_arrow:
    incq %rdx
    movq $2, %r9
    movq $TOK_ARROW, %rax
    jmp finish_tok

.L_chk_single:
    cmpb $'(', %al
    je is_lparen
    cmpb $')', %al
    je is_rparen
    cmpb $'{', %al
    je is_lbrace
    cmpb $'}', %al
    je is_rbrace
    cmpb $'[', %al
    je is_lbracket
    cmpb $']', %al
    je is_rbracket
    cmpb $';', %al
    je is_semicolon
    cmpb $':', %al
    je is_colon
    cmpb $',', %al
    je is_comma
    cmpb $'+', %al
    je is_plus
    cmpb $'-', %al
    je is_minus
    cmpb $'*', %al
    je is_asterisk
    cmpb $'/', %al
    je is_slash
    cmpb $'!', %al
    je is_bang
    cmpb $'<', %al
    je is_lt
    cmpb $'>', %al
    je is_gt

    movq $TOK_ILLEGAL, %rax
    jmp finish_tok

is_lparen:   movq $TOK_LPAREN, %rax; jmp finish_tok
is_rparen:   movq $TOK_RPAREN, %rax; jmp finish_tok
is_lbrace:   movq $TOK_LBRACE, %rax; jmp finish_tok
is_rbrace:   movq $TOK_RBRACE, %rax; jmp finish_tok
is_lbracket: movq $TOK_LBRACKET, %rax; jmp finish_tok
is_rbracket: movq $TOK_RBRACKET, %rax; jmp finish_tok
is_semicolon:movq $TOK_SEMICOLON, %rax; jmp finish_tok
is_colon:    movq $TOK_COLON, %rax; jmp finish_tok
is_comma:    movq $TOK_COMMA, %rax; jmp finish_tok
is_assign:   movq $TOK_ASSIGN, %rax; jmp finish_tok
is_plus:     movq $TOK_PLUS, %rax; jmp finish_tok
is_minus:    movq $TOK_MINUS, %rax; jmp finish_tok
is_asterisk: movq $TOK_ASTERISK, %rax; jmp finish_tok
is_slash:    movq $TOK_SLASH, %rax; jmp finish_tok
is_bang:     movq $TOK_BANG, %rax; jmp finish_tok
is_lt:       movq $TOK_LT, %rax; jmp finish_tok
is_gt:       movq $TOK_GT, %rax; jmp finish_tok
is_dot:      movq $TOK_DOT, %rax; jmp finish_tok

scan_string:
    incq %rdx                # skip opening quote
    movq %rdx, %r8           # start of contents
str_loop:
    cmpq %r13, %rdx
    jge str_end
    movzbq (%r12, %rdx, 1), %rax
    cmpb $'\\', %al
    jne .L_chk_quote
    # skip backslash + escaped char
    addq $2, %rdx
    jmp str_loop
.L_chk_quote:
    cmpb $'"', %al
    je str_close
    incq %rdx
    jmp str_loop
str_close:
    movq %rdx, %r9
    subq %r8, %r9            # len
    incq %rdx                # skip closing quote
    movq $TOK_STRING, %rax
    jmp finish_tok
str_end:
    movq %rdx, %r9
    subq %r8, %r9
    movq $TOK_STRING, %rax
    jmp finish_tok

scan_ident:
    incq %rdx
    cmpq %r13, %rdx
    jge ident_done
    movzbq (%r12, %rdx, 1), %rax
    # [a-zA-Z0-9_]
    cmpb $'a', %al
    jl check_u2
    cmpb $'z', %al
    jle scan_ident
check_u2:
    cmpb $'A', %al
    jl check_num2
    cmpb $'Z', %al
    jle scan_ident
check_num2:
    cmpb $'0', %al
    jl check_un2
    cmpb $'9', %al
    jle scan_ident
check_un2:
    cmpb $'_', %al
    je scan_ident

ident_done:
    movq %rdx, %r9
    subq %r8, %r9            # length

    # Match against keyword table
    leaq kw_table(%rip), %rbx
.L_kw_loop:
    movq (%rbx), %rax        # kw_ptr
    testq %rax, %rax
    jz .L_is_ident
    cmpq 8(%rbx), %r9        # compare length
    jne .L_kw_next

    # Compare %r9 bytes: (%r12 + %r8) vs (%rax)
    xorq %rcx, %rcx
.L_kw_cmp:
    cmpq %r9, %rcx
    jge .L_kw_match
    movb (%r12, %r8, 1), %r10b   # wait: need index r8+rcx
    movq %r8, %r11
    addq %rcx, %r11
    movb (%r12, %r11, 1), %r10b
    cmpb (%rax, %rcx, 1), %r10b
    jne .L_kw_next
    incq %rcx
    jmp .L_kw_cmp

.L_kw_match:
    movq 16(%rbx), %rax      # tok_id
    jmp finish_tok

.L_kw_next:
    addq $24, %rbx
    jmp .L_kw_loop

.L_is_ident:
    movq $TOK_IDENT, %rax
    jmp finish_tok

scan_number:
    incq %rdx
    cmpq %r13, %rdx
    jge num_done
    movzbq (%r12, %rdx, 1), %rax
    cmpb $'0', %al
    jl num_done
    cmpb $'9', %al
    jle scan_number
num_done:
    movq %rdx, %r9
    subq %r8, %r9
    movq $TOK_INT, %rax
    jmp finish_tok

skip_line_comment:
    incq %rdx
    cmpq %r13, %rdx
    jge tok_is_eof
    cmpb $'\n', (%r12, %rdx, 1)
    je next_char
    jmp skip_line_comment

next_char:
    incq %rdx
    jmp skip_spaces

tok_is_eof:
    movq $TOK_EOF, %rax
    xorq %r8, %r8
    xorq %r9, %r9

finish_tok:
    popq %r13
    popq %r12
    popq %rbx
    leave
    ret
