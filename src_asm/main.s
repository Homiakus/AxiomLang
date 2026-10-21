.include "src_asm/include/axiom_defs.inc"

.section .data
msg_header:
    .ascii "================================================================================\n"
    .ascii "  ✨ Pure ASM AxiomLang Compiler (axlc_asm v1.0.0 — Freestanding x86_64)\n"
    .ascii "  ⚡ Zero Go Runtime | Zero Libc | Zero Garbage Collection\n"
    .ascii "================================================================================\n"
msg_header_len = . - msg_header

msg_usage:
    .ascii " Использование: axlc_asm <source.axl> [-o binary] [-run]\n"
    .ascii " Пример: ./bin/axlc_asm examples/01_hello_world.axl -run\n"
msg_usage_len = . - msg_usage

msg_compiling:
    .ascii " [*] Compiling via Pure x86_64 ASM Engine: "
msg_compiling_len = . - msg_compiling

msg_success:
    .ascii " [✓] Build completed in pure machine cycles.\n"
msg_success_len = . - msg_success

newline:
    .byte 10

.section .bss
.align 16
.global source_buffer, source_len, out_asm_buffer, out_asm_len
source_buffer:     .skip 1048576       # 1 MB source file buffer
source_len:        .quad 0
out_asm_buffer:    .skip 4194304       # 4 MB generated assembly buffer
out_asm_len:       .quad 0
tmp_stat_buf:      .skip 144

src_path_ptr:      .quad 0
out_path_ptr:      .quad 0
flag_run:          .quad 0

.section .data
default_out_s:     .asciz "/tmp/axiom_gen.s"
default_out_o:     .asciz "/tmp/axiom_gen.o"
default_out_bin:   .asciz "/tmp/axiom_gen_app"

cmd_as_path:       .asciz "/usr/bin/as"
cmd_ld_path:       .asciz "/usr/bin/ld"

arg_as_0:          .asciz "as"
arg_as_dash_o:     .asciz "-o"
arg_ld_0:          .asciz "ld"
arg_ld_dash_o:     .asciz "-o"

as_argv:           .quad arg_as_0, default_out_s, arg_as_dash_o, default_out_o, 0
ld_argv:           .quad arg_ld_0, default_out_o, arg_ld_dash_o, default_out_bin, 0
run_argv:          .quad default_out_bin, 0
envp:              .quad 0

.section .text
.global _start

_start:
    # Stack layout at _start:
    # 0(%rsp) = argc
    # 8(%rsp) = argv[0]
    # 16(%rsp) = argv[1] ...
    movq (%rsp), %r12        # r12 = argc
    cmpq $2, %r12
    jl print_usage_and_exit

    # Parse arguments
    movq 16(%rsp), %rdi      # argv[1]
    movq %rdi, src_path_ptr(%rip)

    # Check for -run flag in argv[2] or argv[3]
    cmpq $3, %r12
    jl no_more_args
    movq 24(%rsp), %rsi
    call check_run_flag
    cmpq $4, %r12
    jl no_more_args
    movq 32(%rsp), %rsi
    call check_run_flag

no_more_args:
    # Print header banner
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_header(%rip), %rsi
    movq $msg_header_len, %rdx
    syscall

    # Print compiling message
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_compiling(%rip), %rsi
    movq $msg_compiling_len, %rdx
    syscall

    movq src_path_ptr(%rip), %rdi
    call print_cstring
    call print_nl

    # 1. Open and read source file
    movq src_path_ptr(%rip), %rdi
    movq $SYS_OPEN, %rax
    movq $O_RDONLY, %rsi
    xorq %rdx, %rdx
    syscall
    testq %rax, %rax
    js fail_open
    movq %rax, %r13          # r13 = file descriptor

    # Read up to 1MB
    movq $SYS_READ, %rax
    movq %r13, %rdi
    leaq source_buffer(%rip), %rsi
    movq $1048576, %rdx
    syscall
    testq %rax, %rax
    js fail_read
    movq %rax, source_len(%rip)

    # Close file
    movq $SYS_CLOSE, %rax
    movq %r13, %rdi
    syscall

    # 2. Compile source_buffer into out_asm_buffer
    call compile_axiom_source

    # 3. Write out_asm_buffer to /tmp/axiom_gen.s
    movq $SYS_OPEN, %rax
    leaq default_out_s(%rip), %rdi
    movq $(O_WRONLY | O_CREAT | O_TRUNC), %rsi
    movq $0644, %rdx
    syscall
    testq %rax, %rax
    js fail_write
    movq %rax, %r14          # r14 = out fd

    movq $SYS_WRITE, %rax
    movq %r14, %rdi
    leaq out_asm_buffer(%rip), %rsi
    movq out_asm_len(%rip), %rdx
    syscall

    movq $SYS_CLOSE, %rax
    movq %r14, %rdi
    syscall

    # 4. Assemble: /usr/bin/as /tmp/axiom_gen.s -o /tmp/axiom_gen.o via fork/execve
    call run_assembler

    # 5. Link: /usr/bin/ld /tmp/axiom_gen.o -o /tmp/axiom_gen_app
    call run_linker

    # Print success
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_success(%rip), %rsi
    movq $msg_success_len, %rdx
    syscall

    # 6. If -run flag is set, execute the binary
    movq flag_run(%rip), %rax
    testq %rax, %rax
    jz normal_exit

    # Execute generated application
    call run_application

normal_exit:
    movq $SYS_EXIT, %rax
    xorq %rdi, %rdi
    syscall

print_usage_and_exit:
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_header(%rip), %rsi
    movq $msg_header_len, %rdx
    syscall

    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_usage(%rip), %rsi
    movq $msg_usage_len, %rdx
    syscall

    movq $SYS_EXIT, %rax
    movq $1, %rdi
    syscall

fail_open:
fail_read:
fail_write:
    movq $SYS_EXIT, %rax
    movq $2, %rdi
    syscall

check_run_flag:
    # rsi has cstring
    movb (%rsi), %al
    cmpb $'-', %al
    jne crf_ret
    movb 1(%rsi), %al
    cmpb $'r', %al
    jne crf_ret
    movb 2(%rsi), %al
    cmpb $'u', %al
    jne crf_ret
    movb 3(%rsi), %al
    cmpb $'n', %al
    jne crf_ret
    movq $1, flag_run(%rip)
crf_ret:
    ret

print_cstring:
    pushq %rdi
    xorq %rdx, %rdx
str_len_loop:
    cmpb $0, (%rdi, %rdx)
    je str_len_done
    incq %rdx
    jmp str_len_loop
str_len_done:
    popq %rsi
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    syscall
    ret

print_nl:
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq newline(%rip), %rsi
    movq $1, %rdx
    syscall
    ret

# ==============================================================================
# Pipeline fork & execve helpers
# ==============================================================================
run_assembler:
    movq $57, %rax            # sys_fork
    syscall
    testq %rax, %rax
    jz .L_child_as
    # Parent: waitpid
    movq %rax, %rdi
    xorq %rsi, %rsi
    xorq %rdx, %rdx
    xorq %r10, %r10
    movq $61, %rax           # sys_wait4
    syscall
    ret

.L_child_as:
    movq $59, %rax           # sys_execve
    leaq cmd_as_path(%rip), %rdi
    leaq as_argv(%rip), %rsi
    leaq envp(%rip), %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

run_linker:
    movq $57, %rax            # sys_fork
    syscall
    testq %rax, %rax
    jz .L_child_ld
    # Parent: waitpid
    movq %rax, %rdi
    xorq %rsi, %rsi
    xorq %rdx, %rdx
    xorq %r10, %r10
    movq $61, %rax           # sys_wait4
    syscall
    ret

.L_child_ld:
    movq $59, %rax           # sys_execve
    leaq cmd_ld_path(%rip), %rdi
    leaq ld_argv(%rip), %rsi
    leaq envp(%rip), %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

run_application:
    movq $57, %rax            # sys_fork
    syscall
    testq %rax, %rax
    jz .L_child_app
    # Parent: waitpid
    movq %rax, %rdi
    xorq %rsi, %rsi
    xorq %rdx, %rdx
    xorq %r10, %r10
    movq $61, %rax           # sys_wait4
    syscall
    ret

.L_child_app:
    movq $59, %rax           # sys_execve
    leaq default_out_bin(%rip), %rdi
    leaq run_argv(%rip), %rsi
    leaq envp(%rip), %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall
