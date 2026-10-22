.include "src_asm/include/axiom_defs.inc"

.section .text
.global lexer_next_token
.global lexer_skip_whitespace

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
    cmpb $',', %al
    je is_comma
    cmpb $'=', %al
    je is_assign
    cmpb $'+', %al
    je is_plus
    cmpb $'-', %al
    je is_minus

    movq $TOK_ILLEGAL, %rax
    jmp finish_tok

is_lparen:   movq $TOK_LPAREN, %rax; jmp finish_tok
is_rparen:   movq $TOK_RPAREN, %rax; jmp finish_tok
is_lbrace:   movq $TOK_LBRACE, %rax; jmp finish_tok
is_rbrace:   movq $TOK_RBRACE, %rax; jmp finish_tok
is_lbracket: movq $TOK_LBRACKET, %rax; jmp finish_tok
is_rbracket: movq $TOK_RBRACKET, %rax; jmp finish_tok
is_semicolon:movq $TOK_SEMICOLON, %rax; jmp finish_tok
is_comma:    movq $TOK_COMMA, %rax; jmp finish_tok
is_assign:   movq $TOK_ASSIGN, %rax; jmp finish_tok
is_plus:     movq $TOK_PLUS, %rax; jmp finish_tok
is_minus:    movq $TOK_MINUS, %rax; jmp finish_tok

scan_string:
    incq %rdx                # skip opening quote
    movq %rdx, %r8           # start of contents
str_loop:
    cmpq %r13, %rdx
    jge str_end
    movzbq (%r12, %rdx, 1), %rax
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
