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

msg_err_open:
    .ascii " [!] Error: cannot open source file\n"
msg_err_open_len = . - msg_err_open

msg_err_read:
    .ascii " [!] Error: cannot read source file\n"
msg_err_read_len = . - msg_err_read

msg_err_write:
    .ascii " [!] Error: cannot write temporary assembly file\n"
msg_err_write_len = . - msg_err_write

msg_err_compile:
    .ascii " [!] Error: compilation failed ("
msg_err_compile_len = . - msg_err_compile

msg_err_syntax:    .ascii "syntax/AST error"
msg_err_div0:      .ascii "static division by zero"
msg_err_var:       .ascii "undeclared variable"
msg_err_fn:        .ascii "undefined function"
msg_err_cap:       .ascii "capability denied: IO"
msg_err_type:      .ascii "type mismatch"
msg_err_mem:       .ascii "capability denied: Memory"
msg_err_line_pre:  .ascii " at line "
msg_err_line_suf:  .ascii ")\n"

msg_err_as:
    .ascii " [!] Error: assembler (/usr/bin/as) failed\n"
msg_err_as_len = . - msg_err_as

msg_err_ld:
    .ascii " [!] Error: linker (/usr/bin/ld) failed\n"
msg_err_ld_len = . - msg_err_ld

newline:
    .byte 10

.section .bss
.align 16
.global source_buffer, source_len, out_asm_buffer, out_asm_len, out_asm_cap, src_path_ptr
source_buffer:     .quad 0
source_len:        .quad 0
source_cap:        .quad 0
out_asm_buffer:    .quad 0
out_asm_len:       .quad 0
out_asm_cap:       .quad 0
imported_cnt:      .quad 0
imported_paths:    .skip 64 * 256

src_path_ptr:      .quad 0
out_path_ptr:      .quad 0
flag_run:          .quad 0
flag_protect:      .quad 0
child_status:      .quad 0
tmp_s_path:        .skip 64
tmp_o_path:        .skip 64
tmp_bin_path:      .skip 64
tmp_flat_path:     .skip 64
err_num_buf:       .skip 32

.section .data
tmp_prefix:        .asciz "/tmp/axiom_gen_"
tmp_suf_s:         .asciz ".s"
tmp_suf_o:         .asciz ".o"
tmp_suf_bin:       .asciz "_app"
tmp_suf_flat:      .asciz ".bin"

cmd_as_path:       .asciz "/usr/bin/as"
cmd_ld_path:       .asciz "/usr/bin/ld"

arg_as_0:          .asciz "as"
arg_as_dash_o:     .asciz "-o"
arg_ld_0:          .asciz "ld"
arg_ld_dash_o:     .asciz "-o"

flag_str_run:      .asciz "-run"
flag_str_o:        .asciz "-o"
flag_str_protect:  .asciz "--protect"
arg_ld_s:          .asciz "-s"
arg_ld_oformat:    .asciz "--oformat"
arg_ld_binary:     .asciz "binary"
arg_ld_ttext:      .asciz "-Ttext=0"

as_argv:           .quad arg_as_0, tmp_s_path, arg_as_dash_o, tmp_o_path, 0
ld_argv:           .quad arg_ld_0, tmp_o_path, arg_ld_dash_o, tmp_bin_path, 0
ld_strip_argv:     .quad arg_ld_0, arg_ld_s, tmp_o_path, arg_ld_dash_o, tmp_bin_path, 0
ld_flat_argv:      .quad arg_ld_0, arg_ld_oformat, arg_ld_binary, arg_ld_ttext, arg_ld_dash_o, tmp_flat_path, tmp_o_path, 0
run_argv:          .quad tmp_bin_path, 0
envp:              .quad 0

.section .text
.global _start

_start:
    # Stack layout at _start:
    # 0(%rsp) = argc
    # 8(%rsp) = argv[0], 16(%rsp) = argv[1] ...
    movq (%rsp), %r12        # r12 = argc
    cmpq $2, %r12
    jl print_usage_and_exit

    leaq 8(%rsp), %r13       # r13 = &argv[0]
    movq $1, %r14            # i = 1

arg_parse_loop:
    cmpq %r12, %r14
    jge arg_parse_done

    movq (%r13, %r14, 8), %rdi

    # Check "-run"
    leaq flag_str_run(%rip), %rsi
    call cstr_eq
    testq %rax, %rax
    jnz .L_set_run

    # Check "-o"
    movq (%r13, %r14, 8), %rdi
    leaq flag_str_o(%rip), %rsi
    call cstr_eq
    testq %rax, %rax
    jnz .L_set_out

    # Check "--protect"
    movq (%r13, %r14, 8), %rdi
    leaq flag_str_protect(%rip), %rsi
    call cstr_eq
    testq %rax, %rax
    jnz .L_set_protect

    # Otherwise positional source file (first non-flag)
    cmpq $0, src_path_ptr(%rip)
    jne .L_next_arg
    movq (%r13, %r14, 8), %rax
    movq %rax, src_path_ptr(%rip)
    jmp .L_next_arg

.L_set_run:
    movq $1, flag_run(%rip)
    jmp .L_next_arg

.L_set_protect:
    movq $1, flag_protect(%rip)
    movq $0, dwarf_enabled(%rip)
    jmp .L_next_arg

.L_set_out:
    incq %r14
    cmpq %r12, %r14
    jge print_usage_and_exit
    movq (%r13, %r14, 8), %rax
    movq %rax, out_path_ptr(%rip)

.L_next_arg:
    incq %r14
    jmp arg_parse_loop

arg_parse_done:
    cmpq $0, src_path_ptr(%rip)
    je print_usage_and_exit

    # Build PID-unique temp paths: /tmp/axiom_gen_<pid>.{s,o,_app}
    call init_temp_paths

    # If -o was provided, patch ld_argv[3] and run_argv[0]
    movq out_path_ptr(%rip), %rax
    testq %rax, %rax
    jz .L_use_default_out
    leaq ld_argv(%rip), %rbx
    movq %rax, 24(%rbx)
    leaq ld_strip_argv(%rip), %rbx
    movq %rax, 32(%rbx)
    leaq run_argv(%rip), %rbx
    movq %rax, (%rbx)

.L_use_default_out:
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

    # 1. Allocate 8 MB dynamic source_buffer via SYS_MMAP and expand imports
    movq $8388608, %rsi
    movq %rsi, source_cap(%rip)
    movq $SYS_MMAP, %rax
    xorq %rdi, %rdi
    movq $(PROT_READ | PROT_WRITE), %rdx
    movq $(MAP_PRIVATE | MAP_ANONYMOUS), %r10
    movq $-1, %r8
    xorq %r9, %r9
    syscall
    cmpq $0, %rax
    jl fail_read
    movq %rax, source_buffer(%rip)
    movq $0, source_len(%rip)
    movq $0, imported_cnt(%rip)

    movq src_path_ptr(%rip), %rdi
    xorq %rsi, %rsi          # depth = 0
    call expand_source_file

    # Allocate 16 MB dynamic out_asm_buffer via SYS_MMAP
    movq $16777216, %rsi
    movq %rsi, out_asm_cap(%rip)
    movq $SYS_MMAP, %rax
    xorq %rdi, %rdi
    movq $(PROT_READ | PROT_WRITE), %rdx
    movq $(MAP_PRIVATE | MAP_ANONYMOUS), %r10
    movq $-1, %r8
    xorq %r9, %r9
    syscall
    cmpq $0, %rax
    jl fail_write
    movq %rax, out_asm_buffer(%rip)

    # 2. Compile source_buffer into out_asm_buffer
    call compile_axiom_source
    testq %rax, %rax
    jnz fail_compile

    # 3. Write out_asm_buffer to tmp_s_path
    movq $SYS_OPEN, %rax
    leaq tmp_s_path(%rip), %rdi
    movq $(O_WRONLY | O_CREAT | O_TRUNC), %rsi
    movq $0644, %rdx
    syscall
    testq %rax, %rax
    js fail_write
    movq %rax, %r14          # r14 = out fd

    movq $SYS_WRITE, %rax
    movq %r14, %rdi
    movq out_asm_buffer(%rip), %rsi
    movq out_asm_len(%rip), %rdx
    syscall
    testq %rax, %rax
    js fail_write

    movq $SYS_CLOSE, %rax
    movq %r14, %rdi
    syscall

    # 4. Assemble & Link
    call run_assembler
    testq %rax, %rax
    jnz fail_as

    cmpq $0, flag_protect(%rip)
    je .L_link_normal
    call pack_protected_binary
    jmp .L_cleanup_tmp

.L_link_normal:
    call run_linker
    testq %rax, %rax
    jnz fail_ld

.L_cleanup_tmp:
    # Cleanup temp .s and .o files
    movq $SYS_UNLINK, %rax
    leaq tmp_s_path(%rip), %rdi
    syscall
    movq $SYS_UNLINK, %rax
    leaq tmp_o_path(%rip), %rdi
    syscall

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

    call run_application
    movq %rax, %rdi
    movq $SYS_EXIT, %rax
    syscall

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
    leaq msg_err_open(%rip), %rsi
    movq $msg_err_open_len, %rdx
    jmp fail_with_msg
fail_read:
    leaq msg_err_read(%rip), %rsi
    movq $msg_err_read_len, %rdx
    jmp fail_with_msg
fail_write:
    leaq msg_err_write(%rip), %rsi
    movq $msg_err_write_len, %rdx
    jmp fail_with_msg
fail_compile:
    movq $SYS_WRITE, %rax
    movq $STDERR_FILENO, %rdi
    leaq msg_err_compile(%rip), %rsi
    movq $msg_err_compile_len, %rdx
    syscall

    # Print source file path to stderr
    movq src_path_ptr(%rip), %rsi
    testq %rsi, %rsi
    jz .L_fc_no_path
    xorq %rdx, %rdx
.L_fc_plen:
    cmpb $0, (%rsi, %rdx)
    je .L_fc_pwrite
    incq %rdx
    jmp .L_fc_plen
.L_fc_pwrite:
    movq $SYS_WRITE, %rax
    movq $STDERR_FILENO, %rdi
    syscall
    movq $SYS_WRITE, %rax
    movq $STDERR_FILENO, %rdi
    leaq msg_err_line_pre(%rip), %rsi
    movq $1, %rdx            # " "
    syscall
.L_fc_no_path:

    movq chk_err_code(%rip), %rax
    cmpq $2, %rax
    je .L_fc_div0
    cmpq $3, %rax
    je .L_fc_var
    cmpq $4, %rax
    je .L_fc_fn
    cmpq $5, %rax
    je .L_fc_cap
    cmpq $6, %rax
    je .L_fc_type
    cmpq $7, %rax
    je .L_fc_mem
    leaq msg_err_syntax(%rip), %rsi
    movq $16, %rdx
    jmp .L_fc_reason
.L_fc_div0:
    leaq msg_err_div0(%rip), %rsi
    movq $23, %rdx
    jmp .L_fc_reason
.L_fc_var:
    leaq msg_err_var(%rip), %rsi
    movq $19, %rdx
    jmp .L_fc_reason
.L_fc_fn:
    leaq msg_err_fn(%rip), %rsi
    movq $18, %rdx
    jmp .L_fc_reason
.L_fc_cap:
    leaq msg_err_cap(%rip), %rsi
    movq $21, %rdx
    jmp .L_fc_reason
.L_fc_type:
    leaq msg_err_type(%rip), %rsi
    movq $13, %rdx
    jmp .L_fc_reason
.L_fc_mem:
    leaq msg_err_mem(%rip), %rsi
    movq $25, %rdx
.L_fc_reason:
    movq $SYS_WRITE, %rax
    movq $STDERR_FILENO, %rdi
    syscall

    movq chk_err_line(%rip), %rax
    testq %rax, %rax
    jnz .L_fc_has_line
    movq p_line(%rip), %rax
.L_fc_has_line:
    pushq %rax
    movq $SYS_WRITE, %rax
    movq $STDERR_FILENO, %rdi
    leaq msg_err_line_pre(%rip), %rsi
    movq $9, %rdx
    syscall
    popq %rax
    call print_u64_stderr
    movq $SYS_WRITE, %rax
    movq $STDERR_FILENO, %rdi
    leaq msg_err_line_suf(%rip), %rsi
    movq $2, %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $2, %rdi
    syscall

init_temp_paths:
    movq $39, %rax           # sys_getpid
    syscall
    movq %rax, %r12          # pid

    leaq tmp_s_path(%rip), %rdi
    leaq tmp_suf_s(%rip), %rsi
    call build_pid_path

    leaq tmp_o_path(%rip), %rdi
    leaq tmp_suf_o(%rip), %rsi
    call build_pid_path

    leaq tmp_bin_path(%rip), %rdi
    leaq tmp_suf_bin(%rip), %rsi
    call build_pid_path

    leaq tmp_flat_path(%rip), %rdi
    leaq tmp_suf_flat(%rip), %rsi
    call build_pid_path
    ret

build_pid_path:
    # %rdi = dest, %rsi = suffix, %r12 = pid
    pushq %rsi
    leaq tmp_prefix(%rip), %rsi
.L_bpp_pre:
    movb (%rsi), %al
    testb %al, %al
    jz .L_bpp_pid
    movb %al, (%rdi)
    incq %rsi
    incq %rdi
    jmp .L_bpp_pre
.L_bpp_pid:
    movq %r12, %rax
    leaq err_num_buf+31(%rip), %rcx
    xorq %r8, %r8
.L_bpp_div:
    xorq %rdx, %rdx
    movq $10, %r9
    divq %r9
    addb $'0', %dl
    decq %rcx
    movb %dl, (%rcx)
    incq %r8
    testq %rax, %rax
    jnz .L_bpp_div
.L_bpp_cp_pid:
    testq %r8, %r8
    jz .L_bpp_suf
    movb (%rcx), %al
    movb %al, (%rdi)
    incq %rcx
    incq %rdi
    decq %r8
    jmp .L_bpp_cp_pid
.L_bpp_suf:
    popq %rsi
.L_bpp_s_loop:
    movb (%rsi), %al
    movb %al, (%rdi)
    testb %al, %al
    jz .L_bpp_done
    incq %rsi
    incq %rdi
    jmp .L_bpp_s_loop
.L_bpp_done:
    ret

print_u64_stderr:
    leaq err_num_buf+31(%rip), %rsi
    xorq %rdx, %rdx
    testq %rax, %rax
    jnz .L_pus_div
    decq %rsi
    movb $'0', (%rsi)
    incq %rdx
    jmp .L_pus_wr
.L_pus_div:
    testq %rax, %rax
    jz .L_pus_wr
    xorq %rcx, %rcx
    pushq %rdx
    xorq %rdx, %rdx
    movq $10, %rcx
    divq %rcx
    addb $'0', %dl
    decq %rsi
    movb %dl, (%rsi)
    popq %rdx
    incq %rdx
    jmp .L_pus_div
.L_pus_wr:
    movq $SYS_WRITE, %rax
    movq $STDERR_FILENO, %rdi
    syscall
    ret
fail_as:
    leaq msg_err_as(%rip), %rsi
    movq $msg_err_as_len, %rdx
    jmp fail_with_msg
fail_ld:
    leaq msg_err_ld(%rip), %rsi
    movq $msg_err_ld_len, %rdx

fail_with_msg:
    movq $SYS_WRITE, %rax
    movq $STDERR_FILENO, %rdi
    syscall
    movq $SYS_EXIT, %rax
    movq $2, %rdi
    syscall

# ==============================================================================
# cstr_eq(s1: %rdi, s2: %rsi) -> %rax (1 if equal, 0 otherwise)
# ==============================================================================
cstr_eq:
    xorq %rcx, %rcx
.L_ceq_loop:
    movb (%rdi, %rcx), %al
    movb (%rsi, %rcx), %dl
    cmpb %al, %dl
    jne .L_ceq_ne
    testb %al, %al
    jz .L_ceq_eq
    incq %rcx
    jmp .L_ceq_loop
.L_ceq_eq:
    movq $1, %rax
    ret
.L_ceq_ne:
    xorq %rax, %rax
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
# Pipeline fork & execve helpers (returns exit code in %rax, 0 on success)
# ==============================================================================
wait_and_decode:
    # %rdi = child pid
    movq $0, child_status(%rip)
    leaq child_status(%rip), %rsi
    xorq %rdx, %rdx
    xorq %r10, %r10
    movq $61, %rax           # sys_wait4
    syscall
    movl child_status(%rip), %eax
    movl %eax, %ecx
    andl $0x7f, %ecx
    jnz .L_signaled
    shrl $8, %eax
    andl $0xff, %eax
    ret
.L_signaled:
    movq $128, %rax
    addq %rcx, %rax
    ret

run_assembler:
    movq $57, %rax            # sys_fork
    syscall
    testq %rax, %rax
    jz .L_child_as
    movq %rax, %rdi
    jmp wait_and_decode

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
    movq %rax, %rdi
    jmp wait_and_decode

.L_child_ld:
    movq $59, %rax           # sys_execve
    leaq cmd_ld_path(%rip), %rdi
    leaq ld_argv(%rip), %rsi
    leaq envp(%rip), %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

run_linker_flat:
    movq $57, %rax
    syscall
    testq %rax, %rax
    jz .L_child_ldf
    movq %rax, %rdi
    jmp wait_and_decode
.L_child_ldf:
    movq $59, %rax
    leaq cmd_ld_path(%rip), %rdi
    leaq ld_flat_argv(%rip), %rsi
    leaq envp(%rip), %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

run_linker_strip:
    movq $57, %rax
    syscall
    testq %rax, %rax
    jz .L_child_lds
    movq %rax, %rdi
    jmp wait_and_decode
.L_child_lds:
    movq $59, %rax
    leaq cmd_ld_path(%rip), %rdi
    leaq ld_strip_argv(%rip), %rsi
    leaq envp(%rip), %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

# ==============================================================================
# pack_protected_binary()
# Encrypts flat PIC payload into a single continuous LCG polynomial blob
# with PTRACE anti-debug + mmap/mprotect self-decryption stub, then strips ELF.
# ==============================================================================
pack_protected_binary:
    pushq %rbp
    movq %rsp, %rbp
    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15
    subq $160, %rsp          # [0..143]=stat, [144]=mmap_sz, [152]=payload_buf

    call run_linker_flat
    testq %rax, %rax
    jnz fail_ld

    # Open tmp_flat_path
    movq $SYS_OPEN, %rax
    leaq tmp_flat_path(%rip), %rdi
    xorq %rsi, %rsi          # O_RDONLY
    xorq %rdx, %rdx
    syscall
    testq %rax, %rax
    js fail_read
    movq %rax, %r12          # fd

    movq $5, %rax            # sys_fstat
    movq %r12, %rdi
    movq %rsp, %rsi
    syscall
    testq %rax, %rax
    js fail_read
    movq 48(%rsp), %r13      # payload_len

    movq %r13, %rsi
    addq $4095, %rsi
    andq $-4096, %rsi
    movq %rsi, 144(%rsp)
    movq $SYS_MMAP, %rax
    xorq %rdi, %rdi
    movq $(PROT_READ | PROT_WRITE), %rdx
    movq $(MAP_PRIVATE | MAP_ANONYMOUS), %r10
    movq $-1, %r8
    xorq %r9, %r9
    syscall
    cmpq $0, %rax
    jl fail_read
    movq %rax, 152(%rsp)
    movq %rax, %r14          # payload_buf

    xorq %rbx, %rbx
.L_ppb_rd:
    cmpq %r13, %rbx
    jge .L_ppb_rd_done
    movq $SYS_READ, %rax
    movq %r12, %rdi
    leaq (%r14, %rbx), %rsi
    movq %r13, %rdx
    subq %rbx, %rdx
    syscall
    testq %rax, %rax
    jle .L_ppb_rd_done
    addq %rax, %rbx
    jmp .L_ppb_rd
.L_ppb_rd_done:
    movq $SYS_CLOSE, %rax
    movq %r12, %rdi
    syscall
    movq $SYS_UNLINK, %rax
    leaq tmp_flat_path(%rip), %rdi
    syscall

    # Build stub.s in out_asm_buffer (r15 = write pos)
    xorq %r15, %r15

    leaq prot_stub_p1(%rip), %rsi
    movq $prot_stub_p1_len, %rcx
    call .L_ppb_app
    movq %r13, %rax
    addq $65536, %rax
    call .L_ppb_u64

    leaq prot_stub_p2(%rip), %rsi
    movq $prot_stub_p2_len, %rcx
    call .L_ppb_app
    movq %r13, %rax
    call .L_ppb_u64

    leaq prot_stub_p3(%rip), %rsi
    movq $prot_stub_p3_len, %rcx
    call .L_ppb_app
    movq %r13, %rax
    addq $65536, %rax
    call .L_ppb_u64

    leaq prot_stub_p4(%rip), %rsi
    movq $prot_stub_p4_len, %rcx
    call .L_ppb_app

    # Encrypt payload_buf[0..r13-1] with LCG polynomial keystream and emit .byte lines
    xorq %rcx, %rcx          # idx
    movq $0x9E3779B97F4A7C15, %r8
    movq $6364136223846793005, %r9
    movq $1442695040888963407, %r10
.L_ppb_enc:
    cmpq %r13, %rcx
    jge .L_ppb_enc_done
    imulq %r9, %r8
    addq %r10, %r8
    movq %r8, %rax
    shrq $32, %rax
    movq %rcx, %rdx
    imulq $0x9E, %rdx
    xorq %rdx, %rax
    xorb (%r14, %rcx), %al   # al = encrypted byte
    pushq %rax
    pushq %rcx
    leaq prot_byte_pre(%rip), %rsi
    movq $prot_byte_pre_len, %rcx
    call .L_ppb_app
    popq %rcx
    popq %rax
    movzbq %al, %rax
    pushq %rcx
    call .L_ppb_u64
    leaq newline(%rip), %rsi
    movq $1, %rcx
    call .L_ppb_app
    popq %rcx
    incq %rcx
    jmp .L_ppb_enc

.L_ppb_enc_done:
    # Unmap payload_buf
    movq $SYS_MUNMAP, %rax
    movq %r14, %rdi
    movq 144(%rsp), %rsi
    syscall

    # Write stub.s to tmp_s_path
    movq $SYS_OPEN, %rax
    leaq tmp_s_path(%rip), %rdi
    movq $(O_WRONLY | O_CREAT | O_TRUNC), %rsi
    movq $0644, %rdx
    syscall
    testq %rax, %rax
    js fail_write
    movq %rax, %r12

    xorq %rbx, %rbx
    movq out_asm_buffer(%rip), %r14
.L_ppb_wr:
    cmpq %r15, %rbx
    jge .L_ppb_wr_done
    movq $SYS_WRITE, %rax
    movq %r12, %rdi
    leaq (%r14, %rbx), %rsi
    movq %r15, %rdx
    subq %rbx, %rdx
    syscall
    testq %rax, %rax
    js fail_write
    addq %rax, %rbx
    jmp .L_ppb_wr
.L_ppb_wr_done:
    movq $SYS_CLOSE, %rax
    movq %r12, %rdi
    syscall

    call run_assembler
    testq %rax, %rax
    jnz fail_as

    call run_linker_strip
    testq %rax, %rax
    jnz fail_ld

    addq $160, %rsp
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx
    leave
    ret

.L_ppb_app:
    # %rsi = src, %rcx = len, updates %r15
    movq out_asm_buffer(%rip), %rdi
    addq %r15, %rdi
    addq %rcx, %r15
    cld
    rep movsb
    ret

.L_ppb_u64:
    # %rax = u64, appends decimal to out_asm_buffer+%r15
    pushq %r8
    pushq %r9
    subq $32, %rsp
    lea 32(%rsp), %rsi
    xorq %rcx, %rcx
    movq $10, %r8
    testq %rax, %rax
    jnz .L_pu_div
    decq %rsi
    movb $'0', (%rsi)
    incq %rcx
    jmp .L_pu_cp
.L_pu_div:
    xorq %rdx, %rdx
    divq %r8
    addb $'0', %dl
    decq %rsi
    movb %dl, (%rsi)
    incq %rcx
    testq %rax, %rax
    jnz .L_pu_div
.L_pu_cp:
    call .L_ppb_app
    addq $32, %rsp
    popq %r9
    popq %r8
    ret

.section .data
prot_stub_p1:
    .ascii ".global _start\n.section .text\n_start:\n"
    .ascii "    movq $101, %rax\n    xorq %rdi, %rdi\n    xorq %rsi, %rsi\n    xorq %rdx, %rdx\n    xorq %r10, %r10\n    syscall\n"
    .ascii "    testq %rax, %rax\n    js .L_ax_dbg_trap\n"
    .ascii "    movq $9, %rax\n    xorq %rdi, %rdi\n    movq $"
prot_stub_p1_len = . - prot_stub_p1
prot_stub_p2:
    .ascii ", %rsi\n    movq $3, %rdx\n    movq $34, %r10\n    movq $-1, %r8\n    xorq %r9, %r9\n    syscall\n"
    .ascii "    cmpq $0, %rax\n    jl .L_ax_dbg_trap\n    movq %rax, %r12\n    leaq ax_enc_blob(%rip), %r13\n"
    .ascii "    xorq %rcx, %rcx\n    movq $0x9E3779B97F4A7C15, %r8\n    movq $6364136223846793005, %r9\n    movq $1442695040888963407, %r10\n"
    .ascii ".L_ax_dec:\n    cmpq $"
prot_stub_p2_len = . - prot_stub_p2
prot_stub_p3:
    .ascii ", %rcx\n    jge .L_ax_exec\n    imulq %r9, %r8\n    addq %r10, %r8\n    movq %r8, %rax\n    shrq $32, %rax\n"
    .ascii "    movq %rcx, %rdx\n    imulq $0x9E, %rdx\n    xorq %rdx, %rax\n    xorb (%r13, %rcx), %al\n"
    .ascii "    movb %al, (%r12, %rcx)\n    incq %rcx\n    jmp .L_ax_dec\n"
    .ascii ".L_ax_exec:\n    movq $10, %rax\n    movq %r12, %rdi\n    movq $"
prot_stub_p3_len = . - prot_stub_p3
prot_stub_p4:
    .ascii ", %rsi\n    movq $7, %rdx\n    syscall\n    jmp *%r12\n"
    .ascii ".L_ax_dbg_trap:\n    movq $60, %rax\n    movq $1, %rdi\n    syscall\n"
    .ascii ".section .rodata\nax_enc_blob:\n"
prot_stub_p4_len = . - prot_stub_p4
prot_byte_pre: .ascii "    .byte "
prot_byte_pre_len = . - prot_byte_pre
.section .text


run_application:
    movq $57, %rax            # sys_fork
    syscall
    testq %rax, %rax
    jz .L_child_app
    movq %rax, %rdi
    jmp wait_and_decode

.L_child_app:
    movq $59, %rax           # sys_execve
    leaq run_argv(%rip), %rbx
    movq (%rbx), %rdi
    leaq run_argv(%rip), %rsi
    leaq envp(%rip), %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

# ==============================================================================
# expand_source_file(path: %rdi, depth: %rsi)
# Recursively expands `import "path.axl"` directives into source_buffer
# ==============================================================================
expand_source_file:
    pushq %rbp
    movq %rsp, %rbp
    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15
    subq $432, %rsp          # [0..143]=stat, [144..399]=imp_path(256), [400]=depth, [408]=mmap_size, [416]=mmap_ptr

    cmpq $16, %rsi
    jge .L_esf_ret
    movq %rsi, 400(%rsp)     # save depth
    movq %rdi, %r12          # path ptr

    # Check if already imported
    xorq %rcx, %rcx
.L_esf_chk_dup:
    cmpq imported_cnt(%rip), %rcx
    jge .L_esf_rec_path
    movq %rcx, %rax
    imulq $256, %rax
    leaq imported_paths(%rip), %rsi
    addq %rax, %rsi
    movq %r12, %rdi
    pushq %rcx
    call cstr_eq
    popq %rcx
    testq %rax, %rax
    jnz .L_esf_ret
    incq %rcx
    jmp .L_esf_chk_dup

.L_esf_rec_path:
    movq imported_cnt(%rip), %rcx
    cmpq $64, %rcx
    jge .L_esf_open
    imulq $256, %rcx, %rax
    leaq imported_paths(%rip), %rdi
    addq %rax, %rdi
    xorq %rdx, %rdx
.L_esf_cp_p:
    cmpq $255, %rdx
    jge .L_esf_cp_done
    movb (%r12, %rdx), %al
    movb %al, (%rdi, %rdx)
    testb %al, %al
    jz .L_esf_cp_done
    incq %rdx
    jmp .L_esf_cp_p
.L_esf_cp_done:
    movb $0, (%rdi, %rdx)
    incq imported_cnt(%rip)

.L_esf_open:
    movq $SYS_OPEN, %rax
    movq %r12, %rdi
    movq $O_RDONLY, %rsi
    xorq %rdx, %rdx
    syscall
    testq %rax, %rax
    js fail_open
    movq %rax, %r13          # fd

    movq $SYS_FSTAT, %rax
    movq %r13, %rdi
    movq %rsp, %rsi
    syscall
    testq %rax, %rax
    js fail_read
    movq 48(%rsp), %r14      # st_size

    testq %r14, %r14
    jnz .L_esf_mmap
    movq $SYS_CLOSE, %rax
    movq %r13, %rdi
    syscall
    jmp .L_esf_ret

.L_esf_mmap:
    movq %r14, %rsi
    addq $4095, %rsi
    andq $-4096, %rsi
    movq %rsi, 408(%rsp)
    movq $SYS_MMAP, %rax
    xorq %rdi, %rdi
    movq $(PROT_READ | PROT_WRITE), %rdx
    movq $(MAP_PRIVATE | MAP_ANONYMOUS), %r10
    movq $-1, %r8
    xorq %r9, %r9
    syscall
    cmpq $0, %rax
    jl fail_read
    movq %rax, 416(%rsp)     # temp buf
    movq %rax, %r15

    xorq %rbx, %rbx          # bytes read
.L_esf_rd:
    cmpq %r14, %rbx
    jge .L_esf_rd_done
    movq $SYS_READ, %rax
    movq %r13, %rdi
    leaq (%r15, %rbx), %rsi
    movq %r14, %rdx
    subq %rbx, %rdx
    syscall
    testq %rax, %rax
    js fail_read
    jz .L_esf_rd_done
    addq %rax, %rbx
    jmp .L_esf_rd
.L_esf_rd_done:
    movq $SYS_CLOSE, %rax
    movq %r13, %rdi
    syscall

    # Scan temp buf (%r15, len %rbx) line by line
    xorq %r12, %r12          # line start
.L_esf_line:
    cmpq %rbx, %r12
    jge .L_esf_unmap
    movq %r12, %r13          # find '\n'
.L_esf_nl:
    cmpq %rbx, %r13
    jge .L_esf_got_line
    cmpb $10, (%r15, %r13)
    je .L_esf_inc_nl
    incq %r13
    jmp .L_esf_nl
.L_esf_inc_nl:
    incq %r13
.L_esf_got_line:
    # Check if line (after leading spaces/tabs) starts with "import"
    movq %r12, %rax
.L_esf_ws:
    cmpq %r13, %rax
    jge .L_esf_emit_line
    movb (%r15, %rax), %cl
    cmpb $' ', %cl
    je .L_esf_ws_next
    cmpb $'\t', %cl
    je .L_esf_ws_next
    jmp .L_esf_chk_imp
.L_esf_ws_next:
    incq %rax
    jmp .L_esf_ws
.L_esf_chk_imp:
    movq %r13, %rcx
    subq %rax, %rcx
    cmpq $9, %rcx
    jl .L_esf_emit_line
    cmpb $'i', (%r15, %rax)
    jne .L_esf_emit_line
    cmpb $'m', 1(%r15, %rax)
    jne .L_esf_emit_line
    cmpb $'p', 2(%r15, %rax)
    jne .L_esf_emit_line
    cmpb $'o', 3(%r15, %rax)
    jne .L_esf_emit_line
    cmpb $'r', 4(%r15, %rax)
    jne .L_esf_emit_line
    cmpb $'t', 5(%r15, %rax)
    jne .L_esf_emit_line
    movb 6(%r15, %rax), %cl
    cmpb $' ', %cl
    je .L_esf_find_q1
    cmpb $'\t', %cl
    jne .L_esf_emit_line
.L_esf_find_q1:
    addq $7, %rax
.L_esf_q1:
    cmpq %r13, %rax
    jge .L_esf_emit_line
    cmpb $'"', (%r15, %rax)
    je .L_esf_copy_imp
    incq %rax
    jmp .L_esf_q1
.L_esf_copy_imp:
    incq %rax
    leaq 144(%rsp), %rdi
    xorq %rcx, %rcx
.L_esf_imp_lp:
    cmpq %r13, %rax
    jge .L_esf_emit_line
    cmpq $255, %rcx
    jge .L_esf_emit_line
    movb (%r15, %rax), %dl
    cmpb $'"', %dl
    je .L_esf_do_imp
    movb %dl, (%rdi, %rcx)
    incq %rax
    incq %rcx
    jmp .L_esf_imp_lp
.L_esf_do_imp:
    movb $0, (%rdi, %rcx)
    movq 400(%rsp), %rsi
    incq %rsi
    call expand_source_file
    jmp .L_esf_next_line

.L_esf_emit_line:
    movq %r13, %rcx
    subq %r12, %rcx          # line_len
    movq source_len(%rip), %rax
    addq %rcx, %rax
    cmpq source_cap(%rip), %rax
    jge .L_esf_next_line
    movq source_buffer(%rip), %rdi
    addq source_len(%rip), %rdi
    leaq (%r15, %r12), %rsi
    cld
    rep movsb
    movq %rax, source_len(%rip)

.L_esf_next_line:
    movq %r13, %r12
    jmp .L_esf_line

.L_esf_unmap:
    # Ensure trailing newline after each file
    movq source_len(%rip), %rax
    cmpq source_cap(%rip), %rax
    jge .L_esf_munmap
    movq source_buffer(%rip), %rdi
    movb $10, (%rdi, %rax)
    incq source_len(%rip)
.L_esf_munmap:
    movq $SYS_MUNMAP, %rax
    movq 416(%rsp), %rdi
    movq 408(%rsp), %rsi
    syscall

.L_esf_ret:
    addq $432, %rsp
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx
    leave
    ret
