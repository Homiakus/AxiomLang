.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/mca_memory.inc"

.section .data
lsp_init_res:
    .ascii "{\"capabilities\":{\"textDocumentSync\":1,\"hoverProvider\":true}}"
lsp_init_res_len = . - lsp_init_res

lsp_null_res:
    .ascii "null"
lsp_null_res_len = . - lsp_null_res

lsp_hover_res:
    .ascii "{\"contents\":{\"kind\":\"markdown\",\"value\":\"**AxiomLang** Pure x86_64 Symbol\"}}"
lsp_hover_res_len = . - lsp_hover_res

lsp_comp_res:
    .ascii "[{\"label\":\"fn\",\"kind\":14},{\"label\":\"let\",\"kind\":14},{\"label\":\"struct\",\"kind\":14},{\"label\":\"if\",\"kind\":14},{\"label\":\"while\",\"kind\":14},{\"label\":\"for\",\"kind\":14},{\"label\":\"return\",\"kind\":14},{\"label\":\"require\",\"kind\":14},{\"label\":\"println\",\"kind\":3},{\"label\":\"len\",\"kind\":3}]"
lsp_comp_res_len = . - lsp_comp_res

lsp_hdr_cl:   .ascii "Content-Length: "
lsp_hdr_crlf: .ascii "\r\n\r\n"
lsp_rpc_pre:  .ascii "{\"jsonrpc\":\"2.0\",\"id\":"
lsp_rpc_mid:  .ascii ",\"result\":"
lsp_rpc_end:  .ascii "}"
pat_id_key:   .ascii "\"id\":"

lsp_diag_ok_pre:
    .ascii "{\"jsonrpc\":\"2.0\",\"method\":\"textDocument/publishDiagnostics\",\"params\":{\"uri\":\""
lsp_diag_ok_pre_len = . - lsp_diag_ok_pre
lsp_diag_ok_suf:
    .ascii "\",\"diagnostics\":[]}}"
lsp_diag_ok_suf_len = . - lsp_diag_ok_suf
lsp_diag_err_suf:
    .ascii "\",\"diagnostics\":[{\"range\":{\"start\":{\"line\":0,\"character\":0},\"end\":{\"line\":0,\"character\":1}},\"severity\":1,\"message\":\"Static division by zero\"}]}}"
lsp_diag_err_suf_len = . - lsp_diag_err_suf
lsp_uri_def: .ascii "file:///main.axl"
lsp_uri_def_len = . - lsp_uri_def

pat_initialize: .ascii "\"initialize\""
pat_initialize_len = . - pat_initialize

pat_shutdown:   .ascii "\"shutdown\""
pat_shutdown_len = . - pat_shutdown

pat_exit:       .ascii "\"exit\""
pat_exit_len = . - pat_exit

pat_hover:      .ascii "\"textDocument/hover\""
pat_hover_len = . - pat_hover

pat_comp:       .ascii "\"textDocument/completion\""
pat_comp_len = . - pat_comp

pat_didopen:    .ascii "\"textDocument/didOpen\""
pat_didopen_len = . - pat_didopen

pat_didchg:     .ascii "\"textDocument/didChange\""
pat_didchg_len = . - pat_didchg

pat_div0:       .ascii "/ 0"
pat_div0_len = . - pat_div0

.section .bss
.align 16
lsp_in_buf:   .skip 65536
lsp_in_len:   .quad 0
lsp_out_buf:  .skip 65536
lsp_body_buf: .skip 65536
lsp_id_buf:   .skip 32
lsp_id_len:   .quad 0

.section .text
.global _start

_start:
    # Pure x86_64 Freestanding LSP Server with JSON-RPC method dispatch
lsp_read_loop:
    movq $SYS_READ, %rax
    movq $STDIN_FILENO, %rdi
    leaq lsp_in_buf(%rip), %rsi
    movq $65535, %rdx
    syscall

    cmpq $0, %rax
    jle lsp_exit
    movq %rax, lsp_in_len(%rip)

    # Check "exit" notification first
    leaq pat_exit(%rip), %rdi
    movq $pat_exit_len, %rsi
    call buf_contains
    testq %rax, %rax
    jnz lsp_exit

    # Check "initialize"
    leaq pat_initialize(%rip), %rdi
    movq $pat_initialize_len, %rsi
    call buf_contains
    testq %rax, %rax
    jnz .L_send_init

    # Check "shutdown"
    leaq pat_shutdown(%rip), %rdi
    movq $pat_shutdown_len, %rsi
    call buf_contains
    testq %rax, %rax
    jnz .L_send_shutdown

    # Check "textDocument/hover"
    leaq pat_hover(%rip), %rdi
    movq $pat_hover_len, %rsi
    call buf_contains
    testq %rax, %rax
    jnz .L_send_hover

    # Check "textDocument/completion"
    leaq pat_comp(%rip), %rdi
    movq $pat_comp_len, %rsi
    call buf_contains
    testq %rax, %rax
    jnz .L_send_comp

    # Check "textDocument/didOpen" or "textDocument/didChange"
    leaq pat_didopen(%rip), %rdi
    movq $pat_didopen_len, %rsi
    call buf_contains
    testq %rax, %rax
    jnz .L_send_diag

    leaq pat_didchg(%rip), %rdi
    movq $pat_didchg_len, %rsi
    call buf_contains
    testq %rax, %rax
    jnz .L_send_diag

    # Other notifications (initialized, didOpen, etc.): no reply per LSP spec
    jmp lsp_read_loop

.L_send_init:
    leaq lsp_init_res(%rip), %rdi
    movq $lsp_init_res_len, %rsi
    jmp .L_write_rpc

.L_send_shutdown:
    leaq lsp_null_res(%rip), %rdi
    movq $lsp_null_res_len, %rsi
    jmp .L_write_rpc

.L_send_hover:
    leaq lsp_hover_res(%rip), %rdi
    movq $lsp_hover_res_len, %rsi
    jmp .L_write_rpc

.L_send_comp:
    leaq lsp_comp_res(%rip), %rdi
    movq $lsp_comp_res_len, %rsi

.L_write_rpc:
    call send_jsonrpc_response
    jmp lsp_read_loop

.L_send_diag:
    call extract_json_uri    # -> %r12 = uri_ptr, %r13 = uri_len
    leaq lsp_body_buf(%rip), %rdi
    leaq lsp_diag_ok_pre(%rip), %rsi
    movq $lsp_diag_ok_pre_len, %rcx
    movq %rcx, %r14
    cld
    rep movsb
    movq %r12, %rsi
    movq %r13, %rcx
    addq %rcx, %r14
    rep movsb

    pushq %rdi
    leaq pat_div0(%rip), %rdi
    movq $pat_div0_len, %rsi
    call buf_contains
    popq %rdi
    testq %rax, %rax
    jnz .L_sd_err
    leaq lsp_diag_ok_suf(%rip), %rsi
    movq $lsp_diag_ok_suf_len, %rcx
    jmp .L_sd_emit
.L_sd_err:
    leaq lsp_diag_err_suf(%rip), %rsi
    movq $lsp_diag_err_suf_len, %rcx
.L_sd_emit:
    addq %rcx, %r14
    cld
    rep movsb

    # Frame lsp_body_buf (len=%r14) into lsp_out_buf with Content-Length
    leaq lsp_out_buf(%rip), %rdi
    xorq %r15, %r15
    leaq lsp_hdr_cl(%rip), %rsi
    movq $16, %rcx
    addq %rcx, %r15
    rep movsb
    subq $32, %rsp
    leaq 31(%rsp), %rsi
    movq %r14, %rax
    xorq %rcx, %rcx
.L_sd_div:
    xorq %rdx, %rdx
    movq $10, %rbx
    divq %rbx
    addb $'0', %dl
    decq %rsi
    movb %dl, (%rsi)
    incq %rcx
    testq %rax, %rax
    jnz .L_sd_div
    addq %rcx, %r15
    rep movsb
    addq $32, %rsp
    leaq lsp_hdr_crlf(%rip), %rsi
    movq $4, %rcx
    addq %rcx, %r15
    rep movsb
    leaq lsp_body_buf(%rip), %rsi
    movq %r14, %rcx
    addq %rcx, %r15
    rep movsb
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq lsp_out_buf(%rip), %rsi
    movq %r15, %rdx
    syscall
    jmp lsp_read_loop

lsp_exit:
    movq $SYS_EXIT, %rax
    xorq %rdi, %rdi
    syscall

# ==============================================================================
# buf_contains(pat: %rdi, pat_len: %rsi) -> %rax (1 if found in lsp_in_buf, 0 else)
# ==============================================================================
buf_contains:
    pushq %rbx
    pushq %r12
    movq lsp_in_len(%rip), %rcx
    cmpq %rsi, %rcx
    jl .L_bc_no
    subq %rsi, %rcx          # max start index
    leaq lsp_in_buf(%rip), %r12
    xorq %r8, %r8            # i = 0
.L_bc_outer:
    cmpq %rcx, %r8
    jg .L_bc_no
    xorq %r9, %r9            # j = 0
.L_bc_inner:
    cmpq %rsi, %r9
    jge .L_bc_yes
    movq %r8, %rbx
    addq %r9, %rbx
    movb (%r12, %rbx, 1), %al
    cmpb (%rdi, %r9, 1), %al
    jne .L_bc_next
    incq %r9
    jmp .L_bc_inner
.L_bc_next:
    incq %r8
    jmp .L_bc_outer
.L_bc_yes:
    movq $1, %rax
    popq %r12
    popq %rbx
    ret
.L_bc_no:
    xorq %rax, %rax
    popq %r12
    popq %rbx
    ret

# ==============================================================================
# send_jsonrpc_response(result_ptr: %rdi, result_len: %rsi)
# Extracts numeric "id":<n> from lsp_in_buf (defaults to "1") and writes framed JSON-RPC
# ==============================================================================
send_jsonrpc_response:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    movq %rdi, %r12          # result_ptr
    movq %rsi, %r13          # result_len

    # Extract "id" into lsp_id_buf
    call extract_json_id

    # Build JSON body into lsp_body_buf
    leaq lsp_body_buf(%rip), %rdi
    xorq %r14, %r14          # body_len

    leaq lsp_rpc_pre(%rip), %rsi
    movq $22, %rcx           # len("{\"jsonrpc\":\"2.0\",\"id\":")
    rep movsb
    addq $22, %r14

    leaq lsp_id_buf(%rip), %rsi
    movq lsp_id_len(%rip), %rcx
    addq %rcx, %r14
    rep movsb

    leaq lsp_rpc_mid(%rip), %rsi
    movq $10, %rcx           # len(",\"result\":")
    rep movsb
    addq $10, %r14

    movq %r12, %rsi
    movq %r13, %rcx
    addq %rcx, %r14
    rep movsb

    movb $'}', (%rdi)
    incq %r14

    # Build header + body into lsp_out_buf
    leaq lsp_out_buf(%rip), %rdi
    xorq %r15, %r15          # total_len

    leaq lsp_hdr_cl(%rip), %rsi
    movq $16, %rcx           # len("Content-Length: ")
    rep movsb
    addq $16, %r15

    # Convert %r14 (body_len) to ASCII
    subq $32, %rsp
    leaq 31(%rsp), %rsi
    movq %r14, %rax
    xorq %rcx, %rcx
.L_sjr_div:
    xorq %rdx, %rdx
    movq $10, %rbx
    divq %rbx
    addb $'0', %dl
    decq %rsi
    movb %dl, (%rsi)
    incq %rcx
    testq %rax, %rax
    jnz .L_sjr_div
    addq %rcx, %r15
    rep movsb
    addq $32, %rsp

    leaq lsp_hdr_crlf(%rip), %rsi
    movq $4, %rcx
    rep movsb
    addq $4, %r15

    leaq lsp_body_buf(%rip), %rsi
    movq %r14, %rcx
    addq %rcx, %r15
    rep movsb

    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq lsp_out_buf(%rip), %rsi
    movq %r15, %rdx
    syscall

    popq %r15
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

extract_json_id:
    leaq lsp_in_buf(%rip), %rsi
    movq lsp_in_len(%rip), %rcx
    xorq %r8, %r8
    cmpq $5, %rcx
    jl .L_eji_def
    subq $5, %rcx
.L_eji_scan:
    cmpq %rcx, %r8
    jg .L_eji_def
    cmpb $'"', (%rsi, %r8)
    jne .L_eji_next
    cmpb $'i', 1(%rsi, %r8)
    jne .L_eji_next
    cmpb $'d', 2(%rsi, %r8)
    jne .L_eji_next
    cmpb $'"', 3(%rsi, %r8)
    jne .L_eji_next
    cmpb $':', 4(%rsi, %r8)
    jne .L_eji_next
    addq $5, %r8
    addq $5, %rcx
.L_eji_ws:
    cmpq %rcx, %r8
    jge .L_eji_def
    cmpb $' ', (%rsi, %r8)
    jne .L_eji_digits
    incq %r8
    jmp .L_eji_ws
.L_eji_digits:
    leaq lsp_id_buf(%rip), %rdi
    xorq %r9, %r9
.L_eji_dloop:
    cmpq %rcx, %r8
    jge .L_eji_done
    cmpq $20, %r9
    jge .L_eji_done
    movb (%rsi, %r8), %al
    cmpb $'0', %al
    jl .L_eji_done
    cmpb $'9', %al
    jg .L_eji_done
    movb %al, (%rdi, %r9)
    incq %r9
    incq %r8
    jmp .L_eji_dloop
.L_eji_done:
    testq %r9, %r9
    jz .L_eji_def
    movq %r9, lsp_id_len(%rip)
    ret
.L_eji_next:
    incq %r8
    jmp .L_eji_scan
.L_eji_def:
    movb $'1', lsp_id_buf(%rip)
    movq $1, lsp_id_len(%rip)
    ret

extract_json_uri:
    leaq lsp_in_buf(%rip), %rsi
    movq lsp_in_len(%rip), %rcx
    xorq %r8, %r8
    cmpq $7, %rcx
    jl .L_eju_def
    subq $7, %rcx
.L_eju_scan:
    cmpq %rcx, %r8
    jg .L_eju_def
    cmpb $'"', (%rsi, %r8)
    jne .L_eju_next
    cmpb $'u', 1(%rsi, %r8)
    jne .L_eju_next
    cmpb $'r', 2(%rsi, %r8)
    jne .L_eju_next
    cmpb $'i', 3(%rsi, %r8)
    jne .L_eju_next
    cmpb $'"', 4(%rsi, %r8)
    jne .L_eju_next
    cmpb $':', 5(%rsi, %r8)
    jne .L_eju_next
    addq $6, %r8
    addq $7, %rcx
.L_eju_find_q:
    cmpq %rcx, %r8
    jge .L_eju_def
    cmpb $'"', (%rsi, %r8)
    je .L_eju_start
    incq %r8
    jmp .L_eju_find_q
.L_eju_start:
    incq %r8
    leaq (%rsi, %r8), %r12
    xorq %r13, %r13
.L_eju_end:
    cmpq %rcx, %r8
    jge .L_eju_def
    cmpb $'"', (%rsi, %r8)
    je .L_eju_ok
    incq %r8
    incq %r13
    jmp .L_eju_end
.L_eju_next:
    incq %r8
    jmp .L_eju_scan
.L_eju_ok:
    ret
.L_eju_def:
    leaq lsp_uri_def(%rip), %r12
    movq $lsp_uri_def_len, %r13
    ret
