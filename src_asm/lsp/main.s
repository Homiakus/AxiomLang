.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/mca_memory.inc"

.section .data
lsp_header_fmt:
    .ascii "Content-Length: "
lsp_header_fmt_len = . - lsp_header_fmt

lsp_init_response:
    .ascii "{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"capabilities\":{\"textDocumentSync\":1,\"hoverProvider\":true}}}"
lsp_init_response_len = . - lsp_init_response

crlfcrlf:
    .ascii "\r\n\r\n"
crlfcrlf_len = . - crlfcrlf

.section .bss
.align 16
lsp_in_buf: .skip 65536

.section .text
.global _start

_start:
    # Pure x86_64 Freestanding LSP Server
    # Single loop over standard input reading JSON-RPC requests
lsp_read_loop:
    movq $SYS_READ, %rax
    movq $STDIN_FILENO, %rdi
    leaq lsp_in_buf(%rip), %rsi
    movq $65535, %rdx
    syscall

    cmpq $0, %rax
    jle lsp_exit

    # Send hardcoded initialize response
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq lsp_header_fmt(%rip), %rsi
    movq $lsp_header_fmt_len, %rdx
    syscall

    # Output length: 99
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq lsp_len_str(%rip), %rsi
    movq $2, %rdx
    syscall

    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq crlfcrlf(%rip), %rsi
    movq $crlfcrlf_len, %rdx
    syscall

    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq lsp_init_response(%rip), %rsi
    movq $lsp_init_response_len, %rdx
    syscall

    jmp lsp_read_loop

lsp_exit:
    movq $SYS_EXIT, %rax
    xorq %rdi, %rdi
    syscall

.section .data
lsp_len_str: .ascii "99"
